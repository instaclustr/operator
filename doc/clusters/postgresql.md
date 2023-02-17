# PostgreSQL cluster management

## Available spec fields

| Field                 | Type                                                                        | Description                                                                                                                                                                                                                                                                                                                                  |
|-----------------------|-----------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| name                  | string <br /> **required**                                                  | Cluster name. Should have length from 3 to 32 symbols.                                                                                                                                                                                                                                                                                       |
| version               | string <br /> **required**                                                  | PostgreSQL instance version. <br />**Available versions**: `15.1.0`, `14.6.0`, `14.5.0`, `13.9.0`, `13.8.0`.                                                                                                                                                                                                                                 |
| privateNetworkCluster | bool                                                                        | Creates the cluster with private network only, see [Private Network Clusters](https://www.instaclustr.com/support/documentation/useful-information/private-network-clusters/).                                                                                                                                                               |
| slaTier               | string  <br /> **required**                                                 | SLA Tier of the cluster. Non-production clusters may receive lower priority support and reduced SLAs. Production tier is not available when using Developer class nodes. See [SLA Tier](https://www.instaclustr.com/support/documentation/useful-information/sla-tier/) for more information. <br/>**Enum**: `PRODUCTION`, `NON_PRODUCTION`. |
| twoFactorDelete       | Array of objects ([TwoFactorDelete](#TwoFactorDeleteObject))<br />_mutable_ | Contacts that will be contacted when cluster request is sent.                                                                                                                                                                                                                                                                                |
| dataCentres           | Array of objects ([PgDataCentre](#PgDataCentreObject)) <br />_mutable_      | List of data centre settings.                                                                                                                                                                                                                                                                                                                |
| clusterConfigurations | map[string]string<br />_mutable_                                            | PostgreSQL cluster configurations. Cluster nodes will need to be manually reloaded to apply configuration changes. <br />**Format**:<br />clusterConfigurations:<br />- key: value                                                                                                                                                           |
| description           | string<br />_mutable_                                                       | A description of the cluster.                                                                                                                                                                                                                                                                                                                |
| synchronousModeStrict | bool                                                                        | Create the PostgreSQL cluster with the selected replication mode, see [PostgreSQL replication mode](https://www.instaclustr.com/support/documentation/postgresql/options/replication-mode/).                                                                                                                                                 |
| pgRestoreFrom         | Object ([PgRestoreFrom](#PgRestoreFromObject))                              | Triggers a restore cluster operation.                                                                                                                                                                                                                                                                                                        |

### TwoFactorDeleteObject

| Field                   | Type                       | Description                                                                            |
|-------------------------|----------------------------|----------------------------------------------------------------------------------------|
| confirmationPhoneNumber | string                     | The phone number which will be contacted when the cluster is requested to be deleted.  |
| confirmationEmail       | string <br /> **required** | The email address which will be contacted when the cluster is requested to be deleted. |

### PgDataCentreObject

| Field                      | Type                                                                               | Description                                                                                                                                                                                                                                                                                                                                                          |
|----------------------------|------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| name                       | string                                                                             | A logical name for the data centre within a cluster. These names must be unique in the cluster.                                                                                                                                                                                                                                                                      |
| region                     | string <br /> **required**                                                         | Region of the Data Centre. See the description for node size for a compatible Data Centre for a given node size.                                                                                                                                                                                                                                                     |
| cloudProvider              | string    <br /> **required**                                                      | Name of the cloud provider service in which the Data Centre will be provisioned. <br />**Enum**: `AWS_VPC` `GCP` `AZURE` `AZURE_AZ`                                                                                                                                                                                                                                  |
| cloudProviderSettings      | Array of objects ([CloudProviderSettings](#CloudProviderSettingsObject))           | Cloud provider specific settings for the Data Centre                                                                                                                                                                                                                                                                                                                 |
| network                    | string     <br /> **required**                                                     | The private network address block for the Data Centre specified using CIDR address notation. The network must have a prefix length between /12 and /22 and must be part of a private address space.                                                                                                                                                                  |
| nodeSize                   | string     <br /> **required**<br />_mutable_                                      | Size of the nodes provisioned in the Data Centre. Available node sizes, see [Instaclustr API docs NodeSize](https://instaclustr.redoc.ly/Current/tag/Postgresql-Cluster-V2?_ga=2.261654733.1169092465.1661786388-1184637702.1659440305#paths/~1cluster-management~1v2~1resources~1applications~1postgresql~1clusters~1v2~1/post!path=dataCentres/nodeSize&t=request) |
| nodesNumber                | int32  <br /> **required**                                                         | Total number of nodes in the Data Centre. <br/>Available values: [1…5]                                                                                                                                                                                                                                                                                               |
| tags                       | map[string]string                                                                  | List of tags to apply to the Data Centre. Tags are metadata labels which allow you to identify, categorise and filter clusters. This can be useful for grouping together clusters into applications, environments, or any category that you require.<br/>**Format**:<br/>tags:<br/>- key: value                                                                      |
| clientEncryption           | bool                                                                               | Enable client to cluster Encryption.                                                                                                                                                                                                                                                                                                                                 |
| interDataCentreReplication | Array of objects ([InterDataCentreReplication](#InterDataCentreReplicationObject)) |                                                                                                                                                                                                                                                                                                                                                                      |
| intraDataCentreReplication | Array of objects ([IntraDataCentreReplication](#IntraDataCentreReplicationObject)) |                                                                                                                                                                                                                                                                                                                                                                      |
| pgBouncerVersion           | string                                                                             | Version of Pg Bouncer to run on the cluster. Required to enable Pg Bouncer. <br/>**Available versions**: `1.17.0`                                                                                                                                                                                                                                                    |
| poolMode                   | string                                                                             | Creates PgBouncer with the selected mode, see PgBouncer pool modes. Only available with `pgBouncerVersion` filled. <br/>**Enum**: `TRANSACTION` `SESSION` `STATEMENT`                                                                                                                                                                                                |

### CloudProviderSettingsObject

| Field                  | Type   | Description                                                                                                                                                                                                                                                                                                                                                                                                                                 |
|------------------------|--------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| customVirtualNetworkId | string | **AWS**: VPC ID into which the Data Centre will be provisioned. The Data Centre's network allocation must match the IPv4 CIDR block of the specified VPC.<br />**GCP**: Network name or a relative Network or Subnetwork URI e.g. projects/my-project/regions/us-central1/subnetworks/my-subnet. The Data Centre's network allocation must match the IPv4 CIDR block of the specified subnet. <br />Cannot be provided with `resourceGroup` |
| resourceGroup          | string | The name of the Azure Resource Group into which the Data Centre will be provisioned. <br />Cannot be provided with `customVirtualNetworkId` and `diskEncryptionKey`                                                                                                                                                                                                                                                                         |
| diskEncryptionKey      | string | ID of a KMS encryption key to encrypt data on nodes. KMS encryption key must be set in Cluster Resources through the Instaclustr Console before provisioning an encrypted Data Centre. <br />Cannot be provided with `customVirtualNetworkId`                                                                                                                                                                                               |

### InterDataCentreReplicationObject

| Field               | Type                     | Description                                                                                        |
|---------------------|--------------------------|----------------------------------------------------------------------------------------------------|
| isPrimaryDataCentre | bool <br /> **required** | Is this Data centre considered to be the primary (only required if multiple data centres defined). |

### IntraDataCentreReplicationObject

| Field           | Type                     | Description                                                                                                                                                                                                                              |
|-----------------|--------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| replicationMode | bool <br /> **required** | Create the PostgreSQL cluster with the selected replication mode, see [PostgreSQL replication mode](https://www.instaclustr.com/support/documentation/postgresql/options/replication-mode/). <br/>**Enum**: `ASYNCHRONOUS` `SYNCHRONOUS` |

### PgRestoreFromObject

| Field               | Type                                                           | Description                                                                                                                                                                                         |
|---------------------|----------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| clusterId           | string `<uuid>`<br /> **required**                             | Original cluster ID. Backup from that cluster will be used for restore                                                                                                                              |
| clusterNameOverride | string                                                         | The display name of the restored cluster. <br/>By default, the restored cluster will be created with its current name appended with “restored” and the date & time it was requested to be restored. |
| cdcInfos            | Array of objects ([PgRestoreCDCInfo](#PgRestoreCDCInfoObject)) | An optional list of Cluster Data Centres for which custom VPC settings will be used. <br/>This property must be populated for all Cluster Data Centres or none at all.                              |
| pointInTime         | int64   `>= 1420070400000`                                     | Timestamp in milliseconds since epoch. All backed up data will be restored for this point in time.                                                                                                  |

### PgRestoreCDCInfoObject

| Field            | Type            | Description                                                                                                                             |
|------------------|-----------------|-----------------------------------------------------------------------------------------------------------------------------------------|
| cdcId            | string `<uuid>` | Cluster Data Centre ID                                                                                                                  |
| restoreToSameVpc | bool            | Determines if the restored cluster will be allocated to the existing VPC. <br/>Either restoreToSameVpc or customVpcId must be provided. |
| customVpcId      | string          | Custom VPC ID to which the restored cluster will be allocated. <br/>Either restoreToSameVpc or customVpcId must be provided.            |
| customVpcNetwork | string          | CIDR block in which the cluster will be allocated for a custom VPC.                                                                     |

## Cluster creation example

To create a cluster you need to prepare a cluster manifest. Here is an example:
```yaml
# postgresql.yaml file
apiVersion: clusters.instaclustr.com/v1alpha1
kind: PostgreSQL
metadata:
  name: postgresql-sample
  annotations:
    testAnnotation: test
spec:
  name: "examplePostgre"
  version: "14.5.0"
  dataCentres:
    - region: "US_WEST_2"
      network: "10.1.0.0/16"
      cloudProvider: "AWS_VPC"
      nodeSize: "PGS-DEV-t4g.small-5"
      nodesNumber: 2
      clientEncryption: false
      name: "exampleDC1"
      intraDataCentreReplication:
        - replicationMode: "SYNCHRONOUS"
      interDataCentreReplication:
        - isPrimaryDataCentre: true
  description: "example"
  slaTier: "NON_PRODUCTION"
  privateNetworkCluster: false
  synchronousModeStrict: false
```

Next you need to apply this manifest. This will create PostgreSQL custom resource instance:
```console
kubectl apply -f postgresql.yaml
```

Now you can get and describe the instance:
```console
kubectl get postgresqls.clusters.instaclustr.com postgresql-sample
```
```console
kubectl describe postgresqls.clusters.instaclustr.com postgresql-sample
```

Cluster was created on Instaclustr premise if the instance has an id in the status section.

## Cluster update example

To update a cluster you can apply an updated cluster manifest or edit the custom resource instance in kubernetes cluster:
* Apply an updated cluster manifest:
    ```console
    kubectl apply -f postgresql.yaml
    ```
* Edit the custom resource instance:
    ```console
    kubectl edit postgresqls.clusters.instaclustr.com postgresql-sample
    ```
You can only update fields that are **mutable**

## Cluster deletion example

### Cluster deletion

To delete cluster run:
```console
kubectl delete postgresqls.clusters.instaclustr.com postgresql-sample
```

It can take some time to delete all related resources and cluster itself.

### Cluster deletion with twoFactorDelete option enabled

To delete cluster with twoFactorDelete option enabled you need to set the confirmation annotation to true: 
```yaml
Annotations:  
  "instaclustr.com/deletionConfirmed": true
```

And then run:
```console
kubectl delete postgresqls.clusters.instaclustr.com postgresql-sample
```

After that deletion confirmation email will be sent to the email defined in the `confirmationEmail` field of `TwoFactorDelete`. When deletion is confirmed via email, Instaclustr support will delete the cluster and the related cluster resource in kubernetes will be removed soon.

## Update Default User Password

Default user password is stored in a kubernetes secret. It is created before cluster creation. It has name template `default-user-password-<resource-name>`<br />
To update default user password you need to edit `defaultUserPassword` field in the secret:
```console
kubectl edit default-user-password-postgresql-sample
```

## Cluster restore example

To restore a PostgreSQL cluster instance from an existing one you need to prepare the yaml manifest. Here is an example:
```yaml
# postgresql-restore.yaml
apiVersion: clusters.instaclustr.com/v1alpha1
kind: PostgreSQL
metadata:
  name: postgresql-sample-restored
  annotations:
    testAnnotation: test
spec:
  pgRestoreFrom:
    clusterId: 826e00dc-eda5-426a-9185-48fc35f13dfc
    clusterNameOverride: "pgRestoredTest"
    pointInTime: 1676371087358
    clusterNetwork: "10.13.0.0/16"
```

Next, you need to apply this manifest in your K8s cluster. This will create a custom resource instance inside:
```console
kubectl apply -f postgresql-restore.yaml
```

New cluster will be created from the backup of the "restored-from" cluster. Spec will be updated automatically.