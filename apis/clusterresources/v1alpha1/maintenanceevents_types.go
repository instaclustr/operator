/*
Copyright 2022.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/instaclustr/operator/pkg/models"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MaintenanceEventsSpec defines the desired state of MaintenanceEvents
type MaintenanceEventsSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ClusterID          string `json:"clusterId"`
	DayOfWeek          string `json:"dayOfWeek"`
	StartHour          int32  `json:"startHour"`
	DurationInHours    int32  `json:"durationInHours"`
	ScheduledStartTime string `json:"scheduledStartTime,omitempty"`
	EventID            string `json:"eventId,omitempty"`
}

// MaintenanceEventsStatus defines the observed state of MaintenanceEvents
type MaintenanceEventsStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ID                string              `json:"id,omitempty"`
	MaintenanceEvents []*MaintenanceEvent `json:"maintenanceEvents,omitempty"`
}

type MaintenanceEvent struct {
	EventID                   string `json:"id,omitempty"`
	ClusterID                 string `json:"clusterId,omitempty"`
	Description               string `json:"description,omitempty"`
	ExpectedServiceDisruption bool   `json:"expectedServiceDisruption"`
	ScheduledStartTime        string `json:"scheduledStartTime,omitempty"`
	ScheduledEndTime          string `json:"scheduledEndTime,omitempty"`
	ActualStartTime           string `json:"actualStartTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MaintenanceEvents is the Schema for the maintenanceevents API
type MaintenanceEvents struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MaintenanceEventsSpec   `json:"spec,omitempty"`
	Status MaintenanceEventsStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MaintenanceEventsList contains a list of MaintenanceEvents
type MaintenanceEventsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MaintenanceEvents `json:"items"`
}

func (me *MaintenanceEvents) GetJobID(jobName string) string {
	return client.ObjectKeyFromObject(me).String() + "/" + jobName
}

func (me *MaintenanceEvents) NewPatch() client.Patch {
	old := me.DeepCopy()
	old.Annotations[models.ResourceStateAnnotation] = ""
	return client.MergeFrom(old)
}

func init() {
	SchemeBuilder.Register(&MaintenanceEvents{}, &MaintenanceEventsList{})
}
