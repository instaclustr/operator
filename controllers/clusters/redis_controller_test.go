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

package clusters

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/instaclustr/operator/apis/clusters/v1beta1"
	"github.com/instaclustr/operator/controllers/tests"
	"github.com/instaclustr/operator/pkg/instaclustr"
	openapi "github.com/instaclustr/operator/pkg/instaclustr/mock/server/go"
	"github.com/instaclustr/operator/pkg/models"
)

const newRedisNodeSize = "RDS-DEV-t4g.medium-80"

var _ = Describe("Redis Controller", func() {
	redis := v1beta1.Redis{}
	redisManifest := v1beta1.Redis{}

	yfile, err := os.ReadFile("datatest/redis_v1beta1.yaml")
	Expect(err).NotTo(HaveOccurred())

	err = yaml.Unmarshal(yfile, &redisManifest)
	Expect(err).NotTo(HaveOccurred())

	redisNamespacedName := types.NamespacedName{Name: redisManifest.ObjectMeta.Name, Namespace: defaultNS}

	ctx := context.Background()

	When("apply a redis manifest", func() {
		It("should create a redis resources", func() {
			Expect(k8sClient.Create(ctx, &redisManifest)).Should(Succeed())
			done := tests.NewChannelWithTimeout(timeout)

			By("sending redis specification to the Instaclustr API and get ID of created cluster.")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, redisNamespacedName, &redis); err != nil {
					return false
				}

				return redis.Status.ID == openapi.CreatedID
			}).Should(BeTrue())

			<-done
		})
	})

	When("changing a node size", func() {
		It("should update a redis resources", func() {
			Expect(k8sClient.Get(ctx, redisNamespacedName, &redis)).Should(Succeed())

			patch := redis.NewPatch()
			redis.Spec.DataCentres[0].NodeSize = newRedisNodeSize
			Expect(k8sClient.Patch(ctx, &redis, patch)).Should(Succeed())
			By("sending a resize request to the Instaclustr API. And when the resize is completed, " +
				"the status job get new data from the InstAPI and update it in k8s redis resource")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, redisNamespacedName, &redis); err != nil {
					return false
				}

				if len(redis.Status.DataCentres) == 0 || len(redis.Status.DataCentres[0].Nodes) == 0 {
					return false
				}

				return redis.Status.DataCentres[0].Nodes[0].Size == newRedisNodeSize
			}, timeout, interval).Should(BeTrue())
		})
	})

	When("delete the redis resource", func() {
		It("should send delete request to the Instaclustr API", func() {
			Expect(k8sClient.Get(ctx, redisNamespacedName, &redis)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, &redis)).Should(Succeed())
			By("sending delete request to Instaclustr API")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, redisNamespacedName, &redis)
				if err != nil && !k8serrors.IsNotFound(err) {
					return false
				}

				return k8serrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
		})
	})

	When("Deleting the Redis resource by avoiding operator", func() {
		It("should try to get the cluster details and receive StatusNotFound", func() {
			redisManifest2 := redisManifest.DeepCopy()
			redisManifest2.Name += "-test-external-delete"
			redisManifest2.ResourceVersion = ""

			redis2 := v1beta1.Redis{}
			redis2NamespacedName := types.NamespacedName{
				Namespace: redisManifest2.Namespace,
				Name:      redisManifest2.Name,
			}

			Expect(k8sClient.Create(ctx, redisManifest2)).Should(Succeed())

			doneCh := tests.NewChannelWithTimeout(timeout)

			Eventually(func() bool {
				if err := k8sClient.Get(ctx, redis2NamespacedName, &redis2); err != nil {
					return false
				}

				if redis2.Status.State != models.RunningStatus {
					return false
				}

				doneCh <- struct{}{}

				return true
			}, timeout, interval).Should(BeTrue())

			<-doneCh

			By("using all possible ways other than k8s to delete a resource, the k8s operator scheduler should notice the changes and set the status of the deleted resource accordingly")
			Expect(instaClient.DeleteCluster(redis2.Status.ID, instaclustr.RedisEndpoint)).Should(Succeed())

			Eventually(func() bool {
				err := k8sClient.Get(ctx, redis2NamespacedName, &redis2)
				if err != nil {
					return false
				}

				return redis2.Status.State == models.DeletedStatus
			}, timeout, interval).Should(BeTrue())
		})
	})

})
