package rollouts

import (
	"context"
	"fmt"

	"github.com/argoproj-labs/argo-rollouts-manager/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("Resource creation and cleanup tests", func() {

	Context("Resource creation test", func() {
		var (
			ctx context.Context
			a   *v1alpha1.RolloutManager
			r   *RolloutManagerReconciler
		)

		BeforeEach(func() {
			ctx = context.Background()
			a = makeTestRolloutManager()
			r = makeTestReconciler(a)
			err := createNamespace(r, a.Namespace)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Test for reconcileRolloutsServiceAccount function", func() {
			_, err := r.reconcileRolloutsServiceAccount(ctx, a)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Test for reconcileRolloutsRole function", func() {
			role, err := r.reconcileRolloutsRole(ctx, a)
			Expect(err).ToNot(HaveOccurred())

			By("Modify Rules of Role.")
			role.Rules[0].Verbs = append(role.Rules[0].Verbs, "test")
			Expect(r.Client.Update(ctx, role)).ToNot(HaveOccurred())

			By("Reconciler should revert modifications.")
			role, err = r.reconcileRolloutsRole(ctx, a)
			Expect(err).ToNot(HaveOccurred())
			Expect(role.Rules).To(Equal(GetPolicyRules()))
		})

		It("Test for reconcileRolloutsClusterRole function", func() {
			clusterRole, err := r.reconcileRolloutsClusterRole(ctx, a)
			Expect(err).ToNot(HaveOccurred())

			By("Modify Rules of Role.")
			clusterRole.Rules[0].Verbs = append(clusterRole.Rules[0].Verbs, "test")
			Expect(r.Client.Update(ctx, clusterRole)).ToNot(HaveOccurred())

			By("Reconciler should revert modifications.")
			clusterRole, err = r.reconcileRolloutsClusterRole(ctx, a)
			Expect(err).ToNot(HaveOccurred())
			Expect(clusterRole.Rules).To(Equal(GetPolicyRules()))
		})

		It("Test for reconcileRolloutsRoleBinding function", func() {
			sa, err := r.reconcileRolloutsServiceAccount(ctx, a)
			Expect(err).ToNot(HaveOccurred())
			role, err := r.reconcileRolloutsRole(ctx, a)
			Expect(err).ToNot(HaveOccurred())

			Expect(r.reconcileRolloutsRoleBinding(ctx, a, role, sa)).ToNot(HaveOccurred())

			By("Modify Subject of RoleBinding.")
			rb := &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      DefaultArgoRolloutsResourceName,
					Namespace: a.Namespace,
				},
			}
			Expect(fetchObject(ctx, r.Client, a.Namespace, rb.Name, rb)).ToNot(HaveOccurred())

			subTemp := rb.Subjects
			rb.Subjects = append(rb.Subjects, rbacv1.Subject{Kind: rbacv1.ServiceAccountKind, Name: "test", Namespace: "test"})
			Expect(r.Client.Update(ctx, rb)).ToNot(HaveOccurred())

			By("Reconciler should revert modifications.")
			Expect(r.reconcileRolloutsRoleBinding(ctx, a, role, sa)).ToNot(HaveOccurred())
			Expect(fetchObject(ctx, r.Client, a.Namespace, rb.Name, rb)).ToNot(HaveOccurred())
			Expect(rb.Subjects).To(Equal(subTemp))
		})

		It("Test for reconcileRolloutsClusterRoleBinding function", func() {
			sa, err := r.reconcileRolloutsServiceAccount(ctx, a)
			Expect(err).ToNot(HaveOccurred())
			clusterRole, err := r.reconcileRolloutsClusterRole(ctx, a)
			Expect(err).ToNot(HaveOccurred())

			Expect(r.reconcileRolloutsClusterRoleBinding(ctx, a, clusterRole, sa)).ToNot(HaveOccurred())

			By("Modify Subject of ClusterRoleBinding.")
			crb := &rbacv1.ClusterRoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      DefaultArgoRolloutsResourceName,
					Namespace: a.Namespace,
				},
			}
			Expect(fetchObject(ctx, r.Client, a.Namespace, crb.Name, crb)).ToNot(HaveOccurred())

			subTemp := crb.Subjects
			crb.Subjects = append(crb.Subjects, rbacv1.Subject{Kind: rbacv1.ServiceAccountKind, Name: "test", Namespace: "test"})
			Expect(r.Client.Update(ctx, crb)).ToNot(HaveOccurred())

			By("Reconciler should revert modifications.")
			Expect(r.reconcileRolloutsClusterRoleBinding(ctx, a, clusterRole, sa)).ToNot(HaveOccurred())
			Expect(fetchObject(ctx, r.Client, a.Namespace, crb.Name, crb)).ToNot(HaveOccurred())
			Expect(crb.Subjects).To(Equal(subTemp))
		})

		It("Test for reconcileRolloutsAggregateToAdminClusterRole function", func() {
			Expect(r.reconcileRolloutsAggregateToAdminClusterRole(ctx, a)).ToNot(HaveOccurred())

			By("Modify Rules of ClusterRole.")
			clusterRole := &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Name: "argo-rollouts-aggregate-to-admin",
				},
			}
			Expect(fetchObject(ctx, r.Client, "", clusterRole.Name, clusterRole)).ToNot(HaveOccurred())
			clusterRole.Rules[0].Verbs = append(clusterRole.Rules[0].Verbs, "test")
			Expect(r.Client.Update(ctx, clusterRole)).ToNot(HaveOccurred())

			By("Reconciler should revert modifications.")
			Expect(r.reconcileRolloutsAggregateToAdminClusterRole(ctx, a)).ToNot(HaveOccurred())
			Expect(fetchObject(ctx, r.Client, "", clusterRole.Name, clusterRole)).ToNot(HaveOccurred())
			Expect(clusterRole.Rules).To(Equal(getAggregateToAdminPolicyRules()))
		})

		It("Test for reconcileRolloutsAggregateToEditClusterRole function", func() {
			Expect(r.reconcileRolloutsAggregateToEditClusterRole(ctx, a)).ToNot(HaveOccurred())

			By("Modify Rules of ClusterRole.")
			clusterRole := &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Name: "argo-rollouts-aggregate-to-edit",
				},
			}
			Expect(fetchObject(ctx, r.Client, "", clusterRole.Name, clusterRole)).ToNot(HaveOccurred())
			clusterRole.Rules[0].Verbs = append(clusterRole.Rules[0].Verbs, "test")
			Expect(r.Client.Update(ctx, clusterRole)).ToNot(HaveOccurred())

			By("Reconciler should revert modifications.")
			Expect(r.reconcileRolloutsAggregateToEditClusterRole(ctx, a)).ToNot(HaveOccurred())
			Expect(fetchObject(ctx, r.Client, "", clusterRole.Name, clusterRole)).ToNot(HaveOccurred())
			Expect(clusterRole.Rules).To(Equal(getAggregateToEditPolicyRules()))
		})

		It("Test for reconcileRolloutsAggregateToViewClusterRole function", func() {
			Expect(r.reconcileRolloutsAggregateToViewClusterRole(ctx, a)).ToNot(HaveOccurred())

			By("Modify Rules of ClusterRole.")
			clusterRole := &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Name: "argo-rollouts-aggregate-to-view",
				},
			}
			Expect(fetchObject(ctx, r.Client, "", clusterRole.Name, clusterRole)).ToNot(HaveOccurred())
			clusterRole.Rules[0].Verbs = append(clusterRole.Rules[0].Verbs, "test")
			Expect(r.Client.Update(ctx, clusterRole)).ToNot(HaveOccurred())

			By("Reconciler should revert modifications.")
			Expect(r.reconcileRolloutsAggregateToViewClusterRole(ctx, a)).ToNot(HaveOccurred())
			Expect(fetchObject(ctx, r.Client, "", clusterRole.Name, clusterRole)).ToNot(HaveOccurred())
			Expect(clusterRole.Rules).To(Equal(getAggregateToViewPolicyRules()))
		})

		It("Test for reconcileRolloutsMetricsService function", func() {
			Expect(r.reconcileRolloutsMetricsService(ctx, a)).ToNot(HaveOccurred())
		})

		It("Test for reconcileRolloutsSecrets function", func() {
			Expect(r.reconcileRolloutsSecrets(ctx, a)).ToNot(HaveOccurred())
		})
	})

	Context("Resource Cleaup test", func() {
		ctx := context.Background()
		a := makeTestRolloutManager()

		req := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      a.Name,
				Namespace: a.Namespace,
			},
		}
		resources := []runtime.Object{a}

		r := makeTestReconciler(resources...)
		err := createNamespace(r, a.Namespace)
		Expect(err).ToNot(HaveOccurred())

		res, err := r.Reconcile(ctx, req)
		Expect(err).NotTo(HaveOccurred())
		Expect(res.Requeue).Should(BeFalse(), "reconcile should not requeue request")

		err = r.Client.Delete(ctx, a)
		Expect(err).ToNot(HaveOccurred())

		tt := []struct {
			name     string
			resource client.Object
		}{
			{
				fmt.Sprintf("ServiceAccount %s", DefaultArgoRolloutsResourceName),
				&corev1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name:      DefaultArgoRolloutsResourceName,
						Namespace: a.Namespace,
					},
				},
			},
			{
				fmt.Sprintf("Role %s", DefaultArgoRolloutsResourceName),
				&rbacv1.Role{
					ObjectMeta: metav1.ObjectMeta{
						Name:      DefaultArgoRolloutsResourceName,
						Namespace: a.Namespace,
					},
				},
			},
			{
				fmt.Sprintf("RoleBinding %s", DefaultArgoRolloutsResourceName),
				&rbacv1.RoleBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name:      DefaultArgoRolloutsResourceName,
						Namespace: a.Namespace,
					},
				},
			},
			{
				fmt.Sprintf("Secret %s", DefaultRolloutsNotificationSecretName),
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      DefaultRolloutsNotificationSecretName,
						Namespace: a.Namespace,
					},
					Type: corev1.SecretTypeOpaque,
				},
			},
			{
				fmt.Sprintf("Service %s", DefaultArgoRolloutsResourceName),
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      DefaultArgoRolloutsResourceName,
						Namespace: a.Namespace,
					},
				},
			},
		}

		for _, test := range tt {
			When(test.name, func() {
				It("CleanUp all resources created for RolloutManager", func() {
					Expect(fetchObject(ctx, r.Client, a.Namespace, test.name, test.resource)).To(HaveOccurred(), fmt.Sprintf("Expected %s to be deleted", test.name))
				})
			})
		}
	})
})
