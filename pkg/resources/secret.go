package resources

import (
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
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/kind/pkg/log"
)

type secret struct {
	wireguard *v1alpha1.Wireguard
	logger logr.Logger
	agentImage string
	ImagePullPolicy corev1.PullPolicy
	enableIpForwardOnPodInit bool
	targetPort int32
	metricsPort int32
	secretName string
	useWgUserspaceImplementation bool
	client client.Client
	Scheme *runtime.Scheme
}



func(s secret) Type() string {
	return "secret"
}

func(s secret) Name() string {
	return fmt.Sprintf("%s-%s", s.wireguard.Name, s.wireguard.Status.UniqueIdentifier)
}

func(s secret) getSecreteData(ctx context.Context) (map[string][]byte, error) {

	data := map[string][]byte{}

	peers, err := s.getPeersInfo(ctx)
	if err != nil {
		return data, err
	}

	sec, err := s.getExistingSecret(ctx)
	publicKey := ""
	privateKey := ""
	if errors.IsNotFound(err) {

		key, err := wgtypes.GeneratePrivateKey()
		if err!= nil {
			return data, nil
		}

		privateKey = key.String()
		publicKey = key.PublicKey().String()

	} else if err !=nil {
		privateKey = string(sec.Data["privateKey"])
		publicKey = string(sec.Data["publicKey"])
	} else {
		return data, err
	}
	state := agent.State{
		Server:           *s.wireguard.DeepCopy(),
		ServerPrivateKey: privateKey,
		Peers:            peers.Items,
	}


	b, err := json.Marshal(state)
	if err != nil {
		return data, err
	}

	data["state.json"] = b
	data["publicKey"] = []byte(publicKey)
	data["privateKey"] = []byte(privateKey)

	return data, nil
}


func (s *secret) getExistingSecret(ctx context.Context) (*corev1.Secret, error){
	sec := &corev1.Secret{}
	err := s.client.Get(ctx, types.NamespacedName{Name: s.Name(), Namespace: s.wireguard.Namespace}, sec)
	return sec, err
}

func (s *secret) getPeersInfo(ctx context.Context) (*v1alpha1.WireguardPeerList, error){
	// wireguardpeer
	peers := &v1alpha1.WireguardPeerList{}
	// TODO add a label to wireguardpeers and then filter by label here to only get peers of the wg instance we need.
	if err := s.client.List(ctx, peers, client.InNamespace(s.wireguard.Namespace)); err != nil {
		s.logger.Error(err, "Failed to fetch list of peers")
		return peers, err
	}
	return peers, nil
}

func (s *secret) secretForWireguard() *corev1.Secret {

	ls := labelsForWireguard(s.Name())
	dep := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Name(),
			Namespace: s.wireguard.Namespace,
			Labels:    ls,
		},
		Data: map[string][]byte{"state.json": state, "privateKey": []byte(privateKey), "publicKey": []byte(publicKey)},
	}

	ctrl.SetControllerReference(m, dep, r.Scheme)

	return dep

}
