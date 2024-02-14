package rollouts

import (
	"context"
	"testing"

	rolloutsmanagerv1alpha1 "github.com/argoproj-labs/argo-rollouts-manager/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logr "sigs.k8s.io/controller-runtime/pkg/log"
)

func TestCheckForExistingRolloutManager_singleRM(t *testing.T) {

	s := scheme.Scheme
	assert.NoError(t, rolloutsmanagerv1alpha1.AddToScheme(s))
	ctx := context.Background()
	log := logr.Log.WithName("rollouts-controller")
	k8sClient := fake.NewClientBuilder().WithScheme(s).Build()

	// Create only one RolloutManager
	rolloutsManager := rolloutsmanagerv1alpha1.RolloutManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-rm-1",
			Namespace: "test-ns-1",
		},
		Spec: rolloutsmanagerv1alpha1.RolloutManagerSpec{
			NamespaceScoped: false,
		},
	}
	assert.NoError(t, k8sClient.Create(ctx, &rolloutsManager))

	err := checkForExistingRolloutManager(ctx, k8sClient, &rolloutsManager, log)
	assert.NoError(t, err)
}

func TestCheckForExistingRolloutManager_multipleRM_withClusterScoped(t *testing.T) {
	s := scheme.Scheme
	assert.NoError(t, rolloutsmanagerv1alpha1.AddToScheme(s))
	ctx := context.Background()
	log := logr.Log.WithName("rollouts-controller")
	k8sClient := fake.NewClientBuilder().WithScheme(s).Build()

	// Create 1st RM : cluster scoped
	rolloutsManager1 := rolloutsmanagerv1alpha1.RolloutManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-rm-1",
			Namespace: "test-ns-1",
		},
		Spec: rolloutsmanagerv1alpha1.RolloutManagerSpec{
			NamespaceScoped: false,
		},
	}
	assert.NoError(t, k8sClient.Create(ctx, &rolloutsManager1))

	// There should be no errpr as only one RM is created.
	err := checkForExistingRolloutManager(ctx, k8sClient, &rolloutsManager1, log)
	assert.NoError(t, err)

	// Create 2nd RM : namespace scoped
	rolloutsManager2 := rolloutsmanagerv1alpha1.RolloutManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-rm-2",
			Namespace: "test-ns-2",
		},
		Spec: rolloutsmanagerv1alpha1.RolloutManagerSpec{
			NamespaceScoped: true,
		},
	}
	assert.NoError(t, k8sClient.Create(ctx, &rolloutsManager2))

	// 2nd RM should not be allowed
	err = checkForExistingRolloutManager(ctx, k8sClient, &rolloutsManager2, log)

	assert.True(t, doMultipleRolloutManagersExist(err))
	assert.Equal(t, rolloutsManager2.Status.Phase, rolloutsmanagerv1alpha1.PhasePending)
	assert.Equal(t, rolloutsManager2.Status.RolloutController, rolloutsmanagerv1alpha1.PhasePending)

	// Recheck 1st RM and it should also have error now. since multiple RMs are created
	assert.NoError(t, k8sClient.Get(ctx, types.NamespacedName{Name: rolloutsManager2.Name, Namespace: rolloutsManager2.Namespace}, &rolloutsManager2))

	err = checkForExistingRolloutManager(ctx, k8sClient, &rolloutsManager1, log)

	assert.True(t, doMultipleRolloutManagersExist(err))
	assert.Equal(t, rolloutsManager1.Status.Phase, rolloutsmanagerv1alpha1.PhasePending)
	assert.Equal(t, rolloutsManager1.Status.RolloutController, rolloutsmanagerv1alpha1.PhasePending)
}

func TestCheckForExistingRolloutManager_multipleRM_noClusterScoped(t *testing.T) {
	s := scheme.Scheme
	assert.NoError(t, rolloutsmanagerv1alpha1.AddToScheme(s))
	ctx := context.Background()
	log := logr.Log.WithName("rollouts-controller")
	k8sClient := fake.NewClientBuilder().WithScheme(s).Build()

	// Create 1st RM : namespace scoped
	rolloutsManager1 := rolloutsmanagerv1alpha1.RolloutManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-rm-1",
			Namespace: "test-ns-1",
		},
		Spec: rolloutsmanagerv1alpha1.RolloutManagerSpec{
			NamespaceScoped: true,
		},
	}
	assert.NoError(t, k8sClient.Create(ctx, &rolloutsManager1))

	// There should be no errpr as only one RM is created.
	err := checkForExistingRolloutManager(ctx, k8sClient, &rolloutsManager1, log)
	assert.NoError(t, err)

	// Create 2nd RM : namespace scoped
	rolloutsManager2 := rolloutsmanagerv1alpha1.RolloutManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-rm-2",
			Namespace: "test-ns-2",
		},
		Spec: rolloutsmanagerv1alpha1.RolloutManagerSpec{
			NamespaceScoped: true,
		},
	}
	assert.NoError(t, k8sClient.Create(ctx, &rolloutsManager2))

	// 2nd RM should be allowed
	err = checkForExistingRolloutManager(ctx, k8sClient, &rolloutsManager2, log)
	assert.NoError(t, err)

	// Recheck 1st RM and it should still work. since all namespace scoped RMs are created
	assert.NoError(t, k8sClient.Get(ctx, types.NamespacedName{Name: rolloutsManager2.Name, Namespace: rolloutsManager2.Namespace}, &rolloutsManager2))

	err = checkForExistingRolloutManager(ctx, k8sClient, &rolloutsManager1, log)
	assert.NoError(t, err)
}

const (
	testNamespace          = "rollouts"
	testRolloutManagerName = "rollouts"
)

type rolloutManagerOpt func(*rolloutsmanagerv1alpha1.RolloutManager)

func makeTestRolloutManager(opts ...rolloutManagerOpt) *rolloutsmanagerv1alpha1.RolloutManager {
	a := &rolloutsmanagerv1alpha1.RolloutManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testRolloutManagerName,
			Namespace: testNamespace,
		},
	}
	for _, o := range opts {
		o(a)
	}
	return a
}

func makeTestReconciler(t *testing.T, objs ...runtime.Object) *RolloutManagerReconciler {
	s := scheme.Scheme
	assert.NoError(t, rolloutsmanagerv1alpha1.AddToScheme(s))

	cl := fake.NewClientBuilder().WithScheme(s).WithRuntimeObjects(objs...).Build()
	return &RolloutManagerReconciler{
		Client: cl,
		Scheme: s,
	}
}

func createNamespace(r *RolloutManagerReconciler, n string) error {
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: n}}
	return r.Client.Create(context.Background(), ns)
}
