# Kafka Topic resource management

## Available spec fields

| Field                         | Type                               | Description                    |
|-------------------------------|------------------------------------|--------------------------------|
| partitions                    | int32 <br /> **required**          | Topic partition count.         |
| replicationFactor             | int32 <br /> **required**          | Replication factor for Topic.  |
| topic                         | string <br /> **required**         | Kafka Topic name.              |
| clusterId                     | string <br /> **required**         | ID of the Kafka cluster.       |
| configs                       | map[string]string <br /> _mutable_ | List of Kafka topic configs.   |

## Resource create flow
To create a Kafka Topic resource you need to prepare the yaml manifest. Here is an example:
```yaml
# kafka-topic.yaml
apiVersion: kafkamanagement.instaclustr.com/v1alpha1
kind: Topic
metadata:
  name: topic-sample
spec:
  partitions: 3
  replicationFactor: 3
  topic: "topic-test"
  clusterId: "9fcd9820-96dc-49fe-9b58-e9a75fcdd6fc"
```

Next, you need to apply this manifest in your K8s cluster. This will create a custom resource instance inside (more info about an apply command you can find [here](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#apply)):

```console
kubectl apply -f kafka-topic.yaml
```

Now you can get and describe the instance:

```console
kubectl get topics.kafkamanagement.instaclustr.com topic-sample
```
```console
kubectl describe topics.kafkamanagement.instaclustr.com topic-sample
```

After you have applied the entity, the Instaclustr operator will create it on your K8s cluster and send a request to the Instaclustr API. You can be sure, that the resource creation call was sent if the instance has an id field filled in the status section.

## Resource update flow

To update a Kafka Topic you can apply an updated resource manifest or edit the custom resource instance in the K8s cluster:
* Apply an updated cluster manifest:
    ```console
    kubectl apply -f kafka-topic.yaml
    ```
* Edit the custom resource instance:
    ```console
    kubectl edit topics.kafkamanagement.instaclustr.com topic-sample
    ```
You can only update fields that are **mutable**. These fields are marked in the “Available spec fields” table.
You can add [some configs](https://kafka.apache.org/23/documentation.html#topicconfigs) for your Kafka Topic resource. Here is an example:
```yaml
# kafka-topic.yaml
apiVersion: kafkamanagement.instaclustr.com/v1alpha1
kind: Topic
metadata:
  name: topic-sample
spec:
  partitions: 3
  replicationFactor: 3
  topic: "operator-topic-test"
  clusterId: "9fcd9820-96dc-49fe-9b58-e9a75fcdd6fc"
  configs:
    min.insync.replicas: "2"
    retention.ms: "603800000"
```

## Resource delete flow

To delete the Kafka Connect Mirror run:
```console
kubectl delete topics.kafkamanagement.instaclustr.com topic-sample
```

It can take some time to delete the resource.
