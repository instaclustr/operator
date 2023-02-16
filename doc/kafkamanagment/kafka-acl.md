# Kafka ACL management

## Available spec fields

| Field                                                 | Type                                        | Description                                                                                                                                                           |
|-------------------------------------------------------|---------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| userQuery                                         | string <br /> **required** <br /> _mutable_ |  This is the principal without the User: prefix. |
| clusterId                                 | string <br /> **required** <br /> _mutable_ | UUID of the Kafka cluster.            |
| acls                                              | Array of strings <br /> **required**  <br /> _mutable_   | List of ACLs for the given principal.|

### ACLsObject

| Field                                                    | Type                                                        | Description                                                                                                                                                                                                  |
|----------------------------------------------------------|-------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| principal                                                    | string <br /> **required** <br /> _mutable_                 | Specifies the users(s) for which this ACL applies and can include the wildcard *. Valid values must start with "User:" including the wildcard.                                                               |
| permissionType                                          | string <br /> **required** <br /> _mutable_                 | Specifies whether to allow or deny the operation. <br> <br> **Enum**: `ALLOW`, `DENY`.                                                                                                                       |
| host                                           | string <br /> **required** <br /> _mutable_                 | The IP address to which this ACL applies. It takes any string including the wildcard * for all IP addresses.                                                                                                 |
| patternType                                          | string <br /> **required** <br /> _mutable_                 | Indicates the resource-pattern-type  <br> <br> **Enum**: `LITERAL`, `PREFIXED`.                                                                                                                              |
| resourceName                                           | string <br /> **required** <br /> _mutable_  | Any string that fits the resource name, e.g. topic name if the resource type is TOPIC                                                                                                                        |
| operation                                          | string <br /> **required** <br /> _mutable_ | The operation that will be allowed or denied. <br> <br> **Enum**: `ALL`, `READ`, `WRITE`, `CREATE`, `DELETE`, `ALTER`, `DESCRIBE`, `CLUSTER_ACTION`, `DESCRIBE_CONFIGS`, `ALTER_CONFIGS`, `IDEMPOTENT_WRITE`. |
| resourceType                                           | string <br /> **required** <br /> _mutable_  | Specifies the type of resource.  <br> <br> **Enum**: `CLUSTER`, `TOPIC`, `GROUP`, `DELEGATION_TOKEN`, `TRANSACTIONAL_ID`.                                                                                    |


## Resource create flow
To create a Kafka ACL resource you need to prepare the yaml manifest. Here is an example:
```yaml
# kafkaacl.yaml
apiVersion: kafkamanagement.instaclustr.com/v1alpha1
kind: KafkaACL
metadata:
  name: kafkaacl-sample
spec:
  acls:
    - host: "*"
      operation: DESCRIBE
      patternType: LITERAL
      permissionType: ALLOW
      principal: User:test
      resourceName: kafka-cluster
      resourceType: CLUSTER
  clusterId: c1af59c6-ba0e-4cc2-a0f3-65cee17a5f37
  id: c1af59c6-ba0e-4cc2-a0f3-65cee17a5f37_test
  userQuery: test
```

Next, you need to apply this manifest in your K8s cluster. This will create a custom resource instance inside (more info about an apply command you can find [here](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#apply)):

```console
kubectl apply kafkaacl.yaml
```

Now you can get and describe the instance:

```console
kubectl get kafkaacls.kafkamanagement.instaclustr.com kafkaacl-sample
```
```console
kubectl describe kafkaacls.kafkamanagement.instaclustr.com kafkaacl-sample
```

After you have applied the entity, the Instaclustr operator will create it on your K8s cluster and send a request to the Instaclustr API. You can be sure, that the resource creation call was sent if the instance has an id field filled in the status section.

## Resource update flow

To update a Kafka ACL you can apply an updated resource manifest or edit the custom resource instance in the K8s cluster:
* Apply an updated cluster manifest:
    ```console
    kubectl apply kafkaacl.yaml
    ```
* Edit the custom resource instance:
    ```console
    kubectl edit kafkaacls.kafkamanagement.instaclustr.com kafkaacl-sample
    ```
You can only update fields that are **mutable**. These fields are marked in the “Available spec fields” table.

## Resource delete flow

To delete the Kafka ACL run:
```console
kubectl delete kafkaacls.kafkamanagement.instaclustr.com kafkaacl-sample
```

It can take some time to delete the resource.
