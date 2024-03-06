package v1beta1

import (
	"context"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/instaclustr/operator/pkg/instaclustr/mock/appversionsmock"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/validation"
)

func Test_cassandraValidator_ValidateCreate(t *testing.T) {
	api := appversionsmock.NewInstAPI()

	type fields struct {
		Client client.Client
		API    validation.Validation
	}
	type args struct {
		ctx context.Context
		obj runtime.Object
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "PrivateIPBroadcastForDiscovery with public network cluster",
			fields: fields{API: api},
			args: args{
				obj: &Cassandra{
					Spec: CassandraSpec{
						GenericClusterSpec: GenericClusterSpec{
							Name:           "test",
							SLATier:        "NON_PRODUCTION",
							PrivateNetwork: false,
						},
						DataCentres: []*CassandraDataCentre{{
							GenericDataCentreSpec: GenericDataCentreSpec{
								Name:          "test",
								Region:        "AF_SOUTH_1",
								CloudProvider: "AWS_VPC",
								Network:       "10.1.0.0/16",
							},
							NodeSize:                       "test",
							NodesNumber:                    3,
							ReplicationFactor:              3,
							PrivateIPBroadcastForDiscovery: true,
						}},
					},
				},
			},
			wantErr: true,
		},
		{
			name:   "invalid NodesNumber",
			fields: fields{API: api},
			args: args{
				obj: &Cassandra{
					Spec: CassandraSpec{
						GenericClusterSpec: GenericClusterSpec{
							Name:    "test",
							SLATier: "NON_PRODUCTION",
						},
						DataCentres: []*CassandraDataCentre{{
							GenericDataCentreSpec: GenericDataCentreSpec{
								Name:          "test",
								Region:        "AF_SOUTH_1",
								CloudProvider: "AWS_VPC",
								Network:       "10.1.0.0/16",
							},
							NodeSize:          "test",
							NodesNumber:       4,
							ReplicationFactor: 3,
						}},
					},
				},
			},
			wantErr: true,
		},
		{
			name:   "valid cluster",
			fields: fields{API: api},
			args: args{
				obj: &Cassandra{
					Spec: CassandraSpec{
						GenericClusterSpec: GenericClusterSpec{
							Name:    "test",
							SLATier: "NON_PRODUCTION",
						},
						DataCentres: []*CassandraDataCentre{{
							GenericDataCentreSpec: GenericDataCentreSpec{
								Name:          "test",
								Region:        "AF_SOUTH_1",
								CloudProvider: "AWS_VPC",
								Network:       "10.1.0.0/16",
							},
							NodeSize:          "test",
							NodesNumber:       3,
							ReplicationFactor: 3,
						}},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cv := &cassandraValidator{
				Client: tt.fields.Client,
				API:    tt.fields.API,
			}
			if err := cv.ValidateCreate(tt.args.ctx, tt.args.obj); (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_cassandraValidator_ValidateUpdate(t *testing.T) {
	api := appversionsmock.NewInstAPI()

	type fields struct {
		Client client.Client
		API    validation.Validation
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
			name: "update operation with BundledUseOnly enabled",
			fields: fields{
				API: api,
			},
			args: args{
				old: &Cassandra{
					ObjectMeta: v1.ObjectMeta{
						Generation: 1,
					},
					Spec: CassandraSpec{
						BundledUseOnly: true,
					},
					Status: CassandraStatus{
						GenericStatus: GenericStatus{
							ID: "test",
						},
					},
				},
				new: &Cassandra{
					ObjectMeta: v1.ObjectMeta{
						Generation: 2,
					},
					Spec: CassandraSpec{
						BundledUseOnly: true,
					},
					Status: CassandraStatus{
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
				old: &Cassandra{
					ObjectMeta: v1.ObjectMeta{
						Generation: 1,
					},
					Status: CassandraStatus{
						GenericStatus: GenericStatus{
							State:                         models.RunningStatus,
							CurrentClusterOperationStatus: models.NoOperation,
							ID:                            "test",
						},
					},
				},
				new: &Cassandra{
					ObjectMeta: v1.ObjectMeta{
						Generation: 2,
					},
					Status: CassandraStatus{
						GenericStatus: GenericStatus{
							State:                         models.RunningStatus,
							CurrentClusterOperationStatus: models.NoOperation,
							ID:                            "test",
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cv := &cassandraValidator{
				Client: tt.fields.Client,
				API:    tt.fields.API,
			}
			if err := cv.ValidateUpdate(tt.args.ctx, tt.args.old, tt.args.new); (err != nil) != tt.wantErr {
				t.Errorf("ValidateUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCassandraSpec_validateDataCentresUpdate(t *testing.T) {
	type fields struct {
		GenericClusterSpec  GenericClusterSpec
		RestoreFrom         *CassandraRestoreFrom
		DataCentres         []*CassandraDataCentre
		LuceneEnabled       bool
		PasswordAndUserAuth bool
		BundledUseOnly      bool
		PCICompliance       bool
		UserRefs            References
		ResizeSettings      GenericResizeSettings
	}
	type args struct {
		oldSpec CassandraSpec
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "change DC name",
			fields: fields{
				DataCentres: []*CassandraDataCentre{{
					GenericDataCentreSpec: GenericDataCentreSpec{
						Name: "new",
					},
				}},
			},
			args: args{
				oldSpec: CassandraSpec{
					DataCentres: []*CassandraDataCentre{{
						GenericDataCentreSpec: GenericDataCentreSpec{
							Name: "old",
						},
					}}},
			},
			wantErr: true,
		},
		{
			name: "PrivateIPBroadcastForDiscovery with public network cluster",
			fields: fields{
				DataCentres: []*CassandraDataCentre{{
					GenericDataCentreSpec: GenericDataCentreSpec{
						Name: "old",
					},
					NodesNumber:       3,
					ReplicationFactor: 3,
				},
					{
						GenericDataCentreSpec: GenericDataCentreSpec{
							Name:          "new",
							Region:        "AF_SOUTH_1",
							CloudProvider: "AWS_VPC",
							Network:       "10.1.0.0/16",
						},
						NodeSize:                       "test",
						NodesNumber:                    3,
						ReplicationFactor:              3,
						PrivateIPBroadcastForDiscovery: true,
					}},
				GenericClusterSpec: GenericClusterSpec{
					PrivateNetwork: false,
				},
			},
			args: args{
				oldSpec: CassandraSpec{
					DataCentres: []*CassandraDataCentre{{
						GenericDataCentreSpec: GenericDataCentreSpec{
							Name: "old",
						},
						NodesNumber:       3,
						ReplicationFactor: 3,
					}},
					GenericClusterSpec: GenericClusterSpec{
						PrivateNetwork: false,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid NodesNumber",
			fields: fields{
				DataCentres: []*CassandraDataCentre{{
					GenericDataCentreSpec: GenericDataCentreSpec{
						Name: "old",
					},
					NodesNumber:       3,
					ReplicationFactor: 3,
				},
					{
						GenericDataCentreSpec: GenericDataCentreSpec{
							Name:          "new",
							Region:        "AF_SOUTH_1",
							CloudProvider: "AWS_VPC",
							Network:       "10.1.0.0/16",
						},
						NodeSize:          "test",
						NodesNumber:       4,
						ReplicationFactor: 3,
					}},
				GenericClusterSpec: GenericClusterSpec{
					PrivateNetwork: false,
				},
			},
			args: args{
				oldSpec: CassandraSpec{
					DataCentres: []*CassandraDataCentre{{
						GenericDataCentreSpec: GenericDataCentreSpec{
							Name: "old",
						},
						NodesNumber:       3,
						ReplicationFactor: 3,
					}},
					GenericClusterSpec: GenericClusterSpec{
						PrivateNetwork: false,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "change immutable field",
			fields: fields{
				DataCentres: []*CassandraDataCentre{{
					GenericDataCentreSpec: GenericDataCentreSpec{
						Name:   "test",
						Region: "new",
					},
					NodesNumber:       3,
					ReplicationFactor: 3,
				},
				},
				GenericClusterSpec: GenericClusterSpec{
					PrivateNetwork: false,
				},
			},
			args: args{
				oldSpec: CassandraSpec{
					DataCentres: []*CassandraDataCentre{{
						GenericDataCentreSpec: GenericDataCentreSpec{
							Name:   "test",
							Region: "old",
						},
						NodesNumber:       3,
						ReplicationFactor: 3,
					}},
					GenericClusterSpec: GenericClusterSpec{
						PrivateNetwork: false,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "deleting nodes",
			fields: fields{
				DataCentres: []*CassandraDataCentre{{
					GenericDataCentreSpec: GenericDataCentreSpec{
						Name: "test",
					},
					NodesNumber:       3,
					ReplicationFactor: 3,
				},
				},
				GenericClusterSpec: GenericClusterSpec{
					PrivateNetwork: false,
				},
			},
			args: args{
				oldSpec: CassandraSpec{
					DataCentres: []*CassandraDataCentre{{
						GenericDataCentreSpec: GenericDataCentreSpec{
							Name: "test",
						},
						NodesNumber:       6,
						ReplicationFactor: 3,
					}},
					GenericClusterSpec: GenericClusterSpec{
						PrivateNetwork: false,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "unequal debezium",
			fields: fields{
				DataCentres: []*CassandraDataCentre{{
					GenericDataCentreSpec: GenericDataCentreSpec{
						Name: "test",
					},
					NodesNumber:       3,
					ReplicationFactor: 3,
					Debezium:          []*DebeziumCassandraSpec{{}, {}},
				},
				},
			},
			args: args{
				oldSpec: CassandraSpec{
					DataCentres: []*CassandraDataCentre{{
						GenericDataCentreSpec: GenericDataCentreSpec{
							Name: "test",
						},
						NodesNumber:       3,
						ReplicationFactor: 3,
						Debezium:          []*DebeziumCassandraSpec{{}},
					}},
				},
			},
			wantErr: true,
		},
		{
			name: "unequal ShotoverProxy",
			fields: fields{
				DataCentres: []*CassandraDataCentre{{
					GenericDataCentreSpec: GenericDataCentreSpec{
						Name: "test",
					},
					NodesNumber:       3,
					ReplicationFactor: 3,
					ShotoverProxy:     []*ShotoverProxySpec{{}, {}},
				},
				},
			},
			args: args{
				oldSpec: CassandraSpec{
					DataCentres: []*CassandraDataCentre{{
						GenericDataCentreSpec: GenericDataCentreSpec{
							Name: "test",
						},
						NodesNumber:       3,
						ReplicationFactor: 3,
						ShotoverProxy:     []*ShotoverProxySpec{{}},
					}},
				},
			},
			wantErr: true,
		},
		{
			name: "valid case",
			fields: fields{
				DataCentres: []*CassandraDataCentre{{
					GenericDataCentreSpec: GenericDataCentreSpec{
						Name: "old",
					},
					NodesNumber:       3,
					ReplicationFactor: 3,
				},
					{
						GenericDataCentreSpec: GenericDataCentreSpec{
							Name:          "new",
							Region:        "AF_SOUTH_1",
							CloudProvider: "AWS_VPC",
							Network:       "10.1.0.0/16",
						},
						NodeSize:          "test",
						NodesNumber:       3,
						ReplicationFactor: 3,
					}},
				GenericClusterSpec: GenericClusterSpec{
					PrivateNetwork: false,
				},
			},
			args: args{
				oldSpec: CassandraSpec{
					DataCentres: []*CassandraDataCentre{{
						GenericDataCentreSpec: GenericDataCentreSpec{
							Name: "old",
						},
						NodesNumber:       3,
						ReplicationFactor: 3,
					}},
					GenericClusterSpec: GenericClusterSpec{
						PrivateNetwork: false,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := &CassandraSpec{
				GenericClusterSpec:  tt.fields.GenericClusterSpec,
				RestoreFrom:         tt.fields.RestoreFrom,
				DataCentres:         tt.fields.DataCentres,
				LuceneEnabled:       tt.fields.LuceneEnabled,
				PasswordAndUserAuth: tt.fields.PasswordAndUserAuth,
				BundledUseOnly:      tt.fields.BundledUseOnly,
				PCICompliance:       tt.fields.PCICompliance,
				UserRefs:            tt.fields.UserRefs,
				ResizeSettings:      tt.fields.ResizeSettings,
			}
			if err := cs.validateDataCentresUpdate(tt.args.oldSpec); (err != nil) != tt.wantErr {
				t.Errorf("validateDataCentresUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCassandraSpec_validateResizeSettings(t *testing.T) {
	type fields struct {
		GenericClusterSpec  GenericClusterSpec
		RestoreFrom         *CassandraRestoreFrom
		DataCentres         []*CassandraDataCentre
		LuceneEnabled       bool
		PasswordAndUserAuth bool
		BundledUseOnly      bool
		PCICompliance       bool
		UserRefs            References
		ResizeSettings      GenericResizeSettings
	}
	type args struct {
		nodeNumber int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "concurrency greater than number of nodes",
			fields: fields{
				ResizeSettings: []*ResizeSettings{{
					Concurrency: 5,
				}},
			},
			args: args{
				nodeNumber: 3,
			},
			wantErr: true,
		},
		{
			name: "valid case",
			fields: fields{
				ResizeSettings: []*ResizeSettings{{
					Concurrency: 3,
				}, {
					Concurrency: 2,
				}},
			},
			args: args{
				nodeNumber: 3,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CassandraSpec{
				GenericClusterSpec:  tt.fields.GenericClusterSpec,
				RestoreFrom:         tt.fields.RestoreFrom,
				DataCentres:         tt.fields.DataCentres,
				LuceneEnabled:       tt.fields.LuceneEnabled,
				PasswordAndUserAuth: tt.fields.PasswordAndUserAuth,
				BundledUseOnly:      tt.fields.BundledUseOnly,
				PCICompliance:       tt.fields.PCICompliance,
				UserRefs:            tt.fields.UserRefs,
				ResizeSettings:      tt.fields.ResizeSettings,
			}
			if err := c.validateResizeSettings(tt.args.nodeNumber); (err != nil) != tt.wantErr {
				t.Errorf("validateResizeSettings() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
