package models

import "github.com/instaclustr/operator/pkg/models"

const (
	Redis            = "REDIS"
	RedisNodePurpose = "REDIS"
)

const (
	// AWS node types
	t3_small_20_r     = "t3.small-20-r"
	t3_medium_80_r    = "t3.medium-80-r"
	r6g_large_100_r   = "r6g.large-100-r"
	r5_large_100_r    = "r5.large-100-r"
	r6g_xlarge_200_r  = "r6g.xlarge-200-r"
	r5_xlarge_200_r   = "r5.xlarge-200-r"
	r6g_2xlarge_400_r = "r6g.2xlarge-400-r"
	r5_2xlarge_400_r  = "r5.2xlarge-400-r"
	r6g_4xlarge_600_r = "r6g.4xlarge-600-r"
	r5_4xlarge_600_r  = "r5.4xlarge-600-r"
	r5_8xlarge_1000_r = "r5.8xlarge-1000-r"
)

const (
	// Azure node types
	std_ds2_v2_64_r   = "Standard_DS2_v2-64-r"
	std_e2s_v3_100_r  = "Standard_E2s_v3-100-r"
	std_e4s_v3_200_r  = "Standard_E4s_v3-200-r"
	std_e8s_v3_400_r  = "Standard_E8s_v3-400-r"
	std_e16s_v3_800_r = "Standard_E16s_v3-800-r"
)

const (
	// GCP node types
	n1_std_1_30_r       = "n1-standard-1-30-r"
	n1_highmem_2_100_r  = "n1-highmem-2-100-r"
	n1_highmem_4_200_r  = "n1-highmem-4-200-r"
	n1_highmem_8_400_r  = "n1-highmem-8-400-r"
	n1_highmem_16_600_r = "n1-highmem-16-600-r"
)

var RedisAWSNodeTypes = map[string]int{
	t3_small_20_r:     1,
	t3_medium_80_r:    2,
	r6g_large_100_r:   3,
	r5_large_100_r:    4,
	r6g_xlarge_200_r:  5,
	r5_xlarge_200_r:   6,
	r6g_2xlarge_400_r: 7,
	r5_2xlarge_400_r:  8,
	r6g_4xlarge_600_r: 9,
	r5_4xlarge_600_r:  10,
	r5_8xlarge_1000_r: 11,
}
var RedisAzureNodeTypes = map[string]int{
	std_ds2_v2_64_r:   1,
	std_e2s_v3_100_r:  2,
	std_e4s_v3_200_r:  3,
	std_e8s_v3_400_r:  4,
	std_e16s_v3_800_r: 5,
}
var RedisGCPNodeTypes = map[string]int{
	n1_std_1_30_r:       1,
	n1_highmem_2_100_r:  2,
	n1_highmem_4_200_r:  3,
	n1_highmem_8_400_r:  4,
	n1_highmem_16_600_r: 5,
}

type RedisCluster struct {
	models.Cluster `json:",inline"`
	Bundles        []*RedisBundle     `json:"bundles"`
	DataCentres    []*RedisDataCentre `json:"dataCentres,omitempty"`
}

type RedisBundle struct {
	models.Bundle `json:",inline"`
	Options       *RedisOptions `json:"options"`
}

type RedisOptions struct {
	ClientEncryption bool  `json:"clientEncryption,omitempty"`
	MasterNodes      int32 `json:"masterNodes"`
	ReplicaNodes     int32 `json:"replicaNodes"`
	PasswordAuth     bool  `json:"passwordAuth,omitempty"`
}

type RedisDataCentre struct {
	models.DataCentre `json:",inline"`
	Bundles           []*RedisBundle `json:"bundles"`
}
