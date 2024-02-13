package e2e

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/argoproj-labs/argo-rollouts-manager/tests/e2e/fixture"
	"github.com/argoproj-labs/argo-rollouts-manager/tests/e2e/fixture/k8s"
	rmFixture "github.com/argoproj-labs/argo-rollouts-manager/tests/e2e/fixture/rolloutmanager"

	"sigs.k8s.io/controller-runtime/pkg/client"

	rmv1alpha1 "github.com/argoproj-labs/argo-rollouts-manager/api/v1alpha1"

	controllers "github.com/argoproj-labs/argo-rollouts-manager/controllers"
	rv1alpha1 "github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("Cluster Scoped RolloutManager tests", func() {

	Context("Testing cluster scoped RolloutManager behaviour", func() {

		var (
			err       error
			ctx       context.Context
			k8sClient client.Client
			scheme    *runtime.Scheme
		)

		BeforeEach(func() {
			Expect(fixture.EnsureCleanSlate()).To(Succeed())

			k8sClient, scheme, err = fixture.GetE2ETestKubeClient()
			Expect(err).To(Succeed())

			err = rv1alpha1.AddToScheme(scheme)
			Expect(err).To(Succeed())

			ctx = context.Background()
		})

		It("After creating cluster scoped RolloutManager in default namespace i.e argo-rollouts, operator should create appropriate K8s resources and watch argo rollouts CR in different namespace.", func() {

			nsName := "test-ro-ns"
			labels := map[string]string{"app": "test-argo-app"}

			// delete namespace created for test
			defer fixture.DeleteNamespace(ctx, nsName, k8sClient)

			By("Create cluster scoped RolloutManager in default namespace.")
			rolloutsManager, err := createRolloutManager(ctx, k8sClient, "test-rollouts-manager-1", fixture.TestE2ENamespace, false)
			Expect(err).To(Succeed())

			By("Verify that RolloutManager is successfully created.")
			Eventually(rolloutsManager, "1m", "1s").Should(rmFixture.HavePhase(rmv1alpha1.PhaseAvailable))

			By("Verify that Status.Condition is set.")
			Eventually(rolloutsManager, "1m", "1s").Should(rmFixture.HaveConditions(
				metav1.Condition{
					Type:    rmv1alpha1.RolloutManagerConditionTypeSuccess,
					Status:  metav1.ConditionFalse,
					Reason:  rmv1alpha1.RolloutManagerReasonSuccess,
					Message: "",
				}))

			By("Varify that expected resources are created")

			validateArgoRolloutManagerResources(ctx, rolloutsManager, k8sClient)

			By("Verify argo rollout controller able to reconcile CR of other namespace.")

			By("Create a different namespace")
			Expect(createNamespace(ctx, k8sClient, nsName)).To(Succeed())

			By("Create and validate rollouts.")
			validateArgoRolloutsResources(ctx, k8sClient, nsName, labels, 31000, 32000)
		})

		It("After creating cluster scoped RolloutManager in namespace other than argo-rollouts, operator should create appropriate K8s resources and watch argo rollouts CR in another namespace.", func() {

			nsName1 := "test-rom-ns"
			nsName2 := "test-ro-ns"
			labels := map[string]string{"app": "test-argo-app"}

			// delete namespace created for test
			defer fixture.DeleteNamespace(ctx, nsName1, k8sClient)
			defer fixture.DeleteNamespace(ctx, nsName2, k8sClient)

			By("Create a different namespace for rollout manager")
			Expect(createNamespace(ctx, k8sClient, nsName1)).To(Succeed())

			By("Create cluster scoped RolloutManager.")
			rolloutsManager, err := createRolloutManager(ctx, k8sClient, "test-rollouts-manager-1", nsName1, false)
			Expect(err).To(Succeed())

			By("Verify that RolloutManager is successfully created.")
			Eventually(rolloutsManager, "1m", "1s").Should(rmFixture.HavePhase(rmv1alpha1.PhaseAvailable))

			By("Verify that Status.Condition is set.")
			Eventually(rolloutsManager, "1m", "1s").Should(rmFixture.HaveConditions(
				metav1.Condition{
					Type:    rmv1alpha1.RolloutManagerConditionTypeSuccess,
					Status:  metav1.ConditionFalse,
					Reason:  rmv1alpha1.RolloutManagerReasonSuccess,
					Message: "",
				}))

			By("Varify that expected resources are created")

			validateArgoRolloutManagerResources(ctx, rolloutsManager, k8sClient)

			By("Verify argo rollout controller able to reconcile CR of other namespace.")

			By("Create a different namespace for rollout")
			Expect(createNamespace(ctx, k8sClient, nsName2)).To(Succeed())

			By("Create and validate rollouts.")
			validateArgoRolloutsResources(ctx, k8sClient, nsName2, labels, 31000, 32000)
		})

		It("After creating cluster scoped RolloutManager in a namespace, operator should create appropriate K8s resources and watch argo rollouts CR in multiple namespace.", func() {

			nsName1 := "rom-ns-1"
			nsName2 := "ro-ns-1"
			nsName3 := "ro-ns-2"
			labels := map[string]string{"app": "test-argo-app"}

			// delete namespace created for test
			defer fixture.DeleteNamespace(ctx, nsName1, k8sClient)
			defer fixture.DeleteNamespace(ctx, nsName2, k8sClient)
			defer fixture.DeleteNamespace(ctx, nsName3, k8sClient)

			By("Create a different namespace for rollout manager")
			Expect(createNamespace(ctx, k8sClient, nsName1)).To(Succeed())

			By("Create cluster scoped RolloutManager.")
			rolloutsManager, err := createRolloutManager(ctx, k8sClient, "test-rollouts-manager-1", nsName1, false)
			Expect(err).To(Succeed())

			By("Verify that RolloutManager is successfully created.")
			Eventually(rolloutsManager, "1m", "1s").Should(rmFixture.HavePhase(rmv1alpha1.PhaseAvailable))

			By("Verify that Status.Condition is set.")
			Eventually(rolloutsManager, "1m", "1s").Should(rmFixture.HaveConditions(
				metav1.Condition{
					Type:    rmv1alpha1.RolloutManagerConditionTypeSuccess,
					Status:  metav1.ConditionFalse,
					Reason:  rmv1alpha1.RolloutManagerReasonSuccess,
					Message: "",
				}))

			By("Varify that expected resources are created")

			validateArgoRolloutManagerResources(ctx, rolloutsManager, k8sClient)

			By("Verify argo rollout controller able to reconcile CR of other namespace.")

			By("1st rollout : Create a different namespace for rollout")
			Expect(createNamespace(ctx, k8sClient, nsName2)).To(Succeed())

			By("1st rollout : Create active and preview services in new namespace")
			validateArgoRolloutsResources(ctx, k8sClient, nsName2, labels, 31000, 32000)

			By("2nd rollout : Create a another namespace for 2nd rollout")
			Expect(createNamespace(ctx, k8sClient, nsName3)).To(Succeed())

			By("2nd rollout : Create and validate rollouts.")
			validateArgoRolloutsResources(ctx, k8sClient, nsName3, labels, 31001, 32002)
		})

		It("After creating cluster scoped RolloutManager in a namespace, another namespace scoped RolloutManager should not be allowed.", func() {

			nsName := "test-ro-ns"

			// delete namespace created for test
			defer fixture.DeleteNamespace(ctx, nsName, k8sClient)

			By("Create cluster scoped RolloutManager in default namespace.")
			rolloutsManagerCl, err := createRolloutManager(ctx, k8sClient, "test-rollouts-manager-1", fixture.TestE2ENamespace, false)
			Expect(err).To(Succeed())

			By("Verify that RolloutManager is successfully created.")
			Eventually(rolloutsManagerCl, "1m", "1s").Should(rmFixture.HavePhase(rmv1alpha1.PhaseAvailable))

			By("Verify that Status.Condition is set.")
			Eventually(rolloutsManagerCl, "1m", "1s").Should(rmFixture.HaveConditions(
				metav1.Condition{
					Type:    rmv1alpha1.RolloutManagerConditionTypeSuccess,
					Status:  metav1.ConditionFalse,
					Reason:  rmv1alpha1.RolloutManagerReasonSuccess,
					Message: "",
				}))

			By("Create a different namespace")
			Expect(createNamespace(ctx, k8sClient, nsName)).To(Succeed())

			By("Create namespace scoped RolloutManager in different namespace.")
			rolloutsManagerNs, err := createRolloutManager(ctx, k8sClient, "test-rollouts-manager-2", nsName, true)
			Expect(err).To(Succeed())

			By("Verify that RolloutManager is not working.")
			Eventually(rolloutsManagerNs, "1m", "1s").Should(rmFixture.HavePhase(rmv1alpha1.PhasePending))

			By("Verify that Status.Condition is having error message.")
			Eventually(rolloutsManagerNs, "1m", "1s").Should(rmFixture.HaveConditions(
				metav1.Condition{
					Type:    rmv1alpha1.RolloutManagerConditionTypeErrorOccurred,
					Status:  metav1.ConditionTrue,
					Reason:  rmv1alpha1.RolloutManagerReasonMultipleClusterScopedRolloutManager,
					Message: "With a cluster scoped RolloutManager, another RolloutManager is not supported",
				}))

			By("Update cluster scoped RolloutManager, after reconciliation it should also stop working.")

			err = k8sClient.Get(ctx, client.ObjectKeyFromObject(&rolloutsManagerCl), &rolloutsManagerCl)
			Expect(err).To(Succeed())
			rolloutsManagerCl.Spec.Env = append(rolloutsManagerCl.Spec.Env, corev1.EnvVar{Name: "test-name", Value: "test-value"})
			err = k8sClient.Update(ctx, &rolloutsManagerCl)
			Expect(err).To(Succeed())

			By("Verify that now cluster scoped RolloutManager is also not working.")
			Eventually(rolloutsManagerCl, "1m", "1s").Should(rmFixture.HavePhase(rmv1alpha1.PhasePending))

			By("Verify that Status.Condition is now having error message.")
			Eventually(rolloutsManagerCl, "3m", "1s").Should(rmFixture.HaveConditions(
				metav1.Condition{
					Type:    rmv1alpha1.RolloutManagerConditionTypeErrorOccurred,
					Status:  metav1.ConditionTrue,
					Reason:  rmv1alpha1.RolloutManagerReasonMultipleClusterScopedRolloutManager,
					Message: "With a cluster scoped RolloutManager, another RolloutManager is not supported",
				}))
		})

		It("After creating cluster scoped RolloutManager in a namespace, another cluster scoped RolloutManager should not be allowed.", func() {

			nsName := "test-ro-ns"

			// delete namespace created for test
			defer fixture.DeleteNamespace(ctx, nsName, k8sClient)

			By("Create cluster scoped RolloutManager in default namespace.")
			rolloutsManagerCl, err := createRolloutManager(ctx, k8sClient, "test-rollouts-manager-1", fixture.TestE2ENamespace, false)
			Expect(err).To(Succeed())

			By("Verify that RolloutManager is successfully created.")
			Eventually(rolloutsManagerCl, "1m", "1s").Should(rmFixture.HavePhase(rmv1alpha1.PhaseAvailable))

			By("Verify that Status.Condition is set.")
			Eventually(rolloutsManagerCl, "1m", "1s").Should(rmFixture.HaveConditions(
				metav1.Condition{
					Type:    rmv1alpha1.RolloutManagerConditionTypeSuccess,
					Status:  metav1.ConditionFalse,
					Reason:  rmv1alpha1.RolloutManagerReasonSuccess,
					Message: "",
				}))

			By("Create a different namespace")
			Expect(createNamespace(ctx, k8sClient, nsName)).To(Succeed())

			By("Create cluster scoped RolloutManager in different namespace.")
			rolloutsManagerNs, err := createRolloutManager(ctx, k8sClient, "test-rollouts-manager-2", nsName, false)
			Expect(err).To(Succeed())

			By("Verify that RolloutManager is not working.")
			Eventually(rolloutsManagerNs, "1m", "1s").Should(rmFixture.HavePhase(rmv1alpha1.PhasePending))

			By("Verify that Status.Condition is having error message.")
			Eventually(rolloutsManagerNs, "1m", "1s").Should(rmFixture.HaveConditions(
				metav1.Condition{
					Type:    rmv1alpha1.RolloutManagerConditionTypeErrorOccurred,
					Status:  metav1.ConditionTrue,
					Reason:  rmv1alpha1.RolloutManagerReasonMultipleClusterScopedRolloutManager,
					Message: "With a cluster scoped RolloutManager, another RolloutManager is not supported",
				}))

			By("Update first RolloutManager, after reconciliation it should also stop working.")

			err = k8sClient.Get(ctx, client.ObjectKeyFromObject(&rolloutsManagerCl), &rolloutsManagerCl)
			Expect(err).To(Succeed())
			rolloutsManagerCl.Spec.Env = append(rolloutsManagerCl.Spec.Env, corev1.EnvVar{Name: "test-name", Value: "test-value"})
			err = k8sClient.Update(ctx, &rolloutsManagerCl)
			Expect(err).To(Succeed())

			By("Verify that now first RolloutManager is also not working.")
			Eventually(rolloutsManagerCl, "1m", "1s").Should(rmFixture.HavePhase(rmv1alpha1.PhasePending))

			By("Verify that Status.Condition is now having error message.")
			Eventually(rolloutsManagerCl, "3m", "1s").Should(rmFixture.HaveConditions(
				metav1.Condition{
					Type:    rmv1alpha1.RolloutManagerConditionTypeErrorOccurred,
					Status:  metav1.ConditionTrue,
					Reason:  rmv1alpha1.RolloutManagerReasonMultipleClusterScopedRolloutManager,
					Message: "With a cluster scoped RolloutManager, another RolloutManager is not supported",
				}))
		})

		It("After creating namespace scoped RolloutManager, if a cluster scoped RolloutManager is created, both should not be allowed.", func() {

			nsName := "test-ro-ns"

			// delete namespace created for test
			defer fixture.DeleteNamespace(ctx, nsName, k8sClient)

			By("Create namespace scoped RolloutManager in default namespace.")
			rolloutsManagerCl, err := createRolloutManager(ctx, k8sClient, "test-rollouts-manager-1", fixture.TestE2ENamespace, true)
			Expect(err).To(Succeed())

			By("Verify that RolloutManager is successfully created.")
			Eventually(rolloutsManagerCl, "1m", "1s").Should(rmFixture.HavePhase(rmv1alpha1.PhaseAvailable))

			By("Verify that Status.Condition is set.")
			Eventually(rolloutsManagerCl, "1m", "1s").Should(rmFixture.HaveConditions(
				metav1.Condition{
					Type:    rmv1alpha1.RolloutManagerConditionTypeSuccess,
					Status:  metav1.ConditionFalse,
					Reason:  rmv1alpha1.RolloutManagerReasonSuccess,
					Message: "",
				}))

			By("Create a different namespace")
			Expect(createNamespace(ctx, k8sClient, nsName)).To(Succeed())

			By("Create cluster scoped RolloutManager in different namespace.")
			rolloutsManagerNs, err := createRolloutManager(ctx, k8sClient, "test-rollouts-manager-2", nsName, false)
			Expect(err).To(Succeed())

			By("Verify that RolloutManager is not working.")
			Eventually(rolloutsManagerNs, "1m", "1s").Should(rmFixture.HavePhase(rmv1alpha1.PhasePending))

			By("Verify that Status.Condition is having error message.")
			Eventually(rolloutsManagerNs, "1m", "1s").Should(rmFixture.HaveConditions(
				metav1.Condition{
					Type:    rmv1alpha1.RolloutManagerConditionTypeErrorOccurred,
					Status:  metav1.ConditionTrue,
					Reason:  rmv1alpha1.RolloutManagerReasonMultipleClusterScopedRolloutManager,
					Message: "With a cluster scoped RolloutManager, another RolloutManager is not supported",
				}))

			By("Update namespace scoped RolloutManager, after reconciliation it should also stop working.")

			err = k8sClient.Get(ctx, client.ObjectKeyFromObject(&rolloutsManagerCl), &rolloutsManagerCl)
			Expect(err).To(Succeed())
			rolloutsManagerCl.Spec.Env = append(rolloutsManagerCl.Spec.Env, corev1.EnvVar{Name: "test-name", Value: "test-value"})
			err = k8sClient.Update(ctx, &rolloutsManagerCl)
			Expect(err).To(Succeed())

			By("Verify that now namespace scoped RolloutManager is also not working.")
			Eventually(rolloutsManagerCl, "1m", "1s").Should(rmFixture.HavePhase(rmv1alpha1.PhasePending))

			By("Verify that Status.Condition is now having error message.")
			Eventually(rolloutsManagerCl, "3m", "1s").Should(rmFixture.HaveConditions(
				metav1.Condition{
					Type:    rmv1alpha1.RolloutManagerConditionTypeErrorOccurred,
					Status:  metav1.ConditionTrue,
					Reason:  rmv1alpha1.RolloutManagerReasonMultipleClusterScopedRolloutManager,
					Message: "With a cluster scoped RolloutManager, another RolloutManager is not supported",
				}))
		})
	})
})

func validateArgoRolloutManagerResources(ctx context.Context, rolloutsManager rmv1alpha1.RolloutManager, k8sClient client.Client) {

	By("Verify that ServiceAccount is created.")

	Eventually(&corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{Name: controllers.DefaultArgoRolloutsResourceName, Namespace: rolloutsManager.Namespace},
	}, "10s", "1s").Should(k8s.ExistByName(k8sClient))

	By("Verify that ClusterRoles are created.")

	clusterRoles := []string{"argo-rollouts", "argo-rollouts-aggregate-to-admin", "argo-rollouts-aggregate-to-edit", "argo-rollouts-aggregate-to-view"}
	for _, clusterRole := range clusterRoles {
		Eventually(&rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{Name: clusterRole, Namespace: rolloutsManager.Namespace},
		}, "30s", "1s").Should(k8s.ExistByName(k8sClient))
	}

	By("Verify that Service is created.")

	Eventually(&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: controllers.DefaultArgoRolloutsMetricsServiceName, Namespace: rolloutsManager.Namespace},
	}, "10s", "1s").Should(k8s.ExistByName(k8sClient))

	By("Verify that Secret is created.")

	Eventually(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: controllers.DefaultRolloutsNotificationSecretName, Namespace: rolloutsManager.Namespace},
	}, "30s", "1s").Should(k8s.ExistByName(k8sClient))

	By("Verify that argo rollouts deployment is created and it is in Ready state.")

	depl := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: controllers.DefaultArgoRolloutsResourceName, Namespace: rolloutsManager.Namespace},
	}
	Eventually(&depl, "10s", "1s").Should(k8s.ExistByName(k8sClient))
	Eventually(func() bool {
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(&depl), &depl); err != nil {
			return false
		}
		return depl.Status.ReadyReplicas == 1
	}, "3m", "1s").Should(BeTrue())
}

func createNamespace(ctx context.Context, k8sClient client.Client, name string) error {
	return k8sClient.Create(ctx,
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
			Name: name,
		}})
}

func createRolloutManager(ctx context.Context, k8sClient client.Client, name, namespace string, namespaceScoped bool) (rmv1alpha1.RolloutManager, error) {
	rolloutsManager := rmv1alpha1.RolloutManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: rmv1alpha1.RolloutManagerSpec{
			NamespaceScoped: namespaceScoped,
		},
	}
	return rolloutsManager, k8sClient.Create(ctx, &rolloutsManager)
}

func createService(ctx context.Context, k8sClient client.Client, name, namespace string, nodePort int32, selector map[string]string) (corev1.Service, error) {
	service := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeNodePort,
			Selector: selector,
			Ports: []corev1.ServicePort{
				{
					NodePort: nodePort,
					Protocol: corev1.ProtocolTCP,
					Port:     8080,
				},
			},
		},
	}
	return service, k8sClient.Create(ctx, &service)
}

func createArgoRollout(ctx context.Context, k8sClient client.Client, name, namespace, activeService, previewService string, labels map[string]string) (rv1alpha1.Rollout, error) {
	var num int32 = 2
	autoPromotionEnabled := false

	rollout := rv1alpha1.Rollout{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: rv1alpha1.RolloutSpec{
			Replicas: &num,
			Strategy: rv1alpha1.RolloutStrategy{
				BlueGreen: &rv1alpha1.BlueGreenStrategy{
					ActiveService:        activeService,
					PreviewService:       previewService,
					AutoPromotionEnabled: &autoPromotionEnabled,
				},
			},
			RevisionHistoryLimit: &num,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "webserver-simple",
							Image: "docker.io/kostiscodefresh/gitops-canary-app:v1.0",
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 8080,
									Protocol:      corev1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
		},
	}
	return rollout, k8sClient.Create(ctx, &rollout)
}

func validateArgoRolloutsResources(ctx context.Context, k8sClient client.Client, nsName string, labels map[string]string, port1, port2 int32) {

	By("Create active and preview services in new namespace")
	rolloutServiceActive, err := createService(ctx, k8sClient, "rollout-bluegreen-active", nsName, port1, labels)
	Expect(err).To(Succeed())
	Eventually(&rolloutServiceActive, "10s", "1s").Should(k8s.ExistByName(k8sClient))

	rolloutServicePreview, err := createService(ctx, k8sClient, "rollout-bluegreen-preview", nsName, port2, labels)
	Expect(err).To(Succeed())
	Eventually(&rolloutServicePreview, "10s", "1s").Should(k8s.ExistByName(k8sClient))

	By("Create Argo Rollout CR in new namespace and check it is reconciled successfully.")
	rollout, err := createArgoRollout(ctx, k8sClient, "simple-rollout", nsName, rolloutServiceActive.Name, rolloutServicePreview.Name, labels)
	Expect(err).To(Succeed())
	Eventually(func() bool {
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(&rollout), &rollout); err != nil {
			return false
		}
		return rollout.Status.Phase == rv1alpha1.RolloutPhaseHealthy
	}, "3m", "1s").Should(BeTrue())
}
