# Kafka Connect cluster management

## Available spec fields

| Field                 | Type                                                                                          | Description                                                                                                                                                                                                                                                                                                                                  |
|-----------------------|-----------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| name                  | string <br /> **required**                                                                    | Cluster name. Should have length from 3 to 32 symbols.                                                                                                                                                                                                                                                                                       |
| version               | string <br /> **required**                                                                    | Kafka Connect instance version. <br />**Available versions**: `3.1.2`, `3.0.2`, `2.8.2`, `2.7.1`.                                                                                                                                                                                                                                            |
| privateNetworkCluster | bool <br /> **required**                                                                      | Creates the cluster with private network only, see [Private Network Clusters](https://www.instaclustr.com/support/documentation/useful-information/private-network-clusters/).                                                                                                                                                               |
| slaTier               | string <br /> **required**                                                                    | SLA Tier of the cluster. Non-production clusters may receive lower priority support and reduced SLAs. Production tier is not available when using Developer class nodes. See [SLA Tier](https://www.instaclustr.com/support/documentation/useful-information/sla-tier/) for more information. <br/>**Enum**: `PRODUCTION`, `NON_PRODUCTION`. |
| twoFactorDelete       | Array of objects ([TwoFactorDelete](#TwoFactorDeleteObject)) <br />_mutable_                  | Contacts that will be contacted when cluster request is sent.                                                                                                                                                                                                                                                                                |
| targetCluster         | Array of objects ([TargetCluster](#TargetClusterObject)) <br /> **required**                  | Details to connect to a target Kafka Cluster cluster.                                                                                                                                                                                                                                                                                        |
| customConnectors      | Array of objects ([CustomConnectors](#CustomConnectorsObject))                                | Defines the location for custom connector storage and access info.                                                                                                                                                                                                                                                                                       |
| dataCentres           | Array of objects ([KafkaConnectDataCentre](#KafkaConnectDataCentreObject)) <br /> **required** | Object fields are described below as a bulleted list.                                                                                                                                                                                                                                                                                        |

### TwoFactorDeleteObject
| Field                   | Type                                      | Description                                                                            |
|-------------------------|-------------------------------------------|----------------------------------------------------------------------------------------|
| confirmationPhoneNumber | string                     | The phone number which will be contacted when the cluster is requested to be deleted.  |
| confirmationEmail       | string <br /> **required** | The email address which will be contacted when the cluster is requested to be deleted. |

### TargetClusterObject
| Field            | Type                                                         | Description                                                                                                             |
|------------------|--------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------|
| externalCluster  | Array of objects ([ExternalCluster](#ExternalClusterObject)) | Details to connect to a Non-Instaclustr managed cluster. Cannot be provided if targeting an Instaclustr managed cluster.|
| managedCluster   | Array of objects ([ManagedCluster](#ManagedClusterObject))   | Details to connect to a Instaclustr managed cluster. Cannot be provided if targeting an external cluster.|

### ExternalClusterObject
| Field                   | Type   | Description                                                                                                                                                                                                                                                                                                                                 |
|-------------------------|--------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| securityProtocol        | string | Connection information for your Kafka Cluster. These options are analogous to the similarly named options that you would place in your Kafka Connect worker.properties file. Only required if connecting to a Non-Instaclustr managed Kafka Cluster.                                                                                                                                                                                                                                                                                      |
| sslTruststorePassword   | string | Connection information for your Kafka Cluster. These options are analogous to the similarly named options that you would place in your Kafka Connect worker.properties file. Only required if connecting to a Non-Instaclustr managed Kafka Cluster.                                                                                                                                                                                                                                           |
| bootstrapServers        | string | Connection information for your Kafka Cluster. These options are analogous to the similarly named options that you would place in your Kafka Connect worker.properties file. Only required if connecting to a Non-Instaclustr managed Kafka Cluster.                                                                                                                                                           |
| saslJaasConfig          | string | Connection information for your Kafka Cluster. These options are analogous to the similarly named options that you would place in your Kafka Connect worker.properties file. Only required if connecting to a Non-Instaclustr managed Kafka Cluster.                                                                                                                                                                                                                                                                                     |
| saslMechanism           | string | Connection information for your Kafka Cluster. These options are analogous to the similarly named options that you would place in your Kafka Connect worker.properties file. Only required if connecting to a Non-Instaclustr managed Kafka Cluster.                                                                                                                                                                                                                                           |
| sslProtocol             | string | Connection information for your Kafka Cluster. These options are analogous to the similarly named options that you would place in your Kafka Connect worker.properties file. Only required if connecting to a Non-Instaclustr managed Kafka Cluster.                                                                                                                                                              |
| truststore              | string | Base64 encoded version of the TLS trust store (in JKS format) used to connect to your Kafka Cluster. Only required if connecting to a Non-Instaclustr managed Kafka Cluster with TLS enabled.                                                                                                                                                                                                                                                                                      |

### ManagedClusterObject
| Field                   | Type                       | Description                                                                                                                                                                                                                                          |
|-------------------------|----------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| targetKafkaClusterId    | string <br /> **required** | Target kafka cluster id. |
| kafkaConnectVpcType     | string <br /> **required** | Available options are `KAFKA_VPC`, `VPC_PEERED`, `SEPARATE_VPC`.                                                                                                                                                                                           |

### CustomConnectorsObject
| Field            | Type                                                                       | Description                                                                                                             |
|------------------|----------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------|
| azureConnectorSettings  | Array of objects ([AzureConnectorSettings](#AzureConnectorSettingsObject)) | Defines the information to access custom connectors located in an azure storage container. Cannot be provided if custom connectors are stored in GCP or AWS.|
| awsConnectorSettings   | Array of objects ([AWSConnectorSettings](#AwsConnectorSettingsObject))     | Defines the information to access custom connectors located in a S3 bucket. Cannot be provided if custom connectors are stored in GCP or AZURE. Access could be provided via Access and Secret key pair or IAM Role ARN. If neither is provided, access policy is defaulted to be provided later.|
| gcpConnectorSettings   | Array of objects ([GCPConnectorSettings](#GcpConnectorSettingsObject))     | Defines the information to access custom connectors located in a gcp storage container. Cannot be provided if custom connectors are stored in AWS or AZURE.|

### AzureConnectorSettingsObject
| Field                 | Type                       | Description                                                                                                                                                                                                                                                                                                                                 |
|-----------------------|----------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| storageContainerName  | string <br /> **required** | Azure storage container name for Kafka Connect custom connector.                                                                                                                                                                                                                                                                                     |
| storageAccountName    | string <br /> **required** | Azure storage account name to access your Azure bucket for Kafka Connect custom connector.                                                                                                                                                                                                                                           |
| storageAccountKey     | string <br /> **required** | Azure storage account key to access your Azure bucket for Kafka Connect custom connector.                                                                                                                                                                                                                                           |

### AwsConnectorSettingsObject
| Field                 | Type                       | Description                                                                                                                                                                                                                                                                                                                                 |
|-----------------------|----------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| s3RoleArn  | string <br />              | AWS Identity Access Management role that is used for accessing your specified S3 bucket for Kafka Connect custom connector                                                                                                                                                                                                                                                                                    |
| secretKey    | string <br />              | AWS Secret Key associated with the Access Key id that can access your specified S3 bucket for Kafka Connect custom connector.                                                                                                                                                                                                                                           |
| accessKey     | string <br />              | AWS Access Key id that can access your specified S3 bucket for Kafka Connect custom connector.                                                                                                                                                                                                                                           |
| s3BucketName     | string <br /> **required** | S3 bucket name for Kafka Connect custom connector.                                                                                                                                                                                                                                           |

### GcpConnectorSettingsObject
| Field             | Type                       | Description                                                                                                                                                                                                                                                                                                                                 |
|-------------------|----------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| privateKey        | string <br /> **required** | Access information for the GCP Storage bucket for kafka connect custom connectors.                                                                                                                                                                                                                                                                                     |
| clientId          | string <br /> **required** | Access information for the GCP Storage bucket for kafka connect custom connectors.                                                                                                                                                                                                                                           |
| clientEmail       | string <br /> **required** | Access information for the GCP Storage bucket for kafka connect custom connectors.                                                                                                                                                                                                                                           |
| projectId         | string <br /> **required** | Access information for the GCP Storage bucket for kafka connect custom connectors.                                                                                                                                                                                                                                                                                     |
| storageBucketName | string <br /> **required** | Access information for the GCP Storage bucket for kafka connect custom connectors.                                                                                                                                                                                                                                           |
| privateKeyId      | string <br /> **required** | Access information for the GCP Storage bucket for kafka connect custom connectors.                                                                                                                                                                                                                                           |

### KafkaConnectDataCentreObject
| Field                    | Type                                                                    | Description                                                                                                                                                                                                                                                                                                                                                           |
|--------------------------|-------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| name                     | string <br /> **required**                                              | A logical name for the data centre within a cluster. These names must be unique in the cluster.                                                                                                                                                                                                                                                                       |
| region                   | string <br /> **required**                                              | Region of the Data Centre. See the description for node size for a compatible Data Centre for a given node size.                                                                                                                                                                                                                                                      |
| cloudProvider            | string <br /> **required**                                              | Name of the cloud provider service in which the Data Centre will be provisioned. <br />**Enum**: `AWS_VPC` `GCP` `AZURE` `AZURE_AZ`.                                                                                                                                                                                                                                  |
| accountName              | string                                                                  | For customers running in their own account. Your provider account can be found on the Create Cluster page on the Instaclustr Console, or the "Provider Account" property on any existing cluster. For customers provisioning on Instaclustr's cloud provider accounts, this property may be omitted.                                                                  |
| cloudProviderSettings    | Array of objects ([CloudProviderSettings](#CloudProviderSettingsObject))| Cloud provider specific settings for the Data Centre.                                                                                                                                                                                                                                                                                                                 |
| network                  | string <br /> **required**                                              | The private network address block for the Data Centre specified using CIDR address notation. The network must have a prefix length between /12 and /22 and must be part of a private address space.                                                                                                                                                                   |
| nodeSize                 | string <br /> **required**<br />_mutable_                               | Size of the nodes provisioned in the Data Centre. Available node sizes, see [Instaclustr API docs NodeSize](https://instaclustr.redoc.ly/Current/tag/Kafka-Connect-Cluster-V2#paths/~1cluster-management~1v2~1resources~1applications~1kafka-connect~1clusters~1v2~1/post!path=dataCentres/nodeSize&t=request). |
| nodesNumber              | int32 <br /> **required**<br />_mutable_                                | Total number of nodes in the Data Centre. <br/>Available values: [1…5].                                                                                                                                                                                                                                                                                               |
| tags                     | map[string]string                                                       | List of tags to apply to the Data Centre. Tags are metadata labels which allow you to identify, categorise and filter clusters. This can be useful for grouping together clusters into applications, environments, or any category that you require.<br/>**Format**:<br/>tags:<br/>- key: value.                                                                      |
| replicationFactor        | int32 <br /> **required**                                               | Number of racks to use when allocating nodes.                                                                                                                                                                                                                                                                                                                         |                                                                                                                                                                                           

### CloudProviderSettingsObject
| Field                  | Type   | Description                                                                                                                                                                                                                                                                                                                                                                                                                                 |
|------------------------|--------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| customVirtualNetworkId | string | **AWS**: VPC ID into which the Data Centre will be provisioned. The Data Centre's network allocation must match the IPv4 CIDR block of the specified VPC.<br />**GCP**: Network name or a relative Network or Subnetwork URI e.g. projects/my-project/regions/us-central1/subnetworks/my-subnet. The Data Centre's network allocation must match the IPv4 CIDR block of the specified subnet. <br />Cannot be provided with `resourceGroup` |
| resourceGroup          | string | The name of the Azure Resource Group into which the Data Centre will be provisioned. <br />Cannot be provided with `customVirtualNetworkId` and `diskEncryptionKey`                                                                                                                                                                                                                                                                         |
| diskEncryptionKey      | string | ID of a KMS encryption key to encrypt data on nodes. KMS encryption key must be set in Cluster Resources through the Instaclustr Console before provisioning an encrypted Data Centre. <br />Cannot be provided with `customVirtualNetworkId`                                                                                                                                                                                               |

## Cluster create flow
To create a Kafka Connect cluster instance you need to prepare the yaml manifest. Here is an example:

```yaml
# kafkaconnect.yaml

apiVersion: clusters.instaclustr.com/v1alpha1
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

Next, you need to apply this manifest in your K8s cluster. This will create a custom resource instance inside (more info about an apply command you can find [here](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#apply)):
```console
kubectl apply kafkaconnect.yaml
```

Now you can get and describe the instance:
```console
kubectl get kafkaconnects.clusters.instaclustr.com kafkaconnect-sample
```
```console
kubectl describe kafkaconnects.clusters.instaclustr.com kafkaconnect-sample
```

After you have applied the entity, the Instaclustr operator will create it on your K8s cluster and send a request to the Instaclustr API. You can be sure, that the resource creation call was sent if the instance has an id field filled in the status section.

## Cluster update flow
To update a cluster you can apply an updated cluster manifest or edit the custom resource instance in kubernetes cluster:
* Apply an updated cluster manifest:
    ```console
    kubectl apply kafkaconnect.yaml
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

It can take some time to delete the resource.

### Cluster deletion with twoFactorDelete option enabled
To delete cluster with twoFactorDelete option enabled you need to set the confirmation annotation to true:
```yaml
Annotations:  
  "instaclustr.com/deletionConfirmed": true
```

And then simply run:
```console
kubectl delete kafkaconnects.clusters.instaclustr.com kafkaconnect-sample
```

After that, deletion confirmation email will be sent to the email defined in the `confirmationEmail` field of `TwoFactorDelete`. When deletion is confirmed via email, Instaclustr support will delete the cluster and the related cluster resources inside K8s will be also removed.