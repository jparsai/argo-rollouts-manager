package rollouts

import (
	"context"
	"fmt"

	"github.com/argoproj-labs/argo-rollouts-manager/api/v1alpha1"
	rolloutsmanagerv1alpha1 "github.com/argoproj-labs/argo-rollouts-manager/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("RolloutManagerReconciler tests", func() {
	var (
		ctx context.Context
		rm  *v1alpha1.RolloutManager
	)

	BeforeEach(func() {
		ctx = context.Background()
		rm = makeTestRolloutManager()
	})

	It("Should create expected resource when namespace scoped RolloutManager CR is reconcilered.", func() {
		// Make RolloutManager namespace scoped
		rm.Spec.NamespaceScoped = true

		r := makeTestReconciler(rm)
		Expect(createNamespace(r, rm.Namespace)).ToNot(HaveOccurred())

		req := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      rm.Name,
				Namespace: rm.Namespace,
			},
		}

		res, err := r.Reconcile(ctx, req)
		Expect(err).ToNot(HaveOccurred())
		Expect(res.Requeue).Should(BeFalse(), "reconcile should not requeue request")

		By("Check if RolloutManager's Status.Conditions are set.")
		Expect(r.Client.Get(ctx, types.NamespacedName{Name: rm.Name, Namespace: rm.Namespace}, rm)).ToNot(HaveOccurred())
		Expect(rm.Status.Conditions[0].Type == rolloutsmanagerv1alpha1.RolloutManagerConditionTypeSuccess &&
			rm.Status.Conditions[0].Reason == rolloutsmanagerv1alpha1.RolloutManagerReasonSuccess &&
			rm.Status.Conditions[0].Message == "" &&
			rm.Status.Conditions[0].Status == metav1.ConditionFalse).To(BeTrue())

		By("Check expected resources are created.")

		sa := &corev1.ServiceAccount{}
		Expect(r.Client.Get(ctx, types.NamespacedName{
			Name:      DefaultArgoRolloutsResourceName,
			Namespace: testNamespace,
		}, sa)).To(Succeed(), fmt.Sprintf("failed to find the rollouts serviceaccount: %#v\n", err))

		role := &rbacv1.Role{}
		Expect(r.Client.Get(ctx, types.NamespacedName{
			Name:      DefaultArgoRolloutsResourceName,
			Namespace: testNamespace,
		}, role)).To(Succeed(), fmt.Sprintf("failed to find the rollouts role: %#v\n", err))

		roleBinding := &rbacv1.RoleBinding{}
		Expect(r.Client.Get(ctx, types.NamespacedName{
			Name:      DefaultArgoRolloutsResourceName,
			Namespace: testNamespace,
		}, roleBinding)).To(Succeed(), "failed to find the rollouts rolebinding")

		aggregateToAdminClusterRole := &rbacv1.ClusterRole{}
		Expect(r.Client.Get(ctx, types.NamespacedName{
			Name: "argo-rollouts-aggregate-to-admin",
		}, aggregateToAdminClusterRole)).To(Succeed(), fmt.Sprintf("failed to find the aggregateToAdmin ClusterRole: %#v\n", err))

		aggregateToEditClusterRole := &rbacv1.ClusterRole{}
		Expect(r.Client.Get(ctx, types.NamespacedName{
			Name: "argo-rollouts-aggregate-to-edit",
		}, aggregateToEditClusterRole)).To(Succeed(), fmt.Sprintf("failed to find the aggregateToEdit ClusterRole: %#v\n", err))

		aggregateToViewClusterRole := &rbacv1.ClusterRole{}
		Expect(r.Client.Get(ctx, types.NamespacedName{
			Name: "argo-rollouts-aggregate-to-view",
		}, aggregateToViewClusterRole)).To(Succeed(), fmt.Sprintf("failed to find the aggregateToView ClusterRole: %#v\n", err))

		service := &corev1.Service{}
		Expect(r.Client.Get(ctx, types.NamespacedName{
			Name:      DefaultArgoRolloutsMetricsServiceName,
			Namespace: rm.Namespace,
		}, service)).To(Succeed(), fmt.Sprintf("failed to find the rollouts metrics service: %#v\n", err))

		secret := &corev1.Secret{}
		Expect(r.Client.Get(ctx, types.NamespacedName{
			Name:      DefaultRolloutsNotificationSecretName,
			Namespace: rm.Namespace,
		}, secret)).To(Succeed(), fmt.Sprintf("failed to find the rollouts secret: %#v\n", err))
	})

	It("Should create expected resource when cluister scoped RolloutManager CR is reconcilered.", func() {
		r := makeTestReconciler(rm)
		Expect(createNamespace(r, rm.Namespace)).ToNot(HaveOccurred())

		req := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      rm.Name,
				Namespace: rm.Namespace,
			},
		}

		res, err := r.Reconcile(ctx, req)
		Expect(err).ToNot(HaveOccurred())
		Expect(res.Requeue).Should(BeFalse(), "reconcile should not requeue request")

		By("Check if RolloutManager's Status.Conditions are set.")
		Expect(r.Client.Get(ctx, types.NamespacedName{Name: rm.Name, Namespace: rm.Namespace}, rm)).ToNot(HaveOccurred())
		Expect(rm.Status.Conditions[0].Type == rolloutsmanagerv1alpha1.RolloutManagerConditionTypeSuccess &&
			rm.Status.Conditions[0].Reason == rolloutsmanagerv1alpha1.RolloutManagerReasonSuccess &&
			rm.Status.Conditions[0].Message == "" &&
			rm.Status.Conditions[0].Status == metav1.ConditionFalse).To(BeTrue())

		By("Check expected resources are created.")

		sa := &corev1.ServiceAccount{}
		Expect(r.Client.Get(ctx, types.NamespacedName{
			Name:      DefaultArgoRolloutsResourceName,
			Namespace: testNamespace,
		}, sa)).To(Succeed(), fmt.Sprintf("failed to find the rollouts serviceaccount: %#v\n", err))

		ClusterRole := &rbacv1.ClusterRole{}
		Expect(r.Client.Get(ctx, types.NamespacedName{
			Name:      DefaultArgoRolloutsResourceName,
			Namespace: testNamespace,
		}, ClusterRole)).To(Succeed(), fmt.Sprintf("failed to find the rollouts clusterRole: %#v\n", err))

		clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
		Expect(r.Client.Get(ctx, types.NamespacedName{
			Name:      DefaultArgoRolloutsResourceName,
			Namespace: testNamespace,
		}, clusterRoleBinding)).To(Succeed(), "failed to find the rollouts clusterRolebinding")

		aggregateToAdminClusterRole := &rbacv1.ClusterRole{}
		Expect(r.Client.Get(ctx, types.NamespacedName{
			Name: "argo-rollouts-aggregate-to-admin",
		}, aggregateToAdminClusterRole)).To(Succeed(), fmt.Sprintf("failed to find the aggregateToAdmin ClusterRole: %#v\n", err))

		aggregateToEditClusterRole := &rbacv1.ClusterRole{}
		Expect(r.Client.Get(ctx, types.NamespacedName{
			Name: "argo-rollouts-aggregate-to-edit",
		}, aggregateToEditClusterRole)).To(Succeed(), fmt.Sprintf("failed to find the aggregateToEdit ClusterRole: %#v\n", err))

		aggregateToViewClusterRole := &rbacv1.ClusterRole{}
		Expect(r.Client.Get(ctx, types.NamespacedName{
			Name: "argo-rollouts-aggregate-to-view",
		}, aggregateToViewClusterRole)).To(Succeed(), fmt.Sprintf("failed to find the aggregateToView ClusterRole: %#v\n", err))

		service := &corev1.Service{}
		Expect(r.Client.Get(ctx, types.NamespacedName{
			Name:      DefaultArgoRolloutsMetricsServiceName,
			Namespace: rm.Namespace,
		}, service)).To(Succeed(), fmt.Sprintf("failed to find the rollouts metrics service: %#v\n", err))

		secret := &corev1.Secret{}
		Expect(r.Client.Get(ctx, types.NamespacedName{
			Name:      DefaultRolloutsNotificationSecretName,
			Namespace: rm.Namespace,
		}, secret)).To(Succeed(), fmt.Sprintf("failed to find the rollouts secret: %#v\n", err))
	})

	It("Should not allow cluister and namespace scoped RolloutManager CRs together.", func() {
		r := makeTestReconciler(rm)
		Expect(createNamespace(r, rm.Namespace)).ToNot(HaveOccurred())

		req := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      rm.Name,
				Namespace: rm.Namespace,
			},
		}

		res, err := r.Reconcile(ctx, req)
		Expect(err).ToNot(HaveOccurred())
		Expect(res.Requeue).Should(BeFalse(), "reconcile should not requeue request")

		By("Check if RolloutManager's Status.Conditions are set.")
		Expect(r.Client.Get(ctx, types.NamespacedName{Name: rm.Name, Namespace: rm.Namespace}, rm)).ToNot(HaveOccurred())
		Expect(rm.Status.Conditions[0].Type == rolloutsmanagerv1alpha1.RolloutManagerConditionTypeSuccess &&
			rm.Status.Conditions[0].Reason == rolloutsmanagerv1alpha1.RolloutManagerReasonSuccess &&
			rm.Status.Conditions[0].Message == "" &&
			rm.Status.Conditions[0].Status == metav1.ConditionFalse).To(BeTrue())

		rm2 := makeTestRolloutManager()
		rm2.Name = "test-rm"
		rm2.Namespace = "test-ns"

		r2 := makeTestReconciler(rm)

		Expect(createNamespace(r2, rm2.Namespace)).ToNot(HaveOccurred())
		Expect(r.Client.Create(ctx, rm2)).ToNot(HaveOccurred())

		req2 := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      rm2.Name,
				Namespace: rm2.Namespace,
			},
		}

		res2, err := r.Reconcile(ctx, req2)
		Expect(err).To(HaveOccurred())
		Expect(res2.Requeue).Should(BeFalse(), "reconcile should not requeue request")

		Expect(doMultipleRolloutManagersExist(err)).To(BeTrue())

		By("Check if RolloutManager's Status.Conditions are set.")
		Expect(r.Client.Get(ctx, types.NamespacedName{Name: rm2.Name, Namespace: rm2.Namespace}, rm2)).ToNot(HaveOccurred())
		Expect(rm2.Status.Conditions[0].Type == rolloutsmanagerv1alpha1.RolloutManagerConditionTypeErrorOccurred &&
			rm2.Status.Conditions[0].Reason == rolloutsmanagerv1alpha1.RolloutManagerReasonMultipleClusterScopedRolloutManager &&
			rm2.Status.Conditions[0].Message == MultipleRMsNotAllowed &&
			rm2.Status.Conditions[0].Status == metav1.ConditionTrue).To(BeTrue())
	})
})
