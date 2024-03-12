## Inheritance feature for clusters

If you already have created cluster on Instaclustr Console then you may use
this feature to manage the resource via Operator.

To use the feature you should apply the following yaml manifest:

```yaml
apiVersion: clusters.instaclustr.com/v1beta1
kind: KafkaConnect
metadata:
  name: kafkaconnect-inherited
spec:
  inheritsFrom: "265b75a5-f20f-4c3c-8ec2-0d03a8a8ef21"
```

Once you apply the following manifest it will get all cluster details
and update k8s resource spec and status.

Also, there are some specific cases for if you use inheritance feature for
Cadence cluster with AWSArchival enabled. It does the same work as with other
cluster resources, but it also creates a secret which stores your AWS Access Key ID
and AWS Access key which were used to create the cluster.