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
	"fmt"
	"strings"

	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	virtcorev1 "kubevirt.io/api/core/v1"
	cdiv1beta1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/instaclustr/operator/apis/clusters/v1beta1"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/scheduler"
)

//nolint:unused,deadcode
type onPremisesBootstrap struct {
	K8sClient             client.Client
	K8sObject             client.Object
	EventRecorder         record.EventRecorder
	ClusterID             string
	DataCentreID          string
	Nodes                 []*v1beta1.Node
	OnPremisesSpec        *v1beta1.OnPremisesSpec
	ExposePorts           []k8scorev1.ServicePort
	HeadlessPorts         []k8scorev1.ServicePort
	PrivateNetworkCluster bool
}

//nolint:unused,deadcode
func newOnPremisesBootstrap(
	k8sClient client.Client,
	o client.Object,
	e record.EventRecorder,
	clusterID string,
	dataCentreID string,
	nodes []*v1beta1.Node,
	onPremisesSpec *v1beta1.OnPremisesSpec,
	exposePorts,
	headlessPorts []k8scorev1.ServicePort,
	privateNetworkCluster bool,
) *onPremisesBootstrap {
	return &onPremisesBootstrap{
		K8sClient:             k8sClient,
		K8sObject:             o,
		EventRecorder:         e,
		ClusterID:             clusterID,
		DataCentreID:          dataCentreID,
		Nodes:                 nodes,
		OnPremisesSpec:        onPremisesSpec,
		ExposePorts:           exposePorts,
		HeadlessPorts:         headlessPorts,
		PrivateNetworkCluster: privateNetworkCluster,
	}
}

//nolint:unused,deadcode
func handleCreateOnPremisesClusterResources(ctx context.Context, b *onPremisesBootstrap) error {
	if b.DataCentreID == "" {
		return fmt.Errorf("datacenter ID is empty")
	}

	if b.PrivateNetworkCluster {
		err := reconcileSSHGatewayResources(ctx, b)
		if err != nil {
			return err
		}
	}

	err := reconcileNodesResources(ctx, b)
	if err != nil {
		return err
	}

	return nil
}

//nolint:unused,deadcode
func reconcileSSHGatewayResources(ctx context.Context, b *onPremisesBootstrap) error {
	gatewayDVSize, err := resource.ParseQuantity(b.OnPremisesSpec.OSDiskSize)
	if err != nil {
		return err
	}

	gatewayDVName := fmt.Sprintf("%s-%s", models.GatewayDVPrefix, strings.ToLower(b.K8sObject.GetName()))
	gatewayDV, err := createDV(
		ctx,
		b,
		gatewayDVName,
		b.DataCentreID,
		gatewayDVSize,
		true,
	)
	if err != nil {
		return err
	}

	gatewayCPU := resource.Quantity{}
	gatewayCPU.Set(b.OnPremisesSpec.SSHGatewayCPU)

	gatewayMemory, err := resource.ParseQuantity(b.OnPremisesSpec.SSHGatewayMemory)
	if err != nil {
		return err
	}

	gatewayName := fmt.Sprintf("%s-%s", models.GatewayVMPrefix, strings.ToLower(b.K8sObject.GetName()))

	gatewayVM := &virtcorev1.VirtualMachine{}
	err = b.K8sClient.Get(ctx, types.NamespacedName{
		Namespace: b.K8sObject.GetNamespace(),
		Name:      gatewayName,
	}, gatewayVM)
	if client.IgnoreNotFound(err) != nil {
		return err
	}
	if k8serrors.IsNotFound(err) {
		gatewayVM, err = newVM(
			ctx,
			b,
			gatewayName,
			b.DataCentreID,
			models.GatewayRack,
			gatewayDV.Name,
			gatewayCPU,
			gatewayMemory)
		if err != nil {
			return err
		}
		err = b.K8sClient.Create(ctx, gatewayVM)
		if err != nil {
			return err
		}
	}

	gatewaySvcName := fmt.Sprintf("%s-%s", models.GatewaySvcPrefix, gatewayName)
	gatewayExposeService := &k8scorev1.Service{}
	err = b.K8sClient.Get(ctx, types.NamespacedName{
		Namespace: b.K8sObject.GetNamespace(),
		Name:      gatewaySvcName,
	}, gatewayExposeService)

	if client.IgnoreNotFound(err) != nil {
		return err
	}
	if k8serrors.IsNotFound(err) {
		gatewayExposeService = newExposeService(
			b,
			gatewaySvcName,
			gatewayName,
			b.DataCentreID,
		)
		err = b.K8sClient.Create(ctx, gatewayExposeService)
		if err != nil {
			return err
		}
	}

	clusterIPServiceName := fmt.Sprintf("cluster-ip-%s", gatewayName)
	nodeExposeService := &k8scorev1.Service{}
	err = b.K8sClient.Get(ctx, types.NamespacedName{
		Namespace: b.K8sObject.GetNamespace(),
		Name:      clusterIPServiceName,
	}, nodeExposeService)
	if client.IgnoreNotFound(err) != nil {
		return err
	}
	if k8serrors.IsNotFound(err) {
		nodeExposeService = newClusterIPService(
			b,
			clusterIPServiceName,
			gatewayName,
			b.DataCentreID,
		)
		err = b.K8sClient.Create(ctx, nodeExposeService)
		if err != nil {
			return err
		}
	}

	return nil
}

//nolint:unused,deadcode
func reconcileNodesResources(ctx context.Context, b *onPremisesBootstrap) error {
	for i, node := range b.Nodes {
		nodeOSDiskSize, err := resource.ParseQuantity(b.OnPremisesSpec.OSDiskSize)
		if err != nil {
			return err
		}

		nodeOSDiskDVName := fmt.Sprintf("%s-%d-%s", models.NodeOSDVPrefix, i, strings.ToLower(b.K8sObject.GetName()))
		nodeOSDV, err := createDV(
			ctx,
			b,
			nodeOSDiskDVName,
			node.ID,
			nodeOSDiskSize,
			true,
		)
		if err != nil {
			return err
		}

		nodeDataDiskDVSize, err := resource.ParseQuantity(b.OnPremisesSpec.DataDiskSize)
		if err != nil {
			return err
		}

		nodeDataDiskDVName := fmt.Sprintf("%s-%d-%s", models.NodeDVPrefix, i, strings.ToLower(b.K8sObject.GetName()))
		nodeDataDV, err := createDV(
			ctx,
			b,
			nodeDataDiskDVName,
			node.ID,
			nodeDataDiskDVSize,
			false,
		)
		if err != nil {
			return err
		}

		nodeCPU := resource.Quantity{}
		nodeCPU.Set(b.OnPremisesSpec.NodeCPU)

		nodeMemory, err := resource.ParseQuantity(b.OnPremisesSpec.NodeMemory)
		if err != nil {
			return err
		}

		nodeName := fmt.Sprintf("%s-%d-%s", models.NodeVMPrefix, i, strings.ToLower(b.K8sObject.GetName()))

		nodeVM := &virtcorev1.VirtualMachine{}
		err = b.K8sClient.Get(ctx, types.NamespacedName{
			Namespace: b.K8sObject.GetNamespace(),
			Name:      nodeName,
		}, nodeVM)
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		if k8serrors.IsNotFound(err) {
			nodeVM, err = newVM(
				ctx,
				b,
				nodeName,
				node.ID,
				node.Rack,
				nodeOSDV.Name,
				nodeCPU,
				nodeMemory,
				nodeDataDV.Name,
			)
			if err != nil {
				return err
			}
			err = b.K8sClient.Create(ctx, nodeVM)
			if err != nil {
				return err
			}
		}

		if !b.PrivateNetworkCluster {
			nodeExposeName := fmt.Sprintf("%s-%s", models.NodeSvcPrefix, nodeName)
			nodeExposeService := &k8scorev1.Service{}
			err = b.K8sClient.Get(ctx, types.NamespacedName{
				Namespace: b.K8sObject.GetNamespace(),
				Name:      nodeExposeName,
			}, nodeExposeService)
			if client.IgnoreNotFound(err) != nil {
				return err
			}
			if k8serrors.IsNotFound(err) {
				nodeExposeService = newExposeService(
					b,
					nodeExposeName,
					nodeName,
					node.ID,
				)
				err = b.K8sClient.Create(ctx, nodeExposeService)
				if err != nil {
					return err
				}
			}
		}

		headlessServiceName := fmt.Sprintf("%s-%s", models.KubevirtSubdomain, strings.ToLower(b.K8sObject.GetName()))
		headlessSVC := &k8scorev1.Service{}
		err = b.K8sClient.Get(ctx, types.NamespacedName{
			Namespace: b.K8sObject.GetNamespace(),
			Name:      headlessServiceName,
		}, headlessSVC)

		if client.IgnoreNotFound(err) != nil {
			return err
		}
		if k8serrors.IsNotFound(err) {
			headlessSVC = newHeadlessService(
				b,
				headlessServiceName,
			)
			err = b.K8sClient.Create(ctx, headlessSVC)
			if err != nil {
				return err
			}
		}

		clusterIPServiceName := fmt.Sprintf("cluster-ip-%s", nodeName)
		nodeExposeService := &k8scorev1.Service{}
		err = b.K8sClient.Get(ctx, types.NamespacedName{
			Namespace: b.K8sObject.GetNamespace(),
			Name:      clusterIPServiceName,
		}, nodeExposeService)
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		if k8serrors.IsNotFound(err) {
			nodeExposeService = newClusterIPService(
				b,
				clusterIPServiceName,
				nodeName,
				node.ID,
			)
			err = b.K8sClient.Create(ctx, nodeExposeService)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

//nolint:unused,deadcode
func createDV(
	ctx context.Context,
	b *onPremisesBootstrap,
	name,
	nodeID string,
	size resource.Quantity,
	isOSDisk bool,
) (*cdiv1beta1.DataVolume, error) {
	dv := &cdiv1beta1.DataVolume{}
	err := b.K8sClient.Get(ctx, types.NamespacedName{
		Namespace: b.K8sObject.GetNamespace(),
		Name:      name,
	}, dv)
	if client.IgnoreNotFound(err) != nil {
		return nil, err
	}
	if k8serrors.IsNotFound(err) {
		dv = newDataDiskDV(
			b,
			name,
			nodeID,
			size,
			isOSDisk,
		)
		err = b.K8sClient.Create(ctx, dv)
		if err != nil {
			return nil, err
		}
	}
	return dv, nil
}

//nolint:unused,deadcode
func newDataDiskDV(
	b *onPremisesBootstrap,
	name,
	nodeID string,
	storageSize resource.Quantity,
	isOSDisk bool,
) *cdiv1beta1.DataVolume {
	dvSource := &cdiv1beta1.DataVolumeSource{}

	if isOSDisk {
		dvSource.HTTP = &cdiv1beta1.DataVolumeSourceHTTP{URL: b.OnPremisesSpec.OSImageURL}
	} else {
		dvSource.Blank = &cdiv1beta1.DataVolumeBlankImage{}
	}

	return &cdiv1beta1.DataVolume{
		TypeMeta: metav1.TypeMeta{
			Kind:       models.DVKind,
			APIVersion: models.CDIKubevirtV1beta1APIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: b.K8sObject.GetNamespace(),
			Labels: map[string]string{
				models.ClusterIDLabel: b.ClusterID,
				models.NodeIDLabel:    nodeID,
			},
			Finalizers: []string{models.DeletionFinalizer},
		},
		Spec: cdiv1beta1.DataVolumeSpec{
			Source: dvSource,
			PVC: &k8scorev1.PersistentVolumeClaimSpec{
				AccessModes: []k8scorev1.PersistentVolumeAccessMode{
					k8scorev1.ReadWriteOnce,
				},
				Resources: k8scorev1.ResourceRequirements{
					Requests: k8scorev1.ResourceList{
						models.Storage: storageSize,
					},
				},
				StorageClassName: &b.OnPremisesSpec.StorageClassName,
			},
		},
	}
}

//nolint:unused,deadcode
func newVM(
	ctx context.Context,
	b *onPremisesBootstrap,
	vmName,
	nodeID,
	nodeRack,
	OSDiskDVName string,
	cpu,
	memory resource.Quantity,
	storageDVNames ...string,
) (*virtcorev1.VirtualMachine, error) {
	runStrategy := virtcorev1.RunStrategyAlways
	bootOrder1 := uint(1)

	cloudInitSecret := &k8scorev1.Secret{}
	err := b.K8sClient.Get(ctx, types.NamespacedName{
		Namespace: b.OnPremisesSpec.CloudInitScriptRef.Namespace,
		Name:      b.OnPremisesSpec.CloudInitScriptRef.Name,
	}, cloudInitSecret)
	if err != nil {
		return nil, err
	}

	labelSet := map[string]string{
		models.ClusterIDLabel:      b.ClusterID,
		models.NodeIDLabel:         nodeID,
		models.NodeRackLabel:       nodeRack,
		models.KubevirtDomainLabel: vmName,
	}

	if nodeRack != models.GatewayRack {
		labelSet[models.NodeLabel] = models.WorkerNode
	}

	vm := &virtcorev1.VirtualMachine{
		TypeMeta: metav1.TypeMeta{
			Kind:       models.VirtualMachineKind,
			APIVersion: models.KubevirtV1APIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:       vmName,
			Namespace:  b.K8sObject.GetNamespace(),
			Labels:     labelSet,
			Finalizers: []string{models.DeletionFinalizer},
		},
		Spec: virtcorev1.VirtualMachineSpec{
			RunStrategy: &runStrategy,
			Template: &virtcorev1.VirtualMachineInstanceTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labelSet,
				},
				Spec: virtcorev1.VirtualMachineInstanceSpec{
					Hostname:  vmName,
					Subdomain: fmt.Sprintf("%s-%s", models.KubevirtSubdomain, b.K8sObject.GetName()),
					Domain: virtcorev1.DomainSpec{
						Resources: virtcorev1.ResourceRequirements{
							Requests: k8scorev1.ResourceList{
								models.CPU:    cpu,
								models.Memory: memory,
							},
						},
						Devices: virtcorev1.Devices{
							Disks: []virtcorev1.Disk{
								{
									Name:      models.Boot,
									BootOrder: &bootOrder1,
									IO:        models.Native,
									Cache:     models.None,
									DiskDevice: virtcorev1.DiskDevice{
										Disk: &virtcorev1.DiskTarget{
											Bus: models.Virtio,
										},
									},
								},
								{
									Name:       models.CloudInit,
									DiskDevice: virtcorev1.DiskDevice{},
									Cache:      models.None,
								},
							},
							Interfaces: []virtcorev1.Interface{
								{
									Name: models.Default,
									InterfaceBindingMethod: virtcorev1.InterfaceBindingMethod{
										Bridge: &virtcorev1.InterfaceBridge{},
									},
								},
							},
						},
					},
					Volumes: []virtcorev1.Volume{
						{
							Name: models.Boot,
							VolumeSource: virtcorev1.VolumeSource{
								PersistentVolumeClaim: &virtcorev1.PersistentVolumeClaimVolumeSource{
									PersistentVolumeClaimVolumeSource: k8scorev1.PersistentVolumeClaimVolumeSource{
										ClaimName: OSDiskDVName,
									},
								},
							},
						},
						{
							Name: models.CloudInit,
							VolumeSource: virtcorev1.VolumeSource{
								CloudInitNoCloud: &virtcorev1.CloudInitNoCloudSource{
									UserDataSecretRef: &k8scorev1.LocalObjectReference{
										Name: b.OnPremisesSpec.CloudInitScriptRef.Name,
									},
								},
							},
						},
					},
					Networks: []virtcorev1.Network{
						{
							Name: models.Default,
							NetworkSource: virtcorev1.NetworkSource{
								Pod: &virtcorev1.PodNetwork{},
							},
						},
					},
				},
			},
		},
	}

	for i, dvName := range storageDVNames {
		diskName := fmt.Sprintf("%s-%d-%s", models.DataDisk, i, vm.Name)

		vm.Spec.Template.Spec.Domain.Devices.Disks = append(vm.Spec.Template.Spec.Domain.Devices.Disks, virtcorev1.Disk{
			Name:  diskName,
			IO:    models.Native,
			Cache: models.None,
			DiskDevice: virtcorev1.DiskDevice{
				Disk: &virtcorev1.DiskTarget{
					Bus: models.Virtio,
				},
			},
			Serial: models.DataDiskSerial,
		})

		vm.Spec.Template.Spec.Volumes = append(vm.Spec.Template.Spec.Volumes, virtcorev1.Volume{
			Name: diskName,
			VolumeSource: virtcorev1.VolumeSource{
				PersistentVolumeClaim: &virtcorev1.PersistentVolumeClaimVolumeSource{
					PersistentVolumeClaimVolumeSource: k8scorev1.PersistentVolumeClaimVolumeSource{
						ClaimName: dvName,
					},
				},
			},
		})
	}

	return vm, nil
}

//nolint:unused,deadcode
func newExposeService(
	b *onPremisesBootstrap,
	svcName,
	vmName,
	nodeID string,
) *k8scorev1.Service {
	return &k8scorev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       models.ServiceKind,
			APIVersion: models.K8sAPIVersionV1,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcName,
			Namespace: b.K8sObject.GetNamespace(),
			Labels: map[string]string{
				models.ClusterIDLabel: b.ClusterID,
				models.NodeIDLabel:    nodeID,
			},
			Finalizers: []string{models.DeletionFinalizer},
		},
		Spec: k8scorev1.ServiceSpec{
			Ports: b.ExposePorts,
			Selector: map[string]string{
				models.KubevirtDomainLabel: vmName,
				models.NodeIDLabel:         nodeID,
			},
			Type: k8scorev1.ServiceTypeLoadBalancer,
		},
	}
}

//nolint:unused,deadcode
func newHeadlessService(
	b *onPremisesBootstrap,
	svcName string,
) *k8scorev1.Service {
	return &k8scorev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       models.ServiceKind,
			APIVersion: models.K8sAPIVersionV1,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcName,
			Namespace: b.K8sObject.GetNamespace(),
			Labels: map[string]string{
				models.ClusterIDLabel: b.ClusterID,
			},
			Finalizers: []string{models.DeletionFinalizer},
		},
		Spec: k8scorev1.ServiceSpec{
			ClusterIP: k8scorev1.ClusterIPNone,
			Ports:     b.HeadlessPorts,
			Selector: map[string]string{
				models.ClusterIDLabel: b.ClusterID,
				models.NodeLabel:      models.WorkerNode,
			},
		},
	}
}

//nolint:unused,deadcode
func deleteOnPremResources(ctx context.Context, K8sClient client.Client, clusterID, ns string) error {
	vms := &virtcorev1.VirtualMachineList{}
	err := K8sClient.List(ctx, vms, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			models.ClusterIDLabel: clusterID,
		}),
		Namespace: ns,
	})
	if err != nil {
		return err
	}

	for _, vm := range vms.Items {
		err = K8sClient.Delete(ctx, &vm)
		if err != nil {
			return err
		}

		patch := client.MergeFrom(vm.DeepCopy())
		controllerutil.RemoveFinalizer(&vm, models.DeletionFinalizer)
		err = K8sClient.Patch(ctx, &vm, patch)
		if err != nil {
			return err
		}
	}

	vmis := &virtcorev1.VirtualMachineInstanceList{}
	err = K8sClient.List(ctx, vmis, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			models.ClusterIDLabel: clusterID,
		}),
		Namespace: ns,
	})
	if err != nil {
		return err
	}

	for _, vmi := range vmis.Items {
		err = K8sClient.Delete(ctx, &vmi)
		if err != nil {
			return err
		}

		patch := client.MergeFrom(vmi.DeepCopy())
		controllerutil.RemoveFinalizer(&vmi, models.DeletionFinalizer)
		err = K8sClient.Patch(ctx, &vmi, patch)
		if err != nil {
			return err
		}
	}

	dvs := &cdiv1beta1.DataVolumeList{}
	err = K8sClient.List(ctx, dvs, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			models.ClusterIDLabel: clusterID,
		}),
		Namespace: ns,
	})
	if err != nil {
		return err
	}

	for _, dv := range dvs.Items {
		err = K8sClient.Delete(ctx, &dv)
		if err != nil {
			return err
		}

		patch := client.MergeFrom(dv.DeepCopy())
		controllerutil.RemoveFinalizer(&dv, models.DeletionFinalizer)
		err = K8sClient.Patch(ctx, &dv, patch)
		if err != nil {
			return err
		}
	}

	svcs := &k8scorev1.ServiceList{}
	err = K8sClient.List(ctx, svcs, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			models.ClusterIDLabel: clusterID,
		}),
		Namespace: ns,
	})
	if err != nil {
		return err
	}

	for _, svc := range svcs.Items {
		err = K8sClient.Delete(ctx, &svc)
		if err != nil {
			return err
		}

		patch := client.MergeFrom(svc.DeepCopy())
		controllerutil.RemoveFinalizer(&svc, models.DeletionFinalizer)
		err = K8sClient.Patch(ctx, &svc, patch)
		if err != nil {
			return err
		}
	}

	return nil
}

//nolint:unused,deadcode
func newExposePorts(sp []k8scorev1.ServicePort) []k8scorev1.ServicePort {
	var ports []k8scorev1.ServicePort
	ports = []k8scorev1.ServicePort{{
		Name: models.SSH,
		Port: models.Port22,
		TargetPort: intstr.IntOrString{
			Type:   intstr.Int,
			IntVal: models.Port22,
		},
	},
	}

	ports = append(ports, sp...)

	return ports
}

//nolint:unused,deadcode
func newWatchOnPremisesIPsJob(kind string, b *onPremisesBootstrap) scheduler.Job {
	l := log.Log.WithValues("component", fmt.Sprintf("%sOnPremisesIPsCheckerJob", kind))

	return func() error {
		allNodePods := &k8scorev1.PodList{}
		err := b.K8sClient.List(context.Background(), allNodePods, &client.ListOptions{
			LabelSelector: labels.SelectorFromSet(map[string]string{
				models.ClusterIDLabel: b.ClusterID,
				models.NodeLabel:      models.WorkerNode,
			}),
			Namespace: b.K8sObject.GetNamespace(),
		})
		if err != nil {
			l.Error(err, "Cannot get on-premises cluster pods",
				"cluster name", b.K8sObject.GetName(),
				"clusterID", b.ClusterID,
			)

			b.EventRecorder.Eventf(
				b.K8sObject, models.Warning, models.CreationFailed,
				"Fetching on-premises cluster pods is failed. Reason: %v",
				err,
			)
			return err
		}

		if len(allNodePods.Items) != len(b.Nodes) {
			err = fmt.Errorf("the quantity of pods does not match the number of on-premises nodes")
			l.Error(err, "Cannot compare private IPs for the cluster",
				"cluster name", b.K8sObject.GetName(),
				"clusterID", b.ClusterID,
			)

			b.EventRecorder.Eventf(
				b.K8sObject, models.Warning, models.CreationFailed,
				"Comparing of cluster's private IPs is failing. Reason: %v",
				err,
			)
			return err
		}

		for _, node := range b.Nodes {
			nodePods := &k8scorev1.PodList{}
			err = b.K8sClient.List(context.Background(), nodePods, &client.ListOptions{
				LabelSelector: labels.SelectorFromSet(map[string]string{
					models.ClusterIDLabel: b.ClusterID,
					models.NodeIDLabel:    node.ID,
				}),
				Namespace: b.K8sObject.GetNamespace(),
			})
			if err != nil {
				l.Error(err, "Cannot get on-premises cluster pods",
					"cluster name", b.K8sObject.GetName(),
					"clusterID", b.ClusterID,
				)

				b.EventRecorder.Eventf(
					b.K8sObject, models.Warning, models.CreationFailed,
					"Fetching on-premises cluster pods is failed. Reason: %v",
					err,
				)
				return err
			}

			for _, pod := range nodePods.Items {
				if (pod.Status.PodIP != "" && node.PrivateAddress != "") &&
					(pod.Status.PodIP != node.PrivateAddress) {

					err = fmt.Errorf("private IPs was changed")
					l.Error(err, "Node's private IP addresses are not equal",
						"cluster name", b.K8sObject.GetName(),
						"clusterID", b.ClusterID,
						"nodeID", node.ID,
						"nodeIP", node.PrivateAddress,
						"podIP", pod.Status.PodIP,
					)

					b.EventRecorder.Eventf(
						b.K8sObject, models.Warning, models.CreationFailed,
						"The private IP addresses of the node are not matching. Reason: %v",
						err,
					)
					return err
				}
			}

			if !b.PrivateNetworkCluster {
				nodeSVCs := &k8scorev1.ServiceList{}
				err = b.K8sClient.List(context.Background(), nodeSVCs, &client.ListOptions{
					LabelSelector: labels.SelectorFromSet(map[string]string{
						models.ClusterIDLabel: b.ClusterID,
						models.NodeIDLabel:    node.ID,
					}),
					Namespace: b.K8sObject.GetNamespace(),
				})
				if err != nil {
					l.Error(err, "Cannot get services backed by on-premises cluster pods",
						"cluster name", b.K8sObject.GetName(),
						"clusterID", b.ClusterID,
					)

					b.EventRecorder.Eventf(
						b.K8sObject, models.Warning, models.CreationFailed,
						"Fetching services backed by on-premises cluster pods is failed. Reason: %v",
						err,
					)
					return err
				}
				for _, svc := range nodeSVCs.Items {
					if len(svc.Status.LoadBalancer.Ingress) != 0 {
						if (svc.Status.LoadBalancer.Ingress[0].IP != "" && node.PublicAddress != "") &&
							(svc.Status.LoadBalancer.Ingress[0].IP != node.PublicAddress) {

							err = fmt.Errorf("public IPs was changed")
							l.Error(err, "Node's public IP addresses are not equal",
								"cluster name", b.K8sObject.GetName(),
								"clusterID", b.ClusterID,
								"nodeID", node.ID,
								"nodeIP", node.PrivateAddress,
								"svcIP", svc.Status.LoadBalancer.Ingress[0].IP,
							)

							b.EventRecorder.Eventf(
								b.K8sObject, models.Warning, models.CreationFailed,
								"The public IP addresses of the node are not matching. Reason: %v",
								err,
							)
							return err
						}
					}
				}
			}
		}
		return nil
	}
}

//nolint:unused,deadcode
func newClusterIPService(
	b *onPremisesBootstrap,
	svcName,
	vmName,
	nodeID string,
) *k8scorev1.Service {
	return &k8scorev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       models.ServiceKind,
			APIVersion: models.K8sAPIVersionV1,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcName,
			Namespace: b.K8sObject.GetNamespace(),
			Labels: map[string]string{
				models.ClusterIDLabel: b.ClusterID,
				models.NodeIDLabel:    nodeID,
			},
			Finalizers: []string{models.DeletionFinalizer},
		},
		Spec: k8scorev1.ServiceSpec{
			Ports: b.ExposePorts,
			Selector: map[string]string{
				models.KubevirtDomainLabel: vmName,
				models.NodeIDLabel:         nodeID,
			},
			Type: k8scorev1.ServiceTypeClusterIP,
		},
	}
}
