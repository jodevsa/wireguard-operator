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



type deployment struct {
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

func(r deployment) Type() string {
	return "deployment"
}

func(r deployment) Name() string {
	return fmt.Sprintf("%s-%s", r.wireguard.Name, r.wireguard.Status.UniqueIdentifier)
}

func(r deployment) Create(ctx context.Context) error {
	dep := r.deploymentForWireguard()
	r.logger.Info("Creating a new dep", "dep.Namespace", dep.Namespace, "dep.Name", dep.Name)
	err := r.client.Create(ctx, dep)
	if err != nil {
		r.logger.Error(err, "Failed to create new dep", "dep.Namespace", dep.Namespace, "dep.Name", dep.Name)
		return err
	}

	return nil
}

func(r deployment) Update(ctx context.Context) error {
	deployment := &appsv1.Deployment{}
	err := r.client.Get(ctx, types.NamespacedName{Name: r.Name(), Namespace: r.wireguard.Namespace}, deployment)
	if err != nil {
		r.logger.Error(err, "Failed to get deployment")
		return err
	}
	targetDep := r.deploymentForWireguard()

	if !cmp.Equal(deployment, targetDep) {

		r.client.Update(ctx, targetDep)
		if err != nil {
			r.logger.Error(err, "Failed to update deployment", "dep.Namespace", targetDep.Namespace, "dep.Name", targetDep.Name)
			return err
		}

	}

	return nil
}


func labelsForWireguard(name string) map[string]string {
	return map[string]string{"app": "wireguard", "instance": name}
}


func (r deployment) deploymentForWireguard() *appsv1.Deployment {
	ls := labelsForWireguard(r.Name())
	replicas := int32(1)

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Name(),
			Namespace: r.wireguard.Namespace,
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
									SecretName: r.secretName,
								},
							},
						}},
					InitContainers: []corev1.Container{},
					Containers: []corev1.Container{
						{
							SecurityContext: &corev1.SecurityContext{
								Capabilities: &corev1.Capabilities{Add: []corev1.Capability{"NET_ADMIN"}},
							},
							Image:           r.agentImage,
							ImagePullPolicy: r.ImagePullPolicy,
							Name:            "metrics",
							Command:         []string{"/usr/local/bin/prometheus_wireguard_exporter"},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: r.metricsPort,
									Name:          "metrics",
									Protocol:      corev1.ProtocolTCP,
								}},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "socket",
									MountPath: "/var/run/wireguard/",
								},
							},
						},
						{
							SecurityContext: &corev1.SecurityContext{
								Capabilities: &corev1.Capabilities{Add: []corev1.Capability{"NET_ADMIN"}},
							},
							Image:           r.agentImage,
							ImagePullPolicy: r.ImagePullPolicy,
							Name:            "agent",
							Command:         []string{"agent", "--v", "11", "--wg-iface", "wg0", "--wg-listen-port", fmt.Sprintf("%d", port), "--state", "/tmp/wireguard/state.json", "--wg-userspace-implementation-fallback", "wireguard-go"},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: r.targetPort,
									Name:          "wireguard",
									Protocol:      corev1.ProtocolUDP,
								}},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "socket",
									MountPath: "/var/run/wireguard/",
								},
								{
									Name:      "config",
									MountPath: "/tmp/wireguard/",
								}},
						}},
				},
			},
		},
	}

	if r.enableIpForwardOnPodInit {
		privileged := true
		dep.Spec.Template.Spec.InitContainers = append(dep.Spec.Template.Spec.InitContainers,
			corev1.Container{
				SecurityContext: &corev1.SecurityContext{
					Privileged: &privileged,
				},
				Image:           r.agentImage,
				ImagePullPolicy: r.ImagePullPolicy,
				Name:            "sysctl",
				Command:         []string{"/bin/sh"},
				Args:            []string{"-c", "echo 1 > /proc/sys/net/ipv4/ip_forward"},
			})
	}

	if r.useWgUserspaceImplementation {
		for i, c := range dep.Spec.Template.Spec.Containers {
			if c.Name == "agent" {
				dep.Spec.Template.Spec.Containers[i].Command = append(dep.Spec.Template.Spec.Containers[i].Command, "--wg-use-userspace-implementation")
			}
		}
	}

	ctrl.SetControllerReference(r.wireguard, dep, r.Scheme)
	return dep
}