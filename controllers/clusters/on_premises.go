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
	k8scorev1 "k8s.io/api/core/v1"
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

type onPremisesBootstrap struct {
	K8sClient             client.Client
	K8sObject             client.Object
	EventRecorder         record.EventRecorder
	ClusterStatus         v1beta1.ClusterStatus
	OnPremisesSpec        *v1beta1.OnPremisesSpec
	ExposePorts           []k8scorev1.ServicePort
	HeadlessPorts         []k8scorev1.ServicePort
	PrivateNetworkCluster bool
}

func newOnPremisesBootstrap(
	k8sClient client.Client,
	o client.Object,
	e record.EventRecorder,
	status v1beta1.ClusterStatus,
	onPremisesSpec *v1beta1.OnPremisesSpec,
	exposePorts,
	headlessPorts []k8scorev1.ServicePort,
	privateNetworkCluster bool,
) *onPremisesBootstrap {
	return &onPremisesBootstrap{
		K8sClient:             k8sClient,
		K8sObject:             o,
		EventRecorder:         e,
		ClusterStatus:         status,
		OnPremisesSpec:        onPremisesSpec,
		ExposePorts:           exposePorts,
		HeadlessPorts:         headlessPorts,
		PrivateNetworkCluster: privateNetworkCluster,
	}
}

func handleCreateOnPremisesClusterResources(ctx context.Context, b *onPremisesBootstrap) error {
	if len(b.ClusterStatus.DataCentres) < 1 {
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

func handleNodeReplace(ctx context.Context, b *onPremisesBootstrap, newNode, oldNode *v1beta1.Node) error {
	if len(b.ClusterStatus.DataCentres) == 0 {
		return models.ErrZeroDataCentres
	}

	if newNode == nil ||
		oldNode == nil {
		return models.ErrNoNodeToReplace
	}

	objectName := b.K8sObject.GetName()
	vm := &virtcorev1.VirtualMachine{}
	for i := 0; i < len(b.ClusterStatus.DataCentres[0].Nodes); i++ {
		nodeName := fmt.Sprintf("%s-%d-%s", models.NodeVMPrefix, i, objectName)
		err := b.K8sClient.Get(ctx, types.NamespacedName{
			Namespace: b.K8sObject.GetNamespace(),
			Name:      nodeName,
		}, vm)
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		if err == nil {
			continue
		}

		osDVName := fmt.Sprintf("%s-%d-%s", models.NodeOSDVPrefix, i, objectName)
		osDV, err := reconcileDV(ctx, b, newNode.ID, osDVName, models.OSDVLabel)
		if err != nil {
			return err
		}

		storageDVList := &cdiv1beta1.DataVolumeList{}
		err = b.K8sClient.List(ctx, storageDVList, &client.ListOptions{
			LabelSelector: labels.SelectorFromSet(map[string]string{
				models.NodeIDLabel: oldNode.ID,
				models.DVRoleLabel: models.StorageDVLabel,
			}),
		})
		if err != nil {
			return err
		}

		storageDV := &cdiv1beta1.DataVolume{}
		storageDVName := fmt.Sprintf("%s-%d-%s", models.NodeDVPrefix, i, objectName)
		err = b.K8sClient.Get(ctx, types.NamespacedName{
			Namespace: b.K8sObject.GetNamespace(),
			Name:      storageDVName,
		}, storageDV)
		if err != nil {
			return err
		}

		patch := client.MergeFrom(storageDV.DeepCopy())
		storageDV.Labels[models.NodeIDLabel] = newNode.ID
		err = b.K8sClient.Patch(ctx, storageDV, patch)
		if err != nil {
			return err
		}

		svc := &k8scorev1.Service{}
		svcName := fmt.Sprintf("%s-%s", models.ClusterIPServiceRoleLabel, nodeName)
		err = b.K8sClient.Get(ctx, types.NamespacedName{
			Namespace: b.K8sObject.GetNamespace(),
			Name:      svcName,
		}, svc)
		if err != nil {
			return err
		}

		patch = client.MergeFrom(svc.DeepCopy())
		svc.Labels[models.NodeIDLabel] = newNode.ID
		err = b.K8sClient.Patch(ctx, svc, patch)
		if err != nil {
			return err
		}

		_, err = reconcileVM(ctx, b, newNode.ID, nodeName, newNode.Rack, osDV.Name, storageDV.Name)
		if err != nil {
			return err
		}

		break
	}

	return nil
}

func reconcileSSHGatewayResources(ctx context.Context, b *onPremisesBootstrap) error {
	gatewayDVName := fmt.Sprintf("%s-%s", models.GatewayDVPrefix, b.K8sObject.GetName())
	gatewayDV, err := reconcileDV(ctx, b, b.ClusterStatus.DataCentres[0].ID, gatewayDVName, models.OSDVLabel)
	if err != nil {
		return err
	}

	gatewayName := fmt.Sprintf("%s-%s", models.GatewayVMPrefix, b.K8sObject.GetName())
	gatewayVM, err := reconcileVM(
		ctx,
		b,
		b.ClusterStatus.DataCentres[0].ID,
		gatewayName,
		models.GatewayRack,
		gatewayDV.Name)
	if err != nil {
		return err
	}

	_, err = reconcileService(ctx, b, b.ClusterStatus.DataCentres[0].ID, gatewayVM.Name, models.ExposeServiceRoleLabel)
	if err != nil {
		return err
	}

	_, err = reconcileService(ctx, b, b.ClusterStatus.DataCentres[0].ID, gatewayVM.Name, models.ClusterIPServiceRoleLabel)
	if err != nil {
		return err
	}

	return nil
}

func reconcileNodesResources(ctx context.Context, b *onPremisesBootstrap) error {
	for i, node := range b.ClusterStatus.DataCentres[0].Nodes {
		objectName := b.K8sObject.GetName()
		newOSDVName := fmt.Sprintf("%s-%d-%s", models.NodeOSDVPrefix, i, objectName)
		nodeOSDV, err := reconcileDV(ctx, b, node.ID, newOSDVName, models.OSDVLabel)
		if err != nil {
			return err
		}

		nodeDataDiskDVName := fmt.Sprintf("%s-%d-%s", models.NodeDVPrefix, i, objectName)
		nodeDataDV, err := reconcileDV(ctx, b, node.ID, nodeDataDiskDVName, models.StorageDVLabel)
		if err != nil {
			return err
		}

		nodeName := fmt.Sprintf("%s-%d-%s", models.NodeVMPrefix, i, objectName)
		nodeVM, err := reconcileVM(ctx, b, node.ID, nodeName, node.Rack, nodeOSDV.Name, nodeDataDV.Name)
		if err != nil {
			return err
		}

		if !b.PrivateNetworkCluster {
			_, err = reconcileService(ctx, b, node.ID, nodeVM.Name, models.ExposeServiceRoleLabel)
			if err != nil {
				return err
			}
		}

		_, err = reconcileService(ctx, b, node.ID, nodeVM.Name, models.HeadlessServiceRoleLabel)
		if err != nil {
			return err
		}

		_, err = reconcileService(ctx, b, node.ID, nodeVM.Name, models.ClusterIPServiceRoleLabel)
		if err != nil {
			return err
		}
	}

	return nil
}

func reconcileVM(
	ctx context.Context,
	b *onPremisesBootstrap,
	nodeID,
	nodeName,
	rack,
	osDVName string,
	storageDVName ...string) (*virtcorev1.VirtualMachine, error) {
	nodeCPU := resource.Quantity{}
	var memory string
	var err error

	switch rack {
	case models.GatewayRack:
		nodeCPU.Set(b.OnPremisesSpec.SSHGatewayCPU)
		memory = b.OnPremisesSpec.SSHGatewayMemory
	default:
		nodeCPU.Set(b.OnPremisesSpec.NodeCPU)
		memory = b.OnPremisesSpec.NodeMemory
	}

	vmList := &virtcorev1.VirtualMachineList{}
	err = b.K8sClient.List(ctx, vmList, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			models.NodeIDLabel: nodeID,
		}),
	})
	if err != nil {
		return nil, err
	}

	nodeMemory, err := resource.ParseQuantity(memory)
	if err != nil {
		return nil, err
	}

	if len(vmList.Items) == 0 {
		vm, err := newVM(
			ctx,
			b,
			nodeName,
			nodeID,
			rack,
			osDVName,
			nodeCPU,
			nodeMemory,
			storageDVName...,
		)
		if err != nil {
			return nil, err
		}

		err = b.K8sClient.Create(ctx, vm)
		if err != nil {

			return nil, err
		}

		return vm, nil
	}

	return &vmList.Items[0], nil
}

func reconcileDV(ctx context.Context, b *onPremisesBootstrap, nodeID, diskName, diskRole string) (*cdiv1beta1.DataVolume, error) {
	dvList := &cdiv1beta1.DataVolumeList{}
	err := b.K8sClient.List(ctx, dvList, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			models.NodeIDLabel: nodeID,
			models.DVRoleLabel: diskRole,
		}),
	})
	if err != nil {
		return nil, err
	}

	if len(dvList.Items) == 0 {
		dvSize := resource.Quantity{}
		switch diskRole {
		case models.OSDVLabel:
			dvSize, err = resource.ParseQuantity(b.OnPremisesSpec.OSDiskSize)
			if err != nil {
				return nil, err
			}
		case models.StorageDVLabel:
			dvSize, err = resource.ParseQuantity(b.OnPremisesSpec.DataDiskSize)
			if err != nil {
				return nil, err
			}
		}

		dv, err := createDV(
			ctx,
			b,
			diskName,
			nodeID,
			diskRole,
			dvSize,
		)
		if err != nil {
			return nil, err
		}

		return dv, nil
	}

	return &dvList.Items[0], nil
}

func reconcileService(ctx context.Context, b *onPremisesBootstrap, nodeID, nodeName, role string) (*k8scorev1.Service, error) {
	svcList := &k8scorev1.ServiceList{}
	svcLabels := map[string]string{
		models.ServiceRoleLabel: role,
		models.ClusterIDLabel:   b.ClusterStatus.ID,
	}

	var svcName string
	if role == models.HeadlessServiceRoleLabel {
		svcName = fmt.Sprintf("%s-%s", role, b.K8sObject.GetName())
	} else {
		svcName = fmt.Sprintf("%s-%s", role, nodeName)
		svcLabels[models.NodeIDLabel] = nodeID
	}

	err := b.K8sClient.List(ctx, svcList, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(svcLabels),
	})
	if err != nil {
		return nil, err
	}

	svc := &k8scorev1.Service{}
	if len(svcList.Items) == 0 {
		switch role {
		case models.ExposeServiceRoleLabel:
			svc = newExposeService(
				b,
				svcName,
				nodeName,
				nodeID,
			)
		case models.HeadlessServiceRoleLabel:
			svc = newHeadlessService(
				b,
				svcName,
			)
		case models.ClusterIPServiceRoleLabel:
			svc = newClusterIPService(
				b,
				svcName,
				nodeName,
				nodeID,
			)
		}
		err = b.K8sClient.Create(ctx, svc)
		if err != nil {
			return nil, err
		}

		return svc, nil
	}

	err = b.K8sClient.Get(ctx, types.NamespacedName{
		Namespace: b.K8sObject.GetNamespace(),
		Name:      svcName,
	}, svc)
	if err != nil {
		return nil, err
	}

	return svc, nil
}

func createDV(
	ctx context.Context,
	b *onPremisesBootstrap,
	name,
	nodeID,
	diskRole string,
	size resource.Quantity,
) (*cdiv1beta1.DataVolume, error) {
	dv := newDataDiskDV(
		b,
		name,
		nodeID,
		diskRole,
		size,
	)
	err := b.K8sClient.Create(ctx, dv)
	if err != nil {
		return nil, err
	}

	return dv, nil
}

func newDataDiskDV(
	b *onPremisesBootstrap,
	name,
	nodeID,
	diskRole string,
	storageSize resource.Quantity,
) *cdiv1beta1.DataVolume {
	dvSource := &cdiv1beta1.DataVolumeSource{}
	dvLabels := map[string]string{
		models.ClusterIDLabel: b.ClusterStatus.ID,
		models.NodeIDLabel:    nodeID,
		models.DVRoleLabel:    diskRole,
	}

	switch diskRole {
	case models.OSDVLabel:
		dvSource.HTTP = &cdiv1beta1.DataVolumeSourceHTTP{URL: b.OnPremisesSpec.OSImageURL}
	case models.StorageDVLabel:
		dvSource.Blank = &cdiv1beta1.DataVolumeBlankImage{}
	}

	return &cdiv1beta1.DataVolume{
		TypeMeta: metav1.TypeMeta{
			Kind:       models.DVKind,
			APIVersion: models.CDIKubevirtV1beta1APIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			Namespace:  b.K8sObject.GetNamespace(),
			Labels:     dvLabels,
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
		models.ClusterIDLabel:      b.ClusterStatus.ID,
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
								DataVolume: &virtcorev1.DataVolumeSource{
									Name: OSDiskDVName,
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
				DataVolume: &virtcorev1.DataVolumeSource{
					Name: dvName,
				},
			},
		})
	}

	return vm, nil
}

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
				models.ClusterIDLabel:   b.ClusterStatus.ID,
				models.NodeIDLabel:      nodeID,
				models.ServiceRoleLabel: models.ExposeServiceRoleLabel,
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
				models.ClusterIDLabel:   b.ClusterStatus.ID,
				models.ServiceRoleLabel: models.HeadlessServiceRoleLabel,
			},
			Finalizers: []string{models.DeletionFinalizer},
		},
		Spec: k8scorev1.ServiceSpec{
			ClusterIP: k8scorev1.ClusterIPNone,
			Ports:     b.HeadlessPorts,
			Selector: map[string]string{
				models.ClusterIDLabel: b.ClusterStatus.ID,
				models.NodeLabel:      models.WorkerNode,
			},
		},
	}
}

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

func newWatchOnPremisesIPsJob(ctx context.Context, kind string, b *onPremisesBootstrap) scheduler.Job {
	l := log.Log.WithValues("component", fmt.Sprintf("%sOnPremisesIPsCheckerJob", kind))

	return func() error {
		allNodePods := &k8scorev1.PodList{}
		err := b.K8sClient.List(ctx, allNodePods, &client.ListOptions{
			LabelSelector: labels.SelectorFromSet(map[string]string{
				models.ClusterIDLabel: b.ClusterStatus.ID,
				models.NodeLabel:      models.WorkerNode,
			}),
			Namespace: b.K8sObject.GetNamespace(),
		})
		if err != nil {
			l.Error(err, "Cannot get on-premises cluster pods",
				"cluster name", b.K8sObject.GetName(),
				"clusterID", b.ClusterStatus.ID,
			)

			b.EventRecorder.Eventf(
				b.K8sObject, models.Warning, models.CreationFailed,
				"Fetching on-premises cluster pods is failed. Reason: %v",
				err,
			)
			return err
		}

		if len(allNodePods.Items) != len(b.ClusterStatus.DataCentres[0].Nodes) {
			err = fmt.Errorf("the quantity of pods does not match the number of on-premises nodes")
			l.Error(err, "Cannot compare private IPs for the cluster",
				"cluster name", b.K8sObject.GetName(),
				"clusterID", b.ClusterStatus.ID,
			)

			b.EventRecorder.Eventf(
				b.K8sObject, models.Warning, models.CreationFailed,
				"Comparing of cluster's private IPs is failing. Reason: %v",
				err,
			)
			return err
		}

		for _, node := range b.ClusterStatus.DataCentres[0].Nodes {
			nodePods := &k8scorev1.PodList{}
			err = b.K8sClient.List(ctx, nodePods, &client.ListOptions{
				LabelSelector: labels.SelectorFromSet(map[string]string{
					models.ClusterIDLabel: b.ClusterStatus.ID,
					models.NodeIDLabel:    node.ID,
				}),
				Namespace: b.K8sObject.GetNamespace(),
			})
			if err != nil {
				l.Error(err, "Cannot get on-premises cluster pods",
					"cluster name", b.K8sObject.GetName(),
					"clusterID", b.ClusterStatus.ID,
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
						"clusterID", b.ClusterStatus.ID,
						"nodeID", node.ID,
						"nodeIP", node.PrivateAddress,
						"podIP", pod.Status.PodIP,
					)

					b.EventRecorder.Eventf(
						b.K8sObject, models.Warning, models.CreationFailed,
						"The private IP addresses of the node are not matching. Contact Instaclustr support to resolve the issue. Reason: %v",
						err,
					)
					return err
				}
			}

			if !b.PrivateNetworkCluster {
				nodeSVCs := &k8scorev1.ServiceList{}
				err = b.K8sClient.List(ctx, nodeSVCs, &client.ListOptions{
					LabelSelector: labels.SelectorFromSet(map[string]string{
						models.ClusterIDLabel: b.ClusterStatus.ID,
						models.NodeIDLabel:    node.ID,
					}),
					Namespace: b.K8sObject.GetNamespace(),
				})
				if err != nil {
					l.Error(err, "Cannot get services backed by on-premises cluster pods",
						"cluster name", b.K8sObject.GetName(),
						"clusterID", b.ClusterStatus.ID,
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
								"clusterID", b.ClusterStatus.ID,
								"nodeID", node.ID,
								"nodeIP", node.PrivateAddress,
								"svcIP", svc.Status.LoadBalancer.Ingress[0].IP,
							)

							b.EventRecorder.Eventf(
								b.K8sObject, models.Warning, models.CreationFailed,
								"The public IP addresses of the node are not matching. Contact Instaclustr support to resolve the issue. Reason: %v",
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
				models.ClusterIDLabel:   b.ClusterStatus.ID,
				models.NodeIDLabel:      nodeID,
				models.ServiceRoleLabel: models.ClusterIPServiceRoleLabel,
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

func getReplacedNodes(oldNodesDC, newNodesDC *v1beta1.DataCentreStatus) (removedNode, addedNode *v1beta1.Node) {
	for _, oldNode := range oldNodesDC.Nodes {
		for _, newNode := range newNodesDC.Nodes {
			if newNode.PrivateAddress == "" {
				addedNode = newNode
			}

			if oldNode.ID == newNode.ID {
				removedNode = nil
				continue
			}
			removedNode = oldNode
		}
	}

	return
}

// Tries to delete old node resources and returns NotFound error if all resources already deleted
func deleteReplacedNodeResources(ctx context.Context, k8sClient client.Client, oldID string) (deleteDone bool, err error) {
	vmList := &virtcorev1.VirtualMachineList{}
	err = k8sClient.List(ctx, vmList, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			models.NodeIDLabel: oldID,
		}),
	})
	if err != nil {
		return false, err
	}

	dvList := &cdiv1beta1.DataVolumeList{}
	err = k8sClient.List(ctx, dvList, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			models.NodeIDLabel: oldID,
			models.DVRoleLabel: models.OSDVLabel,
		}),
	})
	if err != nil {
		return false, err
	}

	if len(vmList.Items) == 0 &&
		len(dvList.Items) == 0 {
		return true, nil
	}

	for _, vm := range vmList.Items {
		err = k8sClient.Delete(ctx, &vm)
		if err != nil {
			return false, err
		}

		patch := client.MergeFrom(vm.DeepCopy())
		controllerutil.RemoveFinalizer(&vm, models.DeletionFinalizer)
		err = k8sClient.Patch(ctx, &vm, patch)
		if err != nil {
			return false, err
		}
	}

	for _, dv := range dvList.Items {
		err = k8sClient.Delete(ctx, &dv)
		if err != nil {
			return false, err
		}

		patch := client.MergeFrom(dv.DeepCopy())
		controllerutil.RemoveFinalizer(&dv, models.DeletionFinalizer)
		err = k8sClient.Patch(ctx, &dv, patch)
		if err != nil {
			return false, err
		}
	}

	return false, nil
}
