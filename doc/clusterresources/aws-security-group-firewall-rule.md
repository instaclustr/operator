# AWS Security Group Firewall Rule resource management

## Available spec fields

| Field                                        | Type                               | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
|----------------------------------------------|------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| clusterId                                    | string <br /> **required**         | ID of the cluster for the cluster network firewall rule.                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| type                                         | string <br /> **required** <br />  | The type of firewall rule. <br/> <br/>Enum: `APACHE_ZOOKEEPER`, `CADENCE`, `CADENCE_GRPC`, `CADENCE_WEB`, `CASSANDRA`, `CASSANDRA_CQL`, `ELASTICSEARCH`, `KAFKA`, `KAFKA_CONNECT`, `KAFKA_ENCRYPTION`, `KAFKA_MTLS`, `KAFKA_NO_ENCRYPTION`, `KAFKA_REST_PROXY`, `KAFKA_SCHEMA_REGISTRY`, `KARAPACE_REST_PROXY`, `KARAPACE_SCHEMA_REGISTRY`, `OPENSEARCH`, `OPENSEARCH_DASHBOARDS`, `PGBOUNCER`, `POSTGRESQL`, `REDIS`, `SEARCH_DASHBOARDS`, `SECURE_APACHE_ZOOKEEPER`, `SPARK`, `SPARK_JOBSERVER`, `SHOTOVER_PROXY`.  |
| securityGroupId                              | string <br /> **required**         | The security group ID of the AWS security group firewall rule.                                                                                                                                                                                                                                                                                                                                                                                                                                                        |                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |

## Resource create flow
To create an AWS Security Group Firewall Rule resource you need to prepare the yaml manifest. Here is an example:
```yaml
# awssecuritygroupfirewallrule.yaml
apiVersion: clusterresources.instaclustr.com/v1alpha1
kind: AWSSecurityGroupFirewallRule
metadata:
  name: awssecuritygroupfirewallrule-sample
spec:
  securityGroupId: sg-0c19175a3adbbae3a
  clusterId: f33bda01-ff1c-4bb9-9e19-db449c7badf6
  type: REDIS
```

Next, you need to apply this manifest in your K8s cluster. This will create a custom resource instance inside (more info about an apply command you can find [here](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#apply)):

```console
kubectl apply -f awssecuritygroupfirewallrule.yaml
```

Now you can get and describe the instance:

```console
kubectl get awssecuritygroupfirewallrules.clusterresources.instaclustr.com awssecuritygroupfirewallrule-sample
```
```console
kubectl describe awssecuritygroupfirewallrules.clusterresources.instaclustr.com awssecuritygroupfirewallrule-sample
```

After you have applied the entity, the Instaclustr operator will create it on your K8s cluster and send a request to the Instaclustr API. You can be sure, that the resource creation call was sent if the instance has an id field filled in the status section.

## Resource delete flow

To delete the AWS Security Group Firewall Rule run:
```console
kubectl delete awssecuritygroupfirewallrules.clusterresources.instaclustr.com awssecuritygroupfirewallrule-sample
```

It can take some time to delete the resource.
