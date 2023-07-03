# Zookeeper cluster management

## Available spec fields
| Field                 | Type                                                                                     | Description                                                                                                                                                                                                                                                                                                                                  |
|-----------------------|------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| name                  | string <br /> **required**                                                               | Cluster name. Should have length from 3 to 32 symbols.                                                                                                                                                                                                                                                                                       |
| version               | string <br /> **required**                                                               | Zookeeper instance version.                                                                                                                                                                                                                                                                                                                  |
| pciCompliance         | bool <br /> **required**                                                                 | Creates a PCI compliant cluster, see [PCI Compliance](https://www.instaclustr.com/support/documentation/useful-information/pci-compliance/)                                                                                                                                                                                                  |
| privateNetworkCluster | bool <br /> **required**                                                                 | Creates the cluster with private network only, see [Private Network Clusters](https://www.instaclustr.com/support/documentation/useful-information/private-network-clusters/).                                                                                                                                                               |
| slaTier               | string <br /> **required**                                                               | SLA Tier of the cluster. Non-production clusters may receive lower priority support and reduced SLAs. Production tier is not available when using Developer class nodes. See [SLA Tier](https://www.instaclustr.com/support/documentation/useful-information/sla-tier/) for more information. <br/>**Enum**: `PRODUCTION`, `NON_PRODUCTION`. |
| twoFactorDelete       | Array of objects ([TwoFactorDelete](#TwoFactorDeleteObject))                             | Contacts that will be contacted when cluster request is sent.                                                                                                                                                                                                                                                                                |
| dataCentres           | Array of objects ([ZookeeperDataCentre](#ZookeeperDataCentreObject)) <br /> **required** | Object fields are described below as a bulleted list.                                                                                                                                                                                                                                                                                        |

### TwoFactorDeleteObject
| Field                   | Type                       | Description                                                                            |
|-------------------------|----------------------------|----------------------------------------------------------------------------------------|
| confirmationPhoneNumber | string                     | The phone number which will be contacted when the cluster is requested to be deleted.  |
| confirmationEmail       | string <br /> **required** | The email address which will be contacted when the cluster is requested to be deleted. |

### ZookeeperDataCentreObject
| Field                         | Type                                                                     | Description                                                                                                                                                                                                                                                                                                  |
|-------------------------------|--------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| name                          | string <br /> **required**                                               | A logical name for the data centre within a cluster. These names must be unique in the cluster.                                                                                                                                                                                                              |
| region                        | string <br /> **required**                                               | Region of the Data Centre. See the description for node size for a compatible Data Centre for a given node size.                                                                                                                                                                                             |
| cloudProvider                 | string <br /> **required**                                               | Name of the cloud provider service in which the Data Centre will be provisioned. <br />**Enum**: `AWS_VPC` `GCP` `AZURE` `AZURE_AZ`.                                                                                                                                                                         |
| providerAccountName           | string                                                                   | For customers running in their own account. Your provider account can be found on the Create Cluster page on the Instaclustr Console, or the "Provider Account" property on any existing cluster. For customers provisioning on Instaclustr's cloud provider accounts, this property may be omitted.         |
| cloudProviderSettings         | Array of objects ([CloudProviderSettings](#CloudProviderSettingsObject)) | Cloud provider specific settings for the Data Centre.                                                                                                                                                                                                                                                        |
| network                       | string <br /> **required**                                               | The private network address block for the Data Centre specified using CIDR address notation. The network must have a prefix length between /12 and /22 and must be part of a private address space.                                                                                                          |
| nodeSize                      | string <br /> **required**                                               | Size of the nodes provisioned in the Data Centre. Available node sizes, see [Instaclustr API docs NodeSize](https://instaclustr.redoc.ly/Current/tag/Apache-Zookeeper-Cluster-V2#paths/~1cluster-management~1v2~1resources~1applications~1zookeeper~1clusters~1v2/post!path=dataCentres/nodeSize&t=request). |
| nodesNumber                   | int32 <br /> **required**                                                | Total number of nodes in the Data Centre. <br/>Available values: [1…5].                                                                                                                                                                                                                                      |
| tags                          | map[string]string                                                        | List of tags to apply to the Data Centre. Tags are metadata labels which allow you to identify, categorise and filter clusters. This can be useful for grouping together clusters into applications, environments, or any category that you require.<br/>**Format**:<br/>tags:<br/>- key: value.             |
| clientToServerEncryption      | bool <br /> **required**                                                 | Enables Client ⇄ Node Encryption.                                                                                                                                                                                                                                                                            |

### CloudProviderSettingsObject
| Field                  | Type   | Description                                                                                                                                                                                                                                                                                                                                                                                                                                 |
|------------------------|--------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| customVirtualNetworkId | string | **AWS**: VPC ID into which the Data Centre will be provisioned. The Data Centre's network allocation must match the IPv4 CIDR block of the specified VPC.<br />**GCP**: Network name or a relative Network or Subnetwork URI e.g. projects/my-project/regions/us-central1/subnetworks/my-subnet. The Data Centre's network allocation must match the IPv4 CIDR block of the specified subnet. <br />Cannot be provided with `resourceGroup` |
| resourceGroup          | string | The name of the Azure Resource Group into which the Data Centre will be provisioned. <br />Cannot be provided with `customVirtualNetworkId` and `diskEncryptionKey`                                                                                                                                                                                                                                                                         |
| diskEncryptionKey      | string | ID of a KMS encryption key to encrypt data on nodes. KMS encryption key must be set in Cluster Resources through the Instaclustr Console before provisioning an encrypted Data Centre. <br />Cannot be provided with `customVirtualNetworkId`                                                                                                                                                                                               |

## Cluster create flow
To create a Zookeeper cluster instance you need to prepare the yaml manifest. Here is an example:

```yaml
# zookeeper.yaml
apiVersion: clusters.instaclustr.com/v1beta1
kind: Zookeeper
metadata:
  name: zookeeper-sample
spec:
  dataCentres:
    - clientToServerEncryption: true
      cloudProvider: "AWS_VPC"
      name: "DataCentre1"
      network: "10.0.0.0/16"
      nodeSize: "zookeeper-developer-t3.small-20"
      nodesNumber: 3
      region: "US_EAST_1"
  name: "Zookeeper"
  privateNetworkCluster: false
  slaTier: "NON_PRODUCTION"
  version: "3.6.3"
```

Next, you need to apply this manifest in your K8s cluster. This will create a custom resource instance inside (more info about an apply command you can find [here](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#apply)):
```console
kubectl apply -f zookeeper.yaml
```

Now you can get and describe the instance:
```console
kubectl get zookeepers.clusters.instaclustr.com zookeeper-sample
```
```console
kubectl describe zookeepers.clusters.instaclustr.com zookeeper-sample
```

After you have applied the entity, the Instaclustr operator will create it on your K8s cluster and send a request to the Instaclustr API. You can be sure, that the resource creation call was sent if the instance has an id field filled in the status section.

You can check access to the created cluster from your kubernetes cluster and run some simple command to check that it is working with a lot of tools.
All available tools you can find in the Instaclustr console -> Choose your cluster -> Connection Info -> Examples section.

When a cluster is provisioned, a new service will be created along with it that expose public IP addresses. You can use this service name (pattern **{k8s_cluster_name}-service**) instead of public IP addresses and ports to connect to and interact with your cluster.
To do this, the public IP address of your machine must be added to the Firewall Rules tab of your cluster in the Instaclustr console.
![Firewall Rules icon](../images/firewall_rules_screen.png "Firewall Rules icon")
Then, you can use the service name instead of public addresses and port.

### Cluster deletion
To delete cluster run:
```console
kubectl delete zookeepers.clusters.instaclustr.com zookeeper-sample
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
kubectl delete zookeepers.clusters.instaclustr.com zookeeper-sample
```

After that, deletion confirmation email will be sent to the email defined in the `confirmationEmail` field of `TwoFactorDelete`. When deletion is confirmed via email, Instaclustr support will delete the cluster and the related cluster resources inside K8s will be also removed.

