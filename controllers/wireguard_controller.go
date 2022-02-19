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
	"time"

	vpnv1alpha1 "github.com/jodevsa/wireguard-operator/api/v1alpha1"
	"github.com/korylprince/ipnetgen"
	wgtypes "golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// WireguardReconciler reconciles a Wireguard object

const port = 51820

type WireguardReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func labelsForWireguard(name string) map[string]string {
	return map[string]string{"wireguard_cr": name}
}

func (r *WireguardReconciler) ConfigmapForWireguard(m *vpnv1alpha1.Wireguard, hostname string) *corev1.ConfigMap {
	ls := labelsForWireguard(m.Name)
	dep := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name + "-config",
			Namespace: m.Namespace,
			Labels:    ls,
		},
		Data: map[string]string{
			"SERVERURL":       hostname,
			"PUID":            "1000",
			"PGID":            "1000",
			"TZ":              "America/Mexico_City",
			"SERVERPORT":      fmt.Sprint(port),
			"PEERS":           "2",
			"PEERDNS":         "169.254.169.253",
			"ALLOWEDIPS":      "0.0.0.0/0, ::/0",
			"INTERNAL_SUBNET": "10.13.13.0",
		}}

	ctrl.SetControllerReference(m, dep, r.Scheme)
	return dep
}

func (r *WireguardReconciler) getWireguardPeers(ctx context.Context, req ctrl.Request) (*vpnv1alpha1.WireguardPeerList, error) {
	peers := &vpnv1alpha1.WireguardPeerList{}
	if err := r.List(ctx, peers, client.InNamespace(req.Namespace)); err != nil {
		return nil, err
	}

	relatedPeers := &vpnv1alpha1.WireguardPeerList{}

	for _, peer := range peers.Items {
		if peer.Spec.WireguardRef == req.Name {
			relatedPeers.Items = append(relatedPeers.Items, peer)
		}
	}

	return relatedPeers, nil
}

func (r *WireguardReconciler) updateStatus(ctx context.Context, req ctrl.Request, wireguard *vpnv1alpha1.Wireguard, status vpnv1alpha1.WgStatusReport) error {
	newWireguard := wireguard.DeepCopy()
	if newWireguard.Status.Status != status.Status || newWireguard.Status.Message != status.Message {
		newWireguard.Status.Status = status.Status
		newWireguard.Status.Message = status.Message

		if err := r.Status().Update(ctx, newWireguard); err != nil {
			return err
		}
	}
	return nil
}

func getAvaialbleIp(cidr string, usedIps []string) (string, error) {
	gen, err := ipnetgen.New(cidr)
	if err != nil {
		return "", err
	}
	for ip := gen.Next(); ip != nil; ip = gen.Next() {
		used := false
		for _, usedIp := range usedIps {
			if ip.String() == usedIp {
				used = true
				break
			}
		}
		if !used {
			return ip.String(), nil
		}
	}

	return "", fmt.Errorf("No available ip found in %s", cidr)
}

func (r *WireguardReconciler) getUsedIps(peers *vpnv1alpha1.WireguardPeerList) []string {
	usedIps := []string{"10.8.0.0", "10.8.0.1"}
	for _, p := range peers.Items {
		usedIps = append(usedIps, p.Spec.Address)

	}

	return usedIps
}

func (r *WireguardReconciler) updateWireguardPeers(ctx context.Context, req ctrl.Request, wireguard *vpnv1alpha1.Wireguard, serverAddress string, serverPublicKey string, serverMtu string) error {

	peers, err := r.getWireguardPeers(ctx, req)
	if err != nil {
		return err
	}

	usedIps := r.getUsedIps(peers)

	for _, peer := range peers.Items {
		if peer.Spec.Address == "" {
			ip, err := getAvaialbleIp("10.8.0.0/24", usedIps)

			if err != nil {
				return err
			}

			peer.Spec.Address = ip

			if err := r.Update(ctx, &peer); err != nil {
				return err
			}

			usedIps = append(usedIps, ip)
		}

		newConfig := fmt.Sprintf(`
echo "
[Interface]
PrivateKey = $(kubectl get secret %s-peer --template={{.data.privateKey}} -n %s | base64 -d)
Address = %s
DNS = 1.1.1.1`, peer.Name, peer.Namespace, peer.Spec.Address)

		if serverMtu != "" {
			newConfig = newConfig + "\nMTU = " + serverMtu
		}

		newConfig = newConfig + fmt.Sprintf(`

[Peer]
PublicKey = %s
AllowedIPs = 0.0.0.0/0
Endpoint = %s:%s"`, serverPublicKey, serverAddress, wireguard.Status.Port)
		if peer.Status.Config != newConfig || peer.Status.Status != vpnv1alpha1.Ready {
			peer.Status.Config = newConfig
			peer.Status.Status = vpnv1alpha1.Ready
			peer.Status.Message = "Peer configured"
			if err := r.Status().Update(ctx, &peer); err != nil {
				return err
			}
		}
	}

	return nil
}

//+kubebuilder:rbac:groups=vpn.example.com,resources=wireguards,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=vpn.example.com,resources=wireguards/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=vpn.example.com,resources=wireguards/finalizers,verbs=update

//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=deployments,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Wireguard object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *WireguardReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)

	wireguard := &vpnv1alpha1.Wireguard{}
	log.Info(req.NamespacedName.Name)
	err := r.Get(ctx, req.NamespacedName, wireguard)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("wireguard resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get wireguard")
		return ctrl.Result{}, err
	}

	log.Info("processing " + wireguard.Name)

	if wireguard.Status.Status == "" {
		err = r.updateStatus(ctx, req, wireguard, vpnv1alpha1.WgStatusReport{Status: vpnv1alpha1.Pending, Message: "Fetching Wireguard status"})

		if err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{Requeue: true}, nil
	}

	wgConfig := ""

	// wireguardpeer
	peers := &vpnv1alpha1.WireguardPeerList{}
	if err := r.List(ctx, peers, client.InNamespace(req.Namespace)); err != nil {
		log.Error(err, "Failed to fetch list of peers")
		return ctrl.Result{}, err
	}

	for _, peer := range peers.Items {

		if peer.Spec.WireguardRef != wireguard.Name {
			continue
		}
		if peer.Spec.PublicKey == "" {
			continue
		}

		wgConfig = wgConfig + fmt.Sprintf("\n[Peer]\nPublicKey = %s\nallowedIps = %s\n\n", peer.Spec.PublicKey, peer.Spec.Address)
	}

	// svc

	secret := &corev1.Secret{}
	err = r.Get(ctx, types.NamespacedName{Name: wireguard.Name, Namespace: wireguard.Namespace}, secret)
	if err == nil {
		privateKey := string(secret.Data["privateKey"])

		wgConfig = fmt.Sprintf(`
[Interface]
PrivateKey = %s
Address = 10.8.0.1/24
ListenPort = 51820
`, privateKey) + wgConfig

		if string(secret.Data["config"]) != wgConfig {
			log.Info("Updating secret with new config")

			publicKey := string(secret.Data["publicKey"])
			err := r.Update(ctx, r.secretForWireguard(wireguard, privateKey, publicKey, wgConfig))
			if err != nil {
				log.Error(err, "Failed to update secret with new config")
				return ctrl.Result{}, err
			}

			pods := &corev1.PodList{}
			if err := r.List(ctx, pods, client.HasLabels([]string{"SS"})); err != nil {
				log.Error(err, "Failed to fetch list of pods")
				return ctrl.Result{}, err
			}

			for _, pod := range pods.Items {
				if pod.Annotations == nil {
					pod.Annotations = make(map[string]string)
				}

				pod.Annotations["wgConfigLastUpdated"] = time.Now().Format("2006-01-02T15-04-05")
				if err := r.Update(ctx, &pod); err != nil {
					log.Error(err, "Failed to update pod")
					return ctrl.Result{}, err
				}

				log.Info("updated pod")
			}

		}

		wireguard.Annotations["wgConfigLastUpdated"] = time.Now().Format("2006-01-02T15-04-05")
		r.Update(ctx, wireguard)

	}
	if err != nil && errors.IsNotFound(err) {

		key, err := wgtypes.GeneratePrivateKey()

		privateKey := key.String()
		publicKey := key.PublicKey().String()

		if err != nil {
			log.Error(err, "Failed to generate private key")
			return ctrl.Result{}, err
		}

		secret := r.secretForWireguard(wireguard, privateKey, publicKey, wgConfig)

		log.Info("Creating a new secret", "secret.Namespace", secret.Namespace, "secret.Name", secret.Name)
		err = r.Create(ctx, secret)
		if err != nil {
			log.Error(err, "Failed to create new secret", "secret.Namespace", secret.Namespace, "secret.Name", secret.Name)
			return ctrl.Result{}, err
		}

		clientKey, err := wgtypes.GeneratePrivateKey()

		if err != nil {
			log.Error(err, "Failed to generate private key")
			return ctrl.Result{}, err
		}

		clientSecret := r.secretForClient(wireguard, clientKey.String(), clientKey.PublicKey().String())

		log.Info("Creating a new secret", "secret.Namespace", clientSecret.Namespace, "secret.Name", clientSecret.Name)
		err = r.Create(ctx, clientSecret)
		if err != nil {
			log.Error(err, "Failed to create new secret", "secret.Namespace", clientSecret.Namespace, "secret.Name", clientSecret.Name)
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, err
	} else if err != nil {
		log.Error(err, "Failed to get secret")
		return ctrl.Result{}, err
	}

	svcFound := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: wireguard.Name + "-svc", Namespace: wireguard.Namespace}, svcFound)
	if err != nil && errors.IsNotFound(err) {

		svc := r.serviceForWireguard(wireguard)
		log.Info("Creating a new service", "service.Namespace", svc.Namespace, "service.Name", svc.Name)
		err = r.Create(ctx, svc)
		if err != nil {
			log.Error(err, "Failed to create new service", "service.Namespace", svc.Namespace, "service.Name", svc.Name)
			return ctrl.Result{}, err
		}
		// svc created successfully - return and requeue

		err = r.updateStatus(ctx, req, wireguard, vpnv1alpha1.WgStatusReport{Status: vpnv1alpha1.Pending, Message: "Waiting for service to be created"})

		if err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	} else if err != nil {
		log.Error(err, "Failed to get service")
		return ctrl.Result{}, err
	}

	ingressList := svcFound.Status.LoadBalancer.Ingress

	if len(ingressList) == 0 {
		err = r.updateStatus(ctx, req, wireguard, vpnv1alpha1.WgStatusReport{Status: vpnv1alpha1.Pending, Message: "Waiting for service to be ready"})
		if err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	hostname := svcFound.Status.LoadBalancer.Ingress[0].Hostname
	if hostname == "" {
		hostname = svcFound.Status.LoadBalancer.Ingress[0].IP
	}

	if wireguard.Status.Hostname == "" {
		updateWireguard := wireguard.DeepCopy()
		updateWireguard.Status.Hostname = hostname
		updateWireguard.Status.Port = "51820"

		err = r.Status().Update(ctx, updateWireguard)

		if err != nil {
			log.Error(err, "Failed to update wireguard manifest host and port")
			return ctrl.Result{}, err
		}
	}

	// configmap

	configFound := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: wireguard.Name + "-config", Namespace: wireguard.Namespace}, configFound)
	if err != nil && errors.IsNotFound(err) {
		config := r.ConfigmapForWireguard(wireguard, hostname)
		log.Info("Creating a new config", "config.Namespace", config.Namespace, "config.Name", config.Name)
		err = r.Create(ctx, config)
		if err != nil {
			log.Error(err, "Failed to create new dep", "dep.Namespace", config.Namespace, "dep.Name", config.Name)
			return ctrl.Result{}, err
		}

		err = r.updateStatus(ctx, req, wireguard, vpnv1alpha1.WgStatusReport{Status: vpnv1alpha1.Pending, Message: "Waiting for configmap to be created"})

		if err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, err
	} else if err != nil {
		log.Error(err, "Failed to get config")
		return ctrl.Result{}, err
	}

	// deployment

	deploymentFound := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: wireguard.Name + "-dep", Namespace: wireguard.Namespace}, deploymentFound)
	if err != nil && errors.IsNotFound(err) {
		dep := r.deploymentForWireguard(wireguard)
		log.Info("Creating a new dep", "dep.Namespace", dep.Namespace, "dep.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new dep", "dep.Namespace", dep.Namespace, "dep.Name", dep.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{}, err
	} else if err != nil {
		log.Error(err, "Failed to get dep")
		return ctrl.Result{}, err
	}

	if err := r.updateWireguardPeers(ctx, req, wireguard, hostname, string(secret.Data["publicKey"]), wireguard.Spec.Mtu); err != nil {
		return ctrl.Result{}, err
	}

	log.Info("Updated related peers", "wireguard.Namespace", wireguard.Namespace, "wireguard.Name", wireguard.Name)

	err = r.updateStatus(ctx, req, wireguard, vpnv1alpha1.WgStatusReport{Status: vpnv1alpha1.Ready, Message: "VPN is active!"})

	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *WireguardReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&vpnv1alpha1.Wireguard{}).
		Owns(&vpnv1alpha1.WireguardPeer{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}

func (r *WireguardReconciler) serviceForWireguard(m *vpnv1alpha1.Wireguard) *corev1.Service {
	labels := labelsForWireguard(m.Name)
	//timeoutSeconds := int32(120)

	dep := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name + "-svc",
			Namespace: m.Namespace,
			Annotations: map[string]string{
				"service.beta.kubernetes.io/aws-load-balancer-type":            "external",
				"service.beta.kubernetes.io/aws-load-balancer-nlb-target-type": "ip",
				"service.beta.kubernetes.io/aws-load-balancer-scheme":          "internet-facing",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{{
				Protocol:   corev1.ProtocolUDP,
				Port:       port,
				TargetPort: intstr.FromInt(port),
			}},
			Type: corev1.ServiceTypeLoadBalancer,
		},
	}

	ctrl.SetControllerReference(m, dep, r.Scheme)
	return dep
}

func (r *WireguardReconciler) secretForWireguard(m *vpnv1alpha1.Wireguard, privateKey string, publicKey string, config string) *corev1.Secret {
	ls := labelsForWireguard(m.Name)
	dep := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
			Labels:    ls,
		},
		Data: map[string][]byte{"config": []byte(config), "privateKey": []byte(privateKey), "publicKey": []byte(publicKey)},
	}

	ctrl.SetControllerReference(m, dep, r.Scheme)

	return dep

}

func (r *WireguardReconciler) secretForClient(m *vpnv1alpha1.Wireguard, privateKey string, publicKey string) *corev1.Secret {
	ls := labelsForWireguard(m.Name)
	dep := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name + "-client",
			Namespace: m.Namespace,
			Labels:    ls,
		},
		Data: map[string][]byte{"privateKey": []byte(privateKey), "publicKey": []byte(publicKey)},
	}

	ctrl.SetControllerReference(m, dep, r.Scheme)

	return dep

}

func (r *WireguardReconciler) deploymentForWireguard(m *vpnv1alpha1.Wireguard) *appsv1.Deployment {
	ls := labelsForWireguard(m.Name)
	replicas := int32(1)
	trueVal := true

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name + "-dep",
			Namespace: m.Namespace,
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
					Volumes: []corev1.Volume{{

						Name: "config",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: m.Name,
							},
						},
					}},
					Containers: []corev1.Container{{
						SecurityContext: &corev1.SecurityContext{
							Capabilities: &corev1.Capabilities{Add: []corev1.Capability{"NET_ADMIN"}},
							Privileged:   &trueVal,
						},
						Image:           "ghcr.io/jodevsa/wireguard-operator:main",
						ImagePullPolicy: "Always",
						Name:            "wireguard",
						Ports: []corev1.ContainerPort{{
							ContainerPort: port,
							Name:          "wireguard",
							Protocol:      corev1.ProtocolUDP,
						}},
						EnvFrom: []corev1.EnvFromSource{{
							ConfigMapRef: &corev1.ConfigMapEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{Name: m.Name + "-config"},
							},
						}},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "config",
							MountPath: "/tmp/wireguard/",
						}},
						Env: []corev1.EnvVar{
							{Name: "CLIENT_PUBLIC_KEY", ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: m.Name + "-client"}, Key: "publicKey"},
							}},
							{Name: "PUBLIC_KEY", ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: m.Name}, Key: "publicKey"},
							}},
							{Name: "PRIVATE_KEY", ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: m.Name}, Key: "privateKey"},
							}},
						},
					}},
				},
			},
		},
	}

	ctrl.SetControllerReference(m, dep, r.Scheme)
	return dep
}

func (r *WireguardReconciler) pvcForWireguard(m *vpnv1alpha1.Wireguard) *corev1.PersistentVolumeClaim {
	dep := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Name: m.Name + "-pvc", Namespace: m.Namespace},
		Spec: corev1.PersistentVolumeClaimSpec{AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("1Gi")}}},
	}
	ctrl.SetControllerReference(m, dep, r.Scheme)
	return dep
}
