/*
Copyright 2021.

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
	"github.com/jodevsa/wireguard-operator/pkg/api/v1alpha1"
	"github.com/jodevsa/wireguard-operator/pkg/resources"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// WireguardReconciler reconciles a Wireguard object

const port = 51820

const metricsPort = 9586

type WireguardReconciler struct {
	client.Client
	Scheme               *runtime.Scheme
	AgentImage           string
	AgentImagePullPolicy corev1.PullPolicy
}

func labelsForWireguard(name string) map[string]string {
	return map[string]string{"app": "wireguard", "instance": name}
}

func (r *WireguardReconciler) ConfigmapForWireguard(m *v1alpha1.Wireguard) *corev1.ConfigMap {
	ls := labelsForWireguard(m.Name)
	dep := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name + "-config",
			Namespace: m.Namespace,
			Labels:    ls,
		},
	}

	ctrl.SetControllerReference(m, dep, r.Scheme)
	return dep
}

func (r *WireguardReconciler) getNodeIps(ctx context.Context, req ctrl.Request) ([]string, error) {
	nodes := &corev1.NodeList{}
	if err := r.List(ctx, nodes); err != nil {
		return nil, err
	}

	ips := []string{}

	for _, node := range nodes.Items {
		for _, address := range node.Status.Addresses {
			if address.Type == corev1.NodeExternalIP {
				ips = append(ips, address.Address)
			}
		}
	}

	if len(ips) == 0 {
		for _, node := range nodes.Items {
			for _, address := range node.Status.Addresses {
				if address.Type == corev1.NodeInternalIP {
					ips = append(ips, address.Address)
				}
			}
		}
	}

	return ips, nil
}

//+kubebuilder:rbac:groups=vpn.wireguard-operator.io,resources=wireguards,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=vpn.wireguard-operator.io,resources=wireguards/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=vpn.wireguard-operator.io,resources=wireguards/finalizers,verbs=update

//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="apps",resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=nodes,verbs=list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Wireguard object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile

func (r *WireguardReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)

	log.Info("loaded the following wireguard image:" + r.AgentImage)
	wireguard := &v1alpha1.Wireguard{}
	log.Info(req.NamespacedName.Name)
	err := r.Get(ctx, req.NamespacedName, wireguard)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("wireguard resource not found. Ignoring as the resource must have be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get wireguard resource")
		return ctrl.Result{}, err
	}

	log.Info("reconciling " + wireguard.Name)

	if wireguard.Status.Status == "" {
		wireguard.Status.Status = v1alpha1.Pending
		wireguard.Status.Message = "Fetching Wireguard status"
		err = r.Status().Update(ctx, wireguard)

		if err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	secret := resources.Secret{
		Wireguard: wireguard,
		Logger:    log,
		Client:    r.Client,
		Scheme:    r.Scheme,
	}

	deployment := resources.Deployment{
		Wireguard:       wireguard,
		Logger:          log,
		AgentImage:      r.AgentImage,
		ImagePullPolicy: r.AgentImagePullPolicy,
		TargetPort:      port,
		MetricsPort:     metricsPort,
		Client:          r.Client,
		SecretName:      secret.Name(),
		Scheme:          r.Scheme,
	}

	service := resources.Service{
		Wireguard:  wireguard,
		Logger:     log,
		TargetPort: port,
		Client:     r.Client,
		Scheme:     r.Scheme,
	}

	peers := resources.Peers{
		Wireguard: wireguard,
		Logger:    log,
		Client:    r.Client,
	}
	resourcesList := []resources.Resource{
		secret,
		service,
		deployment,
		peers,
	}

	for _, res := range resourcesList {
		log.Info("reconciling resource " + res.Name())

		resourceStatus := v1alpha1.Resource{}
		for _, registeredResource := range wireguard.Status.Resources {
			if registeredResource.Name == res.Name() {
				resourceStatus = registeredResource
				break
			}
		}

		if resourceStatus.Name == "" {
			log.Info("creating resource " + res.Name())
			err = res.Create(ctx)
			if err != nil {
				return ctrl.Result{}, err
			}

			wireguard.Status.Resources = append(
				wireguard.Status.Resources, v1alpha1.Resource{
					Name:   res.Name(),
					Status: v1alpha1.Pending,
				},
			)
			err = r.Status().Update(ctx, wireguard)

			if err != nil {
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, nil
		}

		needsUpdate, err := res.NeedsUpdate(ctx)

		if err != nil {
			return ctrl.Result{}, err
		}

		if needsUpdate {
			log.Info("resource " + res.Name())
			err = res.Update(ctx)
			if err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

		status := v1alpha1.Pending

		converged, err := res.Converged(ctx)
		if err != nil {
			return ctrl.Result{}, err
		}

		if converged {
			status = v1alpha1.Ready
		}

		if status != resourceStatus.Status {
			resourceStatus.Status = status
			err = r.Status().Update(ctx, wireguard)
			return ctrl.Result{}, err
		}
	}

	if wireguard.Status.Status != v1alpha1.Ready {
		wireguard.Status.Status = v1alpha1.Ready
		wireguard.Status.Message = "VPN is active!"
		err = r.Status().Update(ctx, wireguard)

		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *WireguardReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Wireguard{}).
		Owns(&v1alpha1.WireguardPeer{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
