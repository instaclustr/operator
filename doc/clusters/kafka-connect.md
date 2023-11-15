# Kafka Connect cluster management

## Available spec fields

| Field                 | Type                                                                                           | Description                                                                                                                                                                                                                                                                                                                                  |
|-----------------------|------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| name                  | string <br /> **required**                                                                     | Cluster name. Should have length from 3 to 32 symbols.                                                                                                                                                                                                                                                                                       |
| version               | string <br /> **required**                                                                     | Kafka Connect instance version. <br />**Available versions**: `3.1.2`, `3.3.1`, `3.4.1`, `3.5.1`.                                                                                                                                                                                                                                            |
| privateNetworkCluster | bool <br /> **required**                                                                       | Creates the cluster with private network only, see [Private Network Clusters](https://www.instaclustr.com/support/documentation/useful-information/private-network-clusters/).                                                                                                                                                               |
| slaTier               | string <br /> **required**                                                                     | SLA Tier of the cluster. Non-production clusters may receive lower priority support and reduced SLAs. Production tier is not available when using Developer class nodes. See [SLA Tier](https://www.instaclustr.com/support/documentation/useful-information/sla-tier/) for more information. <br/>**Enum**: `PRODUCTION`, `NON_PRODUCTION`. |
| twoFactorDelete       | Array of objects ([TwoFactorDelete](#TwoFactorDeleteObject)) <br />_mutable_                   | Contacts that will be contacted when cluster request is sent.                                                                                                                                                                                                                                                                                |
| targetCluster         | Array of objects ([TargetCluster](#TargetClusterObject)) <br /> **required**                   | Details to connect to a target Kafka Cluster cluster.                                                                                                                                                                                                                                                                                        |
| customConnectors      | Array of objects ([CustomConnectors](#CustomConnectorsObject))                                 | Defines the location for custom connector storage and access info.                                                                                                                                                                                                                                                                           |
| dataCentres           | Array of objects ([KafkaConnectDataCentre](#KafkaConnectDataCentreObject)) <br /> **required** | Object fields are described below as a bulleted list.                                                                                                                                                                                                                                                                                        |
| description           | string <br />                                                                                  | A description of the cluster                                                                                                                                                                                                                                                                                                                 |
| onPremisesSpec        | Object ([OnPremisesSpec](#OnPremisesSpecObject))                                               | Specifies settings to provision on-premises cluster inside K8s cluster.                                                                                                                                                                                                                                                                      |

### TwoFactorDeleteObject
| Field                   | Type                       | Description                                                                            |
|-------------------------|----------------------------|----------------------------------------------------------------------------------------|
| confirmationPhoneNumber | string                     | The phone number which will be contacted when the cluster is requested to be deleted.  |
| confirmationEmail       | string <br /> **required** | The email address which will be contacted when the cluster is requested to be deleted. |

### TargetClusterObject
| Field            | Type                                                         | Description                                                                                                              |
|------------------|--------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------|
| externalCluster  | Array of objects ([ExternalCluster](#ExternalClusterObject)) | Details to connect to a Non-Instaclustr managed cluster. Cannot be provided if targeting an Instaclustr managed cluster. |
| managedCluster   | Array of objects ([ManagedCluster](#ManagedClusterObject))   | Details to connect to a Instaclustr managed cluster. Cannot be provided if targeting an external cluster.                |

### ExternalClusterObject
| Field                 | Type   | Description                                                                                                                                                                                                                                          |
|-----------------------|--------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| securityProtocol      | string | Connection information for your Kafka Cluster. These options are analogous to the similarly named options that you would place in your Kafka Connect worker.properties file. Only required if connecting to a Non-Instaclustr managed Kafka Cluster. |
| sslTruststorePassword | string | Connection information for your Kafka Cluster. These options are analogous to the similarly named options that you would place in your Kafka Connect worker.properties file. Only required if connecting to a Non-Instaclustr managed Kafka Cluster. |
| bootstrapServers      | string | Connection information for your Kafka Cluster. These options are analogous to the similarly named options that you would place in your Kafka Connect worker.properties file. Only required if connecting to a Non-Instaclustr managed Kafka Cluster. |
| saslJaasConfig        | string | Connection information for your Kafka Cluster. These options are analogous to the similarly named options that you would place in your Kafka Connect worker.properties file. Only required if connecting to a Non-Instaclustr managed Kafka Cluster. |
| saslMechanism         | string | Connection information for your Kafka Cluster. These options are analogous to the similarly named options that you would place in your Kafka Connect worker.properties file. Only required if connecting to a Non-Instaclustr managed Kafka Cluster. |
| sslProtocol           | string | Connection information for your Kafka Cluster. These options are analogous to the similarly named options that you would place in your Kafka Connect worker.properties file. Only required if connecting to a Non-Instaclustr managed Kafka Cluster. |
| sslEnabledProtocols   | string | Connection information for your Kafka Cluster. These options are analogous to the similarly named options that you would place in your Kafka Connect worker.properties file. Only required if connecting to a Non-Instaclustr managed Kafka Cluster. |
| truststore            | string | Base64 encoded version of the TLS trust store (in JKS format) used to connect to your Kafka Cluster. Only required if connecting to a Non-Instaclustr managed Kafka Cluster with TLS enabled.                                                        |

### ManagedClusterObject
| Field                   | Type                       | Description                                                             |
|-------------------------|----------------------------|-------------------------------------------------------------------------|
| targetKafkaClusterId    | string <br /> **required** | Target kafka cluster id.                                                |
| kafkaConnectVpcType     | string <br /> **required** | Available options are `KAFKA_VPC`, `VPC_PEERED`, `SEPARATE_VPC`.        |

### CustomConnectorsObject
| Field                    | Type                                                                       | Description                                                                                                                                                                                                                                                                                       |
|--------------------------|----------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| azureConnectorSettings   | Array of objects ([AzureConnectorSettings](#AzureConnectorSettingsObject)) | Defines the information to access custom connectors located in an azure storage container. Cannot be provided if custom connectors are stored in GCP or AWS.                                                                                                                                      |
| awsConnectorSettings     | Array of objects ([AWSConnectorSettings](#AwsConnectorSettingsObject))     | Defines the information to access custom connectors located in a S3 bucket. Cannot be provided if custom connectors are stored in GCP or AZURE. Access could be provided via Access and Secret key pair or IAM Role ARN. If neither is provided, access policy is defaulted to be provided later. |
| gcpConnectorSettings     | Array of objects ([GCPConnectorSettings](#GcpConnectorSettingsObject))     | Defines the information to access custom connectors located in a gcp storage container. Cannot be provided if custom connectors are stored in AWS or AZURE.                                                                                                                                       |

### AzureConnectorSettingsObject
| Field                 | Type                       | Description                                                                                               |
|-----------------------|----------------------------|-----------------------------------------------------------------------------------------------------------|
| storageContainerName  | string <br /> **required** | Azure storage container name for Kafka Connect custom connector.                                          |
| storageAccountName    | string <br /> **required** | Azure storage account name to access your Azure bucket for Kafka Connect custom connector.                |
| storageAccountKey     | string <br /> **required** | Azure storage account key to access your Azure bucket for Kafka Connect custom connector.                 |

### AwsConnectorSettingsObject
| Field        | Type                       | Description                                                                                                                          |
|--------------|----------------------------|--------------------------------------------------------------------------------------------------------------------------------------|
| s3RoleArn    | string <br />              | AWS Identity Access Management role that is used for accessing your specified S3 bucket for Kafka Connect custom connector           |
| secretKey    | string <br />              | AWS Secret Key associated with the Access Key id that can access your specified S3 bucket for Kafka Connect custom connector.        |
| accessKey    | string <br />              | AWS Access Key id that can access your specified S3 bucket for Kafka Connect custom connector.                                       |
| s3BucketName | string <br /> **required** | S3 bucket name for Kafka Connect custom connector.                                                                                   |

### GcpConnectorSettingsObject
| Field             | Type                       | Description                                                                                 |
|-------------------|----------------------------|---------------------------------------------------------------------------------------------|
| privateKey        | string <br /> **required** | Access information for the GCP Storage bucket for kafka connect custom connectors.          |
| clientId          | string <br /> **required** | Access information for the GCP Storage bucket for kafka connect custom connectors.          |
| clientEmail       | string <br /> **required** | Access information for the GCP Storage bucket for kafka connect custom connectors.          |
| projectId         | string <br /> **required** | Access information for the GCP Storage bucket for kafka connect custom connectors.          |
| storageBucketName | string <br /> **required** | Access information for the GCP Storage bucket for kafka connect custom connectors.          |
| privateKeyId      | string <br /> **required** | Access information for the GCP Storage bucket for kafka connect custom connectors.          |

### KafkaConnectDataCentreObject
| Field                    | Type                                                                     | Description                                                                                                                                                                                                                                                                                                      |
|--------------------------|--------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| name                     | string <br /> **required**                                               | A logical name for the data centre within a cluster. These names must be unique in the cluster.                                                                                                                                                                                                                  |
| region                   | string <br /> **required**                                               | Region of the Data Centre. See the description for node size for a compatible Data Centre for a given node size.                                                                                                                                                                                                 |
| cloudProvider            | string <br /> **required**                                               | Name of the cloud provider service in which the Data Centre will be provisioned. <br />**Enum**: `AWS_VPC` `GCP` `AZURE` `AZURE_AZ`.                                                                                                                                                                             |
| accountName              | string                                                                   | For customers running in their own account. Your provider account can be found on the Create Cluster page on the Instaclustr Console, or the "Provider Account" property on any existing cluster. For customers provisioning on Instaclustr's cloud provider accounts, this property may be omitted.             |
| cloudProviderSettings    | Array of objects ([CloudProviderSettings](#CloudProviderSettingsObject)) | Cloud provider specific settings for the Data Centre.                                                                                                                                                                                                                                                            |
| network                  | string <br /> **required**                                               | The private network address block for the Data Centre specified using CIDR address notation. The network must have a prefix length between /12 and /22 and must be part of a private address space.                                                                                                              |
| nodeSize                 | string <br /> **required**<br />_mutable_                                | Size of the nodes provisioned in the Data Centre. Available node sizes, see [Instaclustr API docs NodeSize](https://instaclustr.redoc.ly/Current/tag/Kafka-Connect-Cluster-V2#paths/~1cluster-management~1v2~1resources~1applications~1kafka-connect~1clusters~1v2~1/post!path=dataCentres/nodeSize&t=request).  |
| nodesNumber              | int32 <br /> **required**<br />_mutable_                                 | Total number of nodes in the Data Centre. <br/>Available values: [1…5].                                                                                                                                                                                                                                          |
| tags                     | map[string]string                                                        | List of tags to apply to the Data Centre. Tags are metadata labels which allow you to identify, categorise and filter clusters. This can be useful for grouping together clusters into applications, environments, or any category that you require.<br/>**Format**:<br/>tags:<br/>- key: value.                 |
| replicationFactor        | int32 <br /> **required**                                                | Number of racks to use when allocating nodes.                                                                                                                                                                                                                                                                    |

### CloudProviderSettingsObject
| Field                  | Type   | Description                                                                                                                                                                                                                                                                                                                                                                                                                                 |
|------------------------|--------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| customVirtualNetworkId | string | **AWS**: VPC ID into which the Data Centre will be provisioned. The Data Centre's network allocation must match the IPv4 CIDR block of the specified VPC.<br />**GCP**: Network name or a relative Network or Subnetwork URI e.g. projects/my-project/regions/us-central1/subnetworks/my-subnet. The Data Centre's network allocation must match the IPv4 CIDR block of the specified subnet. <br />Cannot be provided with `resourceGroup` |
| resourceGroup          | string | The name of the Azure Resource Group into which the Data Centre will be provisioned. <br />Cannot be provided with `customVirtualNetworkId` and `diskEncryptionKey`                                                                                                                                                                                                                                                                         |
| diskEncryptionKey      | string | ID of a KMS encryption key to encrypt data on nodes. KMS encryption key must be set in Cluster Resources through the Instaclustr Console before provisioning an encrypted Data Centre. <br />Cannot be provided with `customVirtualNetworkId`                                                                                                                                                                                               |

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
To create a Kafka Connect cluster instance you need to prepare the yaml manifest. Here is an example:

Notice:
- If you choose to omit the cluster name in the specification, the operator will use it from the metadata name instead. In this case make sure that metadata name matches the name pattern, you can check the pattern on instaclustr API specification

```yaml
# kafkaconnect.yaml
apiVersion: clusters.instaclustr.com/v1beta1
kind: KafkaConnect
metadata:
  name: kafkaconnect-sample
spec:
  dataCentres:
    - name: "US_EAST_1_DC_KAFKA"
      nodesNumber: 3
      cloudProvider: "AWS_VPC"
      replicationFactor: 3
      tags:
        tag: "firstTag"
        tag2: "secondTags"
      nodeSize: "KCN-DEV-t4g.medium-30"
      network: "10.3.0.0/16"
      region: "US_EAST_1"
  name: "kafkaConnectOperator"
  version: "3.1.2"
  privateNetworkCluster: false
  slaTier: "NON_PRODUCTION"
  targetCluster:
    - managedCluster:
        - targetKafkaClusterId: "c38ce872-ebf3-49f5-ae9b-f20fccac527e"
          kafkaConnectVpcType: "KAFKA_VPC"
```

Or if you want to create an on-premises cluster:
```yaml
# kafkaconnect.yaml
apiVersion: clusters.instaclustr.com/v1beta1
kind: KafkaConnect
metadata:
  name: kafkaconnect-sample
spec:
  name: "kafkaconnect-sample"
  version: "3.1.2"
  privateNetworkCluster: false
  dataCentres:
    - name: "OnPremKafkaConnect"
      nodesNumber: 3
      cloudProvider: "ONPREMISES" # Don't change if you want to run on-premises
      replicationFactor: 3
      tags:
        tag: "oneTag"
        tag2: "twoTags"
      nodeSize: "KCN-DEV-OP.4.8-200"
      network: "10.1.0.0/16"
      region: "CLIENT_DC" # Don't change if you want to run on-premises
  slaTier: "NON_PRODUCTION"
  targetCluster:
    - managedCluster:
        - targetKafkaClusterId: "34dfc53c-c8c1-4be8-bd2f-cfdb77ec7349"
          kafkaConnectVpcType: "KAFKA_VPC"
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
kubectl apply -f kafkaconnect.yaml
```

Now you can get and describe the instance:
```console
kubectl get kafkaconnects.clusters.instaclustr.com kafkaconnect-sample
```
```console
kubectl describe kafkaconnects.clusters.instaclustr.com kafkaconnect-sample
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
    kubectl apply -f kafkaconnect.yaml
    ```
* Edit the custom resource instance:
    ```console
    kubectl edit kafkaconnects.clusters.instaclustr.com kafkaconnect-sample
    ```
You can only update fields that are **mutable**

## Cluster delete flow

### Cluster deletion
To delete cluster run:
```console
kubectl delete kafkaconnects.clusters.instaclustr.com kafkaconnect-sample
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

## External changes
If you change properties of the cluster (node size, number of nodes, etc.) using the Instaclustr Console UI, 
there is going to be inconsistency in the k8s CRD specification (`Spec`). In this case, you have to reconcile the changes manually, 
for this add the `instaclustr.com/allowSpecAmend: true` annotation to be able to change the k8s resource specification.
