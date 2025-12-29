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
	"fmt"
	"github.com/jodevsa/wireguard-operator/pkg/api/v1alpha1"
	"github.com/jodevsa/wireguard-operator/pkg/resources"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"math/rand"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

// WireguardReconciler reconciles a Wireguard object

const port = 51820
const httpPort = 8080

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

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"0123456789"

func randomString(length int) string {
	return stringWithCharset(length, charset)
}

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func stringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func (r *WireguardReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)

	wireguard := &v1alpha1.Wireguard{}
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

	secret := resources.Secret{
		Wireguard: wireguard,
		Logger:    log,
		Client:    r.Client,
		Scheme:    r.Scheme,
	}

	deployment := resources.Deployment{
		Wireguard:             wireguard,
		Logger:                log,
		AgentImage:            r.AgentImage,
		ImagePullPolicy:       r.AgentImagePullPolicy,
		TargetPort:            port,
		MetricsPort:           metricsPort,
		Client:                r.Client,
		SecretName:            secret.Name(),
		SecretResourceVersion: secret.GetResourceVersion(ctx),
		Scheme:                r.Scheme,
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

	if wireguard.Status.Status == "" {
		wireguard.Status.UniqueIdentifier = randomString(5)

		for _, res := range resourcesList {
			wireguard.Status.Resources = append(
				wireguard.Status.Resources, v1alpha1.Resource{
					Name:   res.Name(),
					Type:   res.Type(),
					Status: v1alpha1.Pending,
				},
			)
		}

		wireguard.Status.Status = v1alpha1.Pending
		wireguard.Status.Message = "Fetching Wireguard status"
		err = r.Status().Update(ctx, wireguard)

		if err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	for _, res := range resourcesList {
		log.Info(fmt.Sprintf("reconciling WG resource %s", res.Name()))

		resourceStatus := &v1alpha1.Resource{}
		for i, _ := range wireguard.Status.Resources {
			registeredResource := &wireguard.Status.Resources[i]
			if registeredResource.Name == res.Name() && registeredResource.Type == res.Type() {
				resourceStatus = registeredResource
				break
			}
		}
		resourceExists, err := res.Exists(ctx)
		if err != nil {
			return ctrl.Result{}, err
		}
		if !resourceExists {
			log.Info(fmt.Sprintf("creating resource '%s' with type '%s'", res.Name(), res.Type()))
			return ctrl.Result{}, res.Create(ctx)
		}
		log.Info(fmt.Sprintf("checking if resource %s with type %s needs an update", res.Name(), res.Type()))
		needsUpdate, err := res.NeedsUpdate(ctx)

		if err != nil {
			log.Info(fmt.Sprintf("unable to check if resource %s with type %s needs an update", res.Name(), res.Type()))
			return ctrl.Result{}, err
		}

		if needsUpdate {
			log.Info(fmt.Sprintf("resource %s with type %s needs to be updated", res.Name(), res.Type()))
			err = res.Update(ctx)
			if err != nil {
				return ctrl.Result{}, err
			}
			log.Info(fmt.Sprintf("resource %s with type %s was updated", res.Name(), res.Type()))
			return ctrl.Result{Requeue: true}, nil
		}

		status := v1alpha1.Pending

		log.Info(fmt.Sprintf("checking if resource %s with type %s has converged", res.Name(), res.Type()))
		converged, err := res.Converged(ctx)
		if err != nil {
			return ctrl.Result{}, err
		}

		if converged {
			log.Info(fmt.Sprintf("resource %s with type %s has converged sucessfully", res.Name(), res.Type()))
			status = v1alpha1.Ready
		} else {
			log.Info(fmt.Sprintf("resource %s with type %s has not converged yet", res.Name(), res.Type()))
			return ctrl.Result{Requeue: true}, err
		}

		if status != resourceStatus.Status {
			resourceStatus.Status = status
			err = r.Status().Update(ctx, wireguard)
			log.Info(fmt.Sprintf("Updating resource status...%v", resourceStatus))
			return ctrl.Result{Requeue: true}, err
		}
	}

	// update WG public key
	publicKey, err := secret.GetPublicKey(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	if wireguard.Status.PublicKey != publicKey {
		wireguard.Status.Status = v1alpha1.Pending
		wireguard.Status.PublicKey = publicKey
		err = r.Status().Update(ctx, wireguard)
		return ctrl.Result{Requeue: true}, err
	}

	// update WG address
	address, port, err := service.GetAddressAndPort(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}
	if wireguard.Status.Address != address || wireguard.Status.Port != port {
		wireguard.Status.Port = port
		wireguard.Status.Address = address
		wireguard.Status.Status = v1alpha1.Pending
		err = r.Status().Update(ctx, wireguard)
		return ctrl.Result{Requeue: true}, err
	}

	if wireguard.Status.Status != v1alpha1.Ready {

		dnsAddress := "1.1.1.1"

		if wireguard.Spec.Dns != "" {
			dnsAddress = wireguard.Spec.Dns
		}
		wireguard.Status.Dns = dnsAddress
		wireguard.Status.Status = v1alpha1.Ready
		wireguard.Status.Message = "VPN is active!"
		err = r.Status().Update(ctx, wireguard)

		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
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
