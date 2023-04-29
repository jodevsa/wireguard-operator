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
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/jodevsa/wireguard-operator/pkg/api/v1alpha1"
	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	configMapName      = "wireguard-peer-config"
	configMapNamespace = "default"
	configMapKey       = "config"
)

type WireguardSidecarReconciler struct {
	client.Client
	Scheme             *runtime.Scheme
	SidecarImage       string
	SidecarImagePullPolicy    corev1.PullPolicy
	RequeueAfter       time.Duration
}

func (r *WireguardSidecarReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&corev1.Pod{}).
        WithEventFilter(predicate.Funcs{
            CreateFunc: func(e event.CreateEvent) bool {
                return r.hasSidecarAnnotation(e.Object)
            },
            UpdateFunc: func(e event.UpdateEvent) bool {
                return r.hasSidecarAnnotation(e.ObjectNew)
            },
            DeleteFunc: func(e event.DeleteEvent) bool {
                // Ignore delete events
                return false
            },
            GenericFunc: func(e event.GenericEvent) bool {
                // Ignore generic events
                return false
            },
        }).
        Complete(r)
}

func (r *WireguardSidecarReconciler) hasAnnotationSidecar(obj runtime.Object) bool {
    pod, ok := obj.(*corev1.Pod)
    if !ok {
        return false
    }

    enable, ok := pod.Annotations["vpn.example.com/sidecar-enable"]
    if !ok || enable != "true" {
        return false
    }

    wgRef, ok := pod.Annotations["vpn.example.com/sidecar-wireguard-ref"]
    if !ok || wgRef == "" {
		wireguardObj := &v1alpha1.Wireguard{}
		err := r.Client.Get(context.Background(), types.NamespacedName{Name: wgRef}, wireguardObj)
		if err != nil {
			r.Log.Error(err, fmt.Sprintf("failed to get Wireguard object %s", wireguardName))
			return false
		}
        r.Log.Error(fmt.Errorf("missing or empty vpn.example.com/sidecar-wireguard-ref annotation for pod %s/%s", pod.Namespace, pod.Name), "failed to reconcile pod")
        return false
    }

    // Create the wireguard peer object
    peer := &v1alpha1.WireguardPeer{
        ObjectMeta: metav1.ObjectMeta{
            Name:      fmt.Sprintf("%s-sidecar", pod.Name),
            Namespace: pod.Namespace,
        },
        Spec: v1alpha1.WireguardPeerSpec{
            WireguardRef: wgRef,
        },
    }

    // Create or update the wireguard peer object in the cluster
    err := r.Client.CreateOrUpdate(context.Background(), peer, func() error {
        return ctrl.SetControllerReference(pod, peer, r.Scheme)
    })
    if err != nil {
        r.Log.Error(err, "failed to create or update WireguardPeer", "peer", peer)
        return false
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
        r.Log.Error(err, "failed to create ConfigMap", "configMap", configMap)
        return false
    }

    // Add the configmap volume to the pod spec
    pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{
        Name: configMapName,
        VolumeSource: corev1.VolumeSource{
            ConfigMap: &corev1.ConfigMapVolumeSource{
                Name: configMapName,
            },
        },
    })

    // Mount the configmap in the sidecar container
    pod.Spec.Containers = append(pod.Spec.Containers, corev1.Container{
        Name:  "wireguard-sidecar",
        Image: r.SidecarImage,
		ImagePullPolicy: r.SidecarImagePullPolicy,
        // Add any required configuration for the sidecar container here
        VolumeMounts: []corev1.VolumeMount{
            {
                Name:      configMapName,
                MountPath: "/etc/wireguard/wg0.conf",
                SubPath:   "wg0.conf",
            },
        },
    })

    return true
}
