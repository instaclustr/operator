package v1beta1

import (
	"reflect"
	"testing"

	clusterresource "github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/pkg/models"
)

func TestKafkaConnectSpec_ManagedClustersFromInstAPI(t *testing.T) {
	type fields struct {
		GenericClusterSpec GenericClusterSpec
		DataCentres        []*KafkaConnectDataCentre
		TargetCluster      []*TargetCluster
		CustomConnectors   []*CustomConnectors
	}
	type args struct {
		iClusters []*models.ManagedCluster
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantClusters []*ManagedCluster
	}{
		{
			name:         "empty response",
			args:         args{nil},
			wantClusters: nil,
		},
		{
			name: "1 managed cluster from inst API, in k8s we have ref to cluster",
			fields: fields{
				TargetCluster: []*TargetCluster{{ManagedCluster: []*ManagedCluster{{ClusterRef: &clusterresource.ClusterRef{Name: "test-ref"}}}}},
			},
			args: args{iClusters: []*models.ManagedCluster{{
				TargetKafkaClusterID: "test-id",
				KafkaConnectVPCType:  "test-vpc",
			}}},
			wantClusters: []*ManagedCluster{{
				TargetKafkaClusterID: "test-id",
				KafkaConnectVPCType:  "test-vpc",
				ClusterRef:           &clusterresource.ClusterRef{Name: "test-ref"},
			}},
		},
		{
			name: "1 managed cluster from inst API, in k8s we don't have ref to cluster",
			args: args{iClusters: []*models.ManagedCluster{{
				TargetKafkaClusterID: "test-id",
				KafkaConnectVPCType:  "test-vpc",
			}}},
			wantClusters: []*ManagedCluster{{
				TargetKafkaClusterID: "test-id",
				KafkaConnectVPCType:  "test-vpc",
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ks := &KafkaConnectSpec{
				GenericClusterSpec: tt.fields.GenericClusterSpec,
				DataCentres:        tt.fields.DataCentres,
				TargetCluster:      tt.fields.TargetCluster,
				CustomConnectors:   tt.fields.CustomConnectors,
			}
			if gotClusters := ks.ManagedClustersFromInstAPI(tt.args.iClusters); !reflect.DeepEqual(gotClusters, tt.wantClusters) {
				t.Errorf("ManagedClustersFromInstAPI() = %v, want %v", gotClusters, tt.wantClusters)
			}
		})
	}
}
