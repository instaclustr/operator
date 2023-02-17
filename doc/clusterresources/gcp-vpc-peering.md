# GCP VPC Peering resource management

## Available spec fields

| Field                                             | Type                               | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
|---------------------------------------------------|------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| peerProjectId                                             | string <br /> **required**         |    The project ID of the owner of the accepter VPC.                                                                                                                                                                                                                            |
| peerSubnets                                           | Array of strings <br /> **required** <br />  | The subnets for the peering VPC. |
| peerVpcNetworkName                             | string <br /> **required**         | The name of the VPC Network you wish to peer to.                                                                                                                                                                                                                                                                                                                                                                                                                                            |                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| cdcId                             | string <br /> **required**         | ID of the Cluster Data Centre.                                                                                                                                                                                                                                                                                                                                                                                                                                   |                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |

## Resource create flow
To create a GCP VPC Peering resource you need to prepare the yaml manifest. Here is an example:
```yaml
# gcpvpcpeering.yaml
apiVersion: clusterresources.instaclustr.com/v1alpha1
kind: GCPVPCPeering
metadata:
  name: gcpvpcpeering-sample
spec:
  cdcId: f3eab841-6952-430d-ba90-1bfc3f15da10
  peerProjectId: example-project123
  peerSubnets:
    - 10.1.0.0/16
  peerVpcNetworkName: network-aabb1122
```

Next, you need to apply this manifest in your K8s cluster. This will create a custom resource instance inside (more info about an apply command you can find [here](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#apply)):

```console
kubectl apply -f gcpvpcpeering.yaml
```

Now you can get and describe the instance:

```console
kubectl get gcpvpcpeerings.clusterresources.instaclustr.com gcpvpcpeering-sample
```
```console
kubectl describe gcpvpcpeerings.clusterresources.instaclustr.com gcpvpcpeering-sample
```

After you have applied the entity, the Instaclustr operator will create it on your K8s cluster and send a request to the Instaclustr API. You can be sure, that the resource creation call was sent if the instance has an id field filled in the status section.

## Resource delete flow

To delete the GCP VPC Peering run:
```console
kubectl delete gcpvpcpeerings.clusterresources.instaclustr.com gcpvpcpeering-sample
```

It can take some time to delete the resource.
