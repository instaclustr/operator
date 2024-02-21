package v1beta1

import (
	"testing"

	"github.com/instaclustr/operator/pkg/models"
)

func TestCluster_ValidateCreation(t *testing.T) {
	type fields struct {
		Name                  string
		Version               string
		PCICompliance         bool
		PrivateNetworkCluster bool
		SLATier               string
		TwoFactorDelete       []*TwoFactorDelete
		Description           string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "empty cluster name",
			fields: fields{
				Name: "",
			},
			wantErr: true,
		},
		{
			name: "more than one two factor delete",
			fields: fields{
				Name: "test",
				TwoFactorDelete: []*TwoFactorDelete{
					{
						Email: "test@mail.com",
						Phone: "12345",
					}, {
						Email: "test@mail.com",
						Phone: "12345",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "unsupported SLAtier",
			fields: fields{
				Name:    "test",
				SLATier: "test",
			},
			wantErr: true,
		},
		{
			name: "valid cluster",
			fields: fields{
				Name:    "test",
				SLATier: "NON_PRODUCTION",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cluster{
				Name:                  tt.fields.Name,
				Version:               tt.fields.Version,
				PCICompliance:         tt.fields.PCICompliance,
				PrivateNetworkCluster: tt.fields.PrivateNetworkCluster,
				SLATier:               tt.fields.SLATier,
				TwoFactorDelete:       tt.fields.TwoFactorDelete,
				Description:           tt.fields.Description,
			}
			if err := c.ValidateCreation(); (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreation() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDataCentre_ValidateCreation(t *testing.T) {
	type fields struct {
		Name                  string
		Region                string
		CloudProvider         string
		ProviderAccountName   string
		CloudProviderSettings []*CloudProviderSettings
		Network               string
		NodeSize              string
		NodesNumber           int
		Tags                  map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "unavailable cloud provider",
			fields: fields{
				CloudProvider: "some unavailable cloud provider",
			},
			wantErr: true,
		},
		{
			name: "unavailable region for AWS cloud provider",
			fields: fields{
				Region:        models.AzureRegions[0],
				CloudProvider: models.AWSVPC,
			},
			wantErr: true,
		},
		{
			name: "unavailable region for Azure cloud provider",
			fields: fields{
				Region:        models.AWSRegions[0],
				CloudProvider: models.AZUREAZ,
			},
			wantErr: true,
		},
		{
			name: "unavailable region for GCP cloud provider",
			fields: fields{
				Region:        models.AzureRegions[0],
				CloudProvider: models.GCP,
			},
			wantErr: true,
		},
		{
			name: "unavailable region for ONPREMISES cloud provider",
			fields: fields{
				Region:        models.AzureRegions[0],
				CloudProvider: models.ONPREMISES,
			},
			wantErr: true,
		},
		{
			name: "cloud provider settings on not RIYOA account",
			fields: fields{
				Region:                models.AWSRegions[0],
				CloudProvider:         models.AWSVPC,
				ProviderAccountName:   models.DefaultAccountName,
				CloudProviderSettings: []*CloudProviderSettings{{}},
			},
			wantErr: true,
		},
		{
			name: "more than one cloud provider settings",
			fields: fields{
				Region:                models.AWSRegions[0],
				CloudProvider:         models.AWSVPC,
				ProviderAccountName:   "custom",
				CloudProviderSettings: []*CloudProviderSettings{{}, {}},
			},
			wantErr: true,
		},
		{
			name: "invalid network",
			fields: fields{
				Region:        models.AWSRegions[0],
				CloudProvider: models.AWSVPC,
				Network:       "test",
			},
			wantErr: true,
		},
		{
			name: "valid DC",
			fields: fields{
				Region:        models.AWSRegions[0],
				CloudProvider: models.AWSVPC,
				Network:       "172.16.0.0/19",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dc := &DataCentre{
				Name:                  tt.fields.Name,
				Region:                tt.fields.Region,
				CloudProvider:         tt.fields.CloudProvider,
				ProviderAccountName:   tt.fields.ProviderAccountName,
				CloudProviderSettings: tt.fields.CloudProviderSettings,
				Network:               tt.fields.Network,
				NodeSize:              tt.fields.NodeSize,
				NodesNumber:           tt.fields.NodesNumber,
				Tags:                  tt.fields.Tags,
			}
			if err := dc.ValidateCreation(); (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreation() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCloudProviderSettings_ValidateCreation(t *testing.T) {
	type fields struct {
		CustomVirtualNetworkID    string
		ResourceGroup             string
		DiskEncryptionKey         string
		BackupBucket              string
		DisableSnapshotAutoExpiry bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "filled resource group & disk encryption key",
			fields: fields{
				ResourceGroup:     "test",
				DiskEncryptionKey: "test",
			},
			wantErr: true,
		},
		{
			name: "filled resource group & custom virtual network ID",
			fields: fields{
				CustomVirtualNetworkID: "test",
				ResourceGroup:          "test",
			},
			wantErr: true,
		},
		{
			name: "valid cloud provider settings",
			fields: fields{
				ResourceGroup: "test",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cps := &CloudProviderSettings{
				CustomVirtualNetworkID:    tt.fields.CustomVirtualNetworkID,
				ResourceGroup:             tt.fields.ResourceGroup,
				DiskEncryptionKey:         tt.fields.DiskEncryptionKey,
				BackupBucket:              tt.fields.BackupBucket,
				DisableSnapshotAutoExpiry: tt.fields.DisableSnapshotAutoExpiry,
			}
			if err := cps.ValidateCreation(); (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreation() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validateReplicationFactor(t *testing.T) {
	type args struct {
		availableReplicationFactors []int
		rf                          int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "invalid replication factor",
			args: args{
				availableReplicationFactors: []int{2, 3, 5},
				rf:                          4,
			},
			wantErr: true,
		},
		{
			name: "valid replication factor",
			args: args{
				availableReplicationFactors: []int{2, 3, 5},
				rf:                          2,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateReplicationFactor(tt.args.availableReplicationFactors, tt.args.rf); (err != nil) != tt.wantErr {
				t.Errorf("validateReplicationFactor() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validateAppVersion(t *testing.T) {
	type args struct {
		versions []*models.AppVersions
		appType  string
		version  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "invalid appVersion",
			args: args{
				versions: []*models.AppVersions{{
					Application: "test",
					Versions:    []string{"1.0.0", "2.2.2"},
				}, {
					Application: "different",
					Versions:    []string{"1.0.0", "2.2.2"},
				}},
				appType: "different",
				version: "3.3.3",
			},
			wantErr: true,
		},
		{
			name: "valid appVersion",
			args: args{
				versions: []*models.AppVersions{{
					Application: "test",
					Versions:    []string{"1.0.0", "2.2.2"},
				}, {
					Application: "different",
					Versions:    []string{"1.0.0", "2.2.2"},
				}},
				appType: "different",
				version: "2.2.2",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateAppVersion(tt.args.versions, tt.args.appType, tt.args.version); (err != nil) != tt.wantErr {
				t.Errorf("validateAppVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validateTwoFactorDelete(t *testing.T) {
	type args struct {
		new []*TwoFactorDelete
		old []*TwoFactorDelete
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "more than one twoFactorDelete",
			args: args{
				new: []*TwoFactorDelete{{}, {}},
				old: nil,
			},
			wantErr: true,
		},
		{
			name: "change twoFactorDelete len",
			args: args{
				new: []*TwoFactorDelete{{}},
				old: []*TwoFactorDelete{{}, {}},
			},
			wantErr: true,
		},
		{
			name: "change twoFactorDelete field",
			args: args{
				new: []*TwoFactorDelete{{
					Email: "new@mail.com",
					Phone: "12345",
				}},
				old: []*TwoFactorDelete{{
					Email: "test@mail.com",
					Phone: "12345",
				}},
			},
			wantErr: true,
		},
		{
			name: "valid twoFactorDelete",
			args: args{
				new: []*TwoFactorDelete{{
					Email: "test@mail.com",
					Phone: "12345",
				}},
				old: []*TwoFactorDelete{{
					Email: "test@mail.com",
					Phone: "12345",
				}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateTwoFactorDelete(tt.args.new, tt.args.old); (err != nil) != tt.wantErr {
				t.Errorf("validateTwoFactorDelete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validateIngestNodes(t *testing.T) {
	type args struct {
		new []*OpenSearchIngestNodes
		old []*OpenSearchIngestNodes
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "change len of the IngestNodes",
			args: args{
				new: []*OpenSearchIngestNodes{{}},
				old: []*OpenSearchIngestNodes{},
			},
			wantErr: true,
		},
		{
			name: "change field of the IngestNode",
			args: args{
				new: []*OpenSearchIngestNodes{{
					NodeSize:  "new",
					NodeCount: 1,
				}},
				old: []*OpenSearchIngestNodes{{
					NodeSize:  "test",
					NodeCount: 1,
				}},
			},
			wantErr: true,
		},
		{
			name: "unchanged IngestNodes",
			args: args{
				new: []*OpenSearchIngestNodes{{
					NodeSize:  "test",
					NodeCount: 1,
				}},
				old: []*OpenSearchIngestNodes{{
					NodeSize:  "test",
					NodeCount: 1,
				}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateIngestNodes(tt.args.new, tt.args.old); (err != nil) != tt.wantErr {
				t.Errorf("validateIngestNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validateClusterManagedNodes(t *testing.T) {
	type args struct {
		new []*ClusterManagerNodes
		old []*ClusterManagerNodes
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "change len of the ClusterManagerNodes",
			args: args{
				new: []*ClusterManagerNodes{{}},
				old: []*ClusterManagerNodes{},
			},
			wantErr: true,
		},
		{
			name: "change field of the ClusterManagerNode",
			args: args{
				new: []*ClusterManagerNodes{{
					NodeSize:         "new",
					DedicatedManager: false,
				}},
				old: []*ClusterManagerNodes{{
					NodeSize:         "test",
					DedicatedManager: false,
				}},
			},
			wantErr: true,
		},
		{
			name: "unchanged ClusterManagerNodes",
			args: args{
				new: []*ClusterManagerNodes{{
					NodeSize:         "test",
					DedicatedManager: false,
				}},
				old: []*ClusterManagerNodes{{
					NodeSize:         "test",
					DedicatedManager: false,
				}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateClusterManagedNodes(tt.args.new, tt.args.old); (err != nil) != tt.wantErr {
				t.Errorf("validateClusterManagedNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validateTagsUpdate(t *testing.T) {
	type args struct {
		new map[string]string
		old map[string]string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "change len of the Tags",
			args: args{
				new: map[string]string{"": ""},
				old: map[string]string{},
			},
			wantErr: true,
		},
		{
			name: "change tag",
			args: args{
				new: map[string]string{"tag1": "new"},
				old: map[string]string{"tag1": "test"},
			},
			wantErr: true,
		},
		{
			name: "unchanged tags",
			args: args{
				new: map[string]string{"tag1": "test"},
				old: map[string]string{"tag1": "test"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateTagsUpdate(tt.args.new, tt.args.old); (err != nil) != tt.wantErr {
				t.Errorf("validateTagsUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validatePrivateLinkUpdate(t *testing.T) {
	type args struct {
		new []*PrivateLink
		old []*PrivateLink
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "change len of the PrivateLink",
			args: args{
				new: []*PrivateLink{{}, {}},
				old: []*PrivateLink{{}},
			},
			wantErr: true,
		},
		{
			name: "change field of the PrivateLink element",
			args: args{
				new: []*PrivateLink{{
					AdvertisedHostname: "new",
				}},
				old: []*PrivateLink{{
					AdvertisedHostname: "old",
				}},
			},
			wantErr: true,
		},
		{
			name: "unchanged PrivateLink",
			args: args{
				new: []*PrivateLink{{
					AdvertisedHostname: "test",
				}},
				old: []*PrivateLink{{
					AdvertisedHostname: "test",
				}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validatePrivateLinkUpdate(tt.args.new, tt.args.old); (err != nil) != tt.wantErr {
				t.Errorf("validatePrivateLinkUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validateSingleConcurrentResize(t *testing.T) {
	type args struct {
		concurrentResizes int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "invalid SingleConcurrentResize",
			args: args{
				concurrentResizes: 2,
			},
			wantErr: true,
		},
		{
			name: "valid SingleConcurrentResize",
			args: args{
				concurrentResizes: 1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateSingleConcurrentResize(tt.args.concurrentResizes); (err != nil) != tt.wantErr {
				t.Errorf("validateSingleConcurrentResize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenericDataCentreSpec_validateImmutableCloudProviderSettingsUpdate(t *testing.T) {
	type fields struct {
		Name                string
		Region              string
		CloudProvider       string
		ProviderAccountName string
		Network             string
		Tags                map[string]string
		AWSSettings         []*AWSSettings
		GCPSettings         []*GCPSettings
		AzureSettings       []*AzureSettings
	}
	type args struct {
		old *GenericDataCentreSpec
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "change AWSSettings",
			fields: fields{
				AWSSettings: []*AWSSettings{{
					DiskEncryptionKey:      "new",
					CustomVirtualNetworkID: "new",
					BackupBucket:           "new",
				}},
				GCPSettings:   nil,
				AzureSettings: nil,
			},
			args: args{
				old: &GenericDataCentreSpec{
					AWSSettings: []*AWSSettings{{
						DiskEncryptionKey:      "test",
						CustomVirtualNetworkID: "test",
						BackupBucket:           "test",
					}},
					GCPSettings:   nil,
					AzureSettings: nil,
				},
			},
			wantErr: true,
		},
		{
			name: "change GCPSettings",
			fields: fields{
				AWSSettings: nil,
				GCPSettings: []*GCPSettings{{
					CustomVirtualNetworkID:    "new",
					DisableSnapshotAutoExpiry: false,
				}},
				AzureSettings: nil,
			},
			args: args{
				old: &GenericDataCentreSpec{
					AWSSettings: nil,
					GCPSettings: []*GCPSettings{{
						CustomVirtualNetworkID:    "test",
						DisableSnapshotAutoExpiry: false,
					}},
					AzureSettings: nil,
				},
			},
			wantErr: true,
		},
		{
			name: "change AzureSettings",
			fields: fields{
				AWSSettings: nil,
				GCPSettings: nil,
				AzureSettings: []*AzureSettings{
					{
						ResourceGroup:          "new",
						CustomVirtualNetworkID: "new",
						StorageNetwork:         "new",
					},
				},
			},
			args: args{
				old: &GenericDataCentreSpec{
					AWSSettings: nil,
					GCPSettings: nil,
					AzureSettings: []*AzureSettings{
						{
							ResourceGroup:          "test",
							CustomVirtualNetworkID: "test",
							StorageNetwork:         "test",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "unchanged GCPSettings",
			fields: fields{
				AWSSettings: nil,
				GCPSettings: []*GCPSettings{{
					CustomVirtualNetworkID:    "test",
					DisableSnapshotAutoExpiry: false,
				}},
				AzureSettings: nil,
			},
			args: args{
				old: &GenericDataCentreSpec{
					AWSSettings: nil,
					GCPSettings: []*GCPSettings{{
						CustomVirtualNetworkID:    "test",
						DisableSnapshotAutoExpiry: false,
					}},
					AzureSettings: nil,
				},
			},
			wantErr: false,
		},
		{
			name: "unchanged AzureSettings",
			fields: fields{
				AWSSettings: nil,
				GCPSettings: nil,
				AzureSettings: []*AzureSettings{
					{
						ResourceGroup:          "test",
						CustomVirtualNetworkID: "test",
						StorageNetwork:         "test",
					},
				},
			},
			args: args{
				old: &GenericDataCentreSpec{
					AWSSettings: nil,
					GCPSettings: nil,
					AzureSettings: []*AzureSettings{
						{
							ResourceGroup:          "test",
							CustomVirtualNetworkID: "test",
							StorageNetwork:         "test",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "unchanged AWSSettings",
			fields: fields{
				AWSSettings: []*AWSSettings{{
					DiskEncryptionKey:      "test",
					CustomVirtualNetworkID: "test",
					BackupBucket:           "test",
				}},
				GCPSettings:   nil,
				AzureSettings: nil,
			},
			args: args{
				old: &GenericDataCentreSpec{
					AWSSettings: []*AWSSettings{{
						DiskEncryptionKey:      "test",
						CustomVirtualNetworkID: "test",
						BackupBucket:           "test",
					}},
					GCPSettings:   nil,
					AzureSettings: nil,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GenericDataCentreSpec{
				Name:                tt.fields.Name,
				Region:              tt.fields.Region,
				CloudProvider:       tt.fields.CloudProvider,
				ProviderAccountName: tt.fields.ProviderAccountName,
				Network:             tt.fields.Network,
				Tags:                tt.fields.Tags,
				AWSSettings:         tt.fields.AWSSettings,
				GCPSettings:         tt.fields.GCPSettings,
				AzureSettings:       tt.fields.AzureSettings,
			}
			if err := s.validateImmutableCloudProviderSettingsUpdate(tt.args.old); (err != nil) != tt.wantErr {
				t.Errorf("validateImmutableCloudProviderSettingsUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenericDataCentreSpec_validateCloudProviderSettings(t *testing.T) {
	type fields struct {
		Name                string
		Region              string
		CloudProvider       string
		ProviderAccountName string
		Network             string
		Tags                map[string]string
		AWSSettings         []*AWSSettings
		GCPSettings         []*GCPSettings
		AzureSettings       []*AzureSettings
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "to many CloudProviderSettings",
			fields: fields{
				AWSSettings:   []*AWSSettings{{}},
				GCPSettings:   []*GCPSettings{{}},
				AzureSettings: nil,
			},
			wantErr: true,
		},
		{
			name: "valid CloudProviderSettings",
			fields: fields{
				AWSSettings:   []*AWSSettings{{}},
				GCPSettings:   nil,
				AzureSettings: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GenericDataCentreSpec{
				Name:                tt.fields.Name,
				Region:              tt.fields.Region,
				CloudProvider:       tt.fields.CloudProvider,
				ProviderAccountName: tt.fields.ProviderAccountName,
				Network:             tt.fields.Network,
				Tags:                tt.fields.Tags,
				AWSSettings:         tt.fields.AWSSettings,
				GCPSettings:         tt.fields.GCPSettings,
				AzureSettings:       tt.fields.AzureSettings,
			}
			if err := s.validateCloudProviderSettings(); (err != nil) != tt.wantErr {
				t.Errorf("validateCloudProviderSettings() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenericDataCentreSpec_hasCloudProviderSettings(t *testing.T) {
	type fields struct {
		Name                string
		Region              string
		CloudProvider       string
		ProviderAccountName string
		Network             string
		Tags                map[string]string
		AWSSettings         []*AWSSettings
		GCPSettings         []*GCPSettings
		AzureSettings       []*AzureSettings
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "no CloudProviderSettings",
			fields: fields{
				AWSSettings:   nil,
				GCPSettings:   nil,
				AzureSettings: nil,
			},
			want: false,
		},
		{
			name: "has CloudProviderSettings",
			fields: fields{
				AWSSettings:   []*AWSSettings{},
				GCPSettings:   nil,
				AzureSettings: nil,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GenericDataCentreSpec{
				Name:                tt.fields.Name,
				Region:              tt.fields.Region,
				CloudProvider:       tt.fields.CloudProvider,
				ProviderAccountName: tt.fields.ProviderAccountName,
				Network:             tt.fields.Network,
				Tags:                tt.fields.Tags,
				AWSSettings:         tt.fields.AWSSettings,
				GCPSettings:         tt.fields.GCPSettings,
				AzureSettings:       tt.fields.AzureSettings,
			}
			if got := s.hasCloudProviderSettings(); got != tt.want {
				t.Errorf("hasCloudProviderSettings() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDataCentre_validateImmutableCloudProviderSettingsUpdate(t *testing.T) {
	type fields struct {
		Name                  string
		Region                string
		CloudProvider         string
		ProviderAccountName   string
		CloudProviderSettings []*CloudProviderSettings
		Network               string
		NodeSize              string
		NodesNumber           int
		Tags                  map[string]string
	}
	type args struct {
		oldSettings []*CloudProviderSettings
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "different len of the CloudProviderSettings",
			fields: fields{
				CloudProviderSettings: []*CloudProviderSettings{{}, {}},
			},
			args:    args{oldSettings: []*CloudProviderSettings{{}}},
			wantErr: true,
		},
		{
			name: "different CloudProviderSettings",
			fields: fields{CloudProviderSettings: []*CloudProviderSettings{{
				CustomVirtualNetworkID: "new",
			}}},
			args: args{oldSettings: []*CloudProviderSettings{{
				CustomVirtualNetworkID: "test",
			}}},
			wantErr: true,
		},
		{
			name: "unchanged CloudProviderSettings",
			fields: fields{CloudProviderSettings: []*CloudProviderSettings{{
				CustomVirtualNetworkID: "test",
			}}},
			args: args{oldSettings: []*CloudProviderSettings{{
				CustomVirtualNetworkID: "test",
			}}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dc := &DataCentre{
				Name:                  tt.fields.Name,
				Region:                tt.fields.Region,
				CloudProvider:         tt.fields.CloudProvider,
				ProviderAccountName:   tt.fields.ProviderAccountName,
				CloudProviderSettings: tt.fields.CloudProviderSettings,
				Network:               tt.fields.Network,
				NodeSize:              tt.fields.NodeSize,
				NodesNumber:           tt.fields.NodesNumber,
				Tags:                  tt.fields.Tags,
			}
			if err := dc.validateImmutableCloudProviderSettingsUpdate(tt.args.oldSettings); (err != nil) != tt.wantErr {
				t.Errorf("validateImmutableCloudProviderSettingsUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenericClusterSpec_ValidateCreation(t *testing.T) {
	type fields struct {
		Name            string
		Version         string
		PrivateNetwork  bool
		SLATier         string
		Description     string
		TwoFactorDelete []*TwoFactorDelete
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "empty cluster name",
			fields: fields{
				Name: "",
			},
			wantErr: true,
		},
		{
			name: "more than one two factor delete",
			fields: fields{
				Name: "test",
				TwoFactorDelete: []*TwoFactorDelete{
					{
						Email: "test@mail.com",
						Phone: "12345",
					}, {
						Email: "test@mail.com",
						Phone: "12345",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "unsupported SLAtier",
			fields: fields{
				Name:    "test",
				SLATier: "test",
			},
			wantErr: true,
		},
		{
			name: "valid cluster",
			fields: fields{
				Name:    "test",
				SLATier: "NON_PRODUCTION",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GenericClusterSpec{
				Name:            tt.fields.Name,
				Version:         tt.fields.Version,
				PrivateNetwork:  tt.fields.PrivateNetwork,
				SLATier:         tt.fields.SLATier,
				Description:     tt.fields.Description,
				TwoFactorDelete: tt.fields.TwoFactorDelete,
			}
			if err := s.ValidateCreation(); (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreation() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenericDataCentreSpec_validateCreation(t *testing.T) {
	type fields struct {
		Name                string
		Region              string
		CloudProvider       string
		ProviderAccountName string
		Network             string
		Tags                map[string]string
		AWSSettings         []*AWSSettings
		GCPSettings         []*GCPSettings
		AzureSettings       []*AzureSettings
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "unavailable cloud provider",
			fields: fields{
				CloudProvider: "some unavailable cloud provider",
			},
			wantErr: true,
		},
		{
			name: "unavailable region for AWS cloud provider",
			fields: fields{
				Region:        models.AzureRegions[0],
				CloudProvider: models.AWSVPC,
			},
			wantErr: true,
		},
		{
			name: "unavailable region for Azure cloud provider",
			fields: fields{
				Region:        models.AWSRegions[0],
				CloudProvider: models.AZUREAZ,
			},
			wantErr: true,
		},
		{
			name: "unavailable region for GCP cloud provider",
			fields: fields{
				Region:        models.AzureRegions[0],
				CloudProvider: models.GCP,
			},
			wantErr: true,
		},
		{
			name: "unavailable region for ONPREMISES cloud provider",
			fields: fields{
				Region:        models.AzureRegions[0],
				CloudProvider: models.ONPREMISES,
			},
			wantErr: true,
		},
		{
			name: "cloud provider settings on not RIYOA account",
			fields: fields{
				Region:              models.AWSRegions[0],
				CloudProvider:       models.AWSVPC,
				ProviderAccountName: models.DefaultAccountName,
				AWSSettings:         []*AWSSettings{},
			},
			wantErr: true,
		},
		{
			name: "more than one cloud provider settings",
			fields: fields{
				Name:                "",
				Region:              models.AWSRegions[0],
				CloudProvider:       models.AWSVPC,
				ProviderAccountName: "custom",
				AWSSettings:         []*AWSSettings{{}},
				GCPSettings:         []*GCPSettings{{}},
			},
			wantErr: true,
		},
		{
			name: "invalid network",
			fields: fields{
				Region:        models.AWSRegions[0],
				CloudProvider: models.AWSVPC,
				Network:       "test",
			},
			wantErr: true,
		},
		{
			name: "valid DC",
			fields: fields{
				Name:                "",
				Region:              models.AWSRegions[0],
				CloudProvider:       models.AWSVPC,
				ProviderAccountName: "",
				Network:             "172.16.0.0/19",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &GenericDataCentreSpec{
				Name:                tt.fields.Name,
				Region:              tt.fields.Region,
				CloudProvider:       tt.fields.CloudProvider,
				ProviderAccountName: tt.fields.ProviderAccountName,
				Network:             tt.fields.Network,
				Tags:                tt.fields.Tags,
				AWSSettings:         tt.fields.AWSSettings,
				GCPSettings:         tt.fields.GCPSettings,
				AzureSettings:       tt.fields.AzureSettings,
			}
			if err := s.validateCreation(); (err != nil) != tt.wantErr {
				t.Errorf("validateCreation() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
