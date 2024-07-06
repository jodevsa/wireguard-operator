package resources

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/jodevsa/wireguard-operator/pkg/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Deployment struct {
	Wireguard                    *v1alpha1.Wireguard
	Logger                       logr.Logger
	AgentImage                   string
	ImagePullPolicy              corev1.PullPolicy
	TargetPort                   int32
	MetricsPort                  int32
	SecretName                   string
	UseWgUserspaceImplementation bool
	Client                       client.Client
	Scheme                       *runtime.Scheme
}

func (r Deployment) Type() string {
	return "deployment"
}

func (r Deployment) Name() string {
	return fmt.Sprintf("%s-%s", r.Wireguard.Name, r.Wireguard.Status.UniqueIdentifier)
}

func (r Deployment) Converged(ctx context.Context) (bool, error) {
	return true, nil
}

func (r Deployment) NeedsUpdate(ctx context.Context) (bool, error) {
	dep := &appsv1.Deployment{}
	err := r.Client.Get(ctx, types.NamespacedName{Name: r.Wireguard.Name + "-dep", Namespace: r.Wireguard.Namespace}, dep)
	if err != nil {
		r.Logger.Error(err, "Failed to get dep", "dep.Namespace", dep.Namespace, "dep.Name", dep.Name)
		return true, err
	}
	// only update if image needs to be updated
	if dep.Spec.Template.Spec.Containers[0].Image != r.AgentImage {
		return true, nil
	}

	return false, nil
}

func (r Deployment) Create(ctx context.Context) error {
	dep := r.deploymentForWireguard()
	r.Logger.Info("Creating a new dep", "dep.Namespace", dep.Namespace, "dep.Name", dep.Name)
	err := r.Client.Create(ctx, dep)
	if err != nil {
		r.Logger.Error(err, "Failed to create new dep", "dep.Namespace", dep.Namespace, "dep.Name", dep.Name)
		return err
	}

	return nil
}

func (r Deployment) Update(ctx context.Context) error {
	deployment := &appsv1.Deployment{}
	err := r.Client.Get(ctx, types.NamespacedName{Name: r.Name(), Namespace: r.Wireguard.Namespace}, deployment)
	if err != nil {
		r.Logger.Error(err, "Failed to get Deployment")
		return err
	}
	targetDep := r.deploymentForWireguard()

	if !cmp.Equal(deployment, targetDep) {

		r.Client.Update(ctx, targetDep)
		if err != nil {
			r.Logger.Error(err, "Failed to update Deployment", "dep.Namespace", targetDep.Namespace, "dep.Name", targetDep.Name)
			return err
		}

	}

	return nil
}

func labelsForWireguard(name string) map[string]string {
	return map[string]string{"app": "Wireguard", "instance": name}
}

func (r Deployment) deploymentForWireguard() *appsv1.Deployment {
	ls := labelsForWireguard(r.Wireguard.Name)
	replicas := int32(1)

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Name(),
			Namespace: r.Wireguard.Namespace,
			Labels:    ls,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "socket",
							VolumeSource: corev1.VolumeSource{

								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						{

							Name: "config",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: r.SecretName,
								},
							},
						}},
					InitContainers: []corev1.Container{},
					Containers: []corev1.Container{
						{
							SecurityContext: &corev1.SecurityContext{
								Capabilities: &corev1.Capabilities{Add: []corev1.Capability{"NET_ADMIN"}},
							},
							Image:           r.AgentImage,
							ImagePullPolicy: r.ImagePullPolicy,
							Name:            "metrics",
							Command:         []string{"/usr/local/bin/prometheus_wireguard_exporter"},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: r.MetricsPort,
									Name:          "metrics",
									Protocol:      corev1.ProtocolTCP,
								}},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "socket",
									MountPath: "/var/run/Wireguard/",
								},
							},
						},
						{
							SecurityContext: &corev1.SecurityContext{
								Capabilities: &corev1.Capabilities{Add: []corev1.Capability{"NET_ADMIN"}},
							},
							Image:           r.AgentImage,
							ImagePullPolicy: r.ImagePullPolicy,
							Name:            "agent",
							Command:         []string{"agent", "--v", "11", "--wg-iface", "wg0", "--wg-listen-port", fmt.Sprintf("%d", r.TargetPort), "--state", "/tmp/Wireguard/state.json", "--wg-userspace-implementation-fallback", "Wireguard-go"},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: r.TargetPort,
									Name:          "Wireguard",
									Protocol:      corev1.ProtocolUDP,
								}},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "socket",
									MountPath: "/var/run/Wireguard/",
								},
								{
									Name:      "config",
									MountPath: "/tmp/Wireguard/",
								}},
						}},
				},
			},
		},
	}

	if r.Wireguard.Spec.EnableIpForwardOnPodInit {
		privileged := true
		dep.Spec.Template.Spec.InitContainers = append(dep.Spec.Template.Spec.InitContainers,
			corev1.Container{
				SecurityContext: &corev1.SecurityContext{
					Privileged: &privileged,
				},
				Image:           r.AgentImage,
				ImagePullPolicy: r.ImagePullPolicy,
				Name:            "sysctl",
				Command:         []string{"/bin/sh"},
				Args:            []string{"-c", "echo 1 > /proc/sys/net/ipv4/ip_forward"},
			})
	}

	if r.UseWgUserspaceImplementation {
		for i, c := range dep.Spec.Template.Spec.Containers {
			if c.Name == "agent" {
				dep.Spec.Template.Spec.Containers[i].Command = append(dep.Spec.Template.Spec.Containers[i].Command, "--wg-use-userspace-implementation")
			}
		}
	}

	ctrl.SetControllerReference(r.Wireguard, dep, r.Scheme)
	return dep
}
