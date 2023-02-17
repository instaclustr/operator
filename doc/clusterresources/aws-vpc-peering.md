# AWS VPC Peering resource management

## Available spec fields

| Field                                             | Type                                                    | Description                                                 |
|---------------------------------------------------|---------------------------------------------------------|-------------------------------------------------------------|
| peerAwsAccountId                                              | string <br /> **required**                              | The AWS account ID of the owner of the accepter VPC.        |
| peerSubnets                                           | Array of strings <br /> **required** <br /> _mutable_ | The subnets for the peering VPC.                            |
| peerVpcId                             | string <br /> **required**                              | ID of the VPC with which the peering connection is created. |
| peerRegion                                           | string <br /> **required**                              | Region code for the accepter VPC.                           |
| cdcId                                       | string <uuid> <br /> **required**                                  | ID of the Cluster Data Centre.       |

## Resource create flow
To create an AWS VPC Peering resource you need to prepare the yaml manifest. Here is an example:
```yaml
# awsvpcpeering.yaml
apiVersion: clusterresources.instaclustr.com/v1alpha1
kind: AWSVPCPeering
metadata:
 name: awsvpcpeering-sample
spec:
 peerAwsAccountId: "789012612380"
 peerSubnets:
   - "172.31.0.0/16"
   - "192.168.0.0/16"
 peerVpcId: "vpc-87256df2"
 peerRegion: "US_EAST_1"
 cdcId: "b8c01eca-a01e-4469-8d2b-705192b35eef"
```

Next, you need to apply this manifest in your K8s cluster. This will create a custom resource instance inside (more info about an apply command you can find [here](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#apply)):

```console
kubectl apply -f awsvpcpeering.yaml
```

Now you can get and describe the instance:

```console
kubectl get awsvpcpeerings.clusterresources.instaclustr.com awsvpcpeering-sample
```
```console
kubectl describe awsvpcpeerings.clusterresources.instaclustr.com awsvpcpeering-sample
```

After you have applied the entity, the Instaclustr operator will create it on your K8s cluster and send a request to the Instaclustr API. You can be sure, that the resource creation call was sent if the instance has an id field filled in the status section.

## Resource update flow

To update an AWS VPC Peering you can apply an updated resource manifest or edit the custom resource instance in the K8s cluster:
* Apply an updated resource manifest:
```console
    kubectl apply -f awsvpcpeering.yaml
```
* Edit the custom resource instance:
```console
    kubectl edit awsvpcpeerings.clusterresources.instaclustr.com awsvpcpeering-sample
```
You can only update fields that are **mutable**. These fields are marked in the “Available spec fields” table.

## Resource delete flow

To delete the AWS VPC Peering run:
```console
kubectl delete awsvpcpeerings.clusterresources.instaclustr.com awsvpcpeering-sample
```

It can take some time to delete the resource.
