package clusters

import (
	"strings"
	"testing"

	"github.com/instaclustr/operator/apis/clusters/v1alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr/api/v1/convertors"
	"github.com/instaclustr/operator/pkg/models"
)

// Unit tests
func TestPGSpecHasRestore(t *testing.T) {
	minimalPGRestoreSpec := v1alpha1.PgSpec{
		PgRestoreFrom: &v1alpha1.PgRestoreFrom{
			ClusterID: "clusterToRestore",
		},
	}

	minimalPGSpec := v1alpha1.PgSpec{
		Cluster: v1alpha1.Cluster{
			Name:    "unitPGTest",
			Version: "pgVersion",
		},
		DataCentres: []*v1alpha1.PgDataCentre{
			{
				DataCentre: v1alpha1.DataCentre{
					Region:        "region",
					CloudProvider: "provider",
					Network:       "network",
					NodeSize:      "nodeSize",
					RacksNumber:   3,
					NodesNumber:   1,
				},
				PostgresqlNodeCount: 3,
			},
		},
	}

	bothCasesFilledPGSpec := v1alpha1.PgSpec{
		PgRestoreFrom: &v1alpha1.PgRestoreFrom{
			ClusterID: "clusterToRestore",
		},
		Cluster: v1alpha1.Cluster{
			Name:    "unitPGTest",
			Version: "pgVersion",
		},
		DataCentres: []*v1alpha1.PgDataCentre{
			{
				DataCentre: v1alpha1.DataCentre{
					Region:        "region",
					CloudProvider: "provider",
					Network:       "network",
					NodeSize:      "nodeSize",
					RacksNumber:   3,
					NodesNumber:   1,
				},
				PostgresqlNodeCount: 3,
			},
		},
	}

	bothCasesFilledRestoreIncorrectPGSpec := v1alpha1.PgSpec{
		PgRestoreFrom: &v1alpha1.PgRestoreFrom{
			ClusterNameOverride: "restoredPGcluster",
		},
		Cluster: v1alpha1.Cluster{
			Name:    "unitPGTest",
			Version: "pgVersion",
		},
		DataCentres: []*v1alpha1.PgDataCentre{
			{
				DataCentre: v1alpha1.DataCentre{
					Region:        "region",
					CloudProvider: "provider",
					Network:       "network",
					NodeSize:      "nodeSize",
					RacksNumber:   3,
					NodesNumber:   1,
				},
				PostgresqlNodeCount: 3,
			},
		},
	}

	incorrectRestoreFilledPGSpec := v1alpha1.PgSpec{
		PgRestoreFrom: &v1alpha1.PgRestoreFrom{
			ClusterNameOverride: "restoredPGcluster",
		},
	}

	tests := []struct {
		name     string
		param    v1alpha1.PgSpec
		expected bool
	}{
		{
			name:     "Check with minimal restore spec",
			param:    minimalPGRestoreSpec,
			expected: true,
		},
		{
			name:     "Check with minimal standart spec",
			param:    minimalPGSpec,
			expected: false,
		},
		{
			name:     "Check with both cases filled",
			param:    bothCasesFilledPGSpec,
			expected: true,
		},
		{
			name:     "Check with both cases filled but restore fields filled incorrectly",
			param:    bothCasesFilledRestoreIncorrectPGSpec,
			expected: false,
		},
		{
			name:     "Check with restore fields filled incorrectly",
			param:    incorrectRestoreFilledPGSpec,
			expected: false,
		},
	}

	err := 0
	for _, uTest := range tests {
		if result := uTest.param.HasRestore(); result != uTest.expected {
			t.Errorf("Result %v not equal to the expected result %v\nTest: %s\nParameter: %+v",
				result, uTest.expected, strings.ToLower(uTest.name), uTest.param)
			err++
		}
	}

	t.Logf("%d%% tests passed with success (%d/%d)", (len(tests)-err)*100/len(tests), len(tests)-err, len(tests))
}

func TestPGToInstAPIv1(t *testing.T) {
	minimalPGSpec := &v1alpha1.PgSpec{
		Cluster: v1alpha1.Cluster{
			Name:    "unitPGTest",
			Version: "pgVersion",
		},
		DataCentres: []*v1alpha1.PgDataCentre{
			{
				DataCentre: v1alpha1.DataCentre{
					Region:        "region",
					CloudProvider: "provider",
					Network:       "network",
					NodeSize:      "nodeSize",
					RacksNumber:   3,
					NodesNumber:   1,
				},
				PostgresqlNodeCount: 3,
			},
		},
	}
	fullyFilledPGSpec := &v1alpha1.PgSpec{
		Cluster: v1alpha1.Cluster{
			Name:                  "unitPGTest",
			Version:               "pgVersion",
			PCICompliance:         true,
			PrivateNetworkCluster: true,
			SLATier:               "slaTier",
			TwoFactorDelete: []*v1alpha1.TwoFactorDelete{
				{
					Email: "Email",
					Phone: "Phone",
				},
			},
		},
		PGBouncerVersion: "version",
		DataCentres: []*v1alpha1.PgDataCentre{
			{
				DataCentre: v1alpha1.DataCentre{
					Name:                "testDC",
					Region:              "region",
					CloudProvider:       "provider",
					ProviderAccountName: "instaclustr",
					CloudProviderSettings: []*v1alpha1.CloudProviderSettings{
						{
							CustomVirtualNetworkID: "id",
							ResourceGroup:          "rg",
							DiskEncryptionKey:      "key",
						},
					},
					Network:     "network",
					NodeSize:    "nodeSize",
					RacksNumber: 3,
					NodesNumber: 1,
					Tags:        map[string]string{"testTag": "test"},
				},
				ClientEncryption:      true,
				PostgresqlNodeCount:   3,
				ReplicationMode:       "mode",
				SynchronousModeStrict: true,
				PoolMode:              "mode",
			},
		},
		ClusterConfigurations: map[string]string{"testConfig": "testValue"},
		Description:           "testCluster",
	}
	noDCsPGSpec := &v1alpha1.PgSpec{
		Cluster: v1alpha1.Cluster{
			Name:    "unitPGTest",
			Version: "pgVersion",
		},
	}

	tests := []struct {
		name     string
		param    *v1alpha1.PgSpec
		expected error
	}{
		{
			name:     "Convert with minimal correct spec",
			param:    minimalPGSpec,
			expected: nil,
		},
		{
			name:     "Convert with fully filled spec",
			param:    fullyFilledPGSpec,
			expected: nil,
		},
		{
			name:     "Convert with data centres not filled",
			param:    noDCsPGSpec,
			expected: models.ZeroDataCentres,
		},
	}

	errCount := 0
	for _, uTest := range tests {
		if _, err := convertors.PgToInstAPI(uTest.param); err != uTest.expected {
			t.Errorf("Result %v not equal to the expected result %v\nTest: %s\nParameter: %+v",
				err, uTest.expected, strings.ToLower(uTest.name), uTest.param)
			errCount++
		}
	}

	t.Logf("%d%% tests passed with success (%d/%d)", (len(tests)-errCount)*100/len(tests), len(tests)-errCount, len(tests))
}

// TODO Integration tests
