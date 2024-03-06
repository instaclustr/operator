package v1beta1

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/instaclustr/operator/pkg/instaclustr/mock/appversionsmock"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/validation"
)

func Test_kafkaValidator_ValidateCreate(t *testing.T) {
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
			name: "invalid nodes number",
			fields: fields{
				API: api,
			},
			args: args{
				obj: &Kafka{
					Spec: KafkaSpec{
						GenericClusterSpec: GenericClusterSpec{
							Name:    "test",
							SLATier: "NON_PRODUCTION",
							Version: "1.0.0",
						},
						DataCentres: []*KafkaDataCentre{{
							GenericDataCentreSpec: GenericDataCentreSpec{
								Name:          "test",
								Region:        "AF_SOUTH_1",
								CloudProvider: models.AWSVPC,
								Network:       "10.1.0.0/16",
							},
							NodesNumber: 4,
							NodeSize:    "test",
						}},
						PartitionsNumber:  1,
						ReplicationFactor: 3,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "privateLint with no PrivateNetwork cluster",
			fields: fields{
				API: api,
			},
			args: args{
				obj: &Kafka{
					Spec: KafkaSpec{
						GenericClusterSpec: GenericClusterSpec{
							Name:           "test",
							SLATier:        "NON_PRODUCTION",
							Version:        "1.0.0",
							PrivateNetwork: false,
						},
						DataCentres: []*KafkaDataCentre{{
							GenericDataCentreSpec: GenericDataCentreSpec{
								Name:          "test",
								Region:        "AF_SOUTH_1",
								CloudProvider: models.AWSVPC,
								Network:       "10.1.0.0/16",
							},
							PrivateLink: PrivateLinkSpec{{
								AdvertisedHostname: "test",
							}},
							NodesNumber: 3,
							NodeSize:    "test",
						}},
						PartitionsNumber:  1,
						ReplicationFactor: 3,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "privateLint with no aws cloudProvider",
			fields: fields{
				API: api,
			},
			args: args{
				obj: &Kafka{
					Spec: KafkaSpec{
						GenericClusterSpec: GenericClusterSpec{
							Name:           "test",
							SLATier:        "NON_PRODUCTION",
							Version:        "1.0.0",
							PrivateNetwork: true,
						},
						DataCentres: []*KafkaDataCentre{{
							GenericDataCentreSpec: GenericDataCentreSpec{
								Name:          "test",
								Region:        "asia-east1",
								CloudProvider: models.GCP,
								Network:       "10.1.0.0/16",
							},
							PrivateLink: PrivateLinkSpec{{
								AdvertisedHostname: "test",
							}},
							NodesNumber: 3,
							NodeSize:    "test",
						}},
						PartitionsNumber:  1,
						ReplicationFactor: 3,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "privateLint with no aws cloudProvider",
			fields: fields{
				API: api,
			},
			args: args{
				obj: &Kafka{
					Spec: KafkaSpec{
						GenericClusterSpec: GenericClusterSpec{
							Name:           "test",
							SLATier:        "NON_PRODUCTION",
							Version:        "1.0.0",
							PrivateNetwork: true,
						},
						DataCentres: []*KafkaDataCentre{{
							GenericDataCentreSpec: GenericDataCentreSpec{
								Name:          "test",
								Region:        "asia-east1",
								CloudProvider: models.GCP,
								Network:       "10.1.0.0/16",
							},
							PrivateLink: PrivateLinkSpec{{
								AdvertisedHostname: "test",
							}},
							NodesNumber: 3,
							NodeSize:    "test",
						}},
						PartitionsNumber:  1,
						ReplicationFactor: 3,
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
				obj: &Kafka{
					Spec: KafkaSpec{
						GenericClusterSpec: GenericClusterSpec{
							Name:           "test",
							SLATier:        "NON_PRODUCTION",
							Version:        "1.0.0",
							PrivateNetwork: true,
						},
						DataCentres: []*KafkaDataCentre{{
							GenericDataCentreSpec: GenericDataCentreSpec{
								Name:          "test",
								Region:        "AF_SOUTH_1",
								CloudProvider: models.AWSVPC,
								Network:       "10.1.0.0/16",
							},
							PrivateLink: PrivateLinkSpec{{
								AdvertisedHostname: "test",
							}},
							NodesNumber: 3,
							NodeSize:    "test",
						}},
						PartitionsNumber:  1,
						ReplicationFactor: 3,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kv := &kafkaValidator{
				API:    tt.fields.API,
				Client: tt.fields.Client,
			}
			if err := kv.ValidateCreate(tt.args.ctx, tt.args.obj); (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_kafkaValidator_ValidateUpdate(t *testing.T) {
	type fields struct {
		API    validation.Validation
		Client client.Client
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
			name:   "update operation with BundledUseOnly enabled",
			fields: fields{},
			args: args{
				new: &Kafka{
					ObjectMeta: metav1.ObjectMeta{
						Generation: 2,
					},
					Spec: KafkaSpec{
						BundledUseOnly: true,
					},
					Status: KafkaStatus{
						GenericStatus: GenericStatus{
							ID: "test",
						},
					},
				},
				old: &Kafka{
					ObjectMeta: metav1.ObjectMeta{
						Generation: 1,
					},
					Spec: KafkaSpec{
						BundledUseOnly: true,
					},
					Status: KafkaStatus{
						GenericStatus: GenericStatus{
							ID: "test",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name:   "valid update",
			fields: fields{},
			args: args{
				new: &Kafka{
					ObjectMeta: metav1.ObjectMeta{
						Generation: 2,
					},
					Status: KafkaStatus{
						GenericStatus: GenericStatus{
							State:                         models.RunningStatus,
							CurrentClusterOperationStatus: models.NoOperation,
							ID:                            "test",
						},
					},
				},
				old: &Kafka{
					ObjectMeta: metav1.ObjectMeta{
						Generation: 1,
					},
					Status: KafkaStatus{
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
			kv := &kafkaValidator{
				API:    tt.fields.API,
				Client: tt.fields.Client,
			}
			if err := kv.ValidateUpdate(tt.args.ctx, tt.args.old, tt.args.new); (err != nil) != tt.wantErr {
				t.Errorf("ValidateUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestKafkaSpec_validateUpdate(t *testing.T) {
	type fields struct {
		GenericClusterSpec        GenericClusterSpec
		ReplicationFactor         int
		PartitionsNumber          int
		AllowDeleteTopics         bool
		AutoCreateTopics          bool
		ClientToClusterEncryption bool
		ClientBrokerAuthWithMTLS  bool
		BundledUseOnly            bool
		PCICompliance             bool
		UserRefs                  References
		DedicatedZookeeper        []*DedicatedZookeeper
		DataCentres               []*KafkaDataCentre
		SchemaRegistry            []*SchemaRegistry
		RestProxy                 []*RestProxy
		KarapaceRestProxy         []*KarapaceRestProxy
		KarapaceSchemaRegistry    []*KarapaceSchemaRegistry
		Kraft                     []*Kraft
		ResizeSettings            GenericResizeSettings
	}
	type args struct {
		old *KafkaSpec
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "update immutable field",
			fields: fields{
				PartitionsNumber: 2,
			},
			args: args{
				old: &KafkaSpec{
					PartitionsNumber: 1,
				},
			},
			wantErr: true,
		},
		{
			name: "change DCs number",
			fields: fields{
				DataCentres: []*KafkaDataCentre{{}, {}},
			},
			args: args{
				old: &KafkaSpec{
					DataCentres: []*KafkaDataCentre{{}},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid nodes number",
			fields: fields{
				DataCentres: []*KafkaDataCentre{{
					NodesNumber: 4,
				}},
				ReplicationFactor: 3,
			},
			args: args{
				old: &KafkaSpec{
					DataCentres: []*KafkaDataCentre{{
						NodesNumber: 3,
					}},
					ReplicationFactor: 3,
				},
			},
			wantErr: true,
		},
		{
			name: "change SchemaRegistry",
			fields: fields{
				SchemaRegistry: []*SchemaRegistry{{
					Version: "2",
				}},
			},
			args: args{
				old: &KafkaSpec{
					SchemaRegistry: []*SchemaRegistry{{
						Version: "1",
					}},
				},
			},
			wantErr: true,
		},
		{
			name: "change KarapaceSchemaRegistry",
			fields: fields{
				KarapaceSchemaRegistry: []*KarapaceSchemaRegistry{{
					Version: "2",
				}},
			},
			args: args{
				old: &KafkaSpec{
					KarapaceSchemaRegistry: []*KarapaceSchemaRegistry{{
						Version: "1",
					}},
				},
			},
			wantErr: true,
		},
		{
			name: "change RestProxy",
			fields: fields{
				RestProxy: []*RestProxy{{
					Version: "2",
				}},
			},
			args: args{
				old: &KafkaSpec{
					RestProxy: []*RestProxy{{
						Version: "1",
					}},
				},
			},
			wantErr: true,
		},
		{
			name: "change Kraft",
			fields: fields{
				Kraft: []*Kraft{{
					ControllerNodeCount: 2,
				}},
			},
			args: args{
				old: &KafkaSpec{
					Kraft: []*Kraft{{
						ControllerNodeCount: 1,
					}},
				},
			},
			wantErr: true,
		},
		{
			name: "change KarapaceRestProxy",
			fields: fields{
				KarapaceRestProxy: []*KarapaceRestProxy{{
					Version: "2",
				}},
			},
			args: args{
				old: &KafkaSpec{
					KarapaceRestProxy: []*KarapaceRestProxy{{
						Version: "1",
					}},
				},
			},
			wantErr: true,
		},
		{
			name: "valid case",
			fields: fields{
				BundledUseOnly: true,
			},
			args: args{
				old: &KafkaSpec{
					BundledUseOnly: true,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ks := &KafkaSpec{
				GenericClusterSpec:        tt.fields.GenericClusterSpec,
				ReplicationFactor:         tt.fields.ReplicationFactor,
				PartitionsNumber:          tt.fields.PartitionsNumber,
				AllowDeleteTopics:         tt.fields.AllowDeleteTopics,
				AutoCreateTopics:          tt.fields.AutoCreateTopics,
				ClientToClusterEncryption: tt.fields.ClientToClusterEncryption,
				ClientBrokerAuthWithMTLS:  tt.fields.ClientBrokerAuthWithMTLS,
				BundledUseOnly:            tt.fields.BundledUseOnly,
				PCICompliance:             tt.fields.PCICompliance,
				UserRefs:                  tt.fields.UserRefs,
				DedicatedZookeeper:        tt.fields.DedicatedZookeeper,
				DataCentres:               tt.fields.DataCentres,
				SchemaRegistry:            tt.fields.SchemaRegistry,
				RestProxy:                 tt.fields.RestProxy,
				KarapaceRestProxy:         tt.fields.KarapaceRestProxy,
				KarapaceSchemaRegistry:    tt.fields.KarapaceSchemaRegistry,
				Kraft:                     tt.fields.Kraft,
				ResizeSettings:            tt.fields.ResizeSettings,
			}
			if err := ks.validateUpdate(tt.args.old); (err != nil) != tt.wantErr {
				t.Errorf("validateUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validateZookeeperUpdate(t *testing.T) {
	type args struct {
		new []*DedicatedZookeeper
		old []*DedicatedZookeeper
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "nil zookeeoer",
			args: args{
				new: nil,
				old: nil,
			},
			want: true,
		},
		{
			name: "change len",
			args: args{
				new: []*DedicatedZookeeper{{}, {}},
				old: []*DedicatedZookeeper{{}},
			},
			want: false,
		},
		{
			name: "change NodesNumber",
			args: args{
				new: []*DedicatedZookeeper{{
					NodesNumber: 2,
				}},
				old: []*DedicatedZookeeper{{
					NodesNumber: 1,
				}},
			},
			want: false,
		},
		{
			name: "unchanged NodesNumber",
			args: args{
				new: []*DedicatedZookeeper{{
					NodesNumber: 1,
				}},
				old: []*DedicatedZookeeper{{
					NodesNumber: 1,
				}},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateZookeeperUpdate(tt.args.new, tt.args.old); got != tt.want {
				t.Errorf("validateZookeeperUpdate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isPrivateLinkValid(t *testing.T) {
	type args struct {
		new []*PrivateLink
		old []*PrivateLink
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "both nil",
			args: args{
				new: nil,
				old: nil,
			},
			want: true,
		},
		{
			name: "change len",
			args: args{
				new: []*PrivateLink{{}, {}},
				old: []*PrivateLink{{}},
			},
			want: false,
		},
		{
			name: "change AdvertisedHostname",
			args: args{
				new: []*PrivateLink{{
					AdvertisedHostname: "new",
				}},
				old: []*PrivateLink{{
					AdvertisedHostname: "old",
				}},
			},
			want: false,
		},
		{
			name: "unchanged AdvertisedHostname",
			args: args{
				new: []*PrivateLink{{
					AdvertisedHostname: "old",
				}},
				old: []*PrivateLink{{
					AdvertisedHostname: "old",
				}},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPrivateLinkValid(tt.args.new, tt.args.old); got != tt.want {
				t.Errorf("isPrivateLinkValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKafkaSpec_validateImmutableDataCentresFieldsUpdate(t *testing.T) {
	type fields struct {
		GenericClusterSpec        GenericClusterSpec
		ReplicationFactor         int
		PartitionsNumber          int
		AllowDeleteTopics         bool
		AutoCreateTopics          bool
		ClientToClusterEncryption bool
		ClientBrokerAuthWithMTLS  bool
		BundledUseOnly            bool
		PCICompliance             bool
		UserRefs                  References
		DedicatedZookeeper        []*DedicatedZookeeper
		DataCentres               []*KafkaDataCentre
		SchemaRegistry            []*SchemaRegistry
		RestProxy                 []*RestProxy
		KarapaceRestProxy         []*KarapaceRestProxy
		KarapaceSchemaRegistry    []*KarapaceSchemaRegistry
		Kraft                     []*Kraft
		ResizeSettings            GenericResizeSettings
	}
	type args struct {
		oldSpec *KafkaSpec
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
				DataCentres: []*KafkaDataCentre{{
					GenericDataCentreSpec: GenericDataCentreSpec{
						Name: "new",
					},
				}},
			},
			args: args{
				oldSpec: &KafkaSpec{
					DataCentres: []*KafkaDataCentre{{
						GenericDataCentreSpec: GenericDataCentreSpec{
							Name: "old",
						},
					}},
				},
			},
			wantErr: true,
		},
		{
			name: "change immutable DC field",
			fields: fields{
				DataCentres: []*KafkaDataCentre{{
					GenericDataCentreSpec: GenericDataCentreSpec{
						Region: "new",
					},
				}},
			},
			args: args{
				oldSpec: &KafkaSpec{
					DataCentres: []*KafkaDataCentre{{
						GenericDataCentreSpec: GenericDataCentreSpec{
							Region: "old",
						},
					}},
				},
			},
			wantErr: true,
		},
		{
			name: "deleting node",
			fields: fields{
				DataCentres: []*KafkaDataCentre{{
					NodesNumber: 1,
				}},
			},
			args: args{
				oldSpec: &KafkaSpec{
					DataCentres: []*KafkaDataCentre{{
						NodesNumber: 2,
					}},
				},
			},
			wantErr: true,
		},
		{
			name: "valid case",
			fields: fields{
				DataCentres: []*KafkaDataCentre{{
					NodesNumber: 2,
				}},
			},
			args: args{
				oldSpec: &KafkaSpec{
					DataCentres: []*KafkaDataCentre{{
						NodesNumber: 1,
					}},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ks := &KafkaSpec{
				GenericClusterSpec:        tt.fields.GenericClusterSpec,
				ReplicationFactor:         tt.fields.ReplicationFactor,
				PartitionsNumber:          tt.fields.PartitionsNumber,
				AllowDeleteTopics:         tt.fields.AllowDeleteTopics,
				AutoCreateTopics:          tt.fields.AutoCreateTopics,
				ClientToClusterEncryption: tt.fields.ClientToClusterEncryption,
				ClientBrokerAuthWithMTLS:  tt.fields.ClientBrokerAuthWithMTLS,
				BundledUseOnly:            tt.fields.BundledUseOnly,
				PCICompliance:             tt.fields.PCICompliance,
				UserRefs:                  tt.fields.UserRefs,
				DedicatedZookeeper:        tt.fields.DedicatedZookeeper,
				DataCentres:               tt.fields.DataCentres,
				SchemaRegistry:            tt.fields.SchemaRegistry,
				RestProxy:                 tt.fields.RestProxy,
				KarapaceRestProxy:         tt.fields.KarapaceRestProxy,
				KarapaceSchemaRegistry:    tt.fields.KarapaceSchemaRegistry,
				Kraft:                     tt.fields.Kraft,
				ResizeSettings:            tt.fields.ResizeSettings,
			}
			if err := ks.validateImmutableDataCentresFieldsUpdate(tt.args.oldSpec); (err != nil) != tt.wantErr {
				t.Errorf("validateImmutableDataCentresFieldsUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
