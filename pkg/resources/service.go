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
	Wireguard  *v1alpha1.Wireguard
	Logger     logr.Logger
	TargetPort int32
	Client     client.Client
	Scheme     *runtime.Scheme
}



func(r Service) Type() string {
	return "service"
}

func(r Service) Name() string {
	return fmt.Sprintf("%s-%s", r.Wireguard.Name, r.Wireguard.Status.UniqueIdentifier)
}



func(s Service) Update(ctx context.Context) error {
	return nil
}

func(s Service) Create(ctx context.Context) error {
	svc := s.serviceForWireguard()
	err := s.Client.Create(ctx, svc)
	if err != nil {
		s.Logger.Error(err, "Failed to create new service", "svc.Namespace", svc.Namespace, "dep.Name", svc.Name)
		return err
	}
	return nil
}




func(s Service) serviceForWireguard() *corev1.Service {
	labels := labelsForWireguard(s.Name())
	dep := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        s.Name(),
			Namespace:   s.Wireguard.Namespace,
			Annotations: s.Wireguard.Spec.ServiceAnnotations,
			Labels:      labels,
		},
		Spec: corev1.ServiceSpec{
			LoadBalancerIP: s.Wireguard.Spec.Address,
			Selector:       labels,
			Ports: []corev1.ServicePort{{
				Protocol:   corev1.ProtocolUDP,
				NodePort:   s.Wireguard.Spec.NodePort,
				Port:       s.TargetPort,
				TargetPort: intstr.FromInt(int(s.TargetPort)),
			}},
			Type: s.Wireguard.Spec.ServiceType,
		},
	}
	ctrl.SetControllerReference(s.Wireguard, dep, s.Scheme)
	return dep
}