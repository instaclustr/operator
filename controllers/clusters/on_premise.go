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
	"k8s.io/client-go/tools/record"
	virtcorev1 "kubevirt.io/api/core/v1"
	cdiv1beta1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/instaclustr/operator/apis/clusters/v1beta1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/scheduler"
)

type onPremiseBootstrap struct {
	IcAdminAPI            instaclustr.IcadminAPI
	K8sClient             client.Client
	EventRecorder         record.EventRecorder
	K8sObject             client.Object
	ClusterID             string
	CdcID                 string
	OnPremisesSpec        *v1beta1.OnPremisesSpec
	ExposePorts           []k8scorev1.ServicePort
	HeadlessPorts         []k8scorev1.ServicePort
	PrivateNetworkCluster bool
}

func newOnPremiseBootstrap(
	icAdminAPI instaclustr.IcadminAPI,
	k8sClient client.Client,
	e record.EventRecorder,
	o client.Object,
	clusterID,
	cdcID string,
	onPremisesSpec *v1beta1.OnPremisesSpec,
	exposePorts,
	headlessPorts []k8scorev1.ServicePort,
	privateNetworkCluster bool,
) *onPremiseBootstrap {
	return &onPremiseBootstrap{
		IcAdminAPI:            icAdminAPI,
		K8sClient:             k8sClient,
		EventRecorder:         e,
		K8sObject:             o,
		ClusterID:             clusterID,
		CdcID:                 cdcID,
		OnPremisesSpec:        onPremisesSpec,
		ExposePorts:           exposePorts,
		HeadlessPorts:         headlessPorts,
		PrivateNetworkCluster: privateNetworkCluster,
	}
}

func createDV(
	ctx context.Context,
	bootstrap *onPremiseBootstrap,
	name,
	nodeID string,
	size resource.Quantity,
	isOSDisk bool,
) (*cdiv1beta1.DataVolume, error) {
	ns := bootstrap.K8sObject.GetNamespace()
	dv := &cdiv1beta1.DataVolume{}
	pvc := &k8scorev1.PersistentVolumeClaim{}
	err := bootstrap.K8sClient.Get(ctx, types.NamespacedName{
		Namespace: ns,
		Name:      name,
	}, pvc)
	if client.IgnoreNotFound(err) != nil {
		return nil, err
	}
	if k8serrors.IsNotFound(err) {
		err = bootstrap.K8sClient.Get(ctx, types.NamespacedName{
			Namespace: ns,
			Name:      name,
		}, dv)
		if client.IgnoreNotFound(err) != nil {
			return nil, err
		}
		if k8serrors.IsNotFound(err) {
			dv = newDataDiskDV(bootstrap, name, nodeID, size, isOSDisk)
			err = bootstrap.K8sClient.Create(ctx, dv)
			if err != nil {
				return nil, err
			}
		}
	}

	return dv, nil
}

func newDataDiskDV(
	bootstrap *onPremiseBootstrap,
	name,
	nodeID string,
	storageSize resource.Quantity,
	isOSDisk bool,
) *cdiv1beta1.DataVolume {
	dvSource := &cdiv1beta1.DataVolumeSource{}

	if isOSDisk {
		dvSource.HTTP = &cdiv1beta1.DataVolumeSourceHTTP{URL: bootstrap.OnPremisesSpec.OSImageURL}
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
			Namespace: bootstrap.K8sObject.GetNamespace(),
			Labels: map[string]string{
				models.ClusterIDLabel: bootstrap.ClusterID,
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
				StorageClassName: &bootstrap.OnPremisesSpec.StorageClassName,
			},
		},
	}
}

func reconcileIgnitionScriptSecret(
	ctx context.Context,
	bootstrap *onPremiseBootstrap,
	nodeName,
	nodeID,
	nodeRack string,
) (string, error) {
	ns := bootstrap.K8sObject.GetNamespace()
	ignitionSecretName := fmt.Sprintf("%s-%s", models.IgnitionScriptSecretPrefix, nodeName)
	ignitionSecret := &k8scorev1.Secret{}
	err := bootstrap.K8sClient.Get(ctx, types.NamespacedName{
		Namespace: ns,
		Name:      ignitionSecretName,
	}, ignitionSecret)
	if client.IgnoreNotFound(err) != nil {
		return "", err
	}
	if k8serrors.IsNotFound(err) {
		script, err := bootstrap.IcAdminAPI.GetIgnitionScript(nodeID)
		if err != nil {
			return "", err
		}

		ignitionSecret = &k8scorev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind:       models.SecretKind,
				APIVersion: models.K8sAPIVersionV1,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      ignitionSecretName,
				Namespace: ns,
				Labels: map[string]string{
					models.ControlledByLabel: bootstrap.K8sObject.GetName(),
					models.ClusterIDLabel:    bootstrap.ClusterID,
					models.NodeIDLabel:       nodeID,
					models.NodeRackLabel:     nodeRack,
				},
				Finalizers: []string{models.DeletionFinalizer},
			},
			StringData: map[string]string{
				models.Script: script,
			},
		}
		err = bootstrap.K8sClient.Create(ctx, ignitionSecret)
		if err != nil {
			return "", err
		}
	}

	return ignitionSecretName, nil
}

func newVM(
	ctx context.Context,
	bootstrap *onPremiseBootstrap,
	vmName,
	nodeID,
	nodeRack,
	OSDiskDVName,
	ignitionSecretName string,
	cpu,
	memory resource.Quantity,
	storageDVNames ...string,
) (*virtcorev1.VirtualMachine, error) {
	runStrategy := virtcorev1.RunStrategyAlways
	bootOrder1 := uint(1)

	cloudInitSecret := &k8scorev1.Secret{}
	err := bootstrap.K8sClient.Get(ctx, types.NamespacedName{
		Namespace: bootstrap.OnPremisesSpec.CloudInitScriptRef.Namespace,
		Name:      bootstrap.OnPremisesSpec.CloudInitScriptRef.Name,
	}, cloudInitSecret)
	if err != nil {
		return nil, err
	}

	vm := &virtcorev1.VirtualMachine{
		TypeMeta: metav1.TypeMeta{
			Kind:       models.VirtualMachineKind,
			APIVersion: models.KubevirtV1APIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      vmName,
			Namespace: bootstrap.K8sObject.GetNamespace(),
			Labels: map[string]string{
				models.ClusterIDLabel:      bootstrap.ClusterID,
				models.NodeIDLabel:         nodeID,
				models.NodeRackLabel:       nodeRack,
				models.KubevirtDomainLabel: vmName,
			},
			Finalizers: []string{models.DeletionFinalizer},
		},
		Spec: virtcorev1.VirtualMachineSpec{
			RunStrategy: &runStrategy,
			Template: &virtcorev1.VirtualMachineInstanceTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						models.ClusterIDLabel:      bootstrap.ClusterID,
						models.NodeIDLabel:         nodeID,
						models.NodeRackLabel:       nodeRack,
						models.KubevirtDomainLabel: vmName,
					},
				},
				Spec: virtcorev1.VirtualMachineInstanceSpec{
					Hostname:  vmName,
					Subdomain: fmt.Sprintf("%s-%s", models.KubevirtSubdomain, bootstrap.K8sObject.GetName()),
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
								{
									Name:       models.IgnitionDisk,
									DiskDevice: virtcorev1.DiskDevice{},
									Serial:     models.IgnitionSerial,
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
										Name: bootstrap.OnPremisesSpec.CloudInitScriptRef.Name,
									},
								},
							},
						},
						{
							Name: models.IgnitionDisk,
							VolumeSource: virtcorev1.VolumeSource{
								Secret: &virtcorev1.SecretVolumeSource{
									SecretName: ignitionSecretName,
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

func newExposeService(
	bootstrap *onPremiseBootstrap,
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
			Namespace: bootstrap.K8sObject.GetNamespace(),
			Labels: map[string]string{
				models.ClusterIDLabel: bootstrap.ClusterID,
				models.NodeIDLabel:    nodeID,
			},
			Finalizers: []string{models.DeletionFinalizer},
		},
		Spec: k8scorev1.ServiceSpec{
			Ports: bootstrap.ExposePorts,
			Selector: map[string]string{
				models.KubevirtDomainLabel: vmName,
				models.NodeIDLabel:         nodeID,
			},
			Type: models.LBType,
		},
	}
}

func newHeadlessService(bootstrap *onPremiseBootstrap, svcName string) *k8scorev1.Service {
	return &k8scorev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       models.ServiceKind,
			APIVersion: models.K8sAPIVersionV1,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcName,
			Namespace: bootstrap.K8sObject.GetNamespace(),
			Labels: map[string]string{
				models.ClusterIDLabel: bootstrap.ClusterID,
			},
			//Finalizers: []string{models.DeletionFinalizer},
		},
		Spec: k8scorev1.ServiceSpec{
			ClusterIP: "None",
			Ports:     bootstrap.HeadlessPorts,
			Selector: map[string]string{
				models.ClusterIDLabel: bootstrap.ClusterID,
			},
		},
	}
}

func reconcileOnPremResources(ctx context.Context, bootstrap *onPremiseBootstrap) error {
	if bootstrap.PrivateNetworkCluster {
		err := reconcileSSHGatewayResources(ctx, bootstrap)
		if err != nil {
			return err
		}
	}

	err := reconcileNodesResources(ctx, bootstrap)
	if err != nil {
		return err
	}

	return nil
}

func reconcileSSHGatewayResources(ctx context.Context, bootstrap *onPremiseBootstrap) error {
	gateways, err := bootstrap.IcAdminAPI.GetGateways(bootstrap.CdcID)
	if err != nil {
		return err
	}

	for i, gateway := range gateways {
		gatewayDVSize, err := resource.ParseQuantity(bootstrap.OnPremisesSpec.OSDiskSize)
		if err != nil {
			return err
		}

		gatewayDVName := fmt.Sprintf("%s-%d-%s", models.GatewayDVPrefix, i, strings.ToLower(bootstrap.K8sObject.GetName()))
		gatewayDV, err := createDV(ctx, bootstrap, gatewayDVName, gateway.ID, gatewayDVSize, true)
		if err != nil {
			return err
		}

		gatewayCPU := resource.Quantity{}
		gatewayCPU.Set(bootstrap.OnPremisesSpec.SSHGatewayCPU)

		gatewayMemory, err := resource.ParseQuantity(bootstrap.OnPremisesSpec.SSHGatewayMemory)
		if err != nil {
			return err
		}

		gatewayVMName := fmt.Sprintf("%s-%d-%s", models.GatewayVMPrefix, i, strings.ToLower(bootstrap.K8sObject.GetName()))
		secretName, err := reconcileIgnitionScriptSecret(
			ctx,
			bootstrap,
			gatewayVMName,
			gateway.ID,
			gateway.Rack)
		if err != nil {
			return err
		}

		gatewayVM := &virtcorev1.VirtualMachine{}
		err = bootstrap.K8sClient.Get(ctx, types.NamespacedName{
			Namespace: bootstrap.K8sObject.GetNamespace(),
			Name:      gatewayVMName,
		}, gatewayVM)
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		if k8serrors.IsNotFound(err) {
			gatewayVM, err = newVM(
				ctx,
				bootstrap,
				gatewayVMName,
				gateway.ID,
				gateway.Rack,
				gatewayDV.Name,
				secretName,
				gatewayCPU,
				gatewayMemory)
			if err != nil {
				return err
			}
			err = bootstrap.K8sClient.Create(ctx, gatewayVM)
			if err != nil {
				return err
			}
		}

		gatewaySvcName := fmt.Sprintf("%s-%s", models.GatewaySvcPrefix, gatewayVMName)
		gatewayExposeService := &k8scorev1.Service{}
		err = bootstrap.K8sClient.Get(ctx, types.NamespacedName{
			Namespace: bootstrap.K8sObject.GetNamespace(),
			Name:      gatewaySvcName,
		}, gatewayExposeService)
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		if k8serrors.IsNotFound(err) {
			gatewayExposeService = newExposeService(bootstrap, gatewaySvcName, gatewayVMName, gateway.ID)

			err = bootstrap.K8sClient.Create(ctx, gatewayExposeService)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func reconcileNodesResources(ctx context.Context, bootstrap *onPremiseBootstrap) error {
	nodes, err := bootstrap.IcAdminAPI.GetOnPremisesNodes(bootstrap.ClusterID)
	if err != nil {
		return err
	}

	for i, node := range nodes {
		nodeOSDiskSize, err := resource.ParseQuantity(bootstrap.OnPremisesSpec.OSDiskSize)
		if err != nil {
			return err
		}

		clusterName := strings.ToLower(bootstrap.K8sObject.GetName())
		nodeOSDiskDVName := fmt.Sprintf("%s-%d-%s", models.NodeOSDVPrefix, i, clusterName)
		nodeOSDV, err := createDV(ctx, bootstrap, nodeOSDiskDVName, node.ID, nodeOSDiskSize, true)
		if err != nil {
			return err
		}

		nodeDataDiskDVSize, err := resource.ParseQuantity(bootstrap.OnPremisesSpec.DataDiskSize)
		if err != nil {
			return err
		}

		nodeDataDiskDVName := fmt.Sprintf("%s-%d-%s", models.NodeDVPrefix, i, clusterName)
		nodeDataDV, err := createDV(ctx, bootstrap, nodeDataDiskDVName, node.ID, nodeDataDiskDVSize, false)
		if err != nil {
			return err
		}

		nodeCPU := resource.Quantity{}
		nodeCPU.Set(bootstrap.OnPremisesSpec.NodeCPU)

		nodeMemory, err := resource.ParseQuantity(bootstrap.OnPremisesSpec.NodeMemory)
		if err != nil {
			return err
		}

		nodeName := fmt.Sprintf("%s-%d-%s", models.NodeVMPrefix, i, clusterName)

		secretName, err := reconcileIgnitionScriptSecret(ctx, bootstrap, nodeName, node.ID, node.Rack)
		if err != nil {
			return err
		}

		ns := bootstrap.K8sObject.GetNamespace()
		nodeVM := &virtcorev1.VirtualMachine{}
		err = bootstrap.K8sClient.Get(ctx, types.NamespacedName{
			Namespace: ns,
			Name:      nodeName,
		}, nodeVM)
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		if k8serrors.IsNotFound(err) {
			nodeVM, err = newVM(
				ctx,
				bootstrap,
				nodeName,
				node.ID,
				node.Rack,
				nodeOSDV.Name,
				secretName,
				nodeCPU,
				nodeMemory,
				nodeDataDV.Name)
			if err != nil {
				return err
			}
			err = bootstrap.K8sClient.Create(ctx, nodeVM)
			if err != nil {
				return err
			}
		}

		if !bootstrap.PrivateNetworkCluster {
			nodeExposeName := fmt.Sprintf("%s-%s", models.NodeSvcPrefix, nodeName)
			nodeExposeService := &k8scorev1.Service{}
			err = bootstrap.K8sClient.Get(ctx, types.NamespacedName{
				Namespace: ns,
				Name:      nodeExposeName,
			}, nodeExposeService)
			if client.IgnoreNotFound(err) != nil {
				return err
			}
			if k8serrors.IsNotFound(err) {
				nodeExposeService = newExposeService(bootstrap, nodeExposeName, nodeName, node.ID)
				err = bootstrap.K8sClient.Create(ctx, nodeExposeService)
				if err != nil {
					return err
				}
			}
		}

		headlessServiceName := fmt.Sprintf("%s-%s", models.KubevirtSubdomain, clusterName)
		headlessSVC := &k8scorev1.Service{}
		err = bootstrap.K8sClient.Get(ctx, types.NamespacedName{
			Namespace: ns,
			Name:      headlessServiceName,
		}, headlessSVC)
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		if k8serrors.IsNotFound(err) {
			headlessSVC = newHeadlessService(bootstrap, headlessServiceName)
			err = bootstrap.K8sClient.Create(ctx, headlessSVC)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

func deleteOnPremResources(ctx context.Context, K8sClient client.Client, clusterID, namespace string) error {
	vms := &virtcorev1.VirtualMachineList{}
	err := K8sClient.List(ctx, vms, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			models.ClusterIDLabel: clusterID,
		}),
		Namespace: namespace,
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
		Namespace: namespace,
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
		Namespace: namespace,
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
		Namespace: namespace,
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

	secrets := &k8scorev1.SecretList{}
	err = K8sClient.List(ctx, secrets, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			models.ClusterIDLabel: clusterID,
		}),
		Namespace: namespace,
	})
	if err != nil {
		return err
	}

	for _, secret := range secrets.Items {
		err = K8sClient.Delete(ctx, &secret)
		if err != nil {
			return err
		}

		patch := client.MergeFrom(secret.DeepCopy())
		controllerutil.RemoveFinalizer(&secret, models.DeletionFinalizer)
		err = K8sClient.Patch(ctx, &secret, patch)
		if err != nil {
			return err
		}
	}

	return nil
}

func newWatchOnPremisesIPsJob(b *onPremiseBootstrap) scheduler.Job {
	l := log.Log.WithValues("component", "cassandraOnPremStatusClusterJob")
	return func() error {
		if b.PrivateNetworkCluster {
			gateways, err := b.IcAdminAPI.GetGateways(b.CdcID)
			if err != nil {
				l.Error(err, "Cannot get Cassandra SSH-gateway nodes from the Instaclustr API",
					"cluster name", b.K8sObject.GetName(),
					"data centre ID", b.CdcID)
				b.EventRecorder.Eventf(b.K8sObject, models.Warning, models.FetchFailed,
					"SSH-gateway nodes fetch from the Instaclustr API is failed. Reason: %v", err)
				return err
			}

			for _, gateway := range gateways {
				gatewayPods := &k8scorev1.PodList{}
				err = b.K8sClient.List(context.Background(), gatewayPods, &client.ListOptions{
					LabelSelector: labels.SelectorFromSet(map[string]string{
						models.ClusterIDLabel: b.ClusterID,
						models.NodeIDLabel:    gateway.ID,
					}),
					Namespace: b.K8sObject.GetNamespace(),
				})
				if err != nil {
					l.Error(err, "Cannot list SSH-gateway pods",
						"cluster name", b.K8sObject.GetName(),
						"clusterID", b.ClusterID)

					b.EventRecorder.Eventf(b.K8sObject, models.Warning, models.CreationFailed,
						"Fetching SSH-gateways is failed. Reason: %v", err)
					return err
				}

				for _, pod := range gatewayPods.Items {
					if (pod.Status.PodIP != "" && gateway.PrivateAddress == "") ||
						(pod.Status.PodIP != "" && pod.Status.PodIP != gateway.PrivateAddress) {
						err = b.IcAdminAPI.SetPrivateGatewayIP(b.CdcID, pod.Status.PodIP)
						if err != nil {
							l.Error(err, "Cannot set Private IP for the SSH-gateway node",
								"cluster name", b.K8sObject.GetName(),
								"clusterID", b.ClusterID)

							b.EventRecorder.Eventf(b.K8sObject, models.Warning, models.CreationFailed,
								"Setting Private IP for the SSH-gateway node is failed. Reason: %v", err)
							return err
						}
					}
				}

				gatewaySVCs := &k8scorev1.ServiceList{}
				err = b.K8sClient.List(context.Background(), gatewaySVCs, &client.ListOptions{
					LabelSelector: labels.SelectorFromSet(map[string]string{
						models.ClusterIDLabel: b.ClusterID,
						models.NodeIDLabel:    gateway.ID,
					}),
					Namespace: b.K8sObject.GetNamespace(),
				})
				if err != nil {
					l.Error(err, "Cannot get services backed by SSH-gateway pods",
						"cluster name", b.K8sObject.GetName(),
						"clusterID", b.ClusterID)
					b.EventRecorder.Eventf(b.K8sObject, models.Warning, models.CreationFailed,
						"Fetching services backed by SSH-gateway pods is failed. Reason: %v", err)
					return err
				}

				for _, svc := range gatewaySVCs.Items {
					if (svc.Status.LoadBalancer.Ingress[0].IP != "" && gateway.PublicAddress == "") ||
						(svc.Status.LoadBalancer.Ingress[0].IP != gateway.PublicAddress) {
						err = b.IcAdminAPI.SetPublicGatewayIP(b.CdcID, svc.Status.LoadBalancer.Ingress[0].IP)
						if err != nil {
							l.Error(err, "Cannot set Public IP for the SSH-gateway node",
								"cluster name", b.K8sObject.GetName(),
								"clusterID", b.ClusterID)

							b.EventRecorder.Eventf(b.K8sObject, models.Warning, models.CreationFailed,
								"Setting Public IP for the SSH-gateway node is failed. Reason: %v", err)
							return err
						}

						l.Info("IPs for on-premises cluster ssh-gateway are set",
							"cluster name", b.K8sObject.GetName(),
							"clusterID", b.ClusterID)

						b.EventRecorder.Eventf(b.K8sObject, models.Normal, models.Created, "SSH-gateway IPs are set")
					}
				}
			}
		}

		request := &v1beta1.OnPremiseNode{}
		nodes, err := b.IcAdminAPI.GetOnPremisesNodes(b.ClusterID)
		if err != nil {
			l.Error(err, "Cannot get Cassandra on-premises nodes from the Instaclustr API",
				"cluster name", b.K8sObject.GetName(),
				"clusterID", b.ClusterID)
			b.EventRecorder.Eventf(b.K8sObject, models.Warning, models.FetchFailed,
				"On-premises nodes fetch from the Instaclustr API is failed. Reason: %v", err)
			return err
		}

		for _, node := range nodes {
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
					"clusterID", b.ClusterID)

				b.EventRecorder.Eventf(b.K8sObject, models.Warning, models.CreationFailed,
					"Fetching on-premises cluster pods is failed. Reason: %v", err)
				return err
			}

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
					"clusterID", b.ClusterID)

				b.EventRecorder.Eventf(b.K8sObject, models.Warning, models.CreationFailed,
					"Fetching services backed by on-premises cluster pods is failed. Reason: %v", err)
				return err
			}

			for _, pod := range nodePods.Items {
				if (pod.Status.PodIP != "" && node.PrivateAddress == "") ||
					(pod.Status.PodIP != "" && pod.Status.PodIP != node.PrivateAddress) {
					request.PrivateAddress = pod.Status.PodIP
				}
			}

			for _, svc := range nodeSVCs.Items {
				if (svc.Status.LoadBalancer.Ingress[0].IP != "" && node.PublicAddress == "") ||
					(svc.Status.LoadBalancer.Ingress[0].IP != node.PublicAddress) {
					request.PublicAddress = svc.Status.LoadBalancer.Ingress[0].IP
				}
			}

			if request.PublicAddress != "" || request.PrivateAddress != "" {
				err = b.IcAdminAPI.SetNodeIPs(node.ID, request)
				if err != nil {
					l.Error(err, "Cannot set IPs for on-premises cluster nodes",
						"cluster name", b.K8sObject.GetName(),
						"clusterID", b.ClusterID)

					b.EventRecorder.Eventf(b.K8sObject, models.Warning, models.CreationFailed,
						"Setting IPs for on-premises cluster nodes is failed. Reason: %v", err)
					return err
				}

				l.Info("IPs for on-premises cluster node are set",
					"cluster name", b.K8sObject.GetName(),
					"clusterID", b.ClusterID,
					"request", request)
				b.EventRecorder.Eventf(b.K8sObject, models.Normal, models.Created, "Nodes IPs are set")
			}
		}

		return nil
	}
}
