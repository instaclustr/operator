# AWS Encryption key resource management

## Available spec fields

| Field                                          | Type                              | Description                                                                                                                                                                                                                                                                                            |
|------------------------------------------------|-----------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| alias                                          | string <br /> **required**        | Encryption key alias for display purposes.                                                                                                                                                                                                                                                             |
| arn                                            | string <br /> **required** <br /> | AWS ARN for the encryption key.                                                                                                                                                                                                                                                                        |
| providerAccountName                            | string <br />                     | For customers running in their own account. Your provider account can be found on the Create Cluster page on the Instaclustr Console, or the "Provider Account" property on any existing cluster. For customers provisioning on Instaclustr's cloud provider accounts, this property may be omitted.   |                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |

## Resource create flow
To create an AWS Encryption key resource you need to prepare the yaml manifest. Here is an example:
```yaml
# awsencryptionkey.yaml
apiVersion: clusterresources.instaclustr.com/v1alpha1
kind: AWSEncryptionKey
metadata:
  name: awsencryptionkey-sample
spec:
  alias: test-1
  arn: arn:aws:kms:us-east-1:152668027680:key/460b469b-9f0a-4801-854a-e23626957d00
```

Next, you need to apply this manifest in your K8s cluster. This will create a custom resource instance inside (more info about an apply command you can find [here](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#apply)):

```console
kubectl apply -f awsencryptionkey.yaml
```

Now you can get and describe the instance:

```console
kubectl get awsencryptionkeys.clusterresources.instaclustr.com awsencryptionkey-sample
```
```console
kubectl describe awsencryptionkeys.clusterresources.instaclustr.com awsencryptionkey-sample
```

After you have applied the entity, the Instaclustr operator will create it on your K8s cluster and send a request to the Instaclustr API. You can be sure, that the resource creation call was sent if the instance has an id field filled in the status section.

## Resource delete flow

To delete the AWS Encryption key run:
```console
kubectl delete awsencryptionkeys.clusterresources.instaclustr.com awsencryptionkey-sample
```

It can take some time to delete the resource.
