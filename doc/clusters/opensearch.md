# OpenSearch cluster management

## Available spec fields

| Field                 | Type                                                                        | Description                                                                                                                                                                                                                                                                                                                                  |
|-----------------------|-----------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| name                  | String <br /> **required**                                                  | Cluster name. Should have length from 3 to 32 symbols.                                                                                                                                                                                                                                                                                       |
| version               | String <br /> **required**                                                  | OpenSearch instance version. <br />**Available versions**: `1.3.7`, `2.2.1`.                                                                                                                                                                                                                                                                 |
| privateNetworkCluster | bool                                                                        | Creates the cluster with private network only, see [Private Network Clusters](https://www.instaclustr.com/support/documentation/useful-information/private-network-clusters/).                                                                                                                                                               |
| slaTier               | String  <br /> **required**                                                 | SLA Tier of the cluster. Non-production clusters may receive lower priority support and reduced SLAs. Production tier is not available when using Developer class nodes. See [SLA Tier](https://www.instaclustr.com/support/documentation/useful-information/sla-tier/) for more information. <br/>**Enum**: `PRODUCTION`, `NON_PRODUCTION`. |
| twoFactorDelete       | Array of objects ([TwoFactorDelete](#TwoFactorDeleteObject))<br />_mutable_ | Contacts that will be contacted when cluster request is sent.                                                                                                                                                                                                                                                                                |
| dataCentres           | Array of objects ([OpenSearchDataCentre](#OpenSearchDataCentreObject))      |                                                                                                                                                                                                                                                                                                                                              |
| description           | String<br />_mutable_                                                       | A description of the cluster.                                                                                                                                                                                                                                                                                                                |
| concurrentResizes     | Int                                                                         | A number, from 1 to the count of nodes in the largest rack, that specifies how many nodes may be resized at the same time.                                                                                                                                                                                                                   |
| notifySupportContacts | bool                                                                        | If set to true will notify Instaclustr support about cluster resize                                                                                                                                                                                                                                                                          |
| privateLink           | Object ([PrivateLink](#PrivateLinkObject))                                  | Creates a PrivateLink cluster, see [PrivateLink](https://www.instaclustr.com/support/documentation/useful-information/privatelink/)                                                                                                                                                                                                          |
| openSearchRestoreFrom | Object ([OpenSearchRestoreFrom](#OpenSearchRestoreFromObject))              | Triggers a restore cluster operation.                                                                                                                                                                                                                                                                                                        |

### TwoFactorDeleteObject

| Field                   | Type                       | Description                                                                            |
|-------------------------|----------------------------|----------------------------------------------------------------------------------------|
| confirmationPhoneNumber | String                     | The phone number which will be contacted when the cluster is requested to be deleted.  |
| confirmationEmail       | String <br /> **required** | The email address which will be contacted when the cluster is requested to be deleted. |

### OpenSearchDataCentreObject

| Field                        | Type                                                                     | Description                                                                                                                                                                                                                                                                                          |
|------------------------------|--------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| name                         | String                                                                   | A logical name for the data centre within a cluster. These names must be unique in the cluster.                                                                                                                                                                                                      |
| region                       | String <br /> **required**                                               | Region of the Data Centre. See the description for node size for a compatible Data Centre for a given node size.                                                                                                                                                                                     |
| cloudProvider                | String    <br /> **required**                                            | Name of the cloud provider service in which the Data Centre will be provisioned. <br />**Enum**: `AWS_VPC` `GCP` `AZURE` `AZURE_AZ`                                                                                                                                                                  |
| cloudProviderSettings        | Array of objects ([CloudProviderSettings](#CloudProviderSettingsObject)) | Cloud provider specific settings for the Data Centre                                                                                                                                                                                                                                                 |
| network                      | String     <br /> **required**                                           | The private network address block for the Data Centre specified using CIDR address notation. The network must have a prefix length between /12 and /22 and must be part of a private address space.                                                                                                  |
| nodeSize                     | String     <br /> **required**<br />_mutable_                            | Size of the nodes provisioned in the Data Centre. Available node sizes, see [Instaclustr API docs NodeSize](https://instaclustr.redoc.ly/Current/tag/Cluster-Resource?_ga=2.261654733.1169092465.1661786388-1184637702.1659440305#operation/extendedProvisionRequestHandler!path=nodeSize&t=request) |
| nodesNumber                  | Int32  <br /> **required**                                               | Total number of nodes in the Data Centre. <br/>**Available values**: [1…100]                                                                                                                                                                                                                         |
| racksNumber                  | Int32   <br /> **required**                                              | Number of racks to use when allocating nodes. <br/>**Available values**: [2…5]                                                                                                                                                                                                                       |
| tags                         | map[string]string                                                        | List of tags to apply to the Data Centre. Tags are metadata labels which allow you to identify, categorise and filter clusters. This can be useful for grouping together clusters into applications, environments, or any category that you require.<br/>**Format**:<br/>tags:<br/>- key: value      |
| dedicatedMasterNodes         | bool                                                                     | This is an optional boolean parameter used to scale the load of the data node by having 3 dedicated master nodes for a particular cluster. We recommend selecting this to true for clusters with 9 and above nodes.                                                                                  |
| masterNodeSize               | string                                                                   | Only to be included if the if `dedicatedMasterNodes` is set to true.                                                                                                                                                                                                                                 |
| openSearchDashboardsNodeSize | string                                                                   | This is an optional string parameter to used to define the OpenSearch Dashboards node size. Only to be included if wish to add an OpenSearch Dashboards node to the cluster.                                                                                                                         |
| indexManagementPlugin        | bool                                                                     | Enables Index Management Plugin. This helps automate recurring index management activities.                                                                                                                                                                                                          |
| alertingPlugin               | bool                                                                     | Enables Alerting Plugin.                                                                                                                                                                                                                                                                             |
| icuPlugin                    | bool                                                                     | Enables ICU Plugin.                                                                                                                                                                                                                                                                                  |
| knnPlugin                    | bool                                                                     | Enables KNN Plugin.                                                                                                                                                                                                                                                                                  |
| notificationsPlugin          | bool                                                                     | Enables Notifications Plugin.                                                                                                                                                                                                                                                                        |
| reportsPlugin                | bool                                                                     | Enables Reports Plugin.                                                                                                                                                                                                                                                                              |

### CloudProviderSettingsObject

| Field                  | Type   | Description                                                                                                                                                                                                                                                                                                                                                                                                                                 |
|------------------------|--------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| customVirtualNetworkId | String | **AWS**: VPC ID into which the Data Centre will be provisioned. The Data Centre's network allocation must match the IPv4 CIDR block of the specified VPC.<br />**GCP**: Network name or a relative Network or Subnetwork URI e.g. projects/my-project/regions/us-central1/subnetworks/my-subnet. The Data Centre's network allocation must match the IPv4 CIDR block of the specified subnet. <br />Cannot be provided with `resourceGroup` |
| resourceGroup          | String | The name of the Azure Resource Group into which the Data Centre will be provisioned. <br />Cannot be provided with `customVirtualNetworkId` and `diskEncryptionKey`                                                                                                                                                                                                                                                                         |
| diskEncryptionKey      | String | ID of a KMS encryption key to encrypt data on nodes. KMS encryption key must be set in Cluster Resources through the Instaclustr Console before provisioning an encrypted Data Centre. <br />Cannot be provided with `customVirtualNetworkId`                                                                                                                                                                                               |

### PrivateLinkObject

| Field            | Type                                 | Description                                                |
|------------------|--------------------------------------|------------------------------------------------------------|
| iamPrincipalARNs | Array of strings <br /> **required** | List of IAM Principal ARNs to add to the endpoint service. |

### OpenSearchRestoreFromObject

| Field               | Type                                                                   | Description                                                                                                                                                                                         |
|---------------------|------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| clusterId           | String `<uuid>`<br /> **required**                                     | Original cluster ID. Backup from that cluster will be used for restore                                                                                                                              |
| clusterNameOverride | String                                                                 | The display name of the restored cluster. <br/>By default, the restored cluster will be created with its current name appended with “restored” and the date & time it was requested to be restored. |
| cdcInfos            | Array of objects ([OpenSearchRestoreCDCInfo](#PgRestoreCDCInfoObject)) | An optional list of Cluster Data Centres for which custom VPC settings will be used. <br/>This property must be populated for all Cluster Data Centres or none at all.                              |
| pointInTime         | Int64   `>= 1420070400000`                                             | Timestamp in milliseconds since epoch. All backed up data will be restored for this point in time.                                                                                                  |
| indexNames          | String                                                                 | Only data for the specified indices will be restored, for the point in time.                                                                                                                        |
| clusterNetwork      | String                                                                 | The cluster network for this cluster to be restored to.                                                                                                                                             |

### OpenSearchRestoreCDCInfoObject

| Field            | Type            | Description                                                                                                                             |
|------------------|-----------------|-----------------------------------------------------------------------------------------------------------------------------------------|
| cdcId            | String `<uuid>` | Cluster Data Centre ID                                                                                                                  |
| restoreToSameVpc | bool            | Determines if the restored cluster will be allocated to the existing VPC. <br/>Either restoreToSameVpc or customVpcId must be provided. |
| customVpcId      | String          | Custom VPC ID to which the restored cluster will be allocated. <br/>Either restoreToSameVpc or customVpcId must be provided.            |
| customVpcNetwork | String          | CIDR block in which the cluster will be allocated for a custom VPC.                                                                     |

## Cluster creation example

To create a cluster you need to prepare a cluster manifest. Here is an example:
```yaml
# opensearch.yaml file
apiVersion: clusters.instaclustr.com/v1alpha1
kind: OpenSearch
metadata:
  name: opensearch-sample
spec:
  name: "k8sOpOS"
  version: "opensearch:2.2.1"
  concurrentResizes: 1
  notifySupportContacts: false
  dataCentres:
    - region: "US_WEST_2"
      network: "10.1.0.0/16"
      cloudProvider: "AWS_VPC"
      nodeSize: "SRH-DEV-t4g.small-5"
      racksNumber: 3
      nodesNumber: 1
      indexManagementPlugin: false
  slaTier: "NON_PRODUCTION"
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

## Cluster deletion example

### Cluster deletion

To delete cluster run:
```console
kubectl delete opensearches.clusters.instaclustr.com opensearch-sample
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
kubectl delete opensearches.clusters.instaclustr.com opensearch-sample
```

After that deletion confirmation email will be sent to the email defined in the `confirmationEmail` field of `TwoFactorDelete`. When deletion is confirmed via email, Instaclustr support will delete the cluster and the related cluster resource in kubernetes will be removed soon.

## Cluster restore example

To restore a OpenSearch cluster instance from an existing one you need to prepare the yaml manifest. Here is an example:
```yaml
# opensearch-restore.yaml
apiVersion: clusters.instaclustr.com/v1alpha1
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