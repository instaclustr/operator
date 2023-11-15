# Kafka cluster management

## Available spec fields

| Field                             | Type                                                                             | Description                                                                                                                                                                                                                                                                                                                                  |
|-----------------------------------|----------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| name                              | string <br /> **required**                                                       | Cluster name. Should have length from 3 to 32 symbols.                                                                                                                                                                                                                                                                                       |
| version                           | string <br /> **required**                                                       | Kafka instance version. <br />**Available versions**: `3.1.2`, `3.3.1`, `3.4.1`, `3.5.1`.                                                                                                                                                                                                                                                    |
| pciCompliance                     | bool <br /> **required**                                                         | Creates a PCI compliant cluster, see [PCI Compliance](https://www.instaclustr.com/support/documentation/useful-information/pci-compliance/)                                                                                                                                                                                                  |
| privateNetworkCluster             | bool <br /> **required**                                                         | Allows topics to be deleted via the kafka-topics tool                                                                                                                                                                                                                                                                                        |
| allowDeleteTopics                 | bool <br /> **required**                                                         | Creates the cluster with private network only, see [Private Network Clusters](https://www.instaclustr.com/support/documentation/useful-information/private-network-clusters/).                                                                                                                                                               |
| slaTier                           | string <br /> **required**                                                       | SLA Tier of the cluster. Non-production clusters may receive lower priority support and reduced SLAs. Production tier is not available when using Developer class nodes. See [SLA Tier](https://www.instaclustr.com/support/documentation/useful-information/sla-tier/) for more information. <br/>**Enum**: `PRODUCTION`, `NON_PRODUCTION`. |
| twoFactorDelete                   | Array of objects ([TwoFactorDelete](#TwoFactorDeleteObject)) <br /> _mutable_    | Contacts that will be contacted when cluster request is sent.                                                                                                                                                                                                                                                                                |
| schemaRegistry                    | Array of objects ([SchemaRegistry](#SchemaRegistryObject))                       | Adds the specified version of Kafka Schema Registry to this Kafka cluster.                                                                                                                                                                                                                                                                   |
| replicationFactor                 | int32 <br /> **required**                                                        | Default Replication factor to use for new topic. Also represents the number of racks to use when allocating nodes.                                                                                                                                                                                                                           |
| partitionsNumber                  | int32 <br /> **required**                                                        | Default number of partitions to use when created new topics.                                                                                                                                                                                                                                                                                 |
| restProxy                         | Array of objects ([RestProxy](#RestProxyObject))                                 | Adds the specified version of Kafka REST Proxy to this Kafka cluster.                                                                                                                                                                                                                                                                        |
| kraft                             | Array of objects ([KafkaKraftSettings](#KafkaKraftSettingsObject))               | Create a KRaft Cluster                                                                                                                                                                                                                                                                                                                       |
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
| description                       | string <br />                                                                    | A description of the cluster                                                                                                                                                                                                                                                                                                                 |
| onPremisesSpec                    | Object ([OnPremisesSpec](#OnPremisesSpecObject))                                 | Specifies settings to provision on-premises cluster inside K8s cluster.                                                                                                                                                                                                                                                                      |

### TwoFactorDeleteObject
| Field                   | Type                              | Description                                                                            |
|-------------------------|-----------------------------------|----------------------------------------------------------------------------------------|
| confirmationPhoneNumber | string                            | The phone number which will be contacted when the cluster is requested to be deleted.  |
| confirmationEmail       | string <br /> **required** <br /> | The email address which will be contacted when the cluster is requested to be deleted. |

### SchemaRegistryObject
| Field     | Type                        | Description                                                                                                |
|-----------|-----------------------------|------------------------------------------------------------------------------------------------------------|
| version   | string <br /> **required**  | Adds the specified version of Kafka Schema Registry to the Kafka cluster. **Available versions:** `5.0.0`. |

### KafkaKraftSettingsObject
| Field               | Type                       | Description                                                         |
|---------------------|----------------------------|---------------------------------------------------------------------|
| controllerNodeCount | int32 <br /> **required**  | Number of KRaft controller nodes (only 3 is currently supported).   |

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
| Field    | Type                       | Description                                                                                                         |
|----------|----------------------------|---------------------------------------------------------------------------------------------------------------------|
| version  | string <br /> **required** | Adds the specified version of Kafka Schema Registry to the Kafka cluster. **Available versions:** `3.4.3`, `3.6.2`. |

### OnPremisesSpecObject

| Field              | Type                                                       | Description                                                                                                                                                                        |
|--------------------|------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| storageClassName   | string <br /> **required**                                 | Name of the storage class that will be used to provision disks for on-premises nodes.                                                                                              |
| osDiskSize         | string <br /> **required**                                 | Disk size on which OS will be installed.                                                                                                                                           |
| dataDiskSize       | string <br /> **required**                                 | Disk size on which on-premises cluster data will be stored.                                                                                                                        |
| sshGatewayCPU      | int64                                                      | Amount of CPU that will be dedicated to provision SSH Gateway node (only for private clusters).                                                                                    |
| sshGatewayMemory   | string                                                     | Amount of RAM that will be dedicated to provision SSH Gateway node (only for private clusters).                                                                                    |
| nodeCPU            | int64 <br /> **required**                                  | Amount of CPU that will be dedicated to provision on-premises worker node.                                                                                                         |
| nodeMemory         | string <br /> **required**                                 | Amount of RAM that will be dedicated to provision on-premises worker node                                                                                                          |
| osImageURL         | string <br /> **required**                                 | OS image URL that will be use to dynamically provision disks with preinstalled OS (more info can be found [here](https://kubevirt.io/2020/KubeVirt-VM-Image-Usage-Patterns.html)). |
| cloudInitScriptRef | Object ([Reference](#ReferenceObject)) <br /> **required** | Reference to the secret with cloud-init script (must be located in the same namespace with the cluster). Example can be found [here](#CloudInitScript).                            |

### ReferenceObject

| Field     | Type   | Description                                       |
|-----------|--------|---------------------------------------------------|
| name      | string | Name of the cloud-init secret.                    |
| namespace | string | Namespace in which the cloud-init secret located. |

## Cluster create flow
To create a Kafka cluster instance you need to prepare the yaml manifest. Here is an example:

Notice:
- If you choose to omit the cluster name in the specification, the operator will use it from the metadata name instead. In this case make sure that metadata name matches the name pattern, you can check the pattern on instaclustr API specification

```yaml
# kafka.yaml
apiVersion: clusters.instaclustr.com/v1beta1
kind: Kafka
metadata:
  name: kafka
spec:
  name: "Kafka-example"
  version: "3.3.1"
  pciCompliance: false
  replicationFactor: 3
  partitionsNumber: 3
  allowDeleteTopics: true
  autoCreateTopics: true
  clientToClusterEncryption: false
  privateNetworkCluster: false
  slaTier: "NON_PRODUCTION"
  bundledUseOnly: true
  clientBrokerAuthWithMtls: true
  dedicatedZookeeper:
    - nodeSize: "KDZ-DEV-t4g.small-30"
      nodesNumber: 3
  twoFactorDelete:
    - email: "example@gmail.com"
      phone: "example"
  karapaceSchemaRegistry:
    - version: "3.2.0"
  schemaRegistry:
    - version: "5.0.0"
  karapaceRestProxy:
    - integrateRestProxyWithSchemaRegistry: true
      version: "3.2.0"
  kraft:
    - controllerNodeCount: 3
  restProxy:
    - integrateRestProxyWithSchemaRegistry: false
      schemaRegistryPassword: "asdfasdf"
      schemaRegistryServerUrl: "schemaRegistryServerUrl"
      "useLocalSchemaRegistry": true
      version: "5.0.0"
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
      accountName: "Custrom"
      cloudProviderSettings:
        - customVirtualNetworkId: "vpc-12345678"
          diskEncryptionKey: "123e4567-e89b-12d3-a456-426614174000"
          resourceGroup: "asdfadfsdfas"
      privateLink:
        - advertisedHostname: "kafka-example-test.com"
  userRefs:
    - name: kafkauser-sample
      namespace: default
  resizeSettings:
    - notifySupportContacts: false
      concurrency: 1
```

Or if you want to create an on-premises cluster:
```yaml
# kafka.yaml
apiVersion: clusters.instaclustr.com/v1beta1
kind: Kafka
metadata:
  name: Kafka-example
spec:
  name: "Kafka-example"
  version: "3.3.1"
  pciCompliance: false
  replicationFactor: 3
  partitionsNumber: 3
  allowDeleteTopics: true
  autoCreateTopics: true
  clientToClusterEncryption: false
  privateNetworkCluster: false
  slaTier: "NON_PRODUCTION"
  onPremisesSpec:
    storageClassName: managed-csi-premium
    osDiskSize: 20Gi
    dataDiskSize: 200Gi
    sshGatewayCPU: 2
    sshGatewayMemory: 4096Mi
    nodeCPU: 2
    nodeMemory: 8192Mi
    osImageURL: "https://s3.amazonaws.com/debian-bucket/debian-11-generic-amd64-20230601-1398.raw"
    cloudInitScriptRef:
      namespace: default
      name: cloud-init-secret
  dataCentres:
    - name: "onPremKafka"
      nodesNumber: 3
      cloudProvider: "ONPREMISES" # Don't change if you want to run on-premises
      tags:
        tag: "oneTag"
        tag2: "twoTags"
      nodeSize: "KFK-DEV-OP.4.8-200"
      network: "10.0.0.0/16"
      region: "CLIENT_DC" # Don't change if you want to run on-premises
  resizeSettings:
    - notifySupportContacts: false
      concurrency: 1
```

Also, don't forget to create cloud-init script firstly.

### CloudInitScript

cloud-init.sh:
```shell
#!/bin/bash

export NEW_PASS="qwerty12345"
export SSH_PUB_KEY=""
export BOOTSTRAP_SSH_KEY="ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAEAQDgeaO3zkY5v1dww3fFONPzUnEgIqJ4kUK0Usu8iFdp+TWIulw9dDeQHa+PdWXP97l5Vv1mG9ipqShEIu7/2bp13KxSblWX4iV1MYZbtimhY3UDOsPn1G3E1Ipis6y+/tosDAm8LoWaGEMcLuE5UjP6gs6K57rCEjkVGjg7vjhIypAMC0J2N2+CxK9o/Y1+LZAec+FL5cmSJoajoo9y/VYJjz/a52jJ93wRafD2uu6ObAl5gkN/+gqY4IJFSMx20sPsIRXdbiBNDqiap56ibHHPKTeZhRbdXvZfjYkHtutXnSM2xn7BjnV8CguISxS3rXlzlzRVYwSUjnKUf5SKBbeyZbCokx271vCeUd5EXfHphvW6FIOna2AI5lpCSYkw5Kho3HaPi2NjXJ9i2dCr1zpcZpCiettDuaEjxR0Cl4Jd6PrAiAEZ0Ns0u2ysVhudshVzQrq6qdd7W9/MLjbDIHdTToNjFLZA6cbE0MQf18LXwJAl+G/QrXgcVaiopUmld+68JL89Xym55LzkMhI0NiDtWufawd/NiZ6jm13Z3+atvnOimdyuqBYeFWgbtcxjs0yN9v7h7PfPh6TbiQYFx37DCfQiIucqx1GWmMijmd7HMY6Nv3UvnoTUTSn4yz1NxhEpC61N+iAInZDpeJEtULSzlEMWlbzL4t5cF+Rm1dFdq3WpZt1mi8F4DgrsgZEuLGAw22RNW3++EWYFUNnJXaYyctPrMpWQktr4DB5nQGIHF92WR8uncxTNUXfWuT29O9e+bFYh1etmq8rsCoLcxN0zFHWvcECK55aE+47lfNAR+HEjuwYW10mGU/pFmO0F9FFmcQRSw4D4tnVUgl3XjKe3bBiTa4lUrzrKkLZ6n9/buW2e7c3jbjmXdPh2R+2Msr/vfuWs9glxQf+CYEbBW6Ye4pekIyI77SaB/bVhaHtXutKxm+QWdNle8aeqiA8Ji1Ml+s75vIg+n5v6viCnl5aV33xHRFpGQJzj2ktsXl9P9d5kgal9eXJYTywC2SnVbZVLb6FGN4kPZTVwX1f+u7v7JCm4YWlbQZtwwiXKjs99AVtQnBWqQvUH5sFUkVXlHA1Y9W6wlup0r+F6URL+7Yw+d0dHByfevrJg3pvmpLb3sEpjIAZodW3dIUReE7Ku3s/q/O9foFnfRBnCcZ2QsnxI5pqNrbrundD1ApOnNXEvICvPXHBBQ44cW0hzAO+WxY5VxyG8y/kXnb48G9efkIQFkNaITJrU9SiOk6bFP4QANdS/pmaSLjJIsHixa+7vmYjRy1SVoQ/39vDUnyCbqKtO56QMH32hQLRO3Vk7NVG6o4dYjFkiaMSaqVlHKMkJQHVzlK2PW9/fjVXfkAHmmhoD debian"
echo "debian:$NEW_PASS" | chpasswd
echo "root:$NEW_PASS" | sudo chpasswd root
sudo echo "$SSH_PUB_KEY" > /home/debian/.ssh/authorized_keys
sudo echo "$BOOTSTRAP_SSH_KEY" >> /home/debian/.ssh/authorized_keys
sudo chown -R debian: /home/debian/.ssh
sudo cp /usr/share/doc/apt/examples/sources.list /etc/apt/sources.list
data_device=$(lsblk -dfn -o NAME,SERIAL | awk '$2 == "DATADISK" {print $1}')
sudo mkfs -t ext4 /dev/"${data_device}"
```

Create the base64 encoded string with the script using `cat cloud-init.sh | base64 -w0` command and create a secret using this yaml manifest:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: cloud-init-secret
data:
  userdata: <base64 encoded string>
```

Use `kubectl apply -f cloud-init-secret.yaml` command to create the secret.

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

### Cluster deletion with twoFactorDelete option enabled
To delete a cluster with the `twoFactorDelete` option enabled you need to do the simple [cluster deletion flow ](#cluster-deletion).
After that, a deletion email will be sent to the email defined in the `confirmationEmail` field of `twoFactorDelete`.
When deletion is confirmed via email, Instaclustr support will delete the cluster.
It can take some time to delete the resource.

If you cancel cluster deletion and want to put cluster on delete again, remove `triggered` from `clusterDeletionAnnotation` annotation like this:

```yaml
Annotations:
  "instaclustr.com/clusterDeletion": ""
```
