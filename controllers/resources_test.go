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
	a := makeTestRolloutManager()

	r := makeTestReconciler(t, a)
	assert.NoError(t, createNamespace(r, a.Namespace))

	role, err := r.reconcileRolloutsRole(context.Background(), a)
	assert.NoError(t, err)

	role.Rules[0].Verbs = append(role.Rules[0].Verbs, "test")
	assert.NoError(t, r.Client.Update(context.Background(), role))

	role, err = r.reconcileRolloutsRole(context.Background(), a)
	assert.NoError(t, err)

	if diff := cmp.Diff(role.Rules, GetPolicyRules()); diff != "" {
		t.Fatalf("failed to reconcile role:\n%s", diff)
	}
}

func TestReconcileRollouts_ClusterRole(t *testing.T) {
	a := makeTestRolloutManager()

	r := makeTestReconciler(t, a)
	assert.NoError(t, createNamespace(r, a.Namespace))

	clusterRole, err := r.reconcileRolloutsClusterRole(context.Background(), a)
	assert.NoError(t, err)

	clusterRole.Rules[0].Verbs = append(clusterRole.Rules[0].Verbs, "test")
	assert.NoError(t, r.Client.Update(context.Background(), clusterRole))

	clusterRole, err = r.reconcileRolloutsClusterRole(context.Background(), a)
	assert.NoError(t, err)
	if diff := cmp.Diff(clusterRole.Rules, GetPolicyRules()); diff != "" {
		t.Fatalf("failed to reconcile clusterRole:\n%s", diff)
	}
}

func TestReconcileRolloutsRoleBinding(t *testing.T) {
	a := makeTestRolloutManager()

	r := makeTestReconciler(t, a)
	assert.NoError(t, createNamespace(r, a.Namespace))

	sa, err := r.reconcileRolloutsServiceAccount(context.Background(), a)
	assert.NoError(t, err)

	role, err := r.reconcileRolloutsRole(context.Background(), a)
	assert.NoError(t, err)

	assert.NoError(t, r.reconcileRolloutsRoleBinding(context.Background(), a, role, sa))

	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DefaultArgoRolloutsResourceName,
			Namespace: a.Namespace,
		},
	}
	assert.NoError(t, fetchObject(context.Background(), r.Client, a.Namespace, rb.Name, rb))

	subTemp := rb.Subjects
	rb.Subjects = append(rb.Subjects, rbacv1.Subject{Kind: rbacv1.ServiceAccountKind, Name: "test", Namespace: "test"})
	assert.NoError(t, r.Client.Update(context.Background(), rb))

	assert.NoError(t, r.reconcileRolloutsRoleBinding(context.Background(), a, role, sa))
	assert.NoError(t, fetchObject(context.Background(), r.Client, a.Namespace, rb.Name, rb))

	if diff := cmp.Diff(rb.Subjects, subTemp); diff != "" {
		t.Fatalf("failed to reconcile roleBinding:\n%s", diff)
	}
}

func TestReconcileRolloutsClusterRoleBinding(t *testing.T) {
	a := makeTestRolloutManager()

	r := makeTestReconciler(t, a)
	assert.NoError(t, createNamespace(r, a.Namespace))

	sa, err := r.reconcileRolloutsServiceAccount(context.Background(), a)
	assert.NoError(t, err)

	clusterRole, err := r.reconcileRolloutsClusterRole(context.Background(), a)
	assert.NoError(t, err)

	assert.NoError(t, r.reconcileRolloutsClusterRoleBinding(context.Background(), a, clusterRole, sa))

	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DefaultArgoRolloutsResourceName,
			Namespace: a.Namespace,
		},
	}

	assert.NoError(t, fetchObject(context.Background(), r.Client, crb.Namespace, crb.Name, crb))

	subTemp := crb.Subjects

	crb.Subjects = append(crb.Subjects, rbacv1.Subject{Kind: rbacv1.ServiceAccountKind, Name: "test", Namespace: "test"})
	assert.NoError(t, r.Client.Update(context.Background(), crb))

	assert.NoError(t, r.reconcileRolloutsClusterRoleBinding(context.Background(), a, clusterRole, sa))
	assert.NoError(t, fetchObject(context.Background(), r.Client, crb.Namespace, crb.Name, crb))

	if diff := cmp.Diff(crb.Subjects, subTemp); diff != "" {
		t.Fatalf("failed to reconcile clusterRoleBinding:\n%s", diff)
	}
}

func TestReconcileAggregateToAdminClusterRole(t *testing.T) {
	a := makeTestRolloutManager()

	r := makeTestReconciler(t, a)
	assert.NoError(t, createNamespace(r, a.Namespace))

	assert.NoError(t, r.reconcileRolloutsAggregateToAdminClusterRole(context.Background(), a))

	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "argo-rollouts-aggregate-to-admin",
		},
	}

	assert.NoError(t, fetchObject(context.Background(), r.Client, "", clusterRole.Name, clusterRole))

	clusterRole.Rules[0].Verbs = append(clusterRole.Rules[0].Verbs, "test")
	assert.NoError(t, r.Client.Update(context.Background(), clusterRole))

	assert.NoError(t, r.reconcileRolloutsAggregateToAdminClusterRole(context.Background(), a))
	assert.NoError(t, fetchObject(context.Background(), r.Client, "", clusterRole.Name, clusterRole))

	if diff := cmp.Diff(clusterRole.Rules, getAggregateToAdminPolicyRules()); diff != "" {
		t.Fatalf("failed to reconcile clusterRole:\n%s", diff)
	}
}

func TestReconcileAggregateToEditClusterRole(t *testing.T) {
	a := makeTestRolloutManager()

	r := makeTestReconciler(t, a)
	assert.NoError(t, createNamespace(r, a.Namespace))

	assert.NoError(t, r.reconcileRolloutsAggregateToEditClusterRole(context.Background(), a))

	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "argo-rollouts-aggregate-to-edit",
		},
	}

	assert.NoError(t, fetchObject(context.Background(), r.Client, "", clusterRole.Name, clusterRole))

	clusterRole.Rules[0].Verbs = append(clusterRole.Rules[0].Verbs, "test")
	assert.NoError(t, r.Client.Update(context.Background(), clusterRole))

	assert.NoError(t, r.reconcileRolloutsAggregateToEditClusterRole(context.Background(), a))
	assert.NoError(t, fetchObject(context.Background(), r.Client, "", clusterRole.Name, clusterRole))

	if diff := cmp.Diff(clusterRole.Rules, getAggregateToEditPolicyRules()); diff != "" {
		t.Fatalf("failed to reconcile clusterRole:\n%s", diff)
	}
}

func TestReconcileAggregateToViewClusterRole(t *testing.T) {
	a := makeTestRolloutManager()

	r := makeTestReconciler(t, a)
	assert.NoError(t, createNamespace(r, a.Namespace))

	assert.NoError(t, r.reconcileRolloutsAggregateToViewClusterRole(context.Background(), a))

	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "argo-rollouts-aggregate-to-view",
		},
	}

	assert.NoError(t, fetchObject(context.Background(), r.Client, "", clusterRole.Name, clusterRole))

	clusterRole.Rules[0].Verbs = append(clusterRole.Rules[0].Verbs, "test")
	assert.NoError(t, r.Client.Update(context.Background(), clusterRole))

	assert.NoError(t, r.reconcileRolloutsAggregateToViewClusterRole(context.Background(), a))
	assert.NoError(t, fetchObject(context.Background(), r.Client, "", clusterRole.Name, clusterRole))

	if diff := cmp.Diff(clusterRole.Rules, getAggregateToViewPolicyRules()); diff != "" {
		t.Fatalf("failed to reconcile clusterRole:\n%s", diff)
	}
}

func TestReconcileRollouts_Service(t *testing.T) {
	a := makeTestRolloutManager()

	r := makeTestReconciler(t, a)
	assert.NoError(t, createNamespace(r, a.Namespace))

	err := r.reconcileRolloutsMetricsService(context.Background(), a)
	assert.NoError(t, err)

	/*
		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      DefaultArgoRolloutsMetricsServiceName,
				Namespace: a.Namespace,
			},
		}
		assert.NoError(t, fetchObject(context.Background(), r.Client, a.Namespace, service.Name, service))

		service.Spec.Ports[0].Port = int32(8091)
		assert.NoError(t, r.Client.Update(context.Background(), service))


		err = r.reconcileRolloutsMetricsService(context.Background(), a)
		assert.NoError(t, err)
		assert.NoError(t, fetchObject(context.Background(), r.Client, a.Namespace, service.Name, service))
		assert.Equal(t, service.Spec.Ports[0].Port, int32(8090))*/

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
