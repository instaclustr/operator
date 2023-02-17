# Cluster Network Firewall Rule resource management

## Available spec fields

| Field                                             | Type                               | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
|---------------------------------------------------|------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| clusterId                                             | string <br /> **required**         | ID of the cluster for the cluster network firewall rule.                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| type                                           | string <br /> **required** <br />  | The type of firewall rule. <br/> <br/>Enum: `APACHE_ZOOKEEPER`, `CADENCE`, `CADENCE_GRPC`, `CADENCE_WEB`, `CASSANDRA`, `CASSANDRA_CQL`, `ELASTICSEARCH`, `KAFKA`, `KAFKA_CONNECT`, `KAFKA_ENCRYPTION`, `KAFKA_MTLS`, `KAFKA_NO_ENCRYPTION`, `KAFKA_REST_PROXY`, `KAFKA_SCHEMA_REGISTRY`, `KARAPACE_REST_PROXY`, `KARAPACE_SCHEMA_REGISTRY`, `OPENSEARCH`, `OPENSEARCH_DASHBOARDS`, `PGBOUNCER`, `POSTGRESQL`, `REDIS`, `SEARCH_DASHBOARDS`, `SECURE_APACHE_ZOOKEEPER`, `SPARK`, `SPARK_JOBSERVER`, `SHOTOVER_PROXY`. |
| network                             | string <br /> **required**         | The network of the cluster network firewall rule.                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |

## Resource create flow
To create a Cluster Network Firewall Rule resource you need to prepare the yaml manifest. Here is an example:
```yaml
# clusternetworkfirewallrule.yaml
apiVersion: clusterresources.instaclustr.com/v1alpha1
kind: ClusterNetworkFirewallRule
metadata:
  name: clusternetworkfirewallrule-sample
spec:
  network: 62.212.64.19/32
  clusterId: bb7166e4-cffa-48e7-8f0a-5ee32873b750
  type: CASSANDRA_CQL
```

Next, you need to apply this manifest in your K8s cluster. This will create a custom resource instance inside (more info about an apply command you can find [here](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#apply)):

```console
kubectl apply -f clusternetworkfirewallrule.yaml
```

Now you can get and describe the instance:

```console
kubectl get clusternetworkfirewallrules.clusterresources.instaclustr.com clusternetworkfirewallrule-sample
```
```console
kubectl describe clusternetworkfirewallrules.clusterresources.instaclustr.com clusternetworkfirewallrule-sample
```

After you have applied the entity, the Instaclustr operator will create it on your K8s cluster and send a request to the Instaclustr API. You can be sure, that the resource creation call was sent if the instance has an id field filled in the status section.

## Resource delete flow

To delete the Cluster Network Firewall Rule run:
```console
kubectl delete clusternetworkfirewallrules.clusterresources.instaclustr.com clusternetworkfirewallrule-sample
```

It can take some time to delete the resource.
