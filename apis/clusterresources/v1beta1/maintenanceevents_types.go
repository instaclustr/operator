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

package v1beta1

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/instaclustr/operator/pkg/validation"
)

type MaintenanceEventRescheduleSpec struct {
	ScheduledStartTime string `json:"scheduledStartTime"`
	ScheduleID         string `json:"scheduleId"`
}

// MaintenanceEventsSpec defines the desired state of MaintenanceEvents
type MaintenanceEventsSpec struct {
	ClusterID                    string                            `json:"clusterId"`
	MaintenanceEventsReschedules []*MaintenanceEventRescheduleSpec `json:"maintenanceEventsReschedule,omitempty"`
}

// MaintenanceEventsStatus defines the observed state of MaintenanceEvents
type MaintenanceEventsStatus struct {
	EventsStatuses []*MaintenanceEventStatus `json:"eventsStatuses,omitempty"`
}

type MaintenanceEventStatus struct {
	ID                    string `json:"id,omitempty"`
	Description           string `json:"description,omitempty"`
	ScheduledStartTime    string `json:"scheduledStartTime,omitempty"`
	ScheduledEndTime      string `json:"scheduledEndTime,omitempty"`
	ScheduledStartTimeMin string `json:"scheduledStartTimeMin,omitempty"`
	ScheduledStartTimeMax string `json:"scheduledStartTimeMax,omitempty"`
	IsFinalized           bool   `json:"isFinalized,omitempty"`
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
	return client.MergeFrom(old)
}

func (me *MaintenanceEvents) AreMEventsStatusesEqual(instMEventsStatuses []*MaintenanceEventStatus) bool {
	if len(instMEventsStatuses) != len(me.Status.EventsStatuses) {
		return false
	}

	for _, instMEvent := range instMEventsStatuses {
		for _, k8sMEvent := range me.Status.EventsStatuses {
			if instMEvent.ID == k8sMEvent.ID {
				if *instMEvent != *k8sMEvent {
					return false
				}

				break
			}
		}
	}

	return true
}

func (mes *MaintenanceEventsSpec) ValidateMaintenanceEventsReschedules() error {
	for _, event := range mes.MaintenanceEventsReschedules {
		if dateValid, err := validation.ValidateISODate(event.ScheduledStartTime); err != nil || !dateValid {
			return fmt.Errorf("scheduledStartTime must be provided in an ISO-8601 formatted UTC string: %v", err)
		}
	}

	return nil
}

func init() {
	SchemeBuilder.Register(&MaintenanceEvents{}, &MaintenanceEventsList{})
}
