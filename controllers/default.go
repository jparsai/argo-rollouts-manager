package rollouts

const (
	// ArgoRolloutsImageEnvName is an environment variable that can be used to deploy a
	// Custom Image of rollouts controller.
	ArgoRolloutsImageEnvName = "ARGO_ROLLOUTS_IMAGE"
	// DefaultArgoRolloutsMetricsServiceName is the default name for rollouts metrics service.
	DefaultArgoRolloutsMetricsServiceName = "argo-rollouts-metrics"
	// ArgoRolloutsDefaultImage is the default image for rollouts controller.
	DefaultArgoRolloutsImage = "quay.io/argoproj/argo-rollouts"
	// ArgoRolloutsDefaultVersion is the default version for the rollouts controller.
	DefaultArgoRolloutsVersion = "v1.6.6" // v1.6.6
	// DefaultArgoRolloutsResourceName is the default name for Rollouts controller resources such as
	// deployment, service, role, rolebinding and serviceaccount.
	DefaultArgoRolloutsResourceName = "argo-rollouts"
	// DefaultRolloutsNotificationSecretName is the default name for Rollouts controller secret resource.
	DefaultRolloutsNotificationSecretName = "argo-rollouts-notification-secret" // #nosec G101
	// DefaultRolloutsServiceSelectorKey is key used by selector
	DefaultRolloutsSelectorKey = "app.kubernetes.io/name"

	// OpenShiftRolloutPluginName is the plugin name for Openshift Route Plugin
	OpenShiftRolloutPluginName = "argoproj-labs/openshift-route-plugin"

	// DefaultRolloutsConfigMapName is the default name of the ConfigMap that contains the Rollouts controller configuration
	DefaultRolloutsConfigMapName = "argo-rollouts-config"

	// NamespaceScopedArgoRolloutsController is an environment variable that can be used to configure scope of Argo Rollouts controller
	// Set true to allow only namespace-scoped Argo Rollouts controller deployment and false for cluster-scoped
	NamespaceScopedArgoRolloutsController = "NAMESPACE_SCOPED_ARGO_ROLLOUTS"
)
