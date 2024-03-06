package v1beta1

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/instaclustr/operator/pkg/instaclustr/mock/appversionsmock"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/validation"
)

func Test_validateOpenSearchNumberOfRacks(t *testing.T) {
	type args struct {
		numberOfRacks int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "invalid NumberOfRacks",
			args: args{
				numberOfRacks: 1,
			},
			wantErr: true,
		},
		{
			name: "valid NumberOfRacks",
			args: args{
				numberOfRacks: 3,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateOpenSearchNumberOfRacks(tt.args.numberOfRacks); (err != nil) != tt.wantErr {
				t.Errorf("validateOpenSearchNumberOfRacks() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_openSearchValidator_ValidateUpdate(t *testing.T) {
	api := appversionsmock.NewInstAPI()

	type fields struct {
		API validation.Validation
	}
	type args struct {
		ctx context.Context
		old runtime.Object
		new runtime.Object
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "update with BundledUseOnly",
			fields: fields{
				API: api,
			},
			args: args{
				old: &OpenSearch{
					Spec: OpenSearchSpec{
						BundledUseOnly: true,
						KNNPlugin:      true,
					},
					Status: OpenSearchStatus{
						GenericStatus: GenericStatus{
							ID: "test",
						},
					},
				},
				new: &OpenSearch{
					Spec: OpenSearchSpec{
						BundledUseOnly: true,
						KNNPlugin:      false,
					},
					Status: OpenSearchStatus{
						GenericStatus: GenericStatus{
							ID: "test",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid case",
			fields: fields{
				API: api,
			},
			args: args{
				old: &OpenSearch{
					ObjectMeta: metav1.ObjectMeta{
						Generation: 1,
					},
					Spec: OpenSearchSpec{
						KNNPlugin: true,
					},
					Status: OpenSearchStatus{
						GenericStatus: GenericStatus{
							ID:                            "test",
							CurrentClusterOperationStatus: models.NoOperation,
							State:                         models.RunningStatus,
						},
					},
				},
				new: &OpenSearch{
					ObjectMeta: metav1.ObjectMeta{
						Generation: 2,
					},
					Spec: OpenSearchSpec{
						KNNPlugin: true,
					},
					Status: OpenSearchStatus{
						GenericStatus: GenericStatus{
							ID:                            "test",
							State:                         models.RunningStatus,
							CurrentClusterOperationStatus: models.NoOperation,
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			osv := &openSearchValidator{
				API: tt.fields.API,
			}
			if err := osv.ValidateUpdate(tt.args.ctx, tt.args.old, tt.args.new); (err != nil) != tt.wantErr {
				t.Errorf("ValidateUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOpenSearchSpec_validateDedicatedManager(t *testing.T) {
	type fields struct {
		GenericClusterSpec       GenericClusterSpec
		RestoreFrom              *OpenSearchRestoreFrom
		DataCentres              []*OpenSearchDataCentre
		DataNodes                []*OpenSearchDataNodes
		Dashboards               []*OpenSearchDashboards
		ClusterManagerNodes      []*ClusterManagerNodes
		ICUPlugin                bool
		AsynchronousSearchPlugin bool
		KNNPlugin                bool
		ReportingPlugin          bool
		SQLPlugin                bool
		NotificationsPlugin      bool
		AnomalyDetectionPlugin   bool
		LoadBalancer             bool
		IndexManagementPlugin    bool
		AlertingPlugin           bool
		BundledUseOnly           bool
		PCICompliance            bool
		UserRefs                 References
		ResizeSettings           []*ResizeSettings
		IngestNodes              []*OpenSearchIngestNodes
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "dedicated manager without data nodes",
			fields: fields{
				ClusterManagerNodes: []*ClusterManagerNodes{{
					DedicatedManager: true,
				}},
				DataNodes: nil,
			},
			wantErr: true,
		},
		{
			name: "valid case",
			fields: fields{
				ClusterManagerNodes: []*ClusterManagerNodes{{
					DedicatedManager: true,
				}},
				DataNodes: []*OpenSearchDataNodes{
					{
						NodesNumber: 3,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oss := &OpenSearchSpec{
				GenericClusterSpec:       tt.fields.GenericClusterSpec,
				RestoreFrom:              tt.fields.RestoreFrom,
				DataCentres:              tt.fields.DataCentres,
				DataNodes:                tt.fields.DataNodes,
				Dashboards:               tt.fields.Dashboards,
				ClusterManagerNodes:      tt.fields.ClusterManagerNodes,
				ICUPlugin:                tt.fields.ICUPlugin,
				AsynchronousSearchPlugin: tt.fields.AsynchronousSearchPlugin,
				KNNPlugin:                tt.fields.KNNPlugin,
				ReportingPlugin:          tt.fields.ReportingPlugin,
				SQLPlugin:                tt.fields.SQLPlugin,
				NotificationsPlugin:      tt.fields.NotificationsPlugin,
				AnomalyDetectionPlugin:   tt.fields.AnomalyDetectionPlugin,
				LoadBalancer:             tt.fields.LoadBalancer,
				IndexManagementPlugin:    tt.fields.IndexManagementPlugin,
				AlertingPlugin:           tt.fields.AlertingPlugin,
				BundledUseOnly:           tt.fields.BundledUseOnly,
				PCICompliance:            tt.fields.PCICompliance,
				UserRefs:                 tt.fields.UserRefs,
				ResizeSettings:           tt.fields.ResizeSettings,
				IngestNodes:              tt.fields.IngestNodes,
			}
			if err := oss.validateDedicatedManager(); (err != nil) != tt.wantErr {
				t.Errorf("validateDedicatedManager() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOpenSearchSpec_validateUpdate(t *testing.T) {
	type fields struct {
		GenericClusterSpec       GenericClusterSpec
		RestoreFrom              *OpenSearchRestoreFrom
		DataCentres              []*OpenSearchDataCentre
		DataNodes                []*OpenSearchDataNodes
		Dashboards               []*OpenSearchDashboards
		ClusterManagerNodes      []*ClusterManagerNodes
		ICUPlugin                bool
		AsynchronousSearchPlugin bool
		KNNPlugin                bool
		ReportingPlugin          bool
		SQLPlugin                bool
		NotificationsPlugin      bool
		AnomalyDetectionPlugin   bool
		LoadBalancer             bool
		IndexManagementPlugin    bool
		AlertingPlugin           bool
		BundledUseOnly           bool
		PCICompliance            bool
		UserRefs                 References
		ResizeSettings           []*ResizeSettings
		IngestNodes              []*OpenSearchIngestNodes
	}
	type args struct {
		oldSpec OpenSearchSpec
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "change immutable field",
			fields: fields{
				KNNPlugin: false,
			},
			args: args{
				oldSpec: OpenSearchSpec{
					KNNPlugin: true,
				},
			},
			wantErr: true,
		},
		{
			name: "valid case",
			fields: fields{
				KNNPlugin: false,
			},
			args: args{
				oldSpec: OpenSearchSpec{
					KNNPlugin: false,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oss := &OpenSearchSpec{
				GenericClusterSpec:       tt.fields.GenericClusterSpec,
				RestoreFrom:              tt.fields.RestoreFrom,
				DataCentres:              tt.fields.DataCentres,
				DataNodes:                tt.fields.DataNodes,
				Dashboards:               tt.fields.Dashboards,
				ClusterManagerNodes:      tt.fields.ClusterManagerNodes,
				ICUPlugin:                tt.fields.ICUPlugin,
				AsynchronousSearchPlugin: tt.fields.AsynchronousSearchPlugin,
				KNNPlugin:                tt.fields.KNNPlugin,
				ReportingPlugin:          tt.fields.ReportingPlugin,
				SQLPlugin:                tt.fields.SQLPlugin,
				NotificationsPlugin:      tt.fields.NotificationsPlugin,
				AnomalyDetectionPlugin:   tt.fields.AnomalyDetectionPlugin,
				LoadBalancer:             tt.fields.LoadBalancer,
				IndexManagementPlugin:    tt.fields.IndexManagementPlugin,
				AlertingPlugin:           tt.fields.AlertingPlugin,
				BundledUseOnly:           tt.fields.BundledUseOnly,
				PCICompliance:            tt.fields.PCICompliance,
				UserRefs:                 tt.fields.UserRefs,
				ResizeSettings:           tt.fields.ResizeSettings,
				IngestNodes:              tt.fields.IngestNodes,
			}
			if err := oss.validateUpdate(tt.args.oldSpec); (err != nil) != tt.wantErr {
				t.Errorf("validateUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOpenSearchSpec_validateImmutableDataCentresUpdate(t *testing.T) {
	type fields struct {
		GenericClusterSpec       GenericClusterSpec
		RestoreFrom              *OpenSearchRestoreFrom
		DataCentres              []*OpenSearchDataCentre
		DataNodes                []*OpenSearchDataNodes
		Dashboards               []*OpenSearchDashboards
		ClusterManagerNodes      []*ClusterManagerNodes
		ICUPlugin                bool
		AsynchronousSearchPlugin bool
		KNNPlugin                bool
		ReportingPlugin          bool
		SQLPlugin                bool
		NotificationsPlugin      bool
		AnomalyDetectionPlugin   bool
		LoadBalancer             bool
		IndexManagementPlugin    bool
		AlertingPlugin           bool
		BundledUseOnly           bool
		PCICompliance            bool
		UserRefs                 References
		ResizeSettings           []*ResizeSettings
		IngestNodes              []*OpenSearchIngestNodes
	}
	type args struct {
		oldDCs []*OpenSearchDataCentre
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "change number of the dcs",
			fields: fields{
				DataCentres: []*OpenSearchDataCentre{{}, {}},
			},
			args: args{
				oldDCs: []*OpenSearchDataCentre{{}},
			},
			wantErr: true,
		},
		{
			name: "change name of the dc",
			fields: fields{
				DataCentres: []*OpenSearchDataCentre{{
					GenericDataCentreSpec: GenericDataCentreSpec{
						Name: "new",
					},
				}},
			},
			args: args{
				oldDCs: []*OpenSearchDataCentre{{
					GenericDataCentreSpec: GenericDataCentreSpec{
						Name: "old",
					},
				}},
			},
			wantErr: true,
		},
		{
			name: "change immutable field",
			fields: fields{
				DataCentres: []*OpenSearchDataCentre{{
					GenericDataCentreSpec: GenericDataCentreSpec{
						Region: "new",
					},
				}},
			},
			args: args{
				oldDCs: []*OpenSearchDataCentre{{
					GenericDataCentreSpec: GenericDataCentreSpec{
						Region: "old",
					},
				}},
			},
			wantErr: true,
		},
		{
			name: "valid case",
			fields: fields{
				DataCentres: []*OpenSearchDataCentre{{
					GenericDataCentreSpec: GenericDataCentreSpec{
						Region: "test",
					},
				}},
			},
			args: args{
				oldDCs: []*OpenSearchDataCentre{{
					GenericDataCentreSpec: GenericDataCentreSpec{
						Region: "test",
					},
				}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oss := &OpenSearchSpec{
				GenericClusterSpec:       tt.fields.GenericClusterSpec,
				RestoreFrom:              tt.fields.RestoreFrom,
				DataCentres:              tt.fields.DataCentres,
				DataNodes:                tt.fields.DataNodes,
				Dashboards:               tt.fields.Dashboards,
				ClusterManagerNodes:      tt.fields.ClusterManagerNodes,
				ICUPlugin:                tt.fields.ICUPlugin,
				AsynchronousSearchPlugin: tt.fields.AsynchronousSearchPlugin,
				KNNPlugin:                tt.fields.KNNPlugin,
				ReportingPlugin:          tt.fields.ReportingPlugin,
				SQLPlugin:                tt.fields.SQLPlugin,
				NotificationsPlugin:      tt.fields.NotificationsPlugin,
				AnomalyDetectionPlugin:   tt.fields.AnomalyDetectionPlugin,
				LoadBalancer:             tt.fields.LoadBalancer,
				IndexManagementPlugin:    tt.fields.IndexManagementPlugin,
				AlertingPlugin:           tt.fields.AlertingPlugin,
				BundledUseOnly:           tt.fields.BundledUseOnly,
				PCICompliance:            tt.fields.PCICompliance,
				UserRefs:                 tt.fields.UserRefs,
				ResizeSettings:           tt.fields.ResizeSettings,
				IngestNodes:              tt.fields.IngestNodes,
			}
			if err := oss.validateImmutableDataCentresUpdate(tt.args.oldDCs); (err != nil) != tt.wantErr {
				t.Errorf("validateImmutableDataCentresUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOpenSearchDataCentre_validateDataNode(t *testing.T) {
	type fields struct {
		GenericDataCentreSpec GenericDataCentreSpec
		PrivateLink           bool
		NumberOfRacks         int
	}
	type args struct {
		nodes []*OpenSearchDataNodes
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "invalid number of data nodes",
			fields: fields{
				NumberOfRacks: 3,
			},
			args: args{
				nodes: []*OpenSearchDataNodes{{
					NodesNumber: 4,
				}},
			},
			wantErr: true,
		},
		{
			name: "valid case",
			fields: fields{
				NumberOfRacks: 3,
			},
			args: args{
				nodes: []*OpenSearchDataNodes{{
					NodesNumber: 9,
				}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dc := &OpenSearchDataCentre{
				GenericDataCentreSpec: tt.fields.GenericDataCentreSpec,
				PrivateLink:           tt.fields.PrivateLink,
				NumberOfRacks:         tt.fields.NumberOfRacks,
			}
			if err := dc.validateDataNode(tt.args.nodes); (err != nil) != tt.wantErr {
				t.Errorf("validateDataNode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validateDataNode(t *testing.T) {
	type args struct {
		newNodes []*OpenSearchDataNodes
		oldNodes []*OpenSearchDataNodes
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "deleting dataNode",
			args: args{
				newNodes: []*OpenSearchDataNodes{{
					NodesNumber: 1,
				}},
				oldNodes: []*OpenSearchDataNodes{{
					NodesNumber: 2,
				}},
			},
			wantErr: true,
		},
		{
			name: "valid case",
			args: args{
				newNodes: []*OpenSearchDataNodes{{
					NodesNumber: 3,
				}},
				oldNodes: []*OpenSearchDataNodes{{
					NodesNumber: 2,
				}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateDataNode(tt.args.newNodes, tt.args.oldNodes); (err != nil) != tt.wantErr {
				t.Errorf("validateDataNode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOpenSearchDataCentre_ValidatePrivateLink(t *testing.T) {
	type fields struct {
		GenericDataCentreSpec GenericDataCentreSpec
		PrivateLink           bool
		NumberOfRacks         int
	}
	type args struct {
		privateNetworkCluster bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "privateLink with the not aws cloudProvider",
			fields: fields{
				GenericDataCentreSpec: GenericDataCentreSpec{
					CloudProvider: models.GCP,
				},
				PrivateLink: true,
			},
			args:    args{},
			wantErr: true,
		},
		{
			name: "privateLink with no privateNetworkCluster",
			fields: fields{
				GenericDataCentreSpec: GenericDataCentreSpec{
					CloudProvider: models.AWSVPC,
				},
				PrivateLink: true,
			},
			args: args{
				privateNetworkCluster: false,
			},
			wantErr: true,
		},
		{
			name: "valid case",
			fields: fields{
				GenericDataCentreSpec: GenericDataCentreSpec{
					CloudProvider: models.AWSVPC,
				},
				PrivateLink: true,
			},
			args: args{
				privateNetworkCluster: true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dc := &OpenSearchDataCentre{
				GenericDataCentreSpec: tt.fields.GenericDataCentreSpec,
				PrivateLink:           tt.fields.PrivateLink,
				NumberOfRacks:         tt.fields.NumberOfRacks,
			}
			if err := dc.ValidatePrivateLink(tt.args.privateNetworkCluster); (err != nil) != tt.wantErr {
				t.Errorf("ValidatePrivateLink() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
