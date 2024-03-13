package e2e

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	utils "github.com/argoproj-labs/argo-rollouts-manager/tests/e2e"
	"github.com/argoproj-labs/argo-rollouts-manager/tests/e2e/fixture"
	rmFixture "github.com/argoproj-labs/argo-rollouts-manager/tests/e2e/fixture/rolloutmanager"

	"sigs.k8s.io/controller-runtime/pkg/client"

	rmv1alpha1 "github.com/argoproj-labs/argo-rollouts-manager/api/v1alpha1"

	controllers "github.com/argoproj-labs/argo-rollouts-manager/controllers"
	rv1alpha1 "github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	corev1 "k8s.io/api/core/v1"
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
			Expect(err).ToNot(HaveOccurred())

			err = rv1alpha1.AddToScheme(scheme)
			Expect(err).ToNot(HaveOccurred())

			ctx = context.Background()
		})

		/*
			In this test a cluster scoped RolloutManager is created in argo-rollouts namespace.
			After creation of RM operator should create required resources (ServiceAccount, ClusterRoles, ClusterRoleBinding, Service, Secret, Deployment) in argo-rollouts namespace.
			Now when a Rollouts CR is created in a different namespace, operator should still be able to reconcile.
		*/
		It("After creating cluster scoped RolloutManager in default namespace i.e argo-rollouts, operator should create appropriate K8s resources and watch argo rollouts CR in different namespace.", func() {

			nsName := "test-ro-ns"
			labels := map[string]string{"app": "test-argo-app"}

			By("Create cluster scoped RolloutManager in default namespace.")
			rolloutsManager, err := utils.CreateRolloutManager(ctx, k8sClient, "test-rollouts-manager-1", fixture.TestE2ENamespace, false)
			Expect(err).ToNot(HaveOccurred())

			By("Verify that RolloutManager is successfully created.")
			Eventually(rolloutsManager, "1m", "1s").Should(rmFixture.HavePhase(rmv1alpha1.PhaseAvailable))

			By("Verify that Status.Condition is set.")
			Eventually(rolloutsManager, "1m", "1s").Should(rmFixture.HaveCondition(
				metav1.Condition{
					Type:    rmv1alpha1.RolloutManagerConditionType,
					Status:  metav1.ConditionTrue,
					Reason:  rmv1alpha1.RolloutManagerReasonSuccess,
					Message: "",
				}))

			By("Verify that expected resources are created.")
			utils.ValidateArgoRolloutManagerResources(ctx, rolloutsManager, k8sClient, false)

			By("Verify argo rollout controller able to reconcile CR of other namespace.")

			By("Create a different namespace.")
			Expect(utils.CreateNamespace(ctx, k8sClient, nsName)).To(Succeed())

			By("Create and validate rollouts.")
			utils.ValidateArgoRolloutsResources(ctx, k8sClient, nsName, labels, 31000, 32000)
		})

		/*
			In this test a cluster scoped RolloutManager is created in namespace other than argo-rollouts.
			After creation of RM operator should create required resources (ServiceAccount, ClusterRoles, ClusterRoleBinding, Service, Secret, Deployment) in other namespace.
			Now when a Rollouts CR is created in a another namespace, operator should still be able to reconcile.
		*/
		It("After creating cluster scoped RolloutManager in namespace other than argo-rollouts, operator should create appropriate K8s resources and watch argo rollouts CR in another namespace.", func() {

			nsName1, nsName2 := "test-rom-ns", "test-ro-ns"
			labels := map[string]string{"app": "test-argo-app"}

			By("Create a different namespace for rollout manager.")
			Expect(utils.CreateNamespace(ctx, k8sClient, nsName1)).To(Succeed())

			By("Create cluster scoped RolloutManager.")
			rolloutsManager, err := utils.CreateRolloutManager(ctx, k8sClient, "test-rollouts-manager-1", nsName1, false)
			Expect(err).ToNot(HaveOccurred())

			By("Verify that RolloutManager is successfully created.")
			Eventually(rolloutsManager, "3m", "1s").Should(rmFixture.HavePhase(rmv1alpha1.PhaseAvailable))

			By("Verify that Status.Condition is set.")
			Eventually(rolloutsManager, "1m", "1s").Should(rmFixture.HaveCondition(
				metav1.Condition{
					Type:    rmv1alpha1.RolloutManagerConditionType,
					Status:  metav1.ConditionTrue,
					Reason:  rmv1alpha1.RolloutManagerReasonSuccess,
					Message: "",
				}))

			By("Verify that expected resources are created.")
			utils.ValidateArgoRolloutManagerResources(ctx, rolloutsManager, k8sClient, false)

			By("Verify argo rollout controller able to reconcile CR of other namespace.")

			By("Create a different namespace for rollout.")
			Expect(utils.CreateNamespace(ctx, k8sClient, nsName2)).To(Succeed())

			By("Create and validate rollouts.")
			utils.ValidateArgoRolloutsResources(ctx, k8sClient, nsName2, labels, 31000, 32000)
		})

		/*
			In this test a cluster scoped RolloutManager is created in a namespace.
			After creation of RM operator should create required resources (ServiceAccount, ClusterRoles, ClusterRoleBinding, Service, Secret, Deployment) in namespace.
			Now when a Rollouts CR is created in multiple namespaces, operator should still be able to reconcile all of them.
		*/
		It("After creating cluster scoped RolloutManager in a namespace, operator should create appropriate K8s resources and watch argo rollouts CR in multiple namespace.", func() {

			nsName1, nsName2, nsName3 := "rom-ns-1", "ro-ns-1", "ro-ns-2"
			labels := map[string]string{"app": "test-argo-app"}

			By("Create a namespace for rollout manager.")
			Expect(utils.CreateNamespace(ctx, k8sClient, nsName1)).To(Succeed())

			By("Create cluster scoped RolloutManager.")
			rolloutsManager, err := utils.CreateRolloutManager(ctx, k8sClient, "test-rollouts-manager-1", nsName1, false)
			Expect(err).ToNot(HaveOccurred())

			By("Verify that RolloutManager is successfully created.")
			Eventually(rolloutsManager, "1m", "1s").Should(rmFixture.HavePhase(rmv1alpha1.PhaseAvailable))

			By("Verify that Status.Condition is set.")
			Eventually(rolloutsManager, "1m", "1s").Should(rmFixture.HaveCondition(
				metav1.Condition{
					Type:    rmv1alpha1.RolloutManagerConditionType,
					Status:  metav1.ConditionTrue,
					Reason:  rmv1alpha1.RolloutManagerReasonSuccess,
					Message: "",
				}))

			By("Verify that expected resources are created.")
			utils.ValidateArgoRolloutManagerResources(ctx, rolloutsManager, k8sClient, false)

			By("Verify argo rollout controller able to reconcile CR of multiple namespaces.")

			By("1st rollout: Create a different namespace for rollout.")
			Expect(utils.CreateNamespace(ctx, k8sClient, nsName2)).To(Succeed())

			By("1st rollout: Create active and preview services in 1st namespace.")
			utils.ValidateArgoRolloutsResources(ctx, k8sClient, nsName2, labels, 31000, 32000)

			By("2nd rollout: Create a another namespace for 2nd rollout.")
			Expect(utils.CreateNamespace(ctx, k8sClient, nsName3)).To(Succeed())

			By("2nd rollout: Create active and preview services in 2nd namespace.")
			utils.ValidateArgoRolloutsResources(ctx, k8sClient, nsName3, labels, 31001, 32002)
		})

		/*
			In this test a cluster scoped RolloutManager is created in a namespace.
			After creation of RM operator should create required resources (ServiceAccount, ClusterRoles, ClusterRoleBinding, Service, Secret, Deployment) in namespace.
			Now when a namespace scoped RolloutManager is created, it should not be accepted by operator, since there in an existing RolloutManager watching entire cluster.
			When 1st cluster scoped RolloutManager is reconciled again it should also have error, because only one cluster scoped or all namespace scoped RolloutManagers are supported.
		*/
		It("Should allow cluster scoped RolloutManager, but not namespace scoped.", func() {

			nsName := "test-ro-ns"

			By("Create cluster scoped RolloutManager in a namespace.")
			rolloutsManagerCl, err := utils.CreateRolloutManager(ctx, k8sClient, "test-rollouts-manager-1", fixture.TestE2ENamespace, false)
			Expect(err).ToNot(HaveOccurred())

			By("Verify that RolloutManager is successfully created.")
			Eventually(rolloutsManagerCl, "1m", "1s").Should(rmFixture.HavePhase(rmv1alpha1.PhaseAvailable))

			By("Verify that Status.Condition is set.")
			Eventually(rolloutsManagerCl, "1m", "1s").Should(rmFixture.HaveCondition(
				metav1.Condition{
					Type:    rmv1alpha1.RolloutManagerConditionType,
					Status:  metav1.ConditionTrue,
					Reason:  rmv1alpha1.RolloutManagerReasonSuccess,
					Message: "",
				}))

			By("Create a different namespace.")
			Expect(utils.CreateNamespace(ctx, k8sClient, nsName)).To(Succeed())

			By("Create namespace scoped RolloutManager in different namespace.")
			rolloutsManagerNs, err := utils.CreateRolloutManager(ctx, k8sClient, "test-rollouts-manager-2", nsName, true)
			Expect(err).ToNot(HaveOccurred())

			By("Verify that RolloutManager is not working.")
			Eventually(rolloutsManagerNs, "1m", "1s").Should(rmFixture.HavePhase(rmv1alpha1.PhaseFailure))

			By("Verify that Status.Condition is having error message.")
			Eventually(rolloutsManagerNs, "1m", "1s").Should(rmFixture.HaveCondition(
				metav1.Condition{
					Type:    rmv1alpha1.RolloutManagerConditionType,
					Status:  metav1.ConditionFalse,
					Reason:  rmv1alpha1.RolloutManagerReasonInvalidScoped,
					Message: controllers.UnsupportedRolloutManagerNamespaceScoped,
				}))

			By("Update cluster scoped RolloutManager, it should still work.")

			Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(&rolloutsManagerCl), &rolloutsManagerCl)).To(Succeed())
			rolloutsManagerCl.Spec.Env = append(rolloutsManagerCl.Spec.Env, corev1.EnvVar{Name: "test-name", Value: "test-value"})
			Expect(k8sClient.Update(ctx, &rolloutsManagerCl)).To(Succeed())

			By("Verify that cluster scoped RolloutManager is still working.")
			Eventually(rolloutsManagerCl, "1m", "1s").Should(rmFixture.HavePhase(rmv1alpha1.PhaseAvailable))

			By("Verify that Status.Condition is not having error message.")
			Eventually(rolloutsManagerCl, "3m", "1s").Should(rmFixture.HaveCondition(
				metav1.Condition{
					Type:    rmv1alpha1.RolloutManagerConditionType,
					Status:  metav1.ConditionTrue,
					Reason:  rmv1alpha1.RolloutManagerReasonSuccess,
					Message: "",
				}))
		})

		/*
			In this test a cluster scoped RolloutManager is created in a namespace.
			After creation of RM operator should create required resources (ServiceAccount, ClusterRoles, ClusterRoleBinding, Service, Secret, Deployment) in namespace.
			Now when another cluster scoped RolloutManager is created, it should not be accepted by operator, since there in an existing RolloutManager watching entire cluster.
			When cluster scoped RolloutManager is reconciled again it should also have error, because only one cluster scoped or all namespace scoped RolloutManagers are supported.
		*/
		It("After creating cluster scoped RolloutManager in a namespace, another cluster scoped RolloutManager should not be allowed.", func() {

			nsName := "test-ro-ns"

			By("Create cluster scoped RolloutManager in a namespace.")
			rolloutsManagerCl, err := utils.CreateRolloutManager(ctx, k8sClient, "test-rollouts-manager-1", fixture.TestE2ENamespace, false)
			Expect(err).ToNot(HaveOccurred())

			By("Verify that RolloutManager is successfully created.")
			Eventually(rolloutsManagerCl, "1m", "1s").Should(rmFixture.HavePhase(rmv1alpha1.PhaseAvailable))

			By("Verify that Status.Condition is set.")
			Eventually(rolloutsManagerCl, "1m", "1s").Should(rmFixture.HaveCondition(
				metav1.Condition{
					Type:    rmv1alpha1.RolloutManagerConditionType,
					Status:  metav1.ConditionTrue,
					Reason:  rmv1alpha1.RolloutManagerReasonSuccess,
					Message: "",
				}))

			By("Create a different namespace.")
			Expect(utils.CreateNamespace(ctx, k8sClient, nsName)).To(Succeed())

			By("Create cluster scoped RolloutManager in different namespace.")
			rolloutsManagerNs, err := utils.CreateRolloutManager(ctx, k8sClient, "test-rollouts-manager-2", nsName, false)
			Expect(err).ToNot(HaveOccurred())

			By("Verify that RolloutManager is not working.")
			Eventually(rolloutsManagerNs, "1m", "1s").Should(rmFixture.HavePhase(rmv1alpha1.PhaseFailure))

			By("Verify that Status.Condition is having error message.")
			Eventually(rolloutsManagerNs, "1m", "1s").Should(rmFixture.HaveCondition(
				metav1.Condition{
					Type:    rmv1alpha1.RolloutManagerConditionType,
					Status:  metav1.ConditionFalse,
					Reason:  rmv1alpha1.RolloutManagerReasonMultipleClusterScopedRolloutManager,
					Message: controllers.UnsupportedRolloutManagerConfiguration,
				}))

			By("Update first RolloutManager, after reconciliation it should also stop working.")

			err = k8sClient.Get(ctx, client.ObjectKeyFromObject(&rolloutsManagerCl), &rolloutsManagerCl)
			Expect(err).ToNot(HaveOccurred())
			rolloutsManagerCl.Spec.Env = append(rolloutsManagerCl.Spec.Env, corev1.EnvVar{Name: "test-name", Value: "test-value"})
			err = k8sClient.Update(ctx, &rolloutsManagerCl)
			Expect(err).ToNot(HaveOccurred())

			By("Verify that now first RolloutManager is also not working.")
			Eventually(rolloutsManagerCl, "1m", "1s").Should(rmFixture.HavePhase(rmv1alpha1.PhaseFailure))

			By("Verify that Status.Condition is now having error message.")
			Eventually(rolloutsManagerCl, "3m", "1s").Should(rmFixture.HaveCondition(
				metav1.Condition{
					Type:    rmv1alpha1.RolloutManagerConditionType,
					Status:  metav1.ConditionFalse,
					Reason:  rmv1alpha1.RolloutManagerReasonMultipleClusterScopedRolloutManager,
					Message: controllers.UnsupportedRolloutManagerConfiguration,
				}))
		})

	})
})
