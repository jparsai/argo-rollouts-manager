package rollouts

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestReconcileRolloutManager_verifyRolloutsResources_namespaceScoped(t *testing.T) {

	ctx := context.Background()
	a := makeTestRolloutManager()

	// make it namespace scoped
	a.Spec.NamespaceScoped = true

	r := makeTestReconciler(t, a)
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

	sa := &corev1.ServiceAccount{}
	if err = r.Client.Get(ctx, types.NamespacedName{
		Name:      DefaultArgoRolloutsResourceName,
		Namespace: testNamespace,
	}, sa); err != nil {
		t.Fatalf("failed to find the rollouts serviceaccount: %#v\n", err)
	}

	role := &rbacv1.Role{}
	if err = r.Client.Get(ctx, types.NamespacedName{
		Name:      DefaultArgoRolloutsResourceName,
		Namespace: testNamespace,
	}, role); err != nil {
		t.Fatalf("failed to find the rollouts role: %#v\n", err)
	}

	rolebinding := &rbacv1.RoleBinding{}
	if err = r.Client.Get(ctx, types.NamespacedName{
		Name:      DefaultArgoRolloutsResourceName,
		Namespace: testNamespace,
	}, rolebinding); err != nil {
		t.Fatalf("failed to find the rollouts rolebinding: %#v\n", err)
	}

	aggregateToAdminClusterRole := &rbacv1.ClusterRole{}
	if err = r.Client.Get(ctx, types.NamespacedName{
		Name: "argo-rollouts-aggregate-to-admin",
	}, aggregateToAdminClusterRole); err != nil {
		t.Fatalf("failed to find the aggregateToAdmin ClusterRole: %#v\n", err)
	}

	aggregateToEditClusterRole := &rbacv1.ClusterRole{}
	if err = r.Client.Get(ctx, types.NamespacedName{
		Name: "argo-rollouts-aggregate-to-edit",
	}, aggregateToEditClusterRole); err != nil {
		t.Fatalf("failed to find the aggregateToEdit ClusterRole: %#v\n", err)
	}

	aggregateToViewClusterRole := &rbacv1.ClusterRole{}
	if err = r.Client.Get(ctx, types.NamespacedName{
		Name: "argo-rollouts-aggregate-to-view",
	}, aggregateToViewClusterRole); err != nil {
		t.Fatalf("failed to find the aggregateToView ClusterRole: %#v\n", err)
	}

	service := &corev1.Service{}
	if err = r.Client.Get(ctx, types.NamespacedName{
		Name:      DefaultArgoRolloutsMetricsServiceName,
		Namespace: a.Namespace,
	}, service); err != nil {
		t.Fatalf("failed to find the rollouts metrics service: %#v\n", err)
	}

	secret := &corev1.Secret{}
	if err = r.Client.Get(ctx, types.NamespacedName{
		Name:      DefaultRolloutsNotificationSecretName,
		Namespace: a.Namespace,
	}, secret); err != nil {
		t.Fatalf("failed to find the rollouts secret: %#v\n", err)
	}
}

func TestReconcileRolloutManager_verifyRolloutsResources_clusterScoped(t *testing.T) {

	ctx := context.Background()
	a := makeTestRolloutManager()

	r := makeTestReconciler(t, a)
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

	sa := &corev1.ServiceAccount{}
	if err = r.Client.Get(ctx, types.NamespacedName{
		Name:      DefaultArgoRolloutsResourceName,
		Namespace: testNamespace,
	}, sa); err != nil {
		t.Fatalf("failed to find the rollouts serviceaccount: %#v\n", err)
	}

	clusterRole := &rbacv1.ClusterRole{}
	if err = r.Client.Get(ctx, types.NamespacedName{
		Name:      DefaultArgoRolloutsResourceName,
		Namespace: testNamespace,
	}, clusterRole); err != nil {
		t.Fatalf("failed to find the rollouts role: %#v\n", err)
	}

	clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
	if err = r.Client.Get(ctx, types.NamespacedName{
		Name: DefaultArgoRolloutsResourceName,
	}, clusterRoleBinding); err != nil {
		t.Fatalf("failed to find the rollouts rolebinding: %#v\n", err)
	}

	aggregateToAdminClusterRole := &rbacv1.ClusterRole{}
	if err = r.Client.Get(ctx, types.NamespacedName{
		Name: "argo-rollouts-aggregate-to-admin",
	}, aggregateToAdminClusterRole); err != nil {
		t.Fatalf("failed to find the aggregateToAdmin ClusterRole: %#v\n", err)
	}

	aggregateToEditClusterRole := &rbacv1.ClusterRole{}
	if err = r.Client.Get(ctx, types.NamespacedName{
		Name: "argo-rollouts-aggregate-to-edit",
	}, aggregateToEditClusterRole); err != nil {
		t.Fatalf("failed to find the aggregateToEdit ClusterRole: %#v\n", err)
	}

	aggregateToViewClusterRole := &rbacv1.ClusterRole{}
	if err = r.Client.Get(ctx, types.NamespacedName{
		Name: "argo-rollouts-aggregate-to-view",
	}, aggregateToViewClusterRole); err != nil {
		t.Fatalf("failed to find the aggregateToView ClusterRole: %#v\n", err)
	}

	service := &corev1.Service{}
	if err = r.Client.Get(ctx, types.NamespacedName{
		Name:      DefaultArgoRolloutsMetricsServiceName,
		Namespace: a.Namespace,
	}, service); err != nil {
		t.Fatalf("failed to find the rollouts metrics service: %#v\n", err)
	}

	secret := &corev1.Secret{}
	if err = r.Client.Get(ctx, types.NamespacedName{
		Name:      DefaultRolloutsNotificationSecretName,
		Namespace: a.Namespace,
	}, secret); err != nil {
		t.Fatalf("failed to find the rollouts secret: %#v\n", err)
	}
}

func TestReconcileAggregateToAdminClusterRole(t *testing.T) {
	a := makeTestRolloutManager()

	r := makeTestReconciler(t, a)
	assert.NoError(t, createNamespace(r, a.Namespace))

	assert.NoError(t, r.reconcileRolloutsAggregateToAdminClusterRole(context.Background(), a))
}

func TestReconcileAggregateToEditClusterRole(t *testing.T) {
	a := makeTestRolloutManager()

	r := makeTestReconciler(t, a)
	assert.NoError(t, createNamespace(r, a.Namespace))

	assert.NoError(t, r.reconcileRolloutsAggregateToEditClusterRole(context.Background(), a))
}

func TestReconcileAggregateToViewClusterRole(t *testing.T) {
	a := makeTestRolloutManager()

	r := makeTestReconciler(t, a)
	assert.NoError(t, createNamespace(r, a.Namespace))

	assert.NoError(t, r.reconcileRolloutsAggregateToViewClusterRole(context.Background(), a))
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
