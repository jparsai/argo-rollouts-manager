package rollouts

import (
	"context"

	rolloutsmanagerv1alpha1 "github.com/argoproj-labs/argo-rollouts-manager/api/v1alpha1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *RolloutManagerReconciler) reconcileRolloutsManager(ctx context.Context, cr *rolloutsmanagerv1alpha1.RolloutManager) (metav1.Condition, error) {

	log.Info("Searching for existing RolloutManager")
	if err := checkForExistingRolloutManager(ctx, r.Client, cr, log); err != nil {
		if multipleRolloutManagersExist(err) {
			return createCondition(err.Error(), rolloutsmanagerv1alpha1.RolloutManagerReasonMultipleClusterScopedRolloutManager), nil
		}
		return createCondition(err.Error()), err
	}

	log.Info("reconciling rollouts ServiceAccount")
	sa, err := r.reconcileRolloutsServiceAccount(ctx, cr)
	if err != nil {
		return createCondition(err.Error()), err
	}

	var role *rbacv1.Role
	var clusterRole *rbacv1.ClusterRole

	if cr.Spec.NamespaceScoped {
		log.Info("reconciling rollouts Roles")
		role, err = r.reconcileRolloutsRole(ctx, cr)
		if err != nil {
			return createCondition(err.Error()), err
		}
	} else {
		log.Info("reconciling rollouts ClusterRoles")
		clusterRole, err = r.reconcileRolloutsClusterRole(ctx, cr)
		if err != nil {
			return createCondition(err.Error()), err
		}
	}

	log.Info("reconciling aggregate-to-admin ClusterRole")
	if err := r.reconcileRolloutsAggregateToAdminClusterRole(ctx, cr); err != nil {
		return createCondition(err.Error()), err
	}

	log.Info("reconciling aggregate-to-edit ClusterRole")
	if err := r.reconcileRolloutsAggregateToEditClusterRole(ctx, cr); err != nil {
		return createCondition(err.Error()), err
	}

	log.Info("reconciling aggregate-to-view ClusterRole")
	if err := r.reconcileRolloutsAggregateToViewClusterRole(ctx, cr); err != nil {
		return createCondition(err.Error()), err
	}

	if cr.Spec.NamespaceScoped {
		log.Info("reconciling rollouts role RoleBindings")
		if err := r.reconcileRolloutsRoleBinding(ctx, cr, role, sa); err != nil {
			return createCondition(err.Error()), err
		}
	} else {
		log.Info("reconciling rollouts clusterRoleBinding")
		if err := r.reconcileRolloutsClusterRoleBinding(ctx, cr, clusterRole, sa); err != nil {
			return createCondition(err.Error()), err
		}
	}

	log.Info("reconciling rollouts Secret")
	if err := r.reconcileRolloutsSecrets(ctx, cr); err != nil {
		return createCondition(err.Error()), err
	}

	// reconcile configMap for plugins
	log.Info("reconciling configMap for plugins")
	if err := r.reconcileConfigMap(ctx, cr); err != nil {
		return err
	}

	log.Info("reconciling rollouts Deployment")
	if err := r.reconcileRolloutsDeployment(ctx, cr, sa); err != nil {
		return createCondition(err.Error()), err
	}

	log.Info("reconciling rollouts metrics service")
	if err := r.reconcileRolloutsMetricsService(ctx, cr); err != nil {
		return createCondition(err.Error()), err
	}

	log.Info("reconciling status of workloads")
	if err := r.reconcileStatus(ctx, cr); err != nil {
		return createCondition(err.Error()), err
	}

	return createCondition(""), nil
}
