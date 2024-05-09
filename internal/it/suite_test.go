package it

import (
	"context"
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
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	kind "sigs.k8s.io/kind/pkg/cluster"
	log2 "sigs.k8s.io/kind/pkg/log"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var k8sClient client.Client
var releasePath string
var agentImage string
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
	cmd := exec.Command("kubectl", "apply",
		"--context", testKindContextName,
		"-n", namespace,
		"-f", "-",
	)
	cmd.Stdin = strings.NewReader(resource)

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return stderr.String(), err
	}
	return strings.TrimSpace(stdout.String()), nil
}

var _ = BeforeSuite(func() {
	releasePath = os.Getenv("WIREGUARD_OPERATOR_RELEASE_PATH")
	agentImage = os.Getenv("AGENT_IMAGE")
	managerImage = os.Getenv("MANAGER_IMAGE")
	kindBinary = os.Getenv("KIND_BIN")
	kubeConfigPath = os.Getenv("KUBE_CONFIG")

	Expect(releasePath).NotTo(Equal(""))
	Expect(agentImage).NotTo(Equal(""))
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
	if _, err := exec.
		Command(kindBinary, "load", "docker-image", managerImage, "--name", testClusterName).
		Output(); err != nil {
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				log.Info(string(exitError.Stderr))
				Expect(err).NotTo(HaveOccurred())
			}
		}

		log.Error(err, "unable to load local image manager:dev")
		Expect(err).NotTo(HaveOccurred())
	}

	if _, err := exec.
		Command(kindBinary, "load", "docker-image", agentImage, "--name", testClusterName).
		Output(); err != nil {
		log.Error(err, "unable to load local image agent:dev")
		return
	}

	// simulate what users exactly do in real life.
	b, err := exec.
		Command("kubectl", "apply", "-f", releasePath, "--context", testKindContextName).
		Output()
	if err != nil {
		log.Error(err, "unable to apply release.yaml")
		return
	}

	expectedResources := []string{
		"namespace/wireguard-system",
		"customresourcedefinition.apiextensions.k8s.io/wireguardpeers.vpn.wireguard-operator.io",
		"customresourcedefinition.apiextensions.k8s.io/wireguards.vpn.wireguard-operator.io",
		"serviceaccount/wireguard-controller-manager",
		"role.rbac.authorization.k8s.io/wireguard-leader-election-role",
		"clusterrole.rbac.authorization.k8s.io/wireguard-manager-role",
		"clusterrole.rbac.authorization.k8s.io/wireguard-metrics-reader",
		"clusterrole.rbac.authorization.k8s.io/wireguard-proxy-role",
		"rolebinding.rbac.authorization.k8s.io/wireguard-leader-election-rolebinding",
		"clusterrolebinding.rbac.authorization.k8s.io/wireguard-manager-rolebinding",
		"clusterrolebinding.rbac.authorization.k8s.io/wireguard-proxy-rolebinding",
		"configmap/wireguard-manager-config",
		"service/wireguard-controller-manager-metrics-service",
		"deployment.apps/wireguard-controller-manager",
	}

	Expect(strings.Split(strings.Trim(strings.ReplaceAll(string(b), " created", ""), "\n"), "\n")).To(BeEquivalentTo(expectedResources))

	err = v1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	k8sClient, err = client.New(c, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())

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
