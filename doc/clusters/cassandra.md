# Cassandra cluster management

## Available spec fields

| Field                 | Type                                                                                     | Description                                                                                                                                                                                                                                                                                                                                  |
|-----------------------|------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| name                  | string <br /> **required**                                                               | Cluster name. Should have length from 3 to 32 symbols.                                                                                                                                                                                                                                                                                       |
| version               | string <br /> **required**                                                               | Cassandra instance version. <br />**Available versions**: `4.0.4`, `3.11.13`.                                                                                                                                                                                                                                                                |
| pciCompliance         | bool <br /> **required**                                                                 | Creates a PCI compliant cluster, see [PCI Compliance](https://www.instaclustr.com/support/documentation/useful-information/pci-compliance/)                                                                                                                                                                                                  |
| privateNetworkCluster | bool <br /> **required**                                                                 | Creates the cluster with private network only, see [Private Network Clusters](https://www.instaclustr.com/support/documentation/useful-information/private-network-clusters/).                                                                                                                                                               |
| slaTier               | string <br /> **required**                                                               | SLA Tier of the cluster. Non-production clusters may receive lower priority support and reduced SLAs. Production tier is not available when using Developer class nodes. See [SLA Tier](https://www.instaclustr.com/support/documentation/useful-information/sla-tier/) for more information. <br/>**Enum**: `PRODUCTION`, `NON_PRODUCTION`. |
| twoFactorDelete       | Array of objects ([TwoFactorDelete](#TwoFactorDeleteObject)) <br /> _mutable_            | Contacts that will be contacted when cluster request is sent.                                                                                                                                                                                                                                                                                |
| spark                 | Array of objects ([Spark](#SparkObject))                                                 | Adds the specified version of Apache Spark to the Cassandra cluster. **Available versions:** `2.3.2`,`3.0.1`                                                                                                                                                                                                                                 |
| luceneEnabled         | bool <br /> **required**                                                                 | Adds Apache Lucene to the Cassandra cluster.                                                                                                                                                                                                                                                                                                 |
| passwordAndUserAuth   | bool <br /> **required**                                                                 | Enables Password Authentication and User Authorization.                                                                                                                                                                                                                                                                                      |
| restoreFrom           | Object ([CassandraRestoreFrom](#CassandraRestoreFromObject))                             | Triggers a restore cluster operation.                                                                                                                                                                                                                                                                                                        |
| dataCentres           | Array of objects ([CassandraDataCentre](#CassandraDataCentreObject)) <br /> **required** | Object fields are described below as a bulleted list.                                                                                                                                                                                                                                                                                        |

### TwoFactorDeleteObject
| Field                   | Type                        | Description                                                                            |
|-------------------------|-----------------------------|----------------------------------------------------------------------------------------|
| confirmationPhoneNumber | string                      | The phone number which will be contacted when the cluster is requested to be deleted.  |
| confirmationEmail       | string <br /> **required**  | The email address which will be contacted when the cluster is requested to be deleted. |

### SparkObject
| Field      | Type                        | Description                                                                                                    |
|------------|-----------------------------|----------------------------------------------------------------------------------------------------------------|
| version    | string <br /> **required**  | Adds the specified version of Apache Spark to the Cassandra cluster. **Available versions:** `2.3.2`, `3.0.1`. |

### CassandraDataCentreObject
| Field                            | Type                                                                     | Description                                                                                                                                                                                                                                                                                            |
|----------------------------------|--------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| name                             | string <br /> **required**                                               | A logical name for the data centre within a cluster. These names must be unique in the cluster.                                                                                                                                                                                                        |
| region                           | string <br /> **required**                                               | Region of the Data Centre. See the description for node size for a compatible Data Centre for a given node size.                                                                                                                                                                                       |
| cloudProvider                    | string <br /> **required**                                               | Name of the cloud provider service in which the Data Centre will be provisioned. <br />**Enum**: `AWS_VPC` `GCP` `AZURE` `AZURE_AZ`.                                                                                                                                                                   |
| accountName                      | string                                                                   | For customers running in their own account. Your provider account can be found on the Create Cluster page on the Instaclustr Console, or the "Provider Account" property on any existing cluster. For customers provisioning on Instaclustr's cloud provider accounts, this property may be omitted.   |
| cloudProviderSettings            | Array of objects ([CloudProviderSettings](#CloudProviderSettingsObject)) | Cloud provider specific settings for the Data Centre.                                                                                                                                                                                                                                                  |
| network                          | string <br /> **required**                                               | The private network address block for the Data Centre specified using CIDR address notation. The network must have a prefix length between /12 and /22 and must be part of a private address space.                                                                                                    |
| nodeSize                         | string <br /> **required**<br />_mutable_                                | Size of the nodes provisioned in the Data Centre. Available node sizes, see [Instaclustr API docs NodeSize](https://instaclustr.redoc.ly/Current/tag/Cassandra-Cluster-V2#paths/~1cluster-management~1v2~1resources~1applications~1cassandra~1clusters~1v2/post!path=dataCentres/nodeSize&t=request).  |
| nodesNumber                      | int32 <br /> **required**<br />_mutable_                                 | Total number of nodes in the Data Centre. <br/>Available values: [1…5].                                                                                                                                                                                                                                |
| tags                             | map[string]string                                                        | List of tags to apply to the Data Centre. Tags are metadata labels which allow you to identify, categorise and filter clusters. This can be useful for grouping together clusters into applications, environments, or any category that you require.<br/>**Format**:<br/>tags:<br/>- key: value.       |
| replicationFactor                | int32 <br /> **required**                                                | Number of racks to use when allocating nodes.                                                                                                                                                                                                                                                          |                                                                                                                                                                                           
| continuousBackup                 | bool <br /> **required**                                                 | Enables commitlog backups and increases the frequency of the default snapshot backups.                                                                                                                                                                                                                 |                                                                                                                                                                                           
| privateIpBroadcastForDiscovery   | bool <br /> **required**                                                 | Enables broadcast of private IPs for auto-discovery.                                                                                                                                                                                                                                                   |                                                                                                                                                                                           
| clientToClusterEncryption        | bool <br /> **required**                                                 | Enables Client ⇄ Node Encryption.                                                                                                                                                                                                                                                                      |                                                                                                                                                                                           


### CloudProviderSettingsObject
| Field                  | Type   | Description                                                                                                                                                                                                                                                                                                                                                                                                                                 |
|------------------------|--------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| customVirtualNetworkId | string | **AWS**: VPC ID into which the Data Centre will be provisioned. The Data Centre's network allocation must match the IPv4 CIDR block of the specified VPC.<br />**GCP**: Network name or a relative Network or Subnetwork URI e.g. projects/my-project/regions/us-central1/subnetworks/my-subnet. The Data Centre's network allocation must match the IPv4 CIDR block of the specified subnet. <br />Cannot be provided with `resourceGroup` |
| resourceGroup          | string | The name of the Azure Resource Group into which the Data Centre will be provisioned. <br />Cannot be provided with `customVirtualNetworkId` and `diskEncryptionKey`                                                                                                                                                                                                                                                                         |
| diskEncryptionKey      | string | ID of a KMS encryption key to encrypt data on nodes. KMS encryption key must be set in Cluster Resources through the Instaclustr Console before provisioning an encrypted Data Centre. <br />Cannot be provided with `customVirtualNetworkId`                                                                                                                                                                                               |

### CassandraRestoreFromObject

| Field                 | Type                                                                         | Description                                                                                                                                                                                         |
|-----------------------|------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| clusterId             | string `<uuid>`<br /> **required**                                           | Original cluster ID. Backup from that cluster will be used for restore                                                                                                                              |
| clusterNameOverride   | string                                                                       | The display name of the restored cluster. <br/>By default, the restored cluster will be created with its current name appended with “restored” and the date & time it was requested to be restored. |
| cdcInfos              | Array of objects ([CassandraRestoreCDCInfo](#CassandraRestoreCDCInfoObject)) | An optional list of Cluster Data Centres for which custom VPC settings will be used. <br/>This property must be populated for all Cluster Data Centres or none at all.                              |
| pointInTime           | int64   `>= 1420070400000`                                                   | Timestamp in milliseconds since epoch. All backed up data will be restored for this point in time.                                                                                                  |
| keyspaceTables        | string                                                                       | A comma separated list of keyspace/table names which follows the format `<keyspace>.<table1>`, `<keyspace>.<table2>`.                                                                               |

### CassandraRestoreCDCInfoObject

| Field            | Type            | Description                                                                                                                             |
|------------------|-----------------|-----------------------------------------------------------------------------------------------------------------------------------------|
| cdcId            | string `<uuid>` | Cluster Data Centre ID                                                                                                                  |
| restoreToSameVpc | bool            | Determines if the restored cluster will be allocated to the existing VPC. <br/>Either restoreToSameVpc or customVpcId must be provided. |
| customVpcId      | string          | Custom VPC ID to which the restored cluster will be allocated. <br/>Either restoreToSameVpc or customVpcId must be provided.            |
| customVpcNetwork | string          | CIDR block in which the cluster will be allocated for a custom VPC.                                                                     |

## Cluster create flow
To create a Cassandra cluster instance you need to prepare the yaml manifest. Here is an example:

```yaml
# cassandra.yaml
apiVersion: clusters.instaclustr.com/v1alpha1
kind: Cassandra
metadata:
  name: cassandra-sample
spec:
  name: "CassandraTEST"
  version: "3.11.13"               # 4.0.1 is deprecated | 4.0.4 and 3.11.13 are supported
  dataCentres:
    - name: "AWS_cassandra"
      region: "US_EAST_1"
      cloudProvider: "AWS_VPC"
      continuousBackup: false
      nodesNumber: 2
      replicationFactor: 2
      privateIpBroadcastForDiscovery: false
      network: "172.16.0.0/19"
      tags:
        "tag": "testTag"
      clientToClusterEncryption: false
      nodeSize: "CAS-DEV-t4g.medium-30"
  pciCompliance: false
  luceneEnabled: false        # can be enabled only on 3.11.13 version of Cassandra
  passwordAndUserAuth: true
  slaTier: "NON_PRODUCTION"
```

Next, you need to apply this manifest in your K8s cluster. This will create a custom resource instance inside (more info about an apply command you can find [here](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#apply)):
```console
kubectl apply -f cassandra.yaml
```

Now you can get and describe the instance:
```console
kubectl get cassandras.clusters.instaclustr.com cassandra-sample
```
```console
kubectl describe cassandras.clusters.instaclustr.com cassandra-sample
```

After you have applied the entity, the Instaclustr operator will create it on your K8s cluster and send a request to the Instaclustr API. You can be sure, that the resource creation call was sent if the instance has an id field filled in the status section.

You can check access to the created Cassandra cluster from your kubernetes cluster and run some simple command to check that it is working with a lot of tools.
All available tools you can find in the Instaclustr console -> Choose your cluster -> Connection Info -> Examples section.
For example, how to connect with cqlsh tool you can see in [this doc](https://www.instaclustr.com/support/documentation/cassandra/using-cassandra/connect-to-cassandra-using-cqlsh/). 

## Cluster update flow
To update a cluster you can apply an updated cluster manifest or edit the custom resource instance in kubernetes cluster:
* Apply an updated cluster manifest:
    ```console
    kubectl apply -f cassandra.yaml
    ```
* Edit the custom resource instance:
    ```console
    kubectl edit cassandras.clusters.instaclustr.com cassandra-sample
    ```
You can only update fields that are **mutable**

## Cluster delete flow

### Cluster deletion
To delete cluster run:
```console
kubectl delete cassandras.clusters.instaclustr.com cassandra-sample
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
kubectl delete cassandras.clusters.instaclustr.com cassandra-sample
```

After that, deletion confirmation email will be sent to the email defined in the `confirmationEmail` field of `TwoFactorDelete`. When deletion is confirmed via email, Instaclustr support will delete the cluster and the related cluster resources inside K8s will be also removed.

## Cluster restore flow
To restore a Cassandra cluster instance from an existing one you need to prepare the yaml manifest. Here is an example:
```yaml
# cassandra-restore.yaml
apiVersion: clusters.instaclustr.com/v1alpha1
kind: Cassandra
metadata:
  name: cassandra-sample-restore
spec:
  restoreFrom:
    clusterId: 855c0704-c17d-421f-977e-6be0d5f86b77
    clusterNameOverride: "cassandraRestored"
    pointInTime: 1675501761821
```

Next, you need to apply this manifest in your K8s cluster. This will create a custom resource instance inside (more info about an apply command you can find [here](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#apply)):
```console
kubectl apply -f cassandra-restore.yaml
```

New cluster will be created from the backup of the "restored-from" cluster. Spec will be updated automatically.