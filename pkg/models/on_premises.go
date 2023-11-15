package models

const (
	ONPREMISES = "ONPREMISES"
	CLIENTDC   = "CLIENT_DC"

	VirtualMachineKind           = "VirtualMachine"
	DVKind                       = "DataVolume"
	ServiceKind                  = "Service"
	KubevirtV1APIVersion         = "kubevirt.io/v1"
	CDIKubevirtV1beta1APIVersion = "cdi.kubevirt.io/v1beta1"

	KubevirtSubdomain          = "kubevirt"
	KubevirtDomainLabel        = "kubevirt.io/domain"
	NodeIDLabel                = "nodeID"
	NodeRackLabel              = "nodeRack"
	NodeLabel                  = "node"
	NodeOSDVPrefix             = "node-os-data-volume-pvc"
	NodeDVPrefix               = "node-data-volume-pvc"
	NodeVMPrefix               = "node-vm"
	NodeSvcPrefix              = "node-service"
	WorkerNode                 = "worker-node"
	GatewayDVPrefix            = "gateway-data-volume-pvc"
	GatewayVMPrefix            = "gateway-vm"
	GatewaySvcPrefix           = "gateway-service"
	GatewayRack                = "ssh-gateway-rack"
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

	SSH    = "ssh"
	Port22 = 22

	// Cassandra

	CassandraInterNode = "cassandra-inter-node"
	CassandraSSL       = "cassandra-ssl"
	CassandraCQL       = "cassandra-cql"
	CassandraJMX       = "cassandra-jmx"
	Port7000           = 7000
	Port7001           = 7001
	Port7199           = 7199
	Port9042           = 9042

	// Kafka

	KafkaClient       = "kafka-client"
	KafkaControlPlane = "kafka-control-plane"
	KafkaBroker       = "kafka-broker"
	Port9092          = 9092
	Port9093          = 9093
	Port9094          = 9094

	// KafkaConnect

	KafkaConnectAPI = "kafka-connect-API"
	Port8083        = 8083

	// Cadence

	CadenceTChannel = "cadence-tchannel"
	CadenceGRPC     = "cadence-grpc"
	CadenceWeb      = "cadence-web"
	CadenceHHTPAPI  = "cadence-http-api"
	Port7933        = 7933
	Port7833        = 7833
	Port8088        = 8088
	Port443         = 433

	// PostgreSQL

	PostgreSQLDB = "postgresql-db"
	Port5432     = 5432

	// Redis

	RedisDB   = "redis-db"
	RedisBus  = "redis-bus"
	Port6379  = 6379
	Port16379 = 16379
)
