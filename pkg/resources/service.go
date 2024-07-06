package resources

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/jodevsa/wireguard-operator/pkg/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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

func (r Service) Type() string {
	return "service"
}

func (r Service) Name() string {
	return fmt.Sprintf("%s-%s", r.Wireguard.Name, r.Wireguard.Status.UniqueIdentifier)
}

func (s Service) Converged(ctx context.Context) (bool, error) {
	svc := &corev1.Service{}
	err := s.Client.Get(ctx, types.NamespacedName{Name: s.Name(), Namespace: s.Wireguard.Namespace}, svc)
	if err != nil {
		s.Logger.Error(err, "Failed to get service", "svc.Namespace", svc.Namespace, "dep.Name", svc.Name)
		return false, err
	}

	if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
		ingressList := svc.Status.LoadBalancer.Ingress
		if len(ingressList) == 0 {
			return false, nil
		}
	}

	if svc.Spec.Type == corev1.ServiceTypeNodePort {
		if len(svc.Spec.Ports) == 0 {
			return false, nil
		}
	}
	return true, nil
}

func (s Service) NeedsUpdate(ctx context.Context) (bool, error) {
	// we don't support updating the service resource yet
	return false, nil
}

func (s Service) Update(ctx context.Context) error {
	return nil
}

func (s Service) Create(ctx context.Context) error {
	svc := s.serviceForWireguard()
	err := s.Client.Create(ctx, svc)
	if err != nil {
		s.Logger.Error(err, "Failed to create new service", "svc.Namespace", svc.Namespace, "dep.Name", svc.Name)
		return err
	}
	return nil
}

func (s Service) serviceType() corev1.ServiceType {
	serviceType := corev1.ServiceTypeLoadBalancer

	if s.Wireguard.Spec.ServiceType != "" {
		serviceType = s.Wireguard.Spec.ServiceType
	}

	return serviceType
}
func (s Service) serviceForWireguard() *corev1.Service {
	labels := labelsForWireguard(s.Name())
	svc := &corev1.Service{
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
			Type: s.serviceType(),
		},
	}
	ctrl.SetControllerReference(s.Wireguard, svc, s.Scheme)
	return svc
}
