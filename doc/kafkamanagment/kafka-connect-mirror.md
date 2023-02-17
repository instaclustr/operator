# Kafka Connect Mirror resource management

## Available spec fields

| Field                                                 | Type                                                                         | Description                                                                                                                                                                                |
|-------------------------------------------------------|------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| targetLatency                                         | int32 <br /> **required** <br /> _mutable_                                   | The latency in milliseconds above which this mirror will be considered out of sync. It can not be less than 1000ms. The suggested initial latency is 30000ms for connectors to be created. |
| kafkaConnectClusterId                                 | string <br /> **required**                                                   | ID of the kafka connect cluster.                                                                                                                                                           |
| maxTasks                                              | int32 <br /> **required**                                                    | Maximum number of tasks for Kafka Connect to use. Should be greater than 0.                                                                                                                |
| sourceCluster                                         | Array of objects ([SourceCluster](#SourceClusterObject)) <br /> **required** | Details to connect to the source kafka cluster.                                                                                                                                            |
| renameMirroredTopics                                  | bool <br /> **required**                                                     | Whether to rename topics as they are mirrored, by prefixing the sourceCluster.alias to the topic name.                                                                                     |
| topicsRegex                                           | string <br /> **required**                                                   | Regex to select which topics to mirror.                                                                                                                                                    |

### SourceClusterObject

| Field                                                    | Type                                                         | Description                                                                                                              |
|----------------------------------------------------------|--------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------|
| alias                                                    | string <br /> **required**                                   | Alias to use for the source kafka cluster. This will be used to rename topics if renameMirroredTopics is turned on.      |
| externalCluster                                          | Array of objects ([ExternalCluster](#ExternalClusterObject)) | Details to connect to a Non-Instaclustr managed cluster. Cannot be provided if targeting an Instaclustr managed cluster. |
| managedCluster                                           | Array of objects ([ManagedCluster](#ManagedClusterObject))   | Details to connect to a Instaclustr managed cluster. Cannot be provided if targeting a non-Instaclustr managed cluster.  |

### ExternalClusterObject

| Field                                                    | Type                        | Description                                                                   |
|----------------------------------------------------------|-----------------------------|-------------------------------------------------------------------------------|
| sourceConnectionProperties                                                    | string <br /> **required**  | Kafka connection properties string used to connect to external kafka cluster. |

### ManagedClusterObject

| Field                                                    | Type                       | Description                                                                   |
|----------------------------------------------------------|----------------------------|-------------------------------------------------------------------------------|
| usePrivateIps                                                    | bool <br /> **required**   | Whether or not to connect to source cluster's private IP addresses. |
| sourceKafkaClusterId                                                    | string <br /> **required** | Source kafka cluster id. |

## Resource create flow
To create a Kafka Connect Mirror resource you need to prepare the yaml manifest. Here is an example:
```yaml
# kafka-connect-mirror.yaml
apiVersion: kafkamanagement.instaclustr.com/v1alpha1
kind: Mirror
metadata:
  name: mirror-sample
spec:
  kafkaConnectClusterId: "897da890-a212-4771-ac13-b08d85e32ad6"
  maxTasks: 3
  renameMirroredTopics: true
  sourceCluster:
    - alias: "source-cluster"
      managedCluster:
        - "sourceKafkaClusterId": "f37f3acb-a65b-4985-944e-fba7a7099b98"
          "usePrivateIps": true
  "targetLatency": 5000
  "topicsRegex": ".*"
```

Next, you need to apply this manifest in your K8s cluster. This will create a custom resource instance inside (more info about an apply command you can find [here](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#apply)):

```console
kubectl apply -f kafka-connect-mirror.yaml
```

Now you can get and describe the instance:

```console
kubectl get mirrors.kafkamanagement.instaclustr.com mirror-sample
```
```console
kubectl describe mirrors.kafkamanagement.instaclustr.com mirror-sample
```

After you have applied the entity, the Instaclustr operator will create it on your K8s cluster and send a request to the Instaclustr API. You can be sure, that the resource creation call was sent if the instance has an id field filled in the status section.

## Resource update flow

To update a Kafka Connect Mirror you can apply an updated resource manifest or edit the custom resource instance in the K8s cluster:
* Apply an updated resource manifest:
    ```console
    kubectl apply -f kafka-connect-mirror.yaml
    ```
* Edit the custom resource instance:
    ```console
    kubectl edit mirrors.kafkamanagement.instaclustr.com mirror-sample
    ```
You can only update fields that are **mutable**. These fields are marked in the “Available spec fields” table.

## Resource delete flow

To delete the Kafka Connect Mirror run:
```console
kubectl delete mirrors.kafkamanagement.instaclustr.com mirror-sample
```

It can take some time to delete the resource.
