/*


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

package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type SchedulingTarget struct {
	Name string `json:"name"`
}

type Schedule struct {
	// +kubebuilder:validation:Enum:=fixed;range
	Type        string `json:"type"`
	Runat       string `json:"runat"`
	Replicas    *int32 `json:"replicas,omitempty"`
	MinReplicas *int32 `json:"minReplicas,omitempty"`
	MaxReplicas int32  `json:"maxReplicas,omitempty"`
	// +kubebuilder:validation:Enum:=cpu;memory
	Metric v1.ResourceName `json:"metric,omitempty"`
}

// ScheduledScalerSpec defines the desired state of ScheduledScaler
type ScheduledScalerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	TimeZone string           `json:"timeZone,omitempty"`
	Target   SchedulingTarget `json:"target"`
	Schedule []Schedule       `json:"schedule"`
}

// ScheduledScalerStatus defines the observed state of ScheduledScaler
type ScheduledScalerStatus struct {
	Phase   string `json:"phase,omitempty"`
	Message string `json:"message,omitempty"`
	Reason  string `json:"reason,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=scheduledscalers,scope=Namespaced,shortName=scsc
// +kubebuilder:printcolumn:name="TARGET",type=string,JSONPath=`.spec.target.name`
// +kubebuilder:printcolumn:name="STATUS",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="AGE",type=date,JSONPath=`.metadata.creationTimestamp`

// ScheduledScaler is the Schema for the scheduledscalers API
type ScheduledScaler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScheduledScalerSpec   `json:"spec,omitempty"`
	Status ScheduledScalerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ScheduledScalerList contains a list of ScheduledScaler
type ScheduledScalerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ScheduledScaler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ScheduledScaler{}, &ScheduledScalerList{})
}
