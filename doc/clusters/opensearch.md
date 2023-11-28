# OpenSearch cluster management

## Available spec fields

| Field                     | Type                                                                                        | Description                                                                                                                                                                                                                                                                                                                                   |
|---------------------------|---------------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| name                      | string <br> **required**                                                                    | Cluster name. Should have length from 3 to 32 symbols.                                                                                                                                                                                                                                                                                        |
| version                   | string <br> **required**                                                                    | OpenSearch instance version.                                                                                                                                                                                                                                                                                                                  |
| privateNetworkCluster     | bool <br> **required**                                                                      | Creates the cluster with private network only, see [Private Network Clusters](https://www.instaclustr.com/support/documentation/useful-information/private-network-clusters/).                                                                                                                                                                |
| slaTier                   | string  <br> **required**                                                                   | SLA Tier of the cluster. Non-production clusters may receive lower priority support and reduced SLAs. Production tier is not available when using Developer class nodes. See [SLA Tier](https://www.instaclustr.com/support/documentation/useful-information/sla-tier/) for more information. <br/>**Enum**: `PRODUCTION`, `NON_PRODUCTION`.  |
| twoFactorDelete           | Array of objects ([TwoFactorDelete](#TwoFactorDeleteObject))                                | Contacts that will be contacted when cluster request is sent.                                                                                                                                                                                                                                                                                 |
| dataCentres               | Array of objects ([OpenSearchDataCentre](#OpenSearchDataCentreObject))    <br> **required** | List of data centre settings.                                                                                                                                                                                                                                                                                                                 |                                                                                                                                                                                                                              
| privateLink               | bool                                                                                        | Creates the cluster with private network only, see [Private Network Clusters](https://www.instaclustr.com/support/documentation/useful-information/private-network-clusters/)                                                                                                                                                                 |
| openSearchRestoreFrom     | Object ([OpenSearchRestoreFrom](#OpenSearchRestoreFromObject))                              | Triggers a restore cluster operation.                                                                                                                                                                                                                                                                                                         |
| bundledUseOnly            | bool                                                                                        | Provision this cluster for Bundled Use only.                                                                                                                                                                                                                                                                                                  |
| pciCompliance             | bool <br> **required**                                                                      | Creates a PCI compliant cluster, see [PCI Compliance](https://www.instaclustr.com/support/documentation/useful-information/pci-compliance/)                                                                                                                                                                                                   |
| clusterManagerNodes       | Array of objects ([ClusterManagerNodes](#ClusterManagerNodes))  <br> **required**           | List of cluster managers node settings                                                                                                                                                                                                                                                                                                        |
| indexManagementPlugin     | bool                                                                                        | Enables Index Management Plugin. This helps automate recurring index management activities.                                                                                                                                                                                                                                                   |
| alertingPlugin            | bool                                                                                        | Enables Alerting Plugin.                                                                                                                                                                                                                                                                                                                      |
| icuPlugin                 | bool                                                                                        | Enables ICU Plugin.                                                                                                                                                                                                                                                                                                                           |
| asynchronousSearchPlugin  | bool                                                                                        | Enables asynchronousSearch plugin.                                                                                                                                                                                                                                                                                                            |
| anomalyDetectionPlugin    | bool                                                                                        | Enables anomalyDetection plugin.                                                                                                                                                                                                                                                                                                              |
| sqlPlugin                 | bool                                                                                        | Enables sql plugin.                                                                                                                                                                                                                                                                                                                           |
| knnPlugin                 | bool                                                                                        | Enables knn plugin.                                                                                                                                                                                                                                                                                                                           |
| notificationsPlugin       | bool                                                                                        | Enables notifications plugin.                                                                                                                                                                                                                                                                                                                 |
| reportingPlugin           | bool                                                                                        | Enables reporting plugin.                                                                                                                                                                                                                                                                                                                     |
| loadBalancer              | bool                                                                                        | Enables Load Balancer.                                                                                                                                                                                                                                                                                                                        |
| dataNodes                 | Array of objects ([DataNodes](#DataNodes))                                                  | List of data node settings                                                                                                                                                                                                                                                                                                                    |
| dashboards                | Array of objects ([Dashboards](#Dashboards))                                                | List of dashboards node settings                                                                                                                                                                                                                                                                                                              |
| description               | string <br />                                                                               | A description of the cluster                                                                                                                                                                                                                                                                                                                  |

### DataNodes 

| Field               | Type                                      | Description       |
|---------------------|-------------------------------------------|-------------------|
| nodeSize            | string   <br> **required** <br> _mutable_ | Size of data node |
| nodesNumber         | int32  <br> **required** <br> _mutable_   | Number of nodes   |


### Dashboards

| Field               | Type                         | Description                                       |
|---------------------|------------------------------|---------------------------------------------------|
| nodeSize            | string    <br> **required**  | Size of the nodes provisioned as Dashboards nodes |
| oidcProvider        | string                       | OIDC provider                                     |
| version             | string     <br> **required** | Version of dashboard to run on the cluster.       |


### ClusterManagerNodes

| Field               | Type                      | Description                                                                                                                                                                                       |
|---------------------|---------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| nodeSize            | string <br> **required**  | Manager node size                                                                                                                                                                                 | 
| dedicatedManager    | bool <br> **required**    | This is a parameter used to scale the load of the data node by having 3 dedicated master nodes for a particular cluster. We recommend selecting this to true for clusters with 9 and above nodes. |


### TwoFactorDeleteObject

| Field                   | Type                     | Description                                                                            |
|-------------------------|--------------------------|----------------------------------------------------------------------------------------|
| confirmationPhoneNumber | string                   | The phone number which will be contacted when the cluster is requested to be deleted.  |
| confirmationEmail       | string <br> **required** | The email address which will be contacted when the cluster is requested to be deleted. |

### OpenSearchDataCentreObject

| Field                 | Type                                                                     | Description                                                                                                                                                                                                                                                                                     |
|-----------------------|--------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| name                  | string                                                                   | A logical name for the data centre within a cluster. These names must be unique in the cluster.                                                                                                                                                                                                 |
| region                | string <br> **required**                                                 | Region of the Data Centre. See the description for node size for a compatible Data Centre for a given node size.                                                                                                                                                                                |
| cloudProvider         | string    <br> **required**                                              | Name of the cloud provider service in which the Data Centre will be provisioned. <br />**Enum**: `AWS_VPC` `GCP` `AZURE` `AZURE_AZ`                                                                                                                                                             |
| accountName           | string    <br> **required**                                              | For customers running in their own account. For customers provisioning on Instaclustr's cloud provider accounts, this property may be omitted.                                                                                                                                                  |
| cloudProviderSettings | Array of objects ([CloudProviderSettings](#CloudProviderSettingsObject)) | Cloud provider specific settings for the Data Centre                                                                                                                                                                                                                                            |
| network               | string     <br> **required**                                             | The private network address block for the Data Centre specified using CIDR address notation. The network must have a prefix length between /12 and /22 and must be part of a private address space.                                                                                             |
| replicationFactor     | int32   <br> **required**                                                | Number of racks to use when allocating nodes. <br/>**Available values**: [2…5]                                                                                                                                                                                                                  |
| tags                  | map[string]string                                                        | List of tags to apply to the Data Centre. Tags are metadata labels which allow you to identify, categorise and filter clusters. This can be useful for grouping together clusters into applications, environments, or any category that you require.<br/>**Format**:<br/>tags:<br/>- key: value |
| privateLink           | bool                                                                     | Create a PrivateLink enabled cluster, see [PrivateLink](https://www.instaclustr.com/support/documentation/useful-information/privatelink/)                                                                                                                                                      |

### CloudProviderSettingsObject

| Field                  | Type   | Description                                                                                                                                                                                                                                                                                                                                                                                                                                 |
|------------------------|--------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| customVirtualNetworkId | string | **AWS**: VPC ID into which the Data Centre will be provisioned. The Data Centre's network allocation must match the IPv4 CIDR block of the specified VPC.<br />**GCP**: Network name or a relative Network or Subnetwork URI e.g. projects/my-project/regions/us-central1/subnetworks/my-subnet. The Data Centre's network allocation must match the IPv4 CIDR block of the specified subnet. <br />Cannot be provided with `resourceGroup` |
| resourceGroup          | string | The name of the Azure Resource Group into which the Data Centre will be provisioned. <br />Cannot be provided with `customVirtualNetworkId` and `diskEncryptionKey`                                                                                                                                                                                                                                                                         |
| diskEncryptionKey      | string | ID of a KMS encryption key to encrypt data on nodes. KMS encryption key must be set in Cluster Resources through the Instaclustr Console before provisioning an encrypted Data Centre. <br />Cannot be provided with `customVirtualNetworkId`                                                                                                                                                                                               |

### OpenSearchRestoreFromObject

| Field               | Type                                                                   | Description                                                                                                                                                                                         |
|---------------------|------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| clusterId           | string `<uuid>`<br> **required**                                       | Original cluster ID. Backup from that cluster will be used for restore                                                                                                                              |
| clusterNameOverride | string                                                                 | The display name of the restored cluster. <br/>By default, the restored cluster will be created with its current name appended with “restored” and the date & time it was requested to be restored. |
| cdcInfos            | Array of objects ([OpenSearchRestoreCDCInfo](#PgRestoreCDCInfoObject)) | An optional list of Cluster Data Centres for which custom VPC settings will be used. <br/>This property must be populated for all Cluster Data Centres or none at all.                              |
| pointInTime         | int64   `>= 1420070400000`                                             | Timestamp in milliseconds since epoch. All backed up data will be restored for this point in time.                                                                                                  |
| indexNames          | string                                                                 | Only data for the specified indices will be restored, for the point in time.                                                                                                                        |
| clusterNetwork      | string                                                                 | The cluster network for this cluster to be restored to.                                                                                                                                             |

### OpenSearchRestoreCDCInfoObject

| Field            | Type            | Description                                                                                                                             |
|------------------|-----------------|-----------------------------------------------------------------------------------------------------------------------------------------|
| cdcId            | string `<uuid>` | Cluster Data Centre ID                                                                                                                  |
| restoreToSameVpc | bool            | Determines if the restored cluster will be allocated to the existing VPC. <br/>Either restoreToSameVpc or customVpcId must be provided. |
| customVpcId      | string          | Custom VPC ID to which the restored cluster will be allocated. <br/>Either restoreToSameVpc or customVpcId must be provided.            |
| customVpcNetwork | string          | CIDR block in which the cluster will be allocated for a custom VPC.                                                                     |

## Cluster creation example

To create a cluster you need to prepare a cluster manifest. Here is an example:
```yaml
# opensearch.yaml file
apiVersion: clusters.instaclustr.com/v1beta1
kind: OpenSearch
metadata:
  name: opensearch-sample
  annotations:
    test.annotation/first: testAnnotation
spec:
  name: opensearch-sample
  alertingPlugin: false
  anomalyDetectionPlugin: false
  asynchronousSearchPlugin: false
  clusterManagerNodes:
    - dedicatedManager: false
      nodeSize: SRH-DEV-t4g.small-30
  dataCentres:
    - cloudProvider: AWS_VPC
      name: AWS_VPC_US_EAST_1
      network: 10.0.0.0/16
      replicationFactor: 3
      privateLink: false
      region: US_EAST_1
  ingestNodes:
    - nodeSize: SRH-DI-PRD-m6g.xlarge-10
      nodeCount: 3
  dataNodes:
    - nodeNumber: 3
      nodeSize: SRH-DEV-t4g.small-5
  icuPlugin: false
  indexManagementPlugin: true
  knnPlugin: false
  loadBalancer: false
  notificationsPlugin: false
  opensearchDashboards:
    - nodeSize: SRH-DEV-t4g.small-5
      oidcProvider: ''
      version: opensearch-dashboards:2.5.0
  version: 2.9.0
  pciCompliance: false
  privateNetworkCluster: false
  reportingPlugin: false
  slaTier: NON_PRODUCTION
  sqlPlugin: false
  resizeSettings:
    - notifySupportContacts: false
      concurrency: 3
```

Next you need to apply this manifest. This will create OpenSearch custom resource instance:
```console
kubectl apply -f opensearch.yaml
```

Now you can get and describe the instance:
```console
kubectl get opensearches.clusters.instaclustr.com opensearch-sample
```
```console
kubectl describe opensearches.clusters.instaclustr.com opensearch-sample
```

Cluster was created on Instaclustr premise if the instance has an id in the status section.

You can check access to the created cluster from your kubernetes cluster and run some simple command to check that it is working with a lot of tools.
All available tools you can find in the Instaclustr console -> Choose your cluster -> Connection Info -> Examples section.

When a cluster is provisioned, a new service will be created along with it that expose public IP addresses. You can use this service name (pattern **{k8s_cluster_name}-service**) instead of public IP addresses and ports to connect to and interact with your cluster.
To do this, the public IP address of your machine must be added to the Firewall Rules tab of your cluster in the Instaclustr console.
![Firewall Rules icon](../images/firewall_rules_screen.png "Firewall Rules icon")
Then, you can use the service name instead of public addresses and port.

## Cluster update example

To update a cluster you can apply an updated cluster manifest or edit the custom resource instance in kubernetes cluster:
* Apply an updated cluster manifest:
    ```console
    kubectl apply -f opensearch.yaml
    ```
* Edit the custom resource instance:
    ```console
    kubectl edit opensearches.clusters.instaclustr.com opensearch-sample
    ```
You can only update fields that are **mutable**

## Cluster delete flow

### Cluster deletion
To delete cluster run:
```console
kubectl delete opensearches.clusters.instaclustr.com opensearch-sample
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

## Cluster restore example

To restore a OpenSearch cluster instance from an existing one you need to prepare the yaml manifest. Here is an example:
```yaml
# opensearch-restore.yaml
apiVersion: clusters.instaclustr.com/v1beta1
kind: OpenSearch
metadata:
  name: opensearch-sample-restore
spec:
  restoreFrom:
    clusterId: "3da96493-cf3a-40b4-ba78-cf33a2290419"
    pointInTime: 1676369341601
    clusterNameOverride: "opensearch-restore-test"
    clusterNetwork: "10.12.0.0/16"
```

Next, you need to apply this manifest in your K8s cluster. This will create a custom resource instance inside:
```console
kubectl apply -f opensearch-restore.yaml
```

New cluster will be created from the backup of the "restored-from" cluster. Spec will be updated automatically.