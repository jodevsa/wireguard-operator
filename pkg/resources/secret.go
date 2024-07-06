package resources

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/jodevsa/wireguard-operator/pkg/agent"
	"github.com/jodevsa/wireguard-operator/pkg/api/v1alpha1"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Secret struct {
	Wireguard *v1alpha1.Wireguard
	Logger    logr.Logger
	Client    client.Client
	Scheme    *runtime.Scheme
}

func (s Secret) Converged(ctx context.Context) (bool, error) {
	return true, nil
}

func (s Secret) NeedsUpdate(ctx context.Context) (bool, error) {

	secret := &corev1.Secret{}
	err := s.Client.Get(ctx, types.NamespacedName{Name: s.Wireguard.Name, Namespace: s.Wireguard.Namespace}, secret)

	if err != nil {
		return true, err
	}

	expectedSecret, err := s.getSecreteData(ctx)

	if err != nil {
		return true, err
	}

	if !bytes.Equal(expectedSecret.Data["state.json"], secret.Data["state.json"]) {
		return true, nil

	}

	return false, nil
}

func (s Secret) Update(ctx context.Context) error {
	sec, err := s.getSecreteData(ctx)

	if err != nil {
		return err
	}
	if err := s.Client.Update(ctx, sec); err != nil {
		return err

	}

	return nil
}

func (s Secret) Create(ctx context.Context) error {
	sec, err := s.getSecreteData(ctx)

	if err != nil {
		return err
	}
	if err := s.Client.Create(ctx, sec); err != nil {
		return err

	}
	return nil
}
func (s Secret) Type() string {
	return "Secret"
}

func (s Secret) Name() string {
	return fmt.Sprintf("%s-%s", s.Wireguard.Name, s.Wireguard.Status.UniqueIdentifier)
}

func (s Secret) getSecreteData(ctx context.Context) (*corev1.Secret, error) {

	data := map[string][]byte{}

	peers, err := s.getPeersInfo(ctx)
	if err != nil {
		return &corev1.Secret{}, err
	}

	sec, err := s.getExistingSecret(ctx)
	publicKey := ""
	privateKey := ""
	if errors.IsNotFound(err) {

		key, err := wgtypes.GeneratePrivateKey()
		if err != nil {
			return &corev1.Secret{}, nil
		}

		privateKey = key.String()
		publicKey = key.PublicKey().String()

	} else if err != nil {
		privateKey = string(sec.Data["privateKey"])
		publicKey = string(sec.Data["publicKey"])
	} else {
		return &corev1.Secret{}, err
	}
	state := agent.State{
		Server:           *s.Wireguard.DeepCopy(),
		ServerPrivateKey: privateKey,
		Peers:            peers.Items,
	}

	b, err := json.Marshal(state)
	if err != nil {
		return &corev1.Secret{}, err
	}

	data["state.json"] = b
	data["publicKey"] = []byte(publicKey)
	data["privateKey"] = []byte(privateKey)

	sec = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Name(),
			Namespace: s.Wireguard.Namespace,
			Labels:    labelsForWireguard(s.Wireguard.Name),
		},
		Data: data,
	}
	ctrl.SetControllerReference(s.Wireguard, sec, s.Scheme)
	return sec, nil
}

func (s *Secret) getExistingSecret(ctx context.Context) (*corev1.Secret, error) {
	sec := &corev1.Secret{}
	err := s.Client.Get(ctx, types.NamespacedName{Name: s.Name(), Namespace: s.Wireguard.Namespace}, sec)
	return sec, err
}

func (s *Secret) getPeersInfo(ctx context.Context) (*v1alpha1.WireguardPeerList, error) {
	// wireguardpeer
	peers := &v1alpha1.WireguardPeerList{}
	// TODO add a label to wireguardpeers and then filter by label here to only get peers of the wg instance we need.
	if err := s.Client.List(ctx, peers, client.InNamespace(s.Wireguard.Namespace)); err != nil {
		s.Logger.Error(err, "Failed to fetch list of peers")
		return peers, err
	}
	return peers, nil
}
