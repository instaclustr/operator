package models

const (
	ONPREMISES        = "ONPREMISES"
	CLIENTDC          = "CLIENT_DC"
	KEEPSTORAGEMODE   = "KEEP_STORAGE"
	NodeReplaceActive = "Active"

	VirtualMachineKind           = "VirtualMachine"
	DVKind                       = "DataVolume"
	ServiceKind                  = "Service"
	KubevirtV1APIVersion         = "kubevirt.io/v1"
	CDIKubevirtV1beta1APIVersion = "cdi.kubevirt.io/v1beta1"

	KubevirtSubdomain          = "kubevirt"
	IPSetAnnotation            = "instaclustr.com/ip-set"
	KubevirtDomainLabel        = "kubevirt.io/domain"
	OperatorLabel              = "instaclustr-operator"
	NodeIDLabel                = "nodeID"
	DVRoleLabel                = "DVRole"
	OSDVRole                   = "OS"
	StorageDVRole              = "Storage"
	NodeRackLabel              = "nodeRack"
	NodeOSDVPrefix             = "node-os-data-volume-pvc"
	NodeDVPrefix               = "node-data-volume-pvc"
	NodeVMPrefix               = "node-vm"
	NodeSvcPrefix              = "node-service"
	GatewayDVPrefix            = "gateway-data-volume-pvc"
	GatewayVMPrefix            = "gateway-vm"
	GatewaySvcPrefix           = "gateway-service"
	IgnitionScriptSecretPrefix = "ignition-script-secret"
	DataDisk                   = "data-disk"

	Boot           = "boot"
	Storage        = "storage"
	CPU            = "cpu"
	Memory         = "memory"
	Virtio         = "virtio"
	Native         = "native"
	None           = "none"
	Script         = "script"
	IgnitionDisk   = "ignition"
	Default        = "default"
	CloudInit      = "cloud-init"
	DataDiskSerial = "DATADISK"
	IgnitionSerial = "IGNITION"

	LBType    = "LoadBalancer"
	SSH       = "ssh"
	InterNode = "inter-node"
	SSL       = "ssl"
	CQLSH     = "cqlsh"
	JMX       = "jmx"

	Port22   = 22
	Port7000 = 7000
	Port7001 = 7001
	Port7199 = 7199
	Port9042 = 9042
)
