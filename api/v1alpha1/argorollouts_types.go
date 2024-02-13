/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"fmt"
	"sort"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RolloutManagerSpec defines the desired state of Argo Rollouts
type RolloutManagerSpec struct {

	// Env lets you specify environment for Rollouts pods
	Env []corev1.EnvVar `json:"env,omitempty"`

	// Extra Command arguments that would append to the Rollouts
	// ExtraCommandArgs will not be added, if one of these commands is already part of the Rollouts command
	// with same or different value.
	ExtraCommandArgs []string `json:"extraCommandArgs,omitempty"`

	// Image defines Argo Rollouts controller image (optional)
	Image string `json:"image,omitempty"`

	// NodePlacement defines NodeSelectors and Taints for Rollouts workloads
	NodePlacement *RolloutsNodePlacementSpec `json:"nodePlacement,omitempty"`

	// Version defines Argo Rollouts controller tag (optional)
	Version string `json:"version,omitempty"`

	// NamespaceScoped lets you specify if rollouts manager has to watch a namespace or the whole cluster
	NamespaceScoped bool `json:"namespaceScoped,omitempty"`
}

// ArgoRolloutsNodePlacementSpec is used to specify NodeSelector and Tolerations for Rollouts workloads
type RolloutsNodePlacementSpec struct {
	// NodeSelector is a field of PodSpec, it is a map of key value pairs used for node selection
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Tolerations allow the pods to schedule onto nodes with matching taints
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

// RolloutManagerStatus defines the observed state of RolloutManager
type RolloutManagerStatus struct {
	// RolloutController is a simple, high-level summary of where the RolloutController component is in its lifecycle.
	// There are three possible RolloutController values:
	// Pending: The RolloutController component has been accepted by the Kubernetes system, but one or more of the required resources have not been created.
	// Running: All of the required Pods for the RolloutController component are in a Ready state.
	// Unknown: The state of the RolloutController component could not be obtained.
	RolloutController RolloutControllerPhase `json:"rolloutController,omitempty"`
	// Phase is a simple, high-level summary of where the RolloutManager is in its lifecycle.
	// There are three possible phase values:
	// Pending: The RolloutManager has been accepted by the Kubernetes system, but one or more of the required resources have not been created.
	// Available: All of the resources for the RolloutManager are ready.
	// Unknown: The state of the RolloutManager phase could not be obtained.
	Phase RolloutControllerPhase `json:"phase,omitempty"`

	// Conditions is an array of the RolloutManager's status conditions
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

type RolloutControllerPhase string

const (
	PhaseAvailable RolloutControllerPhase = "Available"
	PhasePending   RolloutControllerPhase = "Pending"
	PhaseUnknown   RolloutControllerPhase = "Unknown"
	PhaseFailure   RolloutControllerPhase = "Failure"
)

const (
	RolloutManagerConditionTypeSuccess       = "Success"
	RolloutManagerConditionTypeErrorOccurred = "ErrorOccurred"
)

const (
	RolloutManagerReasonSuccess                             = "Success"
	RolloutManagerReasonErrorOccurred                       = "ErrorOccurred"
	RolloutManagerReasonMultipleClusterScopedRolloutManager = "MultipleClusterScopedRolloutManager"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RolloutManager is the Schema for the RolloutManagers API
type RolloutManager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RolloutManagerSpec   `json:"spec,omitempty"`
	Status RolloutManagerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RolloutManagerList contains a list of RolloutManagers
type RolloutManagerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RolloutManager `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RolloutManager{}, &RolloutManagerList{})
}

// SetConditions updates the RolloutManager status conditions for a subset of evaluated types.
// If the RolloutManager has a pre-existing condition of a type that is not in the evaluated list,
// it will be preserved. If the RolloutManager has a pre-existing condition of a type, status, reason that
// is in the evaluated list, but not in the incoming conditions list, it will be removed.
func (status *RolloutManagerStatus) SetConditions(conditions []metav1.Condition) {
	rmConditions := make([]metav1.Condition, 0)
	now := metav1.Now()
	for i := range conditions {
		condition := conditions[i]
		eci := findConditionIndex(status.Conditions, condition.Type)
		if eci >= 0 && status.Conditions[eci].Message == condition.Message && status.Conditions[eci].Reason == condition.Reason && status.Conditions[eci].Status == condition.Status {
			// If we already have a condition of this type, status and reason, only update the timestamp if something
			// has changed.
			rmConditions = append(rmConditions, status.Conditions[eci])
		} else {
			// Otherwise we use the new incoming condition with an updated timestamp:
			condition.LastTransitionTime = now
			rmConditions = append(rmConditions, condition)
		}
	}
	sort.Slice(rmConditions, func(i, j int) bool {
		left := rmConditions[i]
		right := rmConditions[j]
		return fmt.Sprintf("%s/%s/%s/%s/%v", left.Type, left.Message, left.Status, left.Reason, left.LastTransitionTime) < fmt.Sprintf("%s/%s/%s/%s/%v", right.Type, right.Message, right.Status, right.Reason, right.LastTransitionTime)
	})
	status.Conditions = rmConditions
}

func findConditionIndex(conditions []metav1.Condition, t string) int {
	for i := range conditions {
		if conditions[i].Type == t {
			return i
		}
	}
	return -1
}
