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
	//"cmd/go/internal/str"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LocustLoadTestSpec defines the desired state of LocustLoadTest
type LocustLoadTestSpec struct {
	DeploymentName string `json:"deploymentName"`
	//HostURL is the url the loadtest is executed agains
	HostURL string `json:"hosturl"`
	//LocustSpec is the locust file to define tests
	LocustSpec string `json:"locustspec"`
	//TestDuration defines the duration of locaust test to run
	SpecRepository string `json:"specrepository"`
	//NumberOfUsers is the maximum number of users to simulate
	Workers *int32 `json:"workers,omitempty"`
}

// LocustLoadTestStatus defines the observed state of LocustLoadTest
type LocustLoadTestStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	CurrentWorkers int32 `json:"currentworkers,omitempty"`
}

// +kubebuilder:object:root=true

// LocustLoadTest is the Schema for the locustloadtests API
type LocustLoadTest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LocustLoadTestSpec   `json:"spec,omitempty"`
	Status LocustLoadTestStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LocustLoadTestList contains a list of LocustLoadTest
type LocustLoadTestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LocustLoadTest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LocustLoadTest{}, &LocustLoadTestList{})
}
