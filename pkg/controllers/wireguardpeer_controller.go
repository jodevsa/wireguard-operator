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

	wgtypes "golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// WireguardPeerReconciler reconciles a WireguardPeer object
type WireguardPeerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *WireguardPeerReconciler) updateStatus(ctx context.Context, peer *v1alpha1.WireguardPeer, status string, message string) error {
	newPeer := peer.DeepCopy()
	if newPeer.Status.Status != status || newPeer.Status.Message != message {
		newPeer.Status.Status = status
		newPeer.Status.Message = message

		if err := r.Status().Update(ctx, newPeer); err != nil {
			return err
		}
	}
	return nil
}

func (r *WireguardPeerReconciler) secretForPeer(m *v1alpha1.WireguardPeer, privateKey string, publicKey string) *corev1.Secret {
	ls := labelsForWireguard(m.Name)
	dep := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name + "-peer",
			Namespace: m.Namespace,
			Labels:    ls,
		},
		Data: map[string][]byte{"privateKey": []byte(privateKey), "publicKey": []byte(publicKey)},
	}
	// Set Nodered instance as the owner and controller
	ctrl.SetControllerReference(m, dep, r.Scheme)

	return dep

}

//+kubebuilder:rbac:groups=vpn.wireguard-operator.io,resources=wireguardpeers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=vpn.wireguard-operator.io,resources=wireguardpeers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=vpn.wireguard-operator.io,resources=wireguardpeers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the WireguardPeer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile

func (r *WireguardPeerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	peer := &v1alpha1.WireguardPeer{}
	err := r.Get(ctx, req.NamespacedName, peer)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("wireguard peer resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get wireguard peer")
		return ctrl.Result{}, err
	}

	key, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		log.Error(err, "Failed to generate private key")
		return ctrl.Result{}, err
	}

	newPeer := peer.DeepCopy()
	if newPeer.Status.Status == "" {
		err = r.updateStatus(ctx, newPeer, v1alpha1.Pending, "Waiting for wireguard peer to be created")

		if err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{Requeue: true}, nil
	}

	if peer.Spec.PublicKey == "" {
		privateKey := key.String()
		publicKey := key.PublicKey().String()

		secret := r.secretForPeer(peer, privateKey, publicKey)

		log.Info("Creating a new secret", "secret.Namespace", secret.Namespace, "secret.Name", secret.Name)
		err = r.Create(ctx, secret)
		if err != nil {
			log.Error(err, "Failed to create new secret", "secret.Namespace", secret.Namespace, "secret.Name", secret.Name)
			return ctrl.Result{}, err
		}

		newPeer.Spec.PublicKey = publicKey
		newPeer.Spec.PrivateKey = v1alpha1.PrivateKey{
			SecretKeyRef: corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: peer.Name + "-peer"}, Key: "privateKey"}}
		err = r.Update(ctx, newPeer)

		if err != nil {
			log.Error(err, "Failed to create new peer", "secret.Namespace", secret.Namespace, "secret.Name", secret.Name)
			return ctrl.Result{}, err
		}

		return ctrl.Result{Requeue: true}, nil

	}

	wireguard := &v1alpha1.Wireguard{}
	err = r.Get(ctx, types.NamespacedName{Name: newPeer.Spec.WireguardRef, Namespace: newPeer.Namespace}, wireguard)

	if err != nil {
		if errors.IsNotFound(err) {
			err = r.updateStatus(ctx, newPeer, v1alpha1.Error, fmt.Sprintf("Waiting for wireguard resource '%s' to be created", newPeer.Spec.WireguardRef))

			if err != nil {
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, nil
		}

		log.Error(err, "Failed to get wireguard")

		return ctrl.Result{}, err

	}

	if wireguard.Status.Status != v1alpha1.Ready {
		err = r.updateStatus(ctx, newPeer, v1alpha1.Error, fmt.Sprintf("Waiting for %s to be ready", wireguard.Name))

		if err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	if len(newPeer.OwnerReferences) == 0 {
		log.Info("Waiting for owner reference to be set " + wireguard.Name + " " + newPeer.Name)
		ctrl.SetControllerReference(wireguard, newPeer, r.Scheme)

		if newPeer.Labels == nil {
			newPeer.Labels = map[string]string{}
		}
		newPeer.Labels["app"] = "wireguard"
		newPeer.Labels["instance"] = wireguard.Name

		err = r.Update(ctx, newPeer)

		if err != nil {
			log.Error(err, "Failed to update peer with controller reference and labels")
			return ctrl.Result{}, err
		}

		return ctrl.Result{Requeue: true}, nil
	}

	if newPeer.Status.Config == "" {
		err = r.updateStatus(ctx, newPeer, v1alpha1.Pending, "Waiting config to be updated")

		if err != nil {
			return ctrl.Result{}, err
		}

	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *WireguardPeerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.WireguardPeer{}).
		Complete(r)
}
