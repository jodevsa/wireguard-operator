/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/jodevsa/wireguard-operator/pkg/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	configMapName      = "wireguard-peer-config"
	configMapNamespace = "default"
	configMapKey       = "config"
)

type WireguardSidecarReconciler struct {
	client.Client
	Scheme                 *runtime.Scheme
	SidecarImage           string
	SidecarImagePullPolicy corev1.PullPolicy
	RequeueAfter           time.Duration
}

func (r *WireguardSidecarReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &corev1.Pod{}, "metadata.annotations", func(rawObj client.Object) []string {
		pod := rawObj.(*corev1.Pod)
		return []string{pod.ObjectMeta.Annotations["vpn.example.com/enable-sidecar"]}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		WithEventFilter(predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				oldPod := e.ObjectOld.(*corev1.Pod)
				newPod := e.ObjectNew.(*corev1.Pod)
				return oldPod.ObjectMeta.Annotations["vpn.example.com/enable-sidecar"] != newPod.ObjectMeta.Annotations["vpn.example.com/enable-sidecar"]
			},
		}).
		Complete(r)
}

func (r *WireguardSidecarReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var pod corev1.Pod
	if err := r.Get(ctx, req.NamespacedName, &pod); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if pod.ObjectMeta.Annotations["vpn.example.com/enable-sidecar"] != "true" {
		// Pod does not have the desired annotation implement check and garbage collection logic
		return ctrl.Result{}, nil
	}

	// Check if a sidecar container already exists in the pod spec
	hasSidecar := false
	for _, container := range pod.Spec.Containers {
		if container.Name == "wireguard-sidecar" {
			hasSidecar = true
			break
		}
	}

	if !hasSidecar {

		ref, hasRef := pod.ObjectMeta.Annotations["vpn.example.com/sidecar-wireguard-ref"]

		if !hasRef {
			return ctrl.Result{}, fmt.Errorf("%s does not have ref annotation", req.Name)
		}

		wireguard := &v1alpha1.Wireguard{}
		err := r.Get(context.Background(), types.NamespacedName{Name: ref}, wireguard)
		if err != nil {
			if errors.IsNotFound(err) {
				return ctrl.Result{}, fmt.Errorf("Wireguard resource %s not found", req.Name)
			}
			return ctrl.Result{}, err
		}

		peer := &v1alpha1.WireguardPeer{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-sidecar", pod.Name),
				Namespace: pod.Namespace,
			},
			Spec: v1alpha1.WireguardPeerSpec{
				WireguardRef: ref,
			},
		}

		err = r.Client.Create(context.Background(), peer)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("Unable to create peer %s", peer.Name)
		}

		// Create the configmap for the peer status config
		configMapName := fmt.Sprintf("%s-sidecar", pod.Name)
		configMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      configMapName,
				Namespace: pod.Namespace,
			},
			Data: map[string]string{
				"wg0.conf": peer.Status.Config,
			},
		}

		err = r.Client.Create(context.Background(), configMap)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("Unable to create configMap %s", configMap.Name)
		}

		// Add the configmap volume to the pod spec
		pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{
			Name: configMap.Name,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: configMap.Name,
					},
				},
			},
		})

		// Mount the configmap in the sidecar container
		pod.Spec.Containers = append(pod.Spec.Containers, corev1.Container{
			Name:            "wireguard-sidecar",
			Image:           r.SidecarImage,
			ImagePullPolicy: r.SidecarImagePullPolicy,
			VolumeMounts: []corev1.VolumeMount{{
				Name:      configMap.Name,
				MountPath: "/etc/wireguard/wg0.conf",
				SubPath:   "wg0.conf",
			}},
		})

		if err := r.Update(ctx, &pod); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}
