package v1beta1

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/instaclustr/operator/pkg/instaclustr/mock/appversionsmock"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/validation"
)

func Test_redisValidator_ValidateCreate(t *testing.T) {
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
			name: "multiple DataCentres",
			fields: fields{
				API: api,
			},
			args: args{
				ctx: nil,
				obj: &Redis{
					Spec: RedisSpec{
						GenericClusterSpec: GenericClusterSpec{
							Name:    "test",
							SLATier: "NON_PRODUCTION",
						},
						RestoreFrom:         nil,
						ClientEncryption:    false,
						PasswordAndUserAuth: false,
						DataCentres: []*RedisDataCentre{{
							GenericDataCentreSpec: GenericDataCentreSpec{
								Name:                "test",
								Region:              "AF_SOUTH_1",
								CloudProvider:       "AWS_VPC",
								ProviderAccountName: "test",
								Network:             "10.1.0.0/16",
							},
							NodeSize:          "test",
							MasterNodes:       3,
							ReplicaNodes:      0,
							ReplicationFactor: 3,
							PrivateLink:       nil,
						}, {
							GenericDataCentreSpec: GenericDataCentreSpec{
								Name:                "test",
								Region:              "AF_SOUTH_1",
								CloudProvider:       "AWS_VPC",
								ProviderAccountName: "test",
								Network:             "10.1.0.0/16",
							},
							NodeSize:          "test",
							MasterNodes:       3,
							ReplicaNodes:      0,
							ReplicationFactor: 3,
							PrivateLink:       nil,
						}},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid cluster",
			fields: fields{
				API:    api,
				Client: nil,
			},
			args: args{
				ctx: nil,
				obj: &Redis{
					Spec: RedisSpec{
						GenericClusterSpec: GenericClusterSpec{
							Name:    "test",
							SLATier: "NON_PRODUCTION",
						},
						RestoreFrom:         nil,
						ClientEncryption:    false,
						PasswordAndUserAuth: false,
						DataCentres: []*RedisDataCentre{{
							GenericDataCentreSpec: GenericDataCentreSpec{
								Name:                "test",
								Region:              "AF_SOUTH_1",
								CloudProvider:       "AWS_VPC",
								ProviderAccountName: "test",
								Network:             "10.1.0.0/16",
							},
							NodeSize:          "test",
							MasterNodes:       3,
							ReplicaNodes:      0,
							ReplicationFactor: 3,
							PrivateLink:       nil,
						}},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rv := &redisValidator{
				API:    tt.fields.API,
				Client: tt.fields.Client,
			}
			if err := rv.ValidateCreate(tt.args.ctx, tt.args.obj); (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRedisSpec_ValidateUpdate(t *testing.T) {
	type fields struct {
		GenericClusterSpec  GenericClusterSpec
		RestoreFrom         *RedisRestoreFrom
		ClientEncryption    bool
		PasswordAndUserAuth bool
		DataCentres         []*RedisDataCentre
		ResizeSettings      GenericResizeSettings
		UserRefs            References
	}
	type args struct {
		oldSpec RedisSpec
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "changing generic cluster field",
			fields: fields{
				GenericClusterSpec: GenericClusterSpec{
					Name: "new",
				},
			},
			args: args{
				oldSpec: RedisSpec{
					GenericClusterSpec: GenericClusterSpec{
						Name: "old",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "changing specific cluster field",
			fields: fields{
				ClientEncryption: false,
			},
			args: args{
				oldSpec: RedisSpec{
					ClientEncryption: true,
				},
			},
			wantErr: true,
		},
		{
			name: "unchanged field",
			fields: fields{
				GenericClusterSpec: GenericClusterSpec{
					Name: "old",
				},
				ClientEncryption: false,
			},
			args: args{
				oldSpec: RedisSpec{
					GenericClusterSpec: GenericClusterSpec{
						Name: "old",
					},
					ClientEncryption: false,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := &RedisSpec{
				GenericClusterSpec:  tt.fields.GenericClusterSpec,
				RestoreFrom:         tt.fields.RestoreFrom,
				ClientEncryption:    tt.fields.ClientEncryption,
				PasswordAndUserAuth: tt.fields.PasswordAndUserAuth,
				DataCentres:         tt.fields.DataCentres,
				ResizeSettings:      tt.fields.ResizeSettings,
				UserRefs:            tt.fields.UserRefs,
			}
			if err := rs.ValidateUpdate(tt.args.oldSpec); (err != nil) != tt.wantErr {
				t.Errorf("ValidateUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRedisSpec_validateDCsUpdate(t *testing.T) {
	type fields struct {
		GenericClusterSpec  GenericClusterSpec
		RestoreFrom         *RedisRestoreFrom
		ClientEncryption    bool
		PasswordAndUserAuth bool
		DataCentres         []*RedisDataCentre
		ResizeSettings      GenericResizeSettings
		UserRefs            References
	}
	type args struct {
		oldSpec RedisSpec
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "changing imutable DC field",
			fields: fields{
				DataCentres: []*RedisDataCentre{
					{
						ReplicationFactor: 0,
					},
				},
			},
			args: args{
				oldSpec: RedisSpec{
					DataCentres: []*RedisDataCentre{
						{
							ReplicationFactor: 1,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "deleting MasterNodes",
			fields: fields{
				DataCentres: []*RedisDataCentre{
					{
						MasterNodes: 0,
					},
				},
			},
			args: args{
				oldSpec: RedisSpec{
					DataCentres: []*RedisDataCentre{
						{
							MasterNodes: 1,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "deleting ReplicaNodes",
			fields: fields{
				DataCentres: []*RedisDataCentre{
					{
						ReplicaNodes: 0,
					},
				},
			},
			args: args{
				oldSpec: RedisSpec{
					DataCentres: []*RedisDataCentre{
						{
							ReplicaNodes: 1,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "deleting DC",
			fields: fields{
				DataCentres: []*RedisDataCentre{},
			},
			args: args{
				oldSpec: RedisSpec{
					DataCentres: []*RedisDataCentre{{}},
				},
			},
			wantErr: true,
		},
		{
			name: "valid case",
			fields: fields{
				DataCentres: []*RedisDataCentre{{
					MasterNodes: 3,
				}},
			},
			args: args{
				oldSpec: RedisSpec{
					DataCentres: []*RedisDataCentre{{
						MasterNodes: 3,
					}},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := &RedisSpec{
				GenericClusterSpec:  tt.fields.GenericClusterSpec,
				RestoreFrom:         tt.fields.RestoreFrom,
				ClientEncryption:    tt.fields.ClientEncryption,
				PasswordAndUserAuth: tt.fields.PasswordAndUserAuth,
				DataCentres:         tt.fields.DataCentres,
				ResizeSettings:      tt.fields.ResizeSettings,
				UserRefs:            tt.fields.UserRefs,
			}
			if err := rs.validateDCsUpdate(tt.args.oldSpec); (err != nil) != tt.wantErr {
				t.Errorf("validateDCsUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRedisDataCentre_ValidateNodesNumber(t *testing.T) {
	type fields struct {
		GenericDataCentreSpec GenericDataCentreSpec
		NodeSize              string
		MasterNodes           int
		ReplicaNodes          int
		ReplicationFactor     int
		PrivateLink           PrivateLinkSpec
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "more than 100 ReplicaNodes",
			fields: fields{
				ReplicaNodes: 101,
			},
			wantErr: true,
		},
		{
			name: "less than 3 ReplicaNodes",
			fields: fields{
				ReplicaNodes: 2,
			},
			wantErr: true,
		},
		{
			name: "more than 100 MasterNodes",
			fields: fields{
				ReplicaNodes: 101,
			},
			wantErr: true,
		},
		{
			name: "less than 3 MasterNodes",
			fields: fields{
				ReplicaNodes: 2,
			},
			wantErr: true,
		},
		{
			name: "valid case",
			fields: fields{
				ReplicaNodes: 3,
				MasterNodes:  3,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rdc := &RedisDataCentre{
				GenericDataCentreSpec: tt.fields.GenericDataCentreSpec,
				NodeSize:              tt.fields.NodeSize,
				MasterNodes:           tt.fields.MasterNodes,
				ReplicaNodes:          tt.fields.ReplicaNodes,
				ReplicationFactor:     tt.fields.ReplicationFactor,
				PrivateLink:           tt.fields.PrivateLink,
			}
			if err := rdc.ValidateNodesNumber(); (err != nil) != tt.wantErr {
				t.Errorf("ValidateNodesNumber() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRedisDataCentre_ValidatePrivateLink(t *testing.T) {
	type fields struct {
		GenericDataCentreSpec GenericDataCentreSpec
		NodeSize              string
		MasterNodes           int
		ReplicaNodes          int
		ReplicationFactor     int
		PrivateLink           PrivateLinkSpec
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "privateLink for nonAWS cloudPrivider",
			fields: fields{
				GenericDataCentreSpec: GenericDataCentreSpec{
					CloudProvider: "test",
				},
				PrivateLink: PrivateLinkSpec{{}},
			},
			wantErr: true,
		},
		{
			name: "privateLink for AWS cloudPrivider",
			fields: fields{
				GenericDataCentreSpec: GenericDataCentreSpec{
					CloudProvider: models.AWSVPC,
				},
				PrivateLink: PrivateLinkSpec{{}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rdc := &RedisDataCentre{
				GenericDataCentreSpec: tt.fields.GenericDataCentreSpec,
				NodeSize:              tt.fields.NodeSize,
				MasterNodes:           tt.fields.MasterNodes,
				ReplicaNodes:          tt.fields.ReplicaNodes,
				ReplicationFactor:     tt.fields.ReplicationFactor,
				PrivateLink:           tt.fields.PrivateLink,
			}
			if err := rdc.ValidatePrivateLink(); (err != nil) != tt.wantErr {
				t.Errorf("ValidatePrivateLink() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRedisSpec_ValidatePrivateLink(t *testing.T) {
	type fields struct {
		GenericClusterSpec  GenericClusterSpec
		RestoreFrom         *RedisRestoreFrom
		ClientEncryption    bool
		PasswordAndUserAuth bool
		DataCentres         []*RedisDataCentre
		ResizeSettings      GenericResizeSettings
		UserRefs            References
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "PrivateLink enabled for multiDC",
			fields: fields{
				DataCentres: []*RedisDataCentre{{
					PrivateLink: PrivateLinkSpec{{
						AdvertisedHostname: "test",
					}},
				}, {
					PrivateLink: PrivateLinkSpec{{
						AdvertisedHostname: "test",
					}},
				}},
			},
			wantErr: true,
		},
		{
			name: "PrivateLink enabled for single DC",
			fields: fields{
				DataCentres: []*RedisDataCentre{{
					PrivateLink: PrivateLinkSpec{{
						AdvertisedHostname: "test",
					}},
				}},
			},
			wantErr: false,
		},
		{
			name: "no PrivateLink for multiple DC",
			fields: fields{
				DataCentres: []*RedisDataCentre{{}, {}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := &RedisSpec{
				GenericClusterSpec:  tt.fields.GenericClusterSpec,
				RestoreFrom:         tt.fields.RestoreFrom,
				ClientEncryption:    tt.fields.ClientEncryption,
				PasswordAndUserAuth: tt.fields.PasswordAndUserAuth,
				DataCentres:         tt.fields.DataCentres,
				ResizeSettings:      tt.fields.ResizeSettings,
				UserRefs:            tt.fields.UserRefs,
			}
			if err := rs.ValidatePrivateLink(); (err != nil) != tt.wantErr {
				t.Errorf("ValidatePrivateLink() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
