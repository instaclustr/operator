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

func Test_pgValidator_ValidateCreate(t *testing.T) {
	api := appversionsmock.NewInstAPI()

	type fields struct {
		API       validation.Validation
		K8sClient client.Client
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
			name: "multiDC with zero interDataCentreReplication",
			fields: fields{
				API: api,
			},
			args: args{
				obj: &PostgreSQL{
					Spec: PgSpec{
						GenericClusterSpec: GenericClusterSpec{
							Name:    "test",
							SLATier: "NON_PRODUCTION",
						},
						DataCentres: []*PgDataCentre{{
							GenericDataCentreSpec: GenericDataCentreSpec{
								Name:                "test",
								Region:              "AF_SOUTH_1",
								CloudProvider:       "AWS_VPC",
								ProviderAccountName: "test",
								Network:             "10.1.0.0/16",
							},
							NodeSize:    "test",
							NodesNumber: 3,

							IntraDataCentreReplication: []*IntraDataCentreReplication{{
								ReplicationMode: "ASYNCHRONOUS",
							}},
						},
							{
								GenericDataCentreSpec: GenericDataCentreSpec{
									Name:                "test",
									Region:              "AF_SOUTH_1",
									CloudProvider:       "AWS_VPC",
									ProviderAccountName: "test",
									Network:             "10.1.0.0/16",
								},
								NodeSize:    "test",
								NodesNumber: 3,

								IntraDataCentreReplication: []*IntraDataCentreReplication{{
									ReplicationMode: "ASYNCHRONOUS",
								}},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid multiDC",
			fields: fields{
				API: api,
			},
			args: args{
				obj: &PostgreSQL{
					Spec: PgSpec{
						GenericClusterSpec: GenericClusterSpec{
							Name:    "test",
							SLATier: "NON_PRODUCTION",
						},
						DataCentres: []*PgDataCentre{{
							GenericDataCentreSpec: GenericDataCentreSpec{
								Name:                "test",
								Region:              "AF_SOUTH_1",
								CloudProvider:       "AWS_VPC",
								ProviderAccountName: "test",
								Network:             "10.1.0.0/16",
							},
							NodeSize:    "test",
							NodesNumber: 3,

							IntraDataCentreReplication: []*IntraDataCentreReplication{{
								ReplicationMode: "ASYNCHRONOUS",
							}},
							InterDataCentreReplication: []*InterDataCentreReplication{{
								IsPrimaryDataCentre: true,
							}},
						},
							{
								GenericDataCentreSpec: GenericDataCentreSpec{
									Name:                "test",
									Region:              "AF_SOUTH_1",
									CloudProvider:       "AWS_VPC",
									ProviderAccountName: "test",
									Network:             "10.1.0.0/16",
								},
								NodeSize:    "test",
								NodesNumber: 3,

								IntraDataCentreReplication: []*IntraDataCentreReplication{{
									ReplicationMode: "ASYNCHRONOUS",
								}},
								InterDataCentreReplication: []*InterDataCentreReplication{{
									IsPrimaryDataCentre: true,
								}},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid ReplicationMode",
			fields: fields{
				API: api,
			},
			args: args{
				obj: &PostgreSQL{
					Spec: PgSpec{
						GenericClusterSpec: GenericClusterSpec{
							Name:    "test",
							SLATier: "NON_PRODUCTION",
						},
						DataCentres: []*PgDataCentre{{
							GenericDataCentreSpec: GenericDataCentreSpec{
								Name:                "test",
								Region:              "AF_SOUTH_1",
								CloudProvider:       "AWS_VPC",
								ProviderAccountName: "test",
								Network:             "10.1.0.0/16",
							},
							NodeSize:    "test",
							NodesNumber: 3,

							IntraDataCentreReplication: []*IntraDataCentreReplication{{
								ReplicationMode: "test",
							}},
						},
						},
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
				obj: &PostgreSQL{
					Spec: PgSpec{
						GenericClusterSpec: GenericClusterSpec{
							Name:    "test",
							SLATier: "NON_PRODUCTION",
						},
						DataCentres: []*PgDataCentre{{
							GenericDataCentreSpec: GenericDataCentreSpec{
								Name:                "test",
								Region:              "AF_SOUTH_1",
								CloudProvider:       "AWS_VPC",
								ProviderAccountName: "test",
								Network:             "10.1.0.0/16",
							},
							NodeSize:    "test",
							NodesNumber: 3,

							IntraDataCentreReplication: []*IntraDataCentreReplication{{
								ReplicationMode: "ASYNCHRONOUS",
							}},
						},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pgv := &pgValidator{
				API:       tt.fields.API,
				K8sClient: tt.fields.K8sClient,
			}
			if err := pgv.ValidateCreate(tt.args.ctx, tt.args.obj); (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPostgreSQL_ValidateDefaultUserPassword(t *testing.T) {
	type fields struct {
		TypeMeta   v1.TypeMeta
		ObjectMeta v1.ObjectMeta
		Spec       PgSpec
		Status     PgStatus
	}
	type args struct {
		password string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "invalid password",
			fields: fields{},
			args: args{
				password: "test",
			},
			want: false,
		},
		{
			name:   "valid password",
			fields: fields{},
			args: args{
				password: "Te2t!",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pg := &PostgreSQL{
				TypeMeta:   tt.fields.TypeMeta,
				ObjectMeta: tt.fields.ObjectMeta,
				Spec:       tt.fields.Spec,
				Status:     tt.fields.Status,
			}
			if got := pg.ValidateDefaultUserPassword(tt.args.password); got != tt.want {
				t.Errorf("ValidateDefaultUserPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPgDataCentre_ValidatePGBouncer(t *testing.T) {
	type fields struct {
		GenericDataCentreSpec      GenericDataCentreSpec
		ClientEncryption           bool
		InterDataCentreReplication []*InterDataCentreReplication
		IntraDataCentreReplication []*IntraDataCentreReplication
		PGBouncer                  []*PgBouncer
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "invalid pgBouncerVersion",
			fields: fields{
				PGBouncer: []*PgBouncer{{
					PGBouncerVersion: "test",
					PoolMode:         models.PoolModes[0],
				}},
			},
			wantErr: true,
		},
		{
			name: "invalid poolMode",
			fields: fields{
				PGBouncer: []*PgBouncer{{
					PGBouncerVersion: models.PGBouncerVersions[0],
					PoolMode:         "test",
				}},
			},
			wantErr: true,
		},
		{
			name: "valid PgBouncer",
			fields: fields{
				PGBouncer: []*PgBouncer{{
					PGBouncerVersion: models.PGBouncerVersions[0],
					PoolMode:         models.PoolModes[0],
				}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pdc := &PgDataCentre{
				GenericDataCentreSpec:      tt.fields.GenericDataCentreSpec,
				ClientEncryption:           tt.fields.ClientEncryption,
				InterDataCentreReplication: tt.fields.InterDataCentreReplication,
				IntraDataCentreReplication: tt.fields.IntraDataCentreReplication,
				PGBouncer:                  tt.fields.PGBouncer,
			}
			if err := pdc.ValidatePGBouncer(); (err != nil) != tt.wantErr {
				t.Errorf("ValidatePGBouncer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPgSpec_ValidateImmutableFieldsUpdate(t *testing.T) {
	type fields struct {
		GenericClusterSpec    GenericClusterSpec
		SynchronousModeStrict bool
		ClusterConfigurations map[string]string
		PgRestoreFrom         *PgRestoreFrom
		DataCentres           []*PgDataCentre
		UserRefs              []*Reference
		ResizeSettings        []*ResizeSettings
		Extensions            PgExtensions
	}
	type args struct {
		oldSpec PgSpec
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "changing immutable field",
			fields: fields{
				SynchronousModeStrict: false,
			},
			args: args{
				oldSpec: PgSpec{
					SynchronousModeStrict: true,
				},
			},
			wantErr: true,
		},
		{
			name: "changing immutable extensions",
			fields: fields{
				Extensions: []PgExtension{{
					Name:    "test",
					Enabled: true,
				}},
			},
			args: args{
				oldSpec: PgSpec{
					Extensions: []PgExtension{{
						Name:    "test",
						Enabled: false,
					}},
				},
			},
			wantErr: true,
		},
		{
			name: "valid case",
			fields: fields{
				Extensions: []PgExtension{{
					Name:    "test",
					Enabled: false,
				}},
			},
			args: args{
				oldSpec: PgSpec{
					Extensions: []PgExtension{{
						Name:    "test",
						Enabled: false,
					}},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pgs := &PgSpec{
				GenericClusterSpec:    tt.fields.GenericClusterSpec,
				SynchronousModeStrict: tt.fields.SynchronousModeStrict,
				ClusterConfigurations: tt.fields.ClusterConfigurations,
				PgRestoreFrom:         tt.fields.PgRestoreFrom,
				DataCentres:           tt.fields.DataCentres,
				ResizeSettings:        tt.fields.ResizeSettings,
				Extensions:            tt.fields.Extensions,
			}
			if err := pgs.ValidateImmutableFieldsUpdate(tt.args.oldSpec); (err != nil) != tt.wantErr {
				t.Errorf("ValidateImmutableFieldsUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPgSpec_validateImmutableDCsFieldsUpdate(t *testing.T) {
	type fields struct {
		GenericClusterSpec    GenericClusterSpec
		SynchronousModeStrict bool
		ClusterConfigurations map[string]string
		PgRestoreFrom         *PgRestoreFrom
		DataCentres           []*PgDataCentre
		UserRefs              []*Reference
		ResizeSettings        []*ResizeSettings
		Extensions            PgExtensions
	}
	type args struct {
		oldSpec PgSpec
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "changing immutable DC field",
			fields: fields{
				DataCentres: []*PgDataCentre{{
					NodesNumber: 3,
				}},
			},
			args: args{
				oldSpec: PgSpec{
					DataCentres: []*PgDataCentre{{
						NodesNumber: 4,
					}},
				},
			},
			wantErr: true,
		},
		{
			name: "valid case",
			fields: fields{
				DataCentres: []*PgDataCentre{{
					NodesNumber: 3,
				}},
			},
			args: args{
				oldSpec: PgSpec{
					DataCentres: []*PgDataCentre{{
						NodesNumber: 4,
					}},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pgs := &PgSpec{
				GenericClusterSpec:    tt.fields.GenericClusterSpec,
				SynchronousModeStrict: tt.fields.SynchronousModeStrict,
				ClusterConfigurations: tt.fields.ClusterConfigurations,
				PgRestoreFrom:         tt.fields.PgRestoreFrom,
				DataCentres:           tt.fields.DataCentres,
				ResizeSettings:        tt.fields.ResizeSettings,
				Extensions:            tt.fields.Extensions,
			}
			if err := pgs.validateImmutableDCsFieldsUpdate(tt.args.oldSpec); (err != nil) != tt.wantErr {
				t.Errorf("validateImmutableDCsFieldsUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPgDataCentre_validateInterDCImmutableFields(t *testing.T) {
	type fields struct {
		GenericDataCentreSpec      GenericDataCentreSpec
		ClientEncryption           bool
		NodeSize                   string
		NodesNumber                int
		InterDataCentreReplication []*InterDataCentreReplication
		IntraDataCentreReplication []*IntraDataCentreReplication
		PGBouncer                  []*PgBouncer
	}
	type args struct {
		oldInterDC []*InterDataCentreReplication
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "change len of InterDataCentreReplication",
			fields: fields{
				InterDataCentreReplication: []*InterDataCentreReplication{{}, {}},
			},
			args: args{
				oldInterDC: []*InterDataCentreReplication{{}},
			},
			wantErr: true,
		},
		{
			name: "change field of InterDataCentreReplication",
			fields: fields{
				InterDataCentreReplication: []*InterDataCentreReplication{{
					IsPrimaryDataCentre: false,
				}},
			},
			args: args{
				oldInterDC: []*InterDataCentreReplication{{
					IsPrimaryDataCentre: true,
				}},
			},
			wantErr: true,
		},
		{
			name: "valid case",
			fields: fields{
				InterDataCentreReplication: []*InterDataCentreReplication{{
					IsPrimaryDataCentre: false,
				}},
			},
			args: args{
				oldInterDC: []*InterDataCentreReplication{{
					IsPrimaryDataCentre: false,
				}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pdc := &PgDataCentre{
				GenericDataCentreSpec:      tt.fields.GenericDataCentreSpec,
				ClientEncryption:           tt.fields.ClientEncryption,
				NodeSize:                   tt.fields.NodeSize,
				NodesNumber:                tt.fields.NodesNumber,
				InterDataCentreReplication: tt.fields.InterDataCentreReplication,
				IntraDataCentreReplication: tt.fields.IntraDataCentreReplication,
				PGBouncer:                  tt.fields.PGBouncer,
			}
			if err := pdc.validateInterDCImmutableFields(tt.args.oldInterDC); (err != nil) != tt.wantErr {
				t.Errorf("validateInterDCImmutableFields() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPgDataCentre_validateIntraDCImmutableFields(t *testing.T) {
	type fields struct {
		GenericDataCentreSpec      GenericDataCentreSpec
		ClientEncryption           bool
		NodeSize                   string
		NodesNumber                int
		InterDataCentreReplication []*InterDataCentreReplication
		IntraDataCentreReplication []*IntraDataCentreReplication
		PGBouncer                  []*PgBouncer
	}
	type args struct {
		oldIntraDC []*IntraDataCentreReplication
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "change len of IntraDataCentreReplication",
			fields: fields{
				IntraDataCentreReplication: []*IntraDataCentreReplication{{}, {}},
			},
			args: args{
				oldIntraDC: []*IntraDataCentreReplication{{}},
			},
			wantErr: true,
		},
		{
			name: "change field of IntraDataCentreReplication",
			fields: fields{
				IntraDataCentreReplication: []*IntraDataCentreReplication{{
					ReplicationMode: "new",
				}},
			},
			args: args{
				oldIntraDC: []*IntraDataCentreReplication{{
					ReplicationMode: "old",
				}},
			},
			wantErr: true,
		},
		{
			name: "valid case",
			fields: fields{
				IntraDataCentreReplication: []*IntraDataCentreReplication{{
					ReplicationMode: "old",
				}},
			},
			args: args{
				oldIntraDC: []*IntraDataCentreReplication{{
					ReplicationMode: "old",
				}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pdc := &PgDataCentre{
				GenericDataCentreSpec:      tt.fields.GenericDataCentreSpec,
				ClientEncryption:           tt.fields.ClientEncryption,
				NodeSize:                   tt.fields.NodeSize,
				NodesNumber:                tt.fields.NodesNumber,
				InterDataCentreReplication: tt.fields.InterDataCentreReplication,
				IntraDataCentreReplication: tt.fields.IntraDataCentreReplication,
				PGBouncer:                  tt.fields.PGBouncer,
			}
			if err := pdc.validateIntraDCImmutableFields(tt.args.oldIntraDC); (err != nil) != tt.wantErr {
				t.Errorf("validateIntraDCImmutableFields() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
