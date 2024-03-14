package e2e

import (
	"flag"
	"os"

	"path/filepath"
	"strings"
	"testing"

	rolloutsmanagerv1alpha1 "github.com/argoproj-labs/argo-rollouts-manager/api/v1alpha1"
	controllers "github.com/argoproj-labs/argo-rollouts-manager/controllers"
	"github.com/argoproj-labs/argo-rollouts-manager/tests/e2e/fixture"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap/zapcore"
	"k8s.io/client-go/kubernetes/scheme"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true), zap.Level(zapcore.DebugLevel)))

	By("bootstrapping test environment")
	useActualCluster := true
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("../../..", "config", "crd", "bases"),
		},
		UseExistingCluster:    &useActualCluster, // use an actual OpenShift cluster specified in kubeconfig
		ErrorIfCRDPathMissing: true,
	}

	// Set the environment variable for namespace scope of Rollouts
	Expect(os.Setenv(controllers.NamespaceScopedArgoRolloutsController, "true")).To(Succeed())

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	Expect(clientgoscheme.AddToScheme(scheme.Scheme)).NotTo(HaveOccurred())
	Expect(rolloutsmanagerv1alpha1.AddToScheme(scheme.Scheme)).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	err = fixture.EnsureCleanSlate()
	Expect(err).NotTo(HaveOccurred())

	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	opts := zap.Options{}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:                 scheme.Scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "rolloutsmanager.argoproj.io",
	})
	Expect(err).NotTo(HaveOccurred())

	err = (&controllers.RolloutManagerReconciler{
		Client:                                mgr.GetClient(),
		Scheme:                                mgr.GetScheme(),
		OpenShiftRoutePluginLocation:          "https://github.com/argoproj-labs/rollouts-plugin-trafficrouter-openshift/releases/download/commit-2749e0ac96ba00ce6f4af19dc6d5358048227d77/rollouts-plugin-trafficrouter-openshift-linux-amd64",
		NamespaceScopedArgoRolloutsController: strings.ToLower(os.Getenv(controllers.NamespaceScopedArgoRolloutsController)) == "true",
	}).SetupWithManager(mgr)
	Expect(err).NotTo(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = mgr.Start(ctrl.SetupSignalHandler())
		Expect(err).NotTo(HaveOccurred())
	}()
})

var _ = AfterSuite(func() {

	By("delete environment variable")
	Expect(os.Unsetenv(controllers.NamespaceScopedArgoRolloutsController)).To(Succeed())

	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

func TestNamespaScoped(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "NamespaScoped Suite")
}
