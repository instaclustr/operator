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

package exposeservice

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/instaclustr/operator/apis/clusters/v1beta1"
	"github.com/instaclustr/operator/pkg/models"
)

func Create(
	c client.Client,
	clusterName string,
	ns string,
	nodes []*v1beta1.Node,
	targetPort int32,
) error {
	svcName := fmt.Sprintf(models.ExposeServiceNameTemplate, clusterName)
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcName,
			Namespace: ns,
			Labels:    map[string]string{models.ClusterNameLabel: clusterName},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "app",
					Protocol:   v1.ProtocolTCP,
					Port:       80,
					TargetPort: intstr.IntOrString{IntVal: targetPort},
				},
			},
			ClusterIP: v1.ClusterIPNone,
			Type:      v1.ServiceTypeClusterIP,
		},
	}

	addresses := []v1.EndpointAddress{}
	for _, node := range nodes {
		if node.PublicAddress == "" {
			continue
		}

		addresses = append(addresses, v1.EndpointAddress{
			IP: node.PublicAddress,
		})
	}

	endpoints := &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcName,
			Namespace: ns,
			Labels:    map[string]string{models.ClusterNameLabel: clusterName},
		},
		Subsets: []v1.EndpointSubset{
			{
				Addresses: addresses,
				Ports: []v1.EndpointPort{
					{
						Name:     "app",
						Port:     targetPort,
						Protocol: v1.ProtocolTCP,
					},
				},
			},
		},
	}

	namespacedName := client.ObjectKeyFromObject(service)
	err := c.Get(context.Background(), namespacedName, service)
	if k8serrors.IsNotFound(err) {
		err = c.Create(context.Background(), service)
		if err != nil {
			return err
		}

		err = c.Create(context.Background(), endpoints)
		if err != nil {
			return err
		}
	} else {
		err = c.Update(context.Background(), endpoints)
		if err != nil {
			return err
		}
	}

	return nil
}

func Delete(
	c client.Client,
	clusterName string,
	ns string,
) error {
	serviceList := &v1.ServiceList{}
	listOpts := []client.ListOption{
		client.InNamespace(ns),
		client.MatchingLabels{models.ClusterNameLabel: clusterName},
	}

	err := c.List(context.Background(), serviceList, listOpts...)
	if err != nil {
		return err
	}

	for _, svc := range serviceList.Items {
		err = c.Delete(context.Background(), &svc)
		if err != nil {
			return err
		}
	}

	return nil
}

func GetExposeService(
	c client.Client,
	clusterName string,
	ns string,
) (*v1.ServiceList, error) {
	serviceList := &v1.ServiceList{}
	listOpts := []client.ListOption{
		client.InNamespace(ns),
		client.MatchingLabels{models.ClusterNameLabel: clusterName},
	}

	err := c.List(context.Background(), serviceList, listOpts...)
	if err != nil {
		return nil, err
	}

	return serviceList, nil
}

func GetExposeServiceEndpoints(
	c client.Client,
	clusterName string,
	ns string,
) (*v1.EndpointsList, error) {
	endpointsList := &v1.EndpointsList{}
	listOpts := []client.ListOption{
		client.InNamespace(ns),
		client.MatchingLabels{models.ClusterNameLabel: clusterName},
	}

	err := c.List(context.Background(), endpointsList, listOpts...)
	if err != nil {
		return nil, err
	}

	return endpointsList, nil
}
