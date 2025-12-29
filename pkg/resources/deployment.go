package resources

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/jodevsa/wireguard-operator/pkg/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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
	SecretResourceVersion        string
	UseWgUserspaceImplementation bool
	Client                       client.Client
	Scheme                       *runtime.Scheme
}

func (d Deployment) Type() string {
	return "deployment"
}

func (d Deployment) Name() string {
	return fmt.Sprintf("%s-%s", d.Wireguard.Name, d.Wireguard.Status.UniqueIdentifier)
}

func (d Deployment) Converged(ctx context.Context) (bool, error) {
	dep := &appsv1.Deployment{}
	err := d.Client.Get(ctx, types.NamespacedName{Name: d.Name(), Namespace: d.Wireguard.Namespace}, dep)
	if err != nil {
		return false, err
	}

	if dep.Status.ReadyReplicas == *dep.Spec.Replicas {
		return true, nil
	}

	return false, nil

}

func (d Deployment) NeedsUpdate(ctx context.Context) (bool, error) {
	dep := &appsv1.Deployment{}
	err := d.Client.Get(ctx, types.NamespacedName{Name: d.Name(), Namespace: d.Wireguard.Namespace}, dep)
	if err != nil {
		d.Logger.Error(err, "Failed to get dep", "dep.Namespace", dep.Namespace, "dep.Name", dep.Name)
		return true, err
	}
	// only update if image needs to be updated
	if dep.Spec.Template.Spec.Containers[0].Image != d.AgentImage || dep.Annotations["secretResourceVersion"] != d.SecretResourceVersion {
		return true, nil
	}

	return false, nil
}

func (d Deployment) Exists(ctx context.Context) (bool, error) {
	dep := &appsv1.Deployment{}
	err := d.Client.Get(ctx, types.NamespacedName{Name: d.Name(), Namespace: d.Wireguard.Namespace}, dep)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		d.Logger.Error(err, "Failed to get dep", "dep.Namespace", dep.Namespace, "dep.Name", dep.Name)
		return true, err
	}

	return true, nil
}

func (d Deployment) Create(ctx context.Context) error {
	dep := d.deploymentForWireguard()
	d.Logger.Info("Creating a new dep", "dep.Namespace", dep.Namespace, "dep.Name", dep.Name)
	err := d.Client.Create(ctx, dep)
	if err != nil {
		d.Logger.Error(err, "Failed to create new dep", "dep.Namespace", dep.Namespace, "dep.Name", dep.Name)
		return err
	}

	return nil
}

func (d Deployment) Update(ctx context.Context) error {
	deployment := &appsv1.Deployment{}
	err := d.Client.Get(ctx, types.NamespacedName{Name: d.Name(), Namespace: d.Wireguard.Namespace}, deployment)
	if err != nil {
		d.Logger.Error(err, "Failed to get Deployment")
		return err
	}
	targetDep := d.deploymentForWireguard()

	if !cmp.Equal(deployment, targetDep) {

		d.Client.Update(ctx, targetDep)
		if err != nil {
			d.Logger.Error(err, "Failed to update Deployment", "dep.Namespace", targetDep.Namespace, "dep.Name", targetDep.Name)
			return err
		}

	}

	pods := &corev1.PodList{}
	if err := d.Client.List(ctx, pods, client.MatchingLabels{"app": "wireguard", "instance": d.Wireguard.Name}); err != nil {
		d.Logger.Error(err, "Failed to fetch list of pods")
		return err
	}
	for _, pod := range pods.Items {
		if pod.Annotations == nil {
			pod.Annotations = make(map[string]string)
		}
		// this is needed to force k8s to push the new secret to the pod
		pod.Annotations["secretResourceVersion"] = d.SecretResourceVersion
		if err := d.Client.Update(ctx, &pod); err != nil {
			d.Logger.Error(err, "Failed to update pod")
			return err
		}
	}

	return nil
}

func createLabelForInsntance(name string) map[string]string {
	return map[string]string{"app": "wireguard", "instance": name}
}

func (d Deployment) deploymentForWireguard() *appsv1.Deployment {
	ls := createLabelForInsntance(d.Wireguard.Name)
	replicas := int32(1)

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.Name(),
			Namespace: d.Wireguard.Namespace,
			Labels:    ls,
			Annotations: map[string]string{
				"secretResourceVersion": d.SecretResourceVersion,
			},
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
									SecretName: d.SecretName,
								},
							},
						}},
					InitContainers: []corev1.Container{},
					Containers: []corev1.Container{
						{
							SecurityContext: &corev1.SecurityContext{
								Capabilities: &corev1.Capabilities{Add: []corev1.Capability{"NET_ADMIN"}},
							},
							Image:           d.AgentImage,
							ImagePullPolicy: d.ImagePullPolicy,
							Name:            "metrics",
							Command:         []string{"/usr/local/bin/prometheus_wireguard_exporter"},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: d.MetricsPort,
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
							Image:           d.AgentImage,
							ImagePullPolicy: d.ImagePullPolicy,
							Name:            "agent",
							Command:         []string{"agent", "--v", "11", "--wg-iface", "wg0", "--wg-listen-port", fmt.Sprintf("%d", d.TargetPort), "--state", "/tmp/Wireguard/state.json", "--wg-userspace-implementation-fallback", "Wireguard-go"},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: d.TargetPort,
									Name:          "wireguard",
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

	if d.Wireguard.Spec.EnableIpForwardOnPodInit {
		privileged := true
		dep.Spec.Template.Spec.InitContainers = append(dep.Spec.Template.Spec.InitContainers,
			corev1.Container{
				SecurityContext: &corev1.SecurityContext{
					Privileged: &privileged,
				},
				Image:           d.AgentImage,
				ImagePullPolicy: d.ImagePullPolicy,
				Name:            "sysctl",
				Command:         []string{"/bin/sh"},
				Args:            []string{"-c", "echo 1 > /proc/sys/net/ipv4/ip_forward"},
			})
	}

	if d.UseWgUserspaceImplementation {
		for i, c := range dep.Spec.Template.Spec.Containers {
			if c.Name == "agent" {
				dep.Spec.Template.Spec.Containers[i].Command = append(dep.Spec.Template.Spec.Containers[i].Command, "--wg-use-userspace-implementation")
			}
		}
	}

	ctrl.SetControllerReference(d.Wireguard, dep, d.Scheme)
	return dep
}
