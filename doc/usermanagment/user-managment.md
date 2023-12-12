# User Management (Available on Redis, PostgreSQL, OpenSearch, Kafka, Cassandra)

## User creation flow

To create the user, fill redisuser.yaml. We need to create Secret first, and then create the user. You can do it in the same file. Here is an example:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: redis-user-test-1
data:
  username: bXlreXRhCg==
  password: VGVzdDEyMyEK
---
apiVersion: clusterresources.instaclustr.com/v1beta1
kind: RedisUser
metadata:
  name: redisuser-sample-1
spec:
  initialPermissions: "none"
  secretRef:
    name: "redis-user-test-1" #metadata name
    namespace: "default"
```

Next, you need to apply this manifest in your K8s cluster. This will create a custom resource instance inside:
```console
kubectl apply -f redisuser.yaml
```

Now you can get and describe the instance:
```console
kubectl get redisusers.clusterresources.instaclustr.com redisuser-sample-1
```
```console
kubectl describe redisusers.clusterresources.instaclustr.com redisuser-sample-1
```

## To add user references to the cluster add userRef to spec.

## Available spec fields for cluster

| Field            | Type                                                               | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
|------------------|--------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| userRefs         | Array of objects ([UserRefs](#UserRefsObject)) <br /> **required** | Object fields are described below as a bulleted list.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |

### UserRefsObject

| Field               | Type                                  | Description                            |
|---------------------|---------------------------------------|----------------------------------------|
| name                | string <br /> **required**            | User reference name                    |
| namespace           | string <br /> **required** <br />     | Namespace where User reference placed  |

## Here is an example of yaml file:
```yaml
apiVersion: clusters.instaclustr.com/v1beta1
kind: Redis
metadata:
  name: redis-sample
spec:
  userRefs:
      - name: redisuser-sample-1
        namespace: default
      - name: redisuser-sample-2
        namespace: default
      - name: redisuser-sample-3
        namespace: default
```
Next, you need to apply this manifest in your K8s cluster. This will create a custom resource instance inside:
```console
kubectl apply -f redis.yaml
```

Now you can get and describe the instance:
```console
kubectl get redis.clusters.instaclustr.com redis-sample
```
```console
kubectl describe redis.clusters.instaclustr.com redis-sample
```

After you have applied the entity, the Instaclustr operator will create it on your K8s cluster and send a request to the Instaclustr API. You can be sure, that the resource creation call was sent if the instance has an id field filled in the status section.

Edit the custom resource instance:
```console
kubectl edit redis.clusters.instaclustr.com redis-sample
```
You can only update fields that are **mutable**

## User deletion flow

### User deletion form cluster
To delete user from the cluster, remove userRef from "spec", then use following command:
```console
kubectl apply -f redis.yaml
```

## Kafka user certificates

### Creation flow
If you want to use MTLS for a Kafka user you should create a `UserCertificate` resource.
Here is an example of the manifest:
```yaml
apiVersion: kafkamanagement.instaclustr.com/v1beta1
kind: UserCertificate
metadata:
  name: user-cert-from-secret-sample
spec:
  clusterRef:
    name: kafka
    namespace: default
  userRef:
    name: kafka-user
    namespace: default
  secretRef:
    name: csr-1
    namespace: default
    key: kafka-user.csr
  validPeriod: 3
```
It will use a pre-generated certificate request from the provided secret to create the signed certificate.
Remember that the common name in the CSR must match the username of the Kafka User.
You can find the signed certificate in the secret created by the operator. The reference to the secret
is stored in the `UserCertificate` status in the `signedCertSecretRef` field.

Also, you may use the operator to generate a certificate request and create a signed cert.
Manifest example:
```yaml
apiVersion: kafkamanagement.instaclustr.com/v1beta1
kind: UserCertificate
metadata:
  name: user-cert-from-template-sample
spec:
  clusterRef:
    name: kafka
    namespace: default
  userRef:
    name: kafka-user
    namespace: default
  certificateRequestTemplate:
    country: Ireland
    organization: Instaclustr
    organizationalUnit: IC 
  validPeriod: 3
```
It will generate CSR, and create a secret in the same namespace as UserCertificate is created. The secret stores
rsa2048 privateKey and certificate request.

## Deletion flow
To delete `UserCertificate` use the following command:
```console
kubectl delete -f user-certificate.yaml
```
It also deletes secrets created by the operator secrets, because there are `ownerReferecenes` to user certificate resource.

