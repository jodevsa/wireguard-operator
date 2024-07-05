package resources

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/jodevsa/wireguard-operator/pkg/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Service struct {
	wireguard *v1alpha1.Wireguard
	logger logr.Logger
	agentImage string
	targetPort int32
	ImagePullPolicy corev1.PullPolicy
	enableIpForwardOnPodInit bool
	useWgUserspaceImplementation bool
	client client.Client
	Scheme *runtime.Scheme
}



func(r Service) Type() string {
	return "service"
}

func(r Service) Name() string {
	return fmt.Sprintf("%s-%s", r.wireguard.Name, r.wireguard.Status.UniqueIdentifier)
}


func(s Service) Create(ctx context.Context) error {
	svc := s.serviceForWireguard()
	err := s.client.Create(ctx, svc)
	if err != nil {
		s.logger.Error(err, "Failed to create new service", "svc.Namespace", svc.Namespace, "dep.Name", svc.Name)
		return err
	}
	return nil
}




func(s Service) serviceForWireguard() *corev1.Service {
	labels := labelsForWireguard(s.Name())
	dep := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        s.Name(),
			Namespace:   s.wireguard.Namespace,
			Annotations: s.wireguard.Spec.ServiceAnnotations,
			Labels:      labels,
		},
		Spec: corev1.ServiceSpec{
			LoadBalancerIP: s.wireguard.Spec.Address,
			Selector:       labels,
			Ports: []corev1.ServicePort{{
				Protocol:   corev1.ProtocolUDP,
				NodePort:   s.wireguard.Spec.NodePort,
				Port:       s.targetPort,
				TargetPort: intstr.FromInt(int(s.targetPort)),
			}},
			Type: s.wireguard.Spec.ServiceType,
		},
	}
	ctrl.SetControllerReference(s.wireguard, dep, s.Scheme)
	return dep
}