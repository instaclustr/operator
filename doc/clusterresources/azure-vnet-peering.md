# Azure Vnet Peering resource management

## Available spec fields

| Field                                                 | Type                                        | Description                                                                                                |
|-------------------------------------------------------|---------------------------------------------|------------------------------------------------------------------------------------------------------------|
| peerVirtualNetworkName                                | string <br /> **required**                  | The name of the VPC Network you wish to peer to.                                                           |
| peerSubnets                                           | Array of strings <br /> **required** <br /> | The subnets for the peering VPC.                                                                           |
| peerAdObjectId                                        | string                                      | ID of the Active Directory Object to give peering permissions to, required for cross subscription peering. |                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| peerResourceGroup                                     | string <br /> **required**                  | Resource Group Name of the Virtual Network.                                                                |
| peerSubscriptionId                                    | string <br /> **required** <br />           | Subscription ID of the Virtual Network.                                                                    |
| cdcId                                                 | string <br /> **required**                  | ID of the Cluster Data Centre.                                                                             |                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |

## Resource create flow
To create a Azure Vnet Peering resource you need to prepare the yaml manifest. Here is an example:
```yaml
# azurevnetpeering.yaml
apiVersion: clusterresources.instaclustr.com/v1beta1
kind: AzureVNetPeering
metadata:
  name: azurevnetpeering-sample
spec:
  cdcId: f3eab841-6952-430d-ba90-1bfc3f15da10
  peerAdObjectId: 00000000-0000-0000-0000-000000000000
  peerResourceGroup: example-resource-group-123
  peerSubnets:
    - 10.1.0.0/16
    - 10.2.0.0/16
  peerSubscriptionId: 00000000-0000-0000-0000-000000000000
  peerVirtualNetworkName: network-aabb-1122
```

Next, you need to apply this manifest in your K8s cluster. This will create a custom resource instance inside (more info about an apply command you can find [here](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#apply)):

```console
kubectl apply -f azurevnetpeering.yaml
```

Now you can get and describe the instance:

```console
kubectl get azurevnetpeerings.clusterresources.instaclustr.com azurevnetpeering-sample
```
```console
kubectl describe azurevnetpeerings.clusterresources.instaclustr.com azurevnetpeering-sample
```

After you have applied the entity, the Instaclustr operator will create it on your K8s cluster and send a request to the Instaclustr API. You can be sure, that the resource creation call was sent if the instance has an id field filled in the status section.

## Resource delete flow

To delete the Azure Vnet Peering run:
```console
kubectl delete azurevnetpeerings.clusterresources.instaclustr.com azurevnetpeering-sample
```

It can take some time to delete the resource.
