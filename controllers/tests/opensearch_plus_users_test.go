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

package tests

import (
	"context"
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/utils/strings/slices"

	clusterresource "github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/apis/clusters/v1beta1"
	"github.com/instaclustr/operator/pkg/apiextensions"
	openapi "github.com/instaclustr/operator/pkg/instaclustr/mock/server/go"
	"github.com/instaclustr/operator/pkg/models"
)

var _ = Describe("Basic openSearch User controller + Basic openSearch cluster controllers flow", func() {
	var (
		user1 clusterresource.OpenSearchUser
		user2 clusterresource.OpenSearchUser

		userManifest2 clusterresource.OpenSearchUser

		openSearch1 v1beta1.OpenSearch
		openSearch2 v1beta1.OpenSearch

		openSearchManifest v1beta1.OpenSearch

		secret v1.Secret
	)

	openSearchYAML, err := os.ReadFile("../clusters/datatest/opensearch_v1beta1.yaml")
	Expect(err).NotTo(HaveOccurred())

	err = yaml.Unmarshal(openSearchYAML, &openSearchManifest)
	Expect(err).NotTo(HaveOccurred())

	openSearchManifest2 := openSearchManifest.DeepCopy()
	openSearchManifest2.ObjectMeta.Name += "-2"
	openSearchManifest2.Spec.Name += "-2"

	secretManifest := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret-sample-os",
			Namespace: defaultNS,
		},
		StringData: map[string]string{
			"password": "password",
			"username": "username",
		},
	}

	userManifest1 := clusterresource.OpenSearchUser{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "opensearchuser-sample",
			Namespace: defaultNS,
		},
		Spec: clusterresource.OpenSearchUserSpec{
			SecretRef: &apiextensions.ObjectReference{
				Namespace: defaultNS,
				Name:      secretManifest.ObjectMeta.Name,
			},
		},
	}

	secretNamespacedName := types.NamespacedName{Name: secretManifest.ObjectMeta.Name, Namespace: defaultNS}

	userNamespacedName1 := types.NamespacedName{Name: userManifest1.ObjectMeta.Name, Namespace: defaultNS}
	userNamespacedName2 := types.NamespacedName{}

	openSearchNamespacedName1 := types.NamespacedName{Name: openSearchManifest.ObjectMeta.Name, Namespace: defaultNS}
	openSearchNamespacedName2 := types.NamespacedName{Name: openSearchManifest2.ObjectMeta.Name, Namespace: defaultNS}

	clusterID1 := openSearchManifest.Spec.Name + openapi.CreatedID
	clusterID2 := openSearchManifest2.Spec.Name + openapi.CreatedID

	ctx := context.Background()

	When("apply a secret and a openSearch user manifests", func() {
		It("should create both resources and they've got to have a link them through a finalizer", func() {
			Expect(k8sClient.Create(ctx, &secretManifest)).Should(Succeed())
			Expect(k8sClient.Create(ctx, &userManifest1)).Should(Succeed())
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, userNamespacedName1, &user1); err != nil {
					return false
				}

				if err := k8sClient.Get(ctx, secretNamespacedName, &secret); err != nil {
					return false
				}

				if user1.Finalizers == nil {
					return false
				}

				uniqFinalizer := user1.GetDeletionFinalizer()

				return slices.Contains(user1.Finalizers, uniqFinalizer) && slices.Contains(secret.Finalizers, uniqFinalizer)
			}).Should(BeTrue())
		})
	})

	When("apply a openSearch manifest", func() {
		It("should create a openSearch resource", func() {
			Expect(k8sClient.Create(ctx, &openSearchManifest)).Should(Succeed())

			By("sending openSearch specification to the Instaclustr API and get ID of a created cluster")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, openSearchNamespacedName1, &openSearch1); err != nil {
					return false
				}

				return openSearch1.Status.ID == clusterID1
			}).Should(BeTrue())
		})
	})

	When("add the user to a openSearch Reference", func() {
		It("should create the user for the cluster", func() {
			newUsers := []*v1beta1.Reference{{
				Namespace: userManifest1.Namespace,
				Name:      userManifest1.Name,
			}}

			Expect(k8sClient.Get(ctx, openSearchNamespacedName1, &openSearch1)).Should(Succeed())

			patch := openSearch1.NewPatch()
			// adding user
			openSearch1.Spec.UserRefs = newUsers
			Expect(k8sClient.Patch(ctx, &openSearch1, patch)).Should(Succeed())
			done := NewChannelWithTimeout(timeout)
			By("going to openSearch(cluster) controller predicate and put user entity to creation state. " +
				"Finally creates the user for the corresponded cluster")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, userNamespacedName1, &user1); err != nil {
					return false
				}

				if state, exist := user1.Status.ClustersEvents[clusterID1]; exist && state != models.Created {
					return false
				}

				return true
			}, timeout, interval).Should(BeTrue())
			<-done
		})
	})

	When("remove the user from the openSearch Reference", func() {
		It("should delete the user for the cluster", func() {
			Expect(k8sClient.Get(ctx, openSearchNamespacedName1, &openSearch1)).Should(Succeed())

			patch := openSearch1.NewPatch()
			// removing user
			openSearch1.Spec.UserRefs = []*v1beta1.Reference{}
			Expect(k8sClient.Patch(ctx, &openSearch1, patch)).Should(Succeed())
			By("going to openSearch(cluster) controller predicate and put user entity to deletion state. " +
				"Finally deletes the user for the corresponded cluster")
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, userNamespacedName1, &user1); err != nil {
					return false
				}

				if _, exist := user1.Status.ClustersEvents[clusterID1]; exist {
					return false
				}

				return true
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("Create multiple users for the cluster", func() {
		Specify("our users", func() {
			secretManifest2 := v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      secret.Name + "-2",
					Namespace: defaultNS,
				},
				StringData: map[string]string{
					"username": "carlo",
					"password": "qwerty123",
				},
			}
			Expect(k8sClient.Create(ctx, &secretManifest2)).Should(Succeed())

			userManifest2 = clusterresource.OpenSearchUser{
				ObjectMeta: metav1.ObjectMeta{
					Name:      userManifest1.Name + "-2",
					Namespace: defaultNS,
				},
				Spec: clusterresource.OpenSearchUserSpec{
					SecretRef: &apiextensions.ObjectReference{
						Name:      secretManifest2.Name,
						Namespace: secretManifest2.Namespace,
					}},
			}
			Expect(k8sClient.Create(ctx, &userManifest2)).Should(Succeed())

			By("adding the batch of users to the cluster, openSearch(cluster) controller predicate set them creation state")
			newUsers := []*v1beta1.Reference{
				{
					Namespace: userManifest1.Namespace,
					Name:      userManifest1.Name,
				},
				{
					Namespace: userManifest2.Namespace,
					Name:      userManifest2.Name,
				},
			}

			Expect(k8sClient.Get(ctx, openSearchNamespacedName1, &openSearch1)).Should(Succeed())

			patch := openSearch1.NewPatch()
			openSearch1.Spec.UserRefs = newUsers
			Expect(k8sClient.Patch(ctx, &openSearch1, patch)).Should(Succeed())

			userNamespacedName2 = types.NamespacedName{Name: userManifest2.ObjectMeta.Name, Namespace: defaultNS}
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, userNamespacedName1, &user1); err != nil {
					return false
				}
				if state, exist := user1.Status.ClustersEvents[clusterID1]; exist && state != models.Created {
					return false
				}

				if err := k8sClient.Get(ctx, userNamespacedName2, &user2); err != nil {
					return false
				}
				if state, exist := user2.Status.ClustersEvents[clusterID1]; exist && state != models.Created {
					return false
				}

				return true
			}, timeout, interval).Should(BeTrue())
		})
	})

	When("adding new user with exactly same spec (Secret reference)", func() {
		userManifest3 := &clusterresource.OpenSearchUser{
			ObjectMeta: metav1.ObjectMeta{
				Name:      userManifest1.Name + "-3",
				Namespace: defaultNS,
			},
			Spec: *userManifest1.Spec.DeepCopy(),
		}
		It("makes sure that several users may be added to one Secret and "+
			"the Secret has reference on each user respectively", func() {
			Expect(k8sClient.Create(ctx, userManifest3)).Should(Succeed())

			user3 := clusterresource.OpenSearchUser{}
			userNamespacedName3 := types.NamespacedName{Name: userManifest3.ObjectMeta.Name, Namespace: defaultNS}

			Eventually(func() bool {
				if err := k8sClient.Get(ctx, userNamespacedName1, &user1); err != nil {
					return false
				}

				if err := k8sClient.Get(ctx, userNamespacedName3, &user3); err != nil {
					return false
				}

				if err := k8sClient.Get(ctx, secretNamespacedName, &secret); err != nil {
					return false
				}

				if user1.Finalizers == nil {
					return false
				}

				uniqFinalizer := user1.GetDeletionFinalizer()
				uniqFinalizer3 := user3.GetDeletionFinalizer()

				return slices.Contains(user1.Finalizers, uniqFinalizer) && slices.Contains(secret.Finalizers, uniqFinalizer) &&
					slices.Contains(user3.Finalizers, uniqFinalizer3) && slices.Contains(secret.Finalizers, uniqFinalizer3)
			}).Should(BeTrue())
		})
	})

	When("attempting to delete a User entity that is currently associated with a cluster, an error should be returned", func() {
		It("indicates that the user reference needs to be removed from the attached cluster before deletion can proceed", func() {
			Expect(k8sClient.Get(ctx, userNamespacedName2, &user2)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, &user2)).Should(Succeed())

			// wait until send a request and receive the error
			time.Sleep(interval)

			events, err := core.NewForConfigOrDie(cfg).Events("default").
				List(context.TODO(), metav1.ListOptions{
					TypeMeta: metav1.TypeMeta{
						Kind: user2.Kind,
					},
					FieldSelector: "involvedObject.name=" + user2.Name,
				})
			Expect(err).NotTo(HaveOccurred())

			By("iterating through the user events' messages, we assert that the user has not been deleted yet. " +
				"a warning is returned indicating a lingering cluster reference")
			errMsg := "remove it from the clusters specifications first"

			Eventually(func() bool {
				if err := k8sClient.Get(ctx, userNamespacedName2, &user2); err != nil {
					return false
				}

				for _, item := range events.Items {
					if strings.Contains(item.Message, errMsg) {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

		})
	})

	When("choosing to continue working with a user in a deletion state and adding them to a new cluster", func() {
		It("should proceed smoothly without encountering any issues", func() {
			Expect(k8sClient.Get(ctx, userNamespacedName2, &user2)).Should(Succeed())
			By("creating another openSearch cluster manifest with filled user ref, " +
				"we make sure the user creation job works properly and show us that the user is available for use")
			newUsers := []*v1beta1.Reference{{
				Namespace: user2.Namespace,
				Name:      user2.Name,
			}}

			openSearchManifest2.Spec.UserRefs = newUsers
			Expect(k8sClient.Create(ctx, openSearchManifest2)).Should(Succeed())

			Eventually(func() bool {
				if err := k8sClient.Get(ctx, openSearchNamespacedName2, &openSearch2); err != nil {
					return false
				}

				if openSearch2.Status.ID != clusterID2 {
					return false
				}

				if err := k8sClient.Get(ctx, userNamespacedName2, &user2); err != nil {
					return false
				}

				if event, exist := user2.Status.ClustersEvents[clusterID2]; exist && event != models.Created {
					return false
				}

				return true
			}, timeout, interval).Should(BeTrue())
		})
	})

	When("remove the user, which is in a deletion state, from cluster with multiple users", func() {
		It("sends delete user request to user controller", func() {
			Expect(k8sClient.Get(ctx, userNamespacedName2, &user2)).Should(Succeed())
			Expect(k8sClient.Get(ctx, openSearchNamespacedName1, &openSearch1)).Should(Succeed())
			patch := openSearch1.NewPatch()
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, userNamespacedName2, &user2); err != nil {
					return false
				}

				for i := range openSearch1.Spec.UserRefs {
					if user2.Name == openSearch1.Spec.UserRefs[i].Name && user2.Namespace == openSearch1.Spec.UserRefs[i].Namespace {
						openSearch1.Spec.UserRefs = removeUserByIndex(openSearch1.Spec.UserRefs, i)
						Expect(k8sClient.Patch(ctx, &openSearch1, patch)).Should(Succeed())
					}
				}

				if err := k8sClient.Get(ctx, userNamespacedName2, &user2); err != nil {
					return false
				}

				if _, exist := user2.Status.ClustersEvents[clusterID2]; exist {
					return false
				}

				return true
			}, timeout, interval).Should(BeTrue())
		})
	})

	When("deleting a cluster to which (deleting) user is attached", func() {
		It("sends detaching queue to User controller. Since the user has been marked for deletion and "+
			"no longer has any attachments to clusters, the user is subsequently scheduled for deletion", func() {
			Expect(k8sClient.Get(ctx, openSearchNamespacedName2, &openSearch2)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, &openSearch2)).Should(Succeed())
			Eventually(func() bool {
				err := k8sClient.Get(ctx, userNamespacedName2, &user2)
				if err != nil && !k8serrors.IsNotFound(err) {
					return false
				}

				return k8serrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
		})
	})
})
