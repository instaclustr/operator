/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package clusterresources

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/strings/slices"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	clustersv1beta1 "github.com/instaclustr/operator/apis/clusters/v1beta1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
)

func areFirewallRuleStatusesEqual(a, b *v1beta1.FirewallRuleStatus) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil ||
		b == nil ||
		a.ID != b.ID ||
		a.Status != b.Status ||
		a.DeferredReason != b.DeferredReason {
		return false
	}

	return true
}

func arePeeringStatusesEqual(a, b *v1beta1.PeeringStatus) bool {
	if a.ID != b.ID ||
		a.Name != b.Name ||
		a.StatusCode != b.StatusCode ||
		a.FailureReason != b.FailureReason {
		return false
	}

	return true
}

func areEncryptionKeyStatusesEqual(a, b *v1beta1.AWSEncryptionKeyStatus) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil ||
		b == nil ||
		a.ID != b.ID ||
		a.InUse != b.InUse {
		return false
	}

	return true
}

func CheckIfUserExistsOnInstaclustrAPI(username, clusterID, app string, api instaclustr.API) (bool, error) {
	users, err := api.FetchUsers(clusterID, app)
	if err != nil {
		return false, err
	}

	return slices.Contains(users, username), nil
}

func subnetsEqual(subnets1, subnets2 []string) bool {
	if len(subnets1) != len(subnets2) {
		return false
	}

	for _, s1 := range subnets1 {
		var equal bool
		for _, s2 := range subnets2 {
			if s1 == s2 {
				equal = true
			}
		}

		if !equal {
			return false
		}
	}

	return true
}

type ClusterIDProvider interface {
	client.Object
	GetClusterID() string
	GetDataCentreID(cdcName string) string
}

func GetDataCentreID(cl client.Client, ctx context.Context, ref *v1beta1.ClusterRef) (string, error) {
	var obj ClusterIDProvider
	ns := types.NamespacedName{
		Namespace: ref.Namespace,
		Name:      ref.Name,
	}

	switch ref.ClusterKind {
	case models.RedisClusterKind:
		obj = &clustersv1beta1.Redis{}
	case models.OpenSearchKind:
		obj = &clustersv1beta1.OpenSearch{}
	case models.KafkaKind:
		obj = &clustersv1beta1.Kafka{}
	case models.CassandraKind:
		obj = &clustersv1beta1.Cassandra{}
	case models.PgClusterKind:
		obj = &clustersv1beta1.PostgreSQL{}
	case models.ZookeeperClusterKind:
		obj = &clustersv1beta1.Zookeeper{}
	case models.CadenceClusterKind:
		obj = &clustersv1beta1.Cadence{}
	case models.KafkaConnectClusterKind:
		obj = &clustersv1beta1.KafkaConnect{}
	default:
		return "", models.ErrUnsupportedClusterKind
	}
	err := cl.Get(ctx, ns, obj)
	if err != nil {
		return "", err
	}

	return obj.GetDataCentreID(ref.CDCName), nil
}

func GetClusterID(cl client.Client, ctx context.Context, ref *v1beta1.ClusterRef) (string, error) {
	var obj ClusterIDProvider

	switch ref.ClusterKind {
	case models.RedisClusterKind:
		obj = &clustersv1beta1.Redis{}
	case models.OpenSearchKind:
		obj = &clustersv1beta1.OpenSearch{}
	case models.KafkaKind:
		obj = &clustersv1beta1.Kafka{}
	case models.CassandraKind:
		obj = &clustersv1beta1.Cassandra{}
	case models.PgClusterKind:
		obj = &clustersv1beta1.PostgreSQL{}
	case models.ZookeeperClusterKind:
		obj = &clustersv1beta1.Zookeeper{}
	case models.CadenceClusterKind:
		obj = &clustersv1beta1.Cadence{}
	case models.KafkaConnectClusterKind:
		obj = &clustersv1beta1.KafkaConnect{}
	default:
		return "", models.ErrUnsupportedClusterKind
	}

	err := cl.Get(ctx, ref.AsNamespacedName(), obj)
	if err != nil {
		return "", err
	}

	return obj.GetClusterID(), nil
}
