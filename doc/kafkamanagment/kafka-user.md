# Kafka User resource management

## Available spec fields

| Field                                          | Type                                                   | Description                                                                                       |
|------------------------------------------------|--------------------------------------------------------|---------------------------------------------------------------------------------------------------|
| kafkaUserSecretName                            | string <br /> **required** <br /> _mutable_            | Name of the secret in which Kafka User password and Kafka User username are stored.               |
| kafkaUserSecretNamespace                       | string <br /> **required** <br /> _mutable_            | Namespace of the secret in which Kafka User password and Kafka User username are stored.          |
| clusterId                                      | string `<uuid>` <br /> **required**                    | ID of the Kafka cluster.                                                                          |
| initialPermissions                             | string <br /> **required** <br /> _mutable_            | Permissions initially granted to Kafka user upon creation. Enum: `standard`, `read-only`, `none`. |
| options                                        | Object ([Options](#OptionsObject)) <br /> **required** | Initial options used when creating Kafka User.                                                    |

### OptionsObject
| Field                         | Type                                        | Description                                                            |
|-------------------------------|---------------------------------------------|------------------------------------------------------------------------|
| overrideExistingUser          | bool <br /> _mutable_                       | Overwrite user if already exists.                                      |
| saslScramMechanism            | string <br /> **required** <br /> _mutable_ | SASL/SCRAM mechanism for user. Enum: `SCRAM-SHA-256`, `SCRAM-SHA-512`. |

## Resource create flow
To create a Kafka User resource you need to prepare the secret and the yaml manifests. Here is an example of them:
```yaml
# kafka-user-secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: secret-test
data:
  Username: "U2FuY2hv"
  Password: "ZnJlaGcxX3BpbmRmNTRfaW9rMDk="
---
# kafka-user.yaml
apiVersion: kafkamanagement.instaclustr.com/v1beta1
kind: KafkaUser
metadata:
  name: kafkauser-sample
spec:
  kafkaUserSecretName: "secret-test"
  kafkaUserSecretNamespace: "default"
  clusterId: "cc762648-6bc1-470c-8856-20e02f5ebbc2"
  initialPermissions: "standard"
  options:
    overrideExistingUser: true
    saslScramMechanism: "SCRAM-SHA-512"
```

Next, you need to apply this manifest in your K8s cluster. This will create a secret with your user credentials and custom resource instance inside (more info about an apply command you can find [here](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#apply)):

```console
kubectl apply -f kafka-user.yaml
```

Now you can get and describe the instance:

```console
kubectl get kafkausers.kafkamanagement.instaclustr.com kafkauser-sample
```
```console
kubectl describe kafkausers.kafkamanagement.instaclustr.com kafkauser-sample
```

After you have applied the entity, the Instaclustr operator will create it on your K8s cluster and send a request to the Instaclustr API. You can be sure, that the resource creation call was sent if the instance has an id field filled in the status section.
You can change the Kafka User password and username through secret.

## Resource update flow

To update a Kafka User you can apply an updated resource manifest or edit the custom resource instance in the K8s cluster:
* Apply an updated resource manifest:
    ```console
    kubectl apply -f kafka-user.yaml
    ```
* Edit the custom resource instance:
    ```console
    kubectl edit kafkausers.kafkamanagement.instaclustr.com kafkauser-sample
    ```
You can only update fields that are **mutable**. These fields are marked in the “Available spec fields” table.

## Resource delete flow

To delete the Kafka User run:
```console
kubectl delete kafkausers.kafkamanagement.instaclustr.com kafkauser-sample
```

It can take some time to delete the resource.