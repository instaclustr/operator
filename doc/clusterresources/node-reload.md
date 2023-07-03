# Node Reload resource management (available only for PostgreSQL clusters)

## Available spec fields

| Field | Type                                                          | Description                             |
|-------|---------------------------------------------------------------|-----------------------------------------|
| nodes |  Array of objects ([Nodes](#NodesObject)) <br /> **required** | Array with IDs of Nodes to be reloaded. |

### NodesObject
| Field  | Type                         | Description                     |
|--------|------------------------------|---------------------------------|
| nodeID | string <br /> **required**   | ID of the Node to be reloaded.  |

## Resource create flow
To create a Node Reload resource you need to prepare the yaml manifest. Here is an example:
```yaml
# node-reload.yaml
apiVersion: clusterresources.instaclustr.com/v1beta1
kind: NodeReload
metadata:
  name: nodereload-sample
spec:
  nodes:
      - nodeID: "ce567c90-3ce4-49e7-9ac9-fc1f3f91e33e"
      - nodeID: "e7acd145-8329-4577-9eaf-cc8dce4c54eb"
      - nodeID: "2f94dd6d-3a65-4ad6-81bd-d3db8115ad63"
```

Next, you need to apply this manifest in your K8s cluster. This will create a custom resource instance inside (more info about an apply command you can find [here](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#apply)):

```console
kubectl apply -f node-reload.yaml
```

Now you can get and describe the instance:

```console
kubectl get nodereloads.clusterresources.instaclustr.com nodereload-sample
```
```console
kubectl describe nodereloads.clusterresources.instaclustr.com nodereload-sample
```

After you have applied the entity, the Instaclustr operator will create it on your K8s cluster and send some requests (depending on node quantity) to the Instaclustr API. Node Reload resource will be deleted automatically when all the nodes are reloaded.