package rollouts

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestReconcileRollouts_ServiceAccount(t *testing.T) {
	a := makeTestRolloutManager()

	r := makeTestReconciler(t, a)
	assert.NoError(t, createNamespace(r, a.Namespace))

	_, err := r.reconcileRolloutsServiceAccount(context.Background(), a)
	assert.NoError(t, err)
}

func TestReconcileRollouts_Role(t *testing.T) {
	ctx := context.Background()
	a := makeTestRolloutManager()

	r := makeTestReconciler(t, a)
	assert.NoError(t, createNamespace(r, a.Namespace))

	role, err := r.reconcileRolloutsRole(ctx, a)
	assert.NoError(t, err)

	// Modify Rules
	role.Rules[0].Verbs = append(role.Rules[0].Verbs, "test")
	assert.NoError(t, r.Client.Update(ctx, role))

	// Reconciler should revert modifications
	role, err = r.reconcileRolloutsRole(ctx, a)
	assert.NoError(t, err)

	if diff := cmp.Diff(role.Rules, GetPolicyRules()); diff != "" {
		t.Fatalf("failed to reconcile role:\n%s", diff)
	}
}

func TestReconcileRollouts_ClusterRole(t *testing.T) {
	ctx := context.Background()
	a := makeTestRolloutManager()

	r := makeTestReconciler(t, a)
	assert.NoError(t, createNamespace(r, a.Namespace))

	clusterRole, err := r.reconcileRolloutsClusterRole(ctx, a)
	assert.NoError(t, err)

	// Modify Rules
	clusterRole.Rules[0].Verbs = append(clusterRole.Rules[0].Verbs, "test")
	assert.NoError(t, r.Client.Update(ctx, clusterRole))

	// Reconciler should revert modifications
	clusterRole, err = r.reconcileRolloutsClusterRole(ctx, a)
	assert.NoError(t, err)
	if diff := cmp.Diff(clusterRole.Rules, GetPolicyRules()); diff != "" {
		t.Fatalf("failed to reconcile clusterRole:\n%s", diff)
	}
}

func TestReconcileRolloutsRoleBinding(t *testing.T) {
	ctx := context.Background()
	a := makeTestRolloutManager()

	r := makeTestReconciler(t, a)
	assert.NoError(t, createNamespace(r, a.Namespace))

	sa, err := r.reconcileRolloutsServiceAccount(ctx, a)
	assert.NoError(t, err)

	role, err := r.reconcileRolloutsRole(ctx, a)
	assert.NoError(t, err)

	assert.NoError(t, r.reconcileRolloutsRoleBinding(ctx, a, role, sa))

	// Modify Subject
	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DefaultArgoRolloutsResourceName,
			Namespace: a.Namespace,
		},
	}
	assert.NoError(t, fetchObject(ctx, r.Client, a.Namespace, rb.Name, rb))

	subTemp := rb.Subjects
	rb.Subjects = append(rb.Subjects, rbacv1.Subject{Kind: rbacv1.ServiceAccountKind, Name: "test", Namespace: "test"})
	assert.NoError(t, r.Client.Update(ctx, rb))

	// Reconciler should revert modifications
	assert.NoError(t, r.reconcileRolloutsRoleBinding(ctx, a, role, sa))

	assert.NoError(t, fetchObject(ctx, r.Client, a.Namespace, rb.Name, rb))
	if diff := cmp.Diff(rb.Subjects, subTemp); diff != "" {
		t.Fatalf("failed to reconcile roleBinding:\n%s", diff)
	}
}

func TestReconcileRolloutsClusterRoleBinding(t *testing.T) {
	ctx := context.Background()
	a := makeTestRolloutManager()

	r := makeTestReconciler(t, a)
	assert.NoError(t, createNamespace(r, a.Namespace))

	sa, err := r.reconcileRolloutsServiceAccount(ctx, a)
	assert.NoError(t, err)

	clusterRole, err := r.reconcileRolloutsClusterRole(ctx, a)
	assert.NoError(t, err)

	assert.NoError(t, r.reconcileRolloutsClusterRoleBinding(ctx, a, clusterRole, sa))

	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DefaultArgoRolloutsResourceName,
			Namespace: a.Namespace,
		},
	}

	// Modify Subject
	assert.NoError(t, fetchObject(ctx, r.Client, crb.Namespace, crb.Name, crb))

	subTemp := crb.Subjects

	crb.Subjects = append(crb.Subjects, rbacv1.Subject{Kind: rbacv1.ServiceAccountKind, Name: "test", Namespace: "test"})
	assert.NoError(t, r.Client.Update(ctx, crb))

	// Reconciler should revert modifications
	assert.NoError(t, r.reconcileRolloutsClusterRoleBinding(ctx, a, clusterRole, sa))
	assert.NoError(t, fetchObject(ctx, r.Client, crb.Namespace, crb.Name, crb))

	if diff := cmp.Diff(crb.Subjects, subTemp); diff != "" {
		t.Fatalf("failed to reconcile clusterRoleBinding:\n%s", diff)
	}
}

func TestReconcileAggregateToAdminClusterRole(t *testing.T) {
	ctx := context.Background()
	a := makeTestRolloutManager()

	r := makeTestReconciler(t, a)
	assert.NoError(t, createNamespace(r, a.Namespace))

	assert.NoError(t, r.reconcileRolloutsAggregateToAdminClusterRole(ctx, a))

	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "argo-rollouts-aggregate-to-admin",
		},
	}

	// Modify Rules
	assert.NoError(t, fetchObject(ctx, r.Client, "", clusterRole.Name, clusterRole))

	clusterRole.Rules[0].Verbs = append(clusterRole.Rules[0].Verbs, "test")
	assert.NoError(t, r.Client.Update(ctx, clusterRole))

	// Reconciler should revert modifications
	assert.NoError(t, r.reconcileRolloutsAggregateToAdminClusterRole(ctx, a))
	assert.NoError(t, fetchObject(ctx, r.Client, "", clusterRole.Name, clusterRole))

	if diff := cmp.Diff(clusterRole.Rules, getAggregateToAdminPolicyRules()); diff != "" {
		t.Fatalf("failed to reconcile clusterRole:\n%s", diff)
	}
}

func TestReconcileAggregateToEditClusterRole(t *testing.T) {
	ctx := context.Background()
	a := makeTestRolloutManager()

	r := makeTestReconciler(t, a)
	assert.NoError(t, createNamespace(r, a.Namespace))

	assert.NoError(t, r.reconcileRolloutsAggregateToEditClusterRole(ctx, a))

	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "argo-rollouts-aggregate-to-edit",
		},
	}

	// Modify Verbs
	assert.NoError(t, fetchObject(ctx, r.Client, "", clusterRole.Name, clusterRole))
	clusterRole.Rules[0].Verbs = append(clusterRole.Rules[0].Verbs, "test")
	assert.NoError(t, r.Client.Update(ctx, clusterRole))

	// Reconciler should revert modifications
	assert.NoError(t, r.reconcileRolloutsAggregateToEditClusterRole(ctx, a))

	assert.NoError(t, fetchObject(ctx, r.Client, "", clusterRole.Name, clusterRole))
	if diff := cmp.Diff(clusterRole.Rules, getAggregateToEditPolicyRules()); diff != "" {
		t.Fatalf("failed to reconcile clusterRole:\n%s", diff)
	}
}

func TestReconcileAggregateToViewClusterRole(t *testing.T) {
	ctx := context.Background()
	a := makeTestRolloutManager()

	r := makeTestReconciler(t, a)
	assert.NoError(t, createNamespace(r, a.Namespace))

	assert.NoError(t, r.reconcileRolloutsAggregateToViewClusterRole(ctx, a))

	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "argo-rollouts-aggregate-to-view",
		},
	}

	// Modify Rules
	assert.NoError(t, fetchObject(ctx, r.Client, "", clusterRole.Name, clusterRole))

	clusterRole.Rules[0].Verbs = append(clusterRole.Rules[0].Verbs, "test")
	assert.NoError(t, r.Client.Update(ctx, clusterRole))

	// Reconciler should revert modifications
	assert.NoError(t, r.reconcileRolloutsAggregateToViewClusterRole(ctx, a))

	assert.NoError(t, fetchObject(ctx, r.Client, "", clusterRole.Name, clusterRole))
	if diff := cmp.Diff(clusterRole.Rules, getAggregateToViewPolicyRules()); diff != "" {
		t.Fatalf("failed to reconcile clusterRole:\n%s", diff)
	}
}

func TestReconcileRollouts_Service(t *testing.T) {
	ctx := context.Background()
	a := makeTestRolloutManager()

	r := makeTestReconciler(t, a)
	assert.NoError(t, createNamespace(r, a.Namespace))

	err := r.reconcileRolloutsMetricsService(ctx, a)
	assert.NoError(t, err)
}

func TestReconcileRollouts_Secret(t *testing.T) {
	a := makeTestRolloutManager()

	r := makeTestReconciler(t, a)
	assert.NoError(t, createNamespace(r, a.Namespace))

	err := r.reconcileRolloutsSecrets(context.Background(), a)
	assert.NoError(t, err)
}

func TestReconcileRolloutManager_CleanUp(t *testing.T) {

	ctx := context.Background()
	a := makeTestRolloutManager()

	resources := []runtime.Object{a}

	r := makeTestReconciler(t, resources...)
	assert.NoError(t, createNamespace(r, a.Namespace))

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      a.Name,
			Namespace: a.Namespace,
		},
	}
	res, err := r.Reconcile(ctx, req)
	assert.NoError(t, err)
	if res.Requeue {
		t.Fatal("reconcile requeued request")
	}

	err = r.Client.Delete(ctx, a)
	assert.NoError(t, err)

	// check if rollouts resources are deleted
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
		t.Run(test.name, func(t *testing.T) {
			if err = fetchObject(ctx, r.Client, a.Namespace, test.name, test.resource); err == nil {
				t.Errorf("Expected %s to be deleted", test.name)
			}
		})
	}
}
