package rollouts

import (
	"context"
	"fmt"
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

func TestCheckForExistingRolloutManager_multipleRM(t *testing.T) {
	s := scheme.Scheme
	assert.NoError(t, rolloutsmanagerv1alpha1.AddToScheme(s))
	ctx := context.Background()
	log := logr.Log.WithName("rollouts-controller")
	k8sClient := fake.NewClientBuilder().WithScheme(s).Build()

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

	err := checkForExistingRolloutManager(ctx, k8sClient, &rolloutsManager1, log)
	assert.NoError(t, err)

	fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++=")

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
	err = checkForExistingRolloutManager(ctx, k8sClient, &rolloutsManager2, log)

	err = k8sClient.Get(ctx, types.NamespacedName{Name: rolloutsManager2.Name, Namespace: rolloutsManager2.Namespace}, &rolloutsManager2)
	assert.NoError(t, err)

	fmt.Println("rolloutsManager2.Status.Phase == ", rolloutsManager2.Status.Phase)
	fmt.Println("rolloutsManager2.Status.RolloutController == ", rolloutsManager2.Status.RolloutController)

	fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++=")

	err = checkForExistingRolloutManager(ctx, k8sClient, &rolloutsManager1, log)

	fmt.Println("err == ", err)
	fmt.Println("rolloutsManager1.Status.Phase == ", rolloutsManager1.Status.Phase)
	fmt.Println("rolloutsManager1.Status.RolloutController == ", rolloutsManager1.Status.RolloutController)

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
