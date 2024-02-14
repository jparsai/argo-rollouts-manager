package rollouts

import (
	"context"
	"testing"

	rolloutsmanagerv1alpha1 "github.com/argoproj-labs/argo-rollouts-manager/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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

	assert.NoError(t, r.Client.Get(ctx, types.NamespacedName{Name: a.Name, Namespace: a.Namespace}, a))
	assert.True(t,
		a.Status.Conditions[0].Type == rolloutsmanagerv1alpha1.RolloutManagerConditionTypeSuccess &&
			a.Status.Conditions[0].Reason == rolloutsmanagerv1alpha1.RolloutManagerReasonSuccess &&
			a.Status.Conditions[0].Message == "" &&
			a.Status.Conditions[0].Status == metav1.ConditionFalse)

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

	a.DeletionTimestamp = &v1.Time{}
	err = r.deleteRolloutResources(ctx, a)
	assert.NoError(t, err)
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

	assert.NoError(t, r.Client.Get(ctx, types.NamespacedName{Name: a.Name, Namespace: a.Namespace}, a))
	assert.True(t,
		a.Status.Conditions[0].Type == rolloutsmanagerv1alpha1.RolloutManagerConditionTypeSuccess &&
			a.Status.Conditions[0].Reason == rolloutsmanagerv1alpha1.RolloutManagerReasonSuccess &&
			a.Status.Conditions[0].Message == "" &&
			a.Status.Conditions[0].Status == metav1.ConditionFalse)

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
		Name: DefaultArgoRolloutsResourceName, Namespace: testNamespace,
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

func TestReconcileRolloutManager_verifyRolloutsResources_clusterScoped_multiple(t *testing.T) {

	ctx := context.Background()
	rm1 := makeTestRolloutManager()

	r1 := makeTestReconciler(t, rm1)
	assert.NoError(t, createNamespace(r1, rm1.Namespace))

	req1 := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      rm1.Name,
			Namespace: rm1.Namespace,
		},
	}

	res, err := r1.Reconcile(ctx, req1)
	assert.NoError(t, err)
	if res.Requeue {
		t.Fatal("reconcile requeued request")
	}

	assert.NoError(t, r1.Client.Get(ctx, types.NamespacedName{Name: rm1.Name, Namespace: rm1.Namespace}, rm1))
	assert.True(t,
		rm1.Status.Conditions[0].Type == rolloutsmanagerv1alpha1.RolloutManagerConditionTypeSuccess &&
			rm1.Status.Conditions[0].Reason == rolloutsmanagerv1alpha1.RolloutManagerReasonSuccess &&
			rm1.Status.Conditions[0].Message == "" &&
			rm1.Status.Conditions[0].Status == metav1.ConditionFalse)

	rm2 := makeTestRolloutManager()
	rm2.Name = "test-rm"
	rm2.Namespace = "test-ns"

	assert.NoError(t, createNamespace(r1, rm2.Namespace))
	assert.NoError(t, r1.Client.Create(ctx, rm2))

	req2 := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      rm2.Name,
			Namespace: rm2.Namespace,
		},
	}
	res1, err := r1.Reconcile(ctx, req2)

	assert.True(t, doMultipleRolloutManagersExist(err))
	if res1.Requeue {
		t.Fatal("reconcile requeued request")
	}

	assert.NoError(t, r1.Client.Get(ctx, types.NamespacedName{Name: rm2.Name, Namespace: rm2.Namespace}, rm2))
	assert.True(t,
		rm2.Status.Conditions[0].Type == rolloutsmanagerv1alpha1.RolloutManagerConditionTypeErrorOccurred &&
			rm2.Status.Conditions[0].Reason == rolloutsmanagerv1alpha1.RolloutManagerReasonMultipleClusterScopedRolloutManager &&
			rm2.Status.Conditions[0].Message == "With a cluster scoped RolloutManager, another RolloutManager is not supported" &&
			rm2.Status.Conditions[0].Status == metav1.ConditionTrue)
}