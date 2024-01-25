package v1beta1

import (
	"k8s.io/utils/strings/slices"

	"github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/pkg/models"
)

type GenericStatus struct {
	ID                            string `json:"id,omitempty"`
	State                         string `json:"state,omitempty"`
	CurrentClusterOperationStatus string `json:"currentClusterOperationStatus,omitempty"`

	MaintenanceEvents []*v1beta1.ClusteredMaintenanceEventStatus `json:"maintenanceEvents,omitempty"`
}

func (s *GenericStatus) Equals(o *GenericStatus) bool {
	return s.ID == o.ID &&
		s.State == o.State &&
		s.CurrentClusterOperationStatus == o.CurrentClusterOperationStatus
}

func (s *GenericStatus) FromInstAPI(instaModel *models.GenericClusterFields) {
	s.ID = instaModel.ID
	s.State = instaModel.Status
	s.CurrentClusterOperationStatus = instaModel.CurrentClusterOperationStatus
}

func (s *GenericStatus) MaintenanceEventsEqual(events []*v1beta1.ClusteredMaintenanceEventStatus) bool {
	if len(s.MaintenanceEvents) != len(events) {
		return false
	}

	for i := range events {
		if !areEventStatusesEqual(events[i], s.MaintenanceEvents[i]) {
			return false
		}
	}

	return true
}

type GenericDataCentreStatus struct {
	ID     string `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
	Status string `json:"status,omitempty"`

	ResizeOperations []*ResizeOperation `json:"resizeOperations,omitempty"`
}

func (s *GenericDataCentreStatus) Equals(o *GenericDataCentreStatus) bool {
	return s.Name == o.Name &&
		s.Status == o.Status &&
		s.ID == o.ID
}

type Node struct {
	ID             string   `json:"id,omitempty"`
	Size           string   `json:"size,omitempty"`
	PublicAddress  string   `json:"publicAddress,omitempty"`
	PrivateAddress string   `json:"privateAddress,omitempty"`
	Status         string   `json:"status,omitempty"`
	Roles          []string `json:"roles,omitempty"`
	Rack           string   `json:"rack,omitempty"`
}

func (n *Node) Equals(o *Node) bool {
	return n.ID == o.ID &&
		n.Size == o.Size &&
		n.PublicAddress == o.PublicAddress &&
		n.PrivateAddress == o.PrivateAddress &&
		n.Status == o.Status &&
		n.Rack == o.Rack &&
		slices.Equal(n.Roles, o.Roles)
}

func (n *Node) FromInstAPI(instaModel *models.Node) {
	n.ID = instaModel.ID
	n.Size = instaModel.Size
	n.PublicAddress = instaModel.PublicAddress
	n.PrivateAddress = instaModel.PrivateAddress
	n.Status = instaModel.Status
	n.Roles = instaModel.Roles
	n.Rack = instaModel.Rack
}
