package v1beta1

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/pkg/instaclustr/mock/appversionsmock"
	"github.com/instaclustr/operator/pkg/validation"
)

func Test_kafkaConnectValidator_ValidateCreate(t *testing.T) {
	api := appversionsmock.NewInstAPI()

	type fields struct {
		API    validation.Validation
		Client client.Client
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
			name: "ClusterID and ClusterRef are filled",
			fields: fields{
				API: api,
			},
			args: args{
				obj: &KafkaConnect{
					Spec: KafkaConnectSpec{
						GenericClusterSpec: GenericClusterSpec{
							Name:    "test",
							SLATier: "NON_PRODUCTION",
						},
						DataCentres: []*KafkaConnectDataCentre{{
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
						TargetCluster: []*TargetCluster{{
							ManagedCluster: []*ManagedCluster{{
								TargetKafkaClusterID: "test",
								ClusterRef: &v1beta1.ClusterRef{
									Name:      "test",
									Namespace: "test",
								},
								KafkaConnectVPCType: "test",
							}},
						}},
						CustomConnectors: nil,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid TargetKafkaClusterID",
			fields: fields{
				API: api,
			},
			args: args{
				obj: &KafkaConnect{
					Spec: KafkaConnectSpec{
						GenericClusterSpec: GenericClusterSpec{
							Name:    "test",
							SLATier: "NON_PRODUCTION",
						},
						DataCentres: []*KafkaConnectDataCentre{{
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
						TargetCluster: []*TargetCluster{{
							ManagedCluster: []*ManagedCluster{{
								TargetKafkaClusterID: "test",
								KafkaConnectVPCType:  "test",
							}},
						}},
						CustomConnectors: nil,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid KafkaConnectVPCType",
			fields: fields{
				API: api,
			},
			args: args{
				obj: &KafkaConnect{
					Spec: KafkaConnectSpec{
						GenericClusterSpec: GenericClusterSpec{
							Name:    "test",
							SLATier: "NON_PRODUCTION",
						},
						DataCentres: []*KafkaConnectDataCentre{{
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
						TargetCluster: []*TargetCluster{{
							ManagedCluster: []*ManagedCluster{{
								TargetKafkaClusterID: "9b5ba158-12f5-4681-a279-df5c371b417c",
								KafkaConnectVPCType:  "test",
							}},
						}},
						CustomConnectors: nil,
					},
				},
			},
			wantErr: true,
		}, {
			name: "invalid number of nodes",
			fields: fields{
				API: api,
			},
			args: args{
				obj: &KafkaConnect{
					Spec: KafkaConnectSpec{
						GenericClusterSpec: GenericClusterSpec{
							Name:    "test",
							SLATier: "NON_PRODUCTION",
						},
						DataCentres: []*KafkaConnectDataCentre{{
							GenericDataCentreSpec: GenericDataCentreSpec{
								Name:          "test",
								Region:        "AF_SOUTH_1",
								CloudProvider: "AWS_VPC",
								Network:       "10.1.0.0/16",
							},
							NodeSize:          "test",
							NodesNumber:       5,
							ReplicationFactor: 3,
						}},
						TargetCluster: []*TargetCluster{{
							ManagedCluster: []*ManagedCluster{{
								TargetKafkaClusterID: "9b5ba158-12f5-4681-a279-df5c371b417c",
								KafkaConnectVPCType:  "KAFKA_VPC",
							}},
						}},
						CustomConnectors: nil,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid cluster",
			fields: fields{
				API: api,
			},
			args: args{
				obj: &KafkaConnect{
					Spec: KafkaConnectSpec{
						GenericClusterSpec: GenericClusterSpec{
							Name:    "test",
							SLATier: "NON_PRODUCTION",
						},
						DataCentres: []*KafkaConnectDataCentre{{
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
						TargetCluster: []*TargetCluster{{
							ManagedCluster: []*ManagedCluster{{
								TargetKafkaClusterID: "9b5ba158-12f5-4681-a279-df5c371b417c",
								KafkaConnectVPCType:  "KAFKA_VPC",
							}},
						}},
						CustomConnectors: nil,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kcv := &kafkaConnectValidator{
				API:    tt.fields.API,
				Client: tt.fields.Client,
			}
			if err := kcv.ValidateCreate(tt.args.ctx, tt.args.obj); (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestKafkaConnectSpec_validateImmutableDataCentresFieldsUpdate(t *testing.T) {
	type fields struct {
		GenericClusterSpec GenericClusterSpec
		DataCentres        []*KafkaConnectDataCentre
		TargetCluster      []*TargetCluster
		CustomConnectors   []*CustomConnectors
	}
	type args struct {
		oldSpec KafkaConnectSpec
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "deleting nodes",
			fields: fields{
				DataCentres: []*KafkaConnectDataCentre{{
					NodesNumber: 0,
				}},
			},
			args: args{
				oldSpec: KafkaConnectSpec{
					DataCentres: []*KafkaConnectDataCentre{{
						NodesNumber: 1,
					}},
				},
			},
			wantErr: true,
		},
		{
			name: "valid case",
			fields: fields{
				DataCentres: []*KafkaConnectDataCentre{{
					NodesNumber:       3,
					ReplicationFactor: 3,
				}},
			},
			args: args{
				oldSpec: KafkaConnectSpec{
					DataCentres: []*KafkaConnectDataCentre{{
						NodesNumber:       3,
						ReplicationFactor: 3,
					}},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kc := &KafkaConnectSpec{
				GenericClusterSpec: tt.fields.GenericClusterSpec,
				DataCentres:        tt.fields.DataCentres,
				TargetCluster:      tt.fields.TargetCluster,
				CustomConnectors:   tt.fields.CustomConnectors,
			}
			if err := kc.validateImmutableDataCentresFieldsUpdate(tt.args.oldSpec); (err != nil) != tt.wantErr {
				t.Errorf("validateImmutableDataCentresFieldsUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestKafkaConnectSpec_validateImmutableTargetClusterFieldsUpdate(t *testing.T) {
	type fields struct {
		GenericClusterSpec GenericClusterSpec
		DataCentres        []*KafkaConnectDataCentre
		TargetCluster      []*TargetCluster
		CustomConnectors   []*CustomConnectors
	}
	type args struct {
		new []*TargetCluster
		old []*TargetCluster
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "changing length of TargetCluster",
			fields: fields{},
			args: args{
				new: []*TargetCluster{{}},
				old: []*TargetCluster{{}, {}},
			},
			wantErr: true,
		},
		{
			name:   "valid case",
			fields: fields{},
			args: args{
				new: []*TargetCluster{{}},
				old: []*TargetCluster{{}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kc := &KafkaConnectSpec{
				GenericClusterSpec: tt.fields.GenericClusterSpec,
				DataCentres:        tt.fields.DataCentres,
				TargetCluster:      tt.fields.TargetCluster,
				CustomConnectors:   tt.fields.CustomConnectors,
			}
			if err := kc.validateImmutableTargetClusterFieldsUpdate(tt.args.new, tt.args.old); (err != nil) != tt.wantErr {
				t.Errorf("validateImmutableTargetClusterFieldsUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validateImmutableExternalClusterFields(t *testing.T) {
	type args struct {
		new *TargetCluster
		old *TargetCluster
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "changing ExternalCluster field",
			args: args{
				new: &TargetCluster{
					ExternalCluster: []*ExternalCluster{{
						SecurityProtocol: "new",
					}},
				},
				old: &TargetCluster{
					ExternalCluster: []*ExternalCluster{{
						SecurityProtocol: "old",
					}},
				},
			},
			wantErr: true,
		},
		{
			name: "valid case",
			args: args{
				new: &TargetCluster{
					ExternalCluster: []*ExternalCluster{{
						SecurityProtocol: "old",
					}},
				},
				old: &TargetCluster{
					ExternalCluster: []*ExternalCluster{{
						SecurityProtocol: "old",
					}},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateImmutableExternalClusterFields(tt.args.new, tt.args.old); (err != nil) != tt.wantErr {
				t.Errorf("validateImmutableExternalClusterFields() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validateImmutableManagedClusterFields(t *testing.T) {
	type args struct {
		new *TargetCluster
		old *TargetCluster
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "changing ManagedCluster field",
			args: args{
				new: &TargetCluster{
					ManagedCluster: []*ManagedCluster{{
						TargetKafkaClusterID: "new",
					}},
				},
				old: &TargetCluster{
					ManagedCluster: []*ManagedCluster{{
						TargetKafkaClusterID: "old",
					}},
				},
			},
			wantErr: true,
		},
		{
			name: "valid case",
			args: args{
				new: &TargetCluster{
					ManagedCluster: []*ManagedCluster{{
						TargetKafkaClusterID: "old",
					}},
				},
				old: &TargetCluster{
					ManagedCluster: []*ManagedCluster{{
						TargetKafkaClusterID: "old",
					}},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateImmutableManagedClusterFields(tt.args.new, tt.args.old); (err != nil) != tt.wantErr {
				t.Errorf("validateImmutableManagedClusterFields() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
