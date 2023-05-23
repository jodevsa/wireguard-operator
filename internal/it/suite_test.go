package it

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/go-logr/stdr"
	"github.com/jodevsa/wireguard-operator/pkg/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v12 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	kind "sigs.k8s.io/kind/pkg/cluster"
	log2 "sigs.k8s.io/kind/pkg/log"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var releasePath string
var agentImage string
var sidecarImage string
var managerImage string
var kindBinary string
var kubeConfigPath string

var testProvider = kind.NewProvider(
	kind.ProviderWithDocker())

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t,
		"Controller Suite")
}

const (
	Timeout             = time.Second * 120
	Interval            = time.Second * 1
	testClusterName     = "wg-kind-test"
	testKindContextName = "kind-" + testClusterName
	TestNamespace       = "default"
)

func waitForDeploymentTobeReady(name string, namespace string) {
	Eventually(func() int {
		deploymentKey := types.NamespacedName{
			Namespace: namespace,
			Name:      name,
		}

		deployment := &v12.Deployment{}
		k8sClient.Get(context.Background(), deploymentKey, deployment)
		return int(deployment.Status.ReadyReplicas)
	}, Timeout, Interval).Should(Equal(1))

}

func WaitForWireguardToBeReady(name string, namespace string) {
	Eventually(func() string {
		wgKey := types.NamespacedName{
			Namespace: namespace,
			Name:      name,
		}
		wg := &v1alpha1.Wireguard{}
		k8sClient.Get(context.Background(), wgKey, wg)
		return wg.Status.Status
	}, Timeout, Interval).Should(Equal(v1alpha1.Ready))

	waitForDeploymentTobeReady(name+"-dep", namespace)
}
func WaitForPeerToBeReady(name string, namespace string) {
	Eventually(func() string {
		wgKey := types.NamespacedName{
			Namespace: namespace,
			Name:      name,
		}
		wg := &v1alpha1.WireguardPeer{}
		k8sClient.Get(context.Background(), wgKey, wg)
		return wg.Status.Status
	}, Timeout, Interval).Should(Equal(v1alpha1.Ready))

}

func KubectlApply(resource string, namespace string) (string, error) {
	bashCommand := fmt.Sprintf("echo \"%s\" | kubectl apply -n %s --context %s -f -", strings.TrimSpace(strings.ReplaceAll(resource, "\"", "\\\"")), namespace, testKindContextName)

	cmd := exec.Command("bash", "-c", bashCommand)

	o, err := cmd.Output()

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return string(exitError.Stderr), err
		}
		return "", err
	}

	return strings.TrimSpace(string(o)), nil
}

var _ = BeforeSuite(func() {
	releasePath = os.Getenv("WIREGUARD_OPERATOR_RELEASE_PATH")
	agentImage = os.Getenv("AGENT_IMAGE")
	sidecarImage = os.Getenv("SIDECAR_IMAGE")
	managerImage = os.Getenv("MANAGER_IMAGE")
	kindBinary = os.Getenv("KIND_BIN")
	kubeConfigPath = os.Getenv("KUBE_CONFIG")

	Expect(releasePath).NotTo(Equal(""))
	Expect(agentImage).NotTo(Equal(""))
	Expect(sidecarImage).NotTo(Equal(""))
	Expect(releasePath).NotTo(Equal(""))
	Expect(managerImage).NotTo(Equal(""))
	Expect(kindBinary).NotTo(Equal(""))
	Expect(kubeConfigPath).NotTo(Equal(""))

	config := v1alpha4.Cluster{
		Nodes: []v1alpha4.Node{
			{
				Role: v1alpha4.ControlPlaneRole,
				ExtraPortMappings: []v1alpha4.PortMapping{
					{
						HostPort:      31820,
						ContainerPort: 31820,
						Protocol:      v1alpha4.PortMappingProtocolUDP,
					},
				},
			},
		},
	}

	log := stdr.NewWithOptions(log.New(os.Stderr, "", log.LstdFlags), stdr.Options{LogCaller: stdr.All})

	By("bootstrapping test environment")

	provider := kind.NewProvider(
		kind.ProviderWithLogger(log2.NoopLogger{}))

	err := provider.Create(testClusterName, kind.CreateWithV1Alpha4Config(&config))
	if err != nil {
		log.Error(err, "unable to create kind cluster")
		return
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeConfigPath},
		&clientcmd.ConfigOverrides{
			CurrentContext: testKindContextName,
		})

	c, err := clientConfig.ClientConfig()
	if err != nil {
		log.Error(err, "unable to create kind cluster")
		return
	}

	// load locally built images
	cmd := exec.Command(kindBinary, "load", "docker-image", managerImage, "--name", testClusterName)
	b, err := cmd.Output()
	if err != nil {
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				log.Info(string(exitError.Stderr))
				Expect(err).NotTo(HaveOccurred())
			}
		}

		log.Error(err, "unable to load local image manager:dev")
		Expect(err).NotTo(HaveOccurred())
	}
	cmd = exec.Command(kindBinary, "load", "docker-image", agentImage, "--name", testClusterName)
	b, err = cmd.Output()
	if err != nil {
		log.Error(err, "unable to load local image for agent")
		return
	}
	cmd = exec.Command(kindBinary, "load", "docker-image", sidecarImage, "--name", testClusterName)
	b, err = cmd.Output()
	if err != nil {
		log.Error(err, "unable to load local image for sidecar")
		return
	}

	// simulate what users exactly do in real life.
	cmd = exec.Command("kubectl", "apply", "-f", releasePath, "--context", testKindContextName)
	b, err = cmd.Output()

	if err != nil {
		log.Error(err, "unable to apply release.yaml")
		return
	}

	expectedResources := []string{
		"namespace/wireguard-system",
		"customresourcedefinition.apiextensions.k8s.io/wireguardpeers.vpn.example.com",
		"customresourcedefinition.apiextensions.k8s.io/wireguards.vpn.example.com",
		"serviceaccount/wireguard-wireguard-controller-manager",
		"role.rbac.authorization.k8s.io/wireguard-wireguard-leader-election-role",
		"clusterrole.rbac.authorization.k8s.io/wireguard-wireguard-manager-role",
		"clusterrole.rbac.authorization.k8s.io/wireguard-wireguard-metrics-reader",
		"clusterrole.rbac.authorization.k8s.io/wireguard-wireguard-proxy-role",
		"rolebinding.rbac.authorization.k8s.io/wireguard-wireguard-leader-election-rolebinding",
		"clusterrolebinding.rbac.authorization.k8s.io/wireguard-wireguard-manager-rolebinding",
		"clusterrolebinding.rbac.authorization.k8s.io/wireguard-wireguard-proxy-rolebinding",
		"configmap/wireguard-wireguard-manager-config",
		"service/wireguard-hello-kubernetes",
		"service/wireguard-wireguard-controller-manager-metrics-service",
		"deployment.apps/wireguard-hello-kubernetes",
		"deployment.apps/wireguard-wireguard-controller-manager",
	}

	Expect(strings.Split(strings.Trim(strings.ReplaceAll(string(b), " created", ""), "\n"), "\n")).To(BeEquivalentTo(expectedResources))

	err = v1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	k8sClient, err = client.New(c, client.Options{Scheme: scheme.Scheme})

	// wait until operator is ready
	Eventually(func() int {
		deploymentKey := types.NamespacedName{
			Namespace: "wireguard-system",
			Name:      "wireguard-controller-manager",
		}

		deployment := &v12.Deployment{}
		k8sClient.Get(context.Background(), deploymentKey, deployment)
		return int(deployment.Status.ReadyReplicas)
	}, Timeout, Interval).Should(Equal(1))

	go func() {
		defer GinkgoRecover()
	}()

}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testProvider.Delete(testClusterName, kubeConfigPath)
	Expect(err).NotTo(HaveOccurred())
})
