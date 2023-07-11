# Kafka cluster management

## Available spec fields

| Field                             | Type                                                                             | Description                                                                                                                                                                                                                                                                                                                                  |
|-----------------------------------|----------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| name                              | string <br /> **required**                                                       | Cluster name. Should have length from 3 to 32 symbols.                                                                                                                                                                                                                                                                                       |
| version                           | string <br /> **required**                                                       | Kafka instance version. <br />**Available versions**: `3.0.2`, `3.1.2`, `2.8.2`.                                                                                                                                                                                                                                                             |
| pciCompliance                     | bool <br /> **required**                                                         | Creates a PCI compliant cluster, see [PCI Compliance](https://www.instaclustr.com/support/documentation/useful-information/pci-compliance/)                                                                                                                                                                                                  |
| privateNetworkCluster             | bool <br /> **required**                                                         | Creates the cluster with private network only, see [Private Network Clusters](https://www.instaclustr.com/support/documentation/useful-information/private-network-clusters/).                                                                                                                                                               |
| slaTier                           | string <br /> **required**                                                       | SLA Tier of the cluster. Non-production clusters may receive lower priority support and reduced SLAs. Production tier is not available when using Developer class nodes. See [SLA Tier](https://www.instaclustr.com/support/documentation/useful-information/sla-tier/) for more information. <br/>**Enum**: `PRODUCTION`, `NON_PRODUCTION`. |
| twoFactorDelete                   | Array of objects ([TwoFactorDelete](#TwoFactorDeleteObject)) <br /> _mutable_    | Contacts that will be contacted when cluster request is sent.                                                                                                                                                                                                                                                                                |
| schemaRegistry                    | Array of objects ([SchemaRegistry](#SchemaRegistryObject))                       | Adds the specified version of Kafka Schema Registry to this Kafka cluster.                                                                                                                                                                                                                                                                   |
| replicationFactor           | int32 <br /> **required**                                                        | Default Replication factor to use for new topic. Also represents the number of racks to use when allocating nodes.                                                                                                                                                                                                                           |
| partitionsNumber                  | int32 <br /> **required**                                                        | Default number of partitions to use when created new topics.                                                                                                                                                                                                                                                                                 |
| restProxy                         | Array of objects ([RestProxy](#RestProxyObject))                                 | Adds the specified version of Kafka REST Proxy to this Kafka cluster.                                                                                                                                                                                                                                                                        |
| allowDeleteTopics                 | bool <br /> **required**                                                         | Allows topics to be deleted via the kafka-topics tool.                                                                                                                                                                                                                                                                                       |
| autoCreateTopics                  | bool <br /> **required**                                                         | Allows topics to be auto created by brokers when messages are published to a non-existent topic.                                                                                                                                                                                                                                             |
| clientToClusterEncryption         | bool <br /> **required**                                                         | Enables Client ⇄ Cluster Encryption.                                                                                                                                                                                                                                                                                                         |
| dataCentres                       | Array of objects ([KafkaDataCentre](#KafkaDataCentreObject)) <br /> **required** | Object fields are described below as a bulleted list.                                                                                                                                                                                                                                                                                        |
| dedicatedZookeeper                | Array of objects ([DedicatedZookeeper](#DedicatedZookeeperObject))               | Provision additional dedicated nodes for Apache Zookeeper to run on. Zookeeper nodes will be co-located with Kafka if this is not provided.                                                                                                                                                                                                  |
| clientBrokerAuthWithMtls          | bool                                                                             | Enables Client ⇄ Broker Authentication with mTLS.                                                                                                                                                                                                                                                                                            |
| clientAuthBrokerWithoutEncryption | bool                                                                             | Enables Client ⇄ Broker Authentication without Encryption.                                                                                                                                                                                                                                                                                   |
| clientAuthBrokerWithEncryption    | bool                                                                             | Enables Client ⇄ Broker Authentication with Encryption.                                                                                                                                                                                                                                                                                      |
| karapaceRestProxy                 | Array of objects ([KarapaceRestProxy](#KarapaceRestProxyObject))                 | Adds the specified version of Kafka Karapace REST Proxy to this Kafka cluster.                                                                                                                                                                                                                                                               |
| karapaceSchemaRegistry            | Array of objects ([KarapaceSchemaRegistry](#KarapaceSchemaRegistryObject))       | Adds the specified version of Kafka Karapace Schema Registry to this Kafka cluster.                                                                                                                                                                                                                                                          |
| bundledUseOnly                    | bool                                                                             | Provision this cluster for [Bundled Use only](https://www.instaclustr.com/support/documentation/cadence/getting-started-with-cadence/bundled-use-only-cluster-deployments/).                                                                                                                                                                 |

### TwoFactorDeleteObject
| Field                   | Type                              | Description                                                                            |
|-------------------------|-----------------------------------|----------------------------------------------------------------------------------------|
| confirmationPhoneNumber | string                            | The phone number which will be contacted when the cluster is requested to be deleted.  |
| confirmationEmail       | string <br /> **required** <br /> | The email address which will be contacted when the cluster is requested to be deleted. |

### SchemaRegistryObject
| Field     | Type                        | Description                                                                                                |
|-----------|-----------------------------|------------------------------------------------------------------------------------------------------------|
| version   | string <br /> **required**  | Adds the specified version of Kafka Schema Registry to the Kafka cluster. **Available versions:** `5.0.0`. |

### RestProxyObject
| Field                                | Type                       | Description                                                                                                                                                          |
|--------------------------------------|----------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| integrateRestProxyWithSchemaRegistry | bool <br /> **required**   | Enables Integration of the REST proxy with a Schema registry.                                                                                                        |
| useLocalSchemaRegistry               | bool                       | Integrates the REST proxy with the Schema registry attached to this cluster. Requires 'integrateRestProxyWithSchemaRegistry' to be 'true'.                           |
| schemaRegistryServerUrl              | string                     | URL of the Kafka schema registry to integrate with. Requires 'integrateRestProxyWithSchemaRegistry' to be 'true' and useLocalSchemaRegistry to be 'false'.           |
| schemaRegistryUsername               | string                     | Username to use when connecting to the Kafka schema registry. Requires 'integrateRestProxyWithSchemaRegistry' to be 'true' and useLocalSchemaRegistry to be 'false'. |
| schemaRegistryPassword               | string                     | Password to use when connecting to the Kafka schema registry. Requires 'integrateRestProxyWithSchemaRegistry' to be 'true' and useLocalSchemaRegistry to be 'false'. |
| version                              | string <br /> **required** | Adds the specified version of Kafka REST Proxy to the Kafka cluster. **Available versions:** `5.0.4`, `5.0.0`.                                                       |

### KafkaDataCentreObject
| Field                   | Type                                                                     | Description                                                                                                                                                                                                                                                                                          |
|-------------------------|--------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| name                    | string <br /> **required**                                               | A logical name for the data centre within a cluster. These names must be unique in the cluster.                                                                                                                                                                                                      |
| region                  | string <br /> **required**                                               | Region of the Data Centre. See the description for node size for a compatible Data Centre for a given node size.                                                                                                                                                                                     |
| cloudProvider           | string <br /> **required**                                               | Name of the cloud provider service in which the Data Centre will be provisioned. <br />**Enum**: `AWS_VPC` `GCP` `AZURE` `AZURE_AZ`.                                                                                                                                                                 |
| accountName             | string                                                                   | For customers running in their own account. Your provider account can be found on the Create Cluster page on the Instaclustr Console, or the "Provider Account" property on any existing cluster. For customers provisioning on Instaclustr's cloud provider accounts, this property may be omitted. |
| cloudProviderSettings   | Array of objects ([CloudProviderSettings](#CloudProviderSettingsObject)) | Cloud provider specific settings for the Data Centre.                                                                                                                                                                                                                                                |
| network                 | string <br /> **required**                                               | The private network address block for the Data Centre specified using CIDR address notation. The network must have a prefix length between /12 and /22 and must be part of a private address space.                                                                                                  |
| nodeSize                | string <br /> **required**<br />_mutable_                                | Size of the nodes provisioned in the Data Centre. Available node sizes, see [Instaclustr API docs NodeSize](https://instaclustr.redoc.ly/Current/tag/Kafka-Cluster-V2#paths/~1cluster-management~1v2~1resources~1applications~1kafka~1clusters~1v2/post!path=dataCentres/nodeSize&t=request).        |
| nodesNumber             | int32 <br /> **required**<br />_mutable_                                 | Total number of nodes in the Data Centre. <br/>Available values: [1…5].                                                                                                                                                                                                                              |
| tags                    | map[string]string                                                        | List of tags to apply to the Data Centre. Tags are metadata labels which allow you to identify, categorise and filter clusters. This can be useful for grouping together clusters into applications, environments, or any category that you require.<br/>**Format**:<br/>tags:<br/>- key: value.     |
| privateLink             | Array of objects ([PrivateLink](#PrivateLinkObject))                     | Create a PrivateLink enabled cluster, see [PrivateLink](https://www.instaclustr.com/support/documentation/useful-information/privatelink/).                                                                                                                                                          |

### PrivateLinkObject
| Field               | Type                                                                     | Description                                                                                                                                                                                                                                                              |
|---------------------|--------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| advertisedHostname  | string <br /> **required**                                               | The hostname to be used to connect to the PrivateLink cluster. `>= 3 characters`                                                                                                                                                                                         |


### CloudProviderSettingsObject
| Field                  | Type   | Description                                                                                                                                                                                                                                                                                                                                                                                                                                 |
|------------------------|--------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| customVirtualNetworkId | string | **AWS**: VPC ID into which the Data Centre will be provisioned. The Data Centre's network allocation must match the IPv4 CIDR block of the specified VPC.<br />**GCP**: Network name or a relative Network or Subnetwork URI e.g. projects/my-project/regions/us-central1/subnetworks/my-subnet. The Data Centre's network allocation must match the IPv4 CIDR block of the specified subnet. <br />Cannot be provided with `resourceGroup` |
| resourceGroup          | string | The name of the Azure Resource Group into which the Data Centre will be provisioned. <br />Cannot be provided with `customVirtualNetworkId` and `diskEncryptionKey`                                                                                                                                                                                                                                                                         |
| diskEncryptionKey      | string | ID of a KMS encryption key to encrypt data on nodes. KMS encryption key must be set in Cluster Resources through the Instaclustr Console before provisioning an encrypted Data Centre. <br />Cannot be provided with `customVirtualNetworkId`                                                                                                                                                                                               |

### DedicatedZookeeperObject
| Field        | Type                                        | Description                                                  |
|--------------|---------------------------------------------|--------------------------------------------------------------|
| nodeSize     | string <br /> **required** <br /> _mutable_ | Size of the nodes provisioned as dedicated Zookeeper nodes.  |
| nodesNumber  | int32 <br /> **required**                   | Number of dedicated Zookeeper node count, it must be 3 or 5. |

### KarapaceRestProxyObject
| Field                                | Type                       | Description                                                                                           |
|--------------------------------------|----------------------------|-------------------------------------------------------------------------------------------------------|
| integrateRestProxyWithSchemaRegistry | bool <br /> **required**   | Enables Integration of the Karapace REST proxy with the local Karapace Schema registry.               |
| version                              | string <br /> **required** | Adds the specified version of Kafka REST Proxy to the Kafka cluster. **Available versions:** `3.4.3`. |

### KarapaceSchemaRegistryObject
| Field    | Type                       | Description                                                                                                |
|----------|----------------------------|------------------------------------------------------------------------------------------------------------|
| version  | string <br /> **required** | Adds the specified version of Kafka Schema Registry to the Kafka cluster. **Available versions:** `3.4.3`. |

## Cluster create flow
To create a Kafka cluster instance you need to prepare the yaml manifest. Here is an example:

```yaml
# kafka.yaml
apiVersion: clusters.instaclustr.com/v1beta1
kind: Kafka
metadata:
  name: kafka-sample
spec:
  name: "kafka"
  version: "2.8.2"
  pciCompliance: false
  replicationFactor: 3
  partitionsNumber: 3
  allowDeleteTopics: true
  autoCreateTopics: true
  clientToClusterEncryption: true
  privateNetworkCluster: true
  slaTier: "NON_PRODUCTION"
  karapaceSchemaRegistry:
    - version: "3.2.0"
  karapaceRestProxy:
    - integrateRestProxyWithSchemaRegistry: true
      version: "3.2.0"
  dataCentres:
    - name: "AWS_VPC_US_EAST_1"
      nodesNumber: 3
      cloudProvider: "AWS_VPC"
      tags:
        tag: "oneTag"
        tag2: "twoTags"
      nodeSize: "KFK-DEV-t4g.small-5"
      network: "10.0.0.0/16"
      region: "US_EAST_1"
  dedicatedZookeeper:
    - nodeSize: "KDZ-DEV-t4g.small-30"
      nodesNumber: 3
```

Next, you need to apply this manifest in your K8s cluster. This will create a custom resource instance inside (more info about an apply command you can find [here](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#apply)):
```console
kubectl apply -f kafka.yaml
```

Now you can get and describe the instance:
```console
kubectl get kafkas.clusters.instaclustr.com kafka-sample
```
```console
kubectl describe kafkas.clusters.instaclustr.com kafka-sample
```

After you have applied the entity, the Instaclustr operator will create it on your K8s cluster and send a request to the Instaclustr API. You can be sure, that the resource creation call was sent if the instance has an id field filled in the status section.

You can check access to the created cluster from your kubernetes cluster and run some simple command to check that it is working with a lot of tools.
All available tools you can find in the Instaclustr console -> Choose your cluster -> Connection Info -> Examples section.

When a cluster is provisioned, a new service will be created along with it that expose public IP addresses. You can use this service name (pattern **{k8s_cluster_name}-service**) instead of public IP addresses and ports to connect to and interact with your cluster.
To do this, the public IP address of your machine must be added to the Firewall Rules tab of your cluster in the Instaclustr console.
![Firewall Rules icon](../images/firewall_rules_screen.png "Firewall Rules icon")
Then, you can use the service name instead of public addresses and port.

## Cluster update flow
To update a cluster you can apply an updated cluster manifest or edit the custom resource instance in kubernetes cluster:
* Apply an updated cluster manifest:
    ```console
    kubectl apply -f kafka.yaml
    ```
* Edit the custom resource instance:
    ```console
    kubectl edit kafkas.clusters.instaclustr.com kafka-sample
    ```
You can only update fields that are **mutable**

## Cluster delete flow

### Cluster deletion
To delete cluster run:
```console
kubectl delete kafkas.clusters.instaclustr.com kafka-sample
```

It can take some time to delete the resource.

### Cluster deletion with twoFactorDelete option enabled
To delete cluster with twoFactorDelete option enabled you need to set the confirmation annotation to true:
```yaml
Annotations:  
  "instaclustr.com/deletionConfirmed": true
```

And then simply run:
```console
kubectl delete kafkas.clusters.instaclustr.com kafka-sample
```

After that, deletion confirmation email will be sent to the email defined in the `confirmationEmail` field of `TwoFactorDelete`. When deletion is confirmed via email, Instaclustr support will delete the cluster and the related cluster resources inside K8s will be also removed.