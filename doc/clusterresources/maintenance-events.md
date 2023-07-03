# AWS VPC Peering resource management

## Available spec fields

| Field                           | Type                                                                                   | Description                                                          |
|---------------------------------|----------------------------------------------------------------------------------------|----------------------------------------------------------------------|
| clusterId                       | string <br /> **required**                                                             | Cluster Id for the cluster that this exclusion window relates to.    |
| exclusionWindows                | Array of objects ([ExclusionWindows](#ExclusionWindowsObject))                         | Exclusion Window array. Allows to create one or more windows.        |
| maintenanceEventsReschedule     | Array of objects ([MaintenanceEventsReschedules](#MaintenanceEventsReschedulesObject)) | Maintenance Events array. Allows to reschedule maintenance events.   |

### ExclusionWindowsObject
| Field            | Type                                         | Description                                                 |
|------------------|----------------------------------------------|-------------------------------------------------------------|
| dayOfWeek        | string <br /> **required** <br /> _mutable_  | The day of the week that this exclusion window starts on.   |
| startHour        | int32 <br /> **required**  <br /> _mutable_  | The hour of the day that this exclusion window starts on.   |
| durationInHours  | int32 <br /> **required**  <br /> _mutable_  | The duration (in hours) of this exclusion window.           |

### MaintenanceEventsReschedulesObject
| Field                        | Type                                                 | Description                                                                      |
|------------------------------|------------------------------------------------------|----------------------------------------------------------------------------------|
| scheduledStartTime           | string <br /> **required** <br /> _mutable_          | The time this maintenance event should be rescheduled to start at (in UTC time). |
| scheduleId                   | string `<uuid>` <br /> **required** <br /> _mutable_ | **Example:** `2c7067d6-c70a-11ea-87d0-0242ac130003`.                             |

## Resource create flow
To create an Exclusion window resource you need to prepare the yaml manifest. Here is an example:
```yaml
# maintenance-event.yaml
apiVersion: clusterresources.instaclustr.com/v1beta1
kind: MaintenanceEvents
metadata:
  name: maintenanceevents-sample
spec:
  clusterId: "2217f85d-f4a8-4da4-b7d7-1b39c5fc625e"
  exclusionWindows:
    - dayOfWeek: "MONDAY"
      startHour: 1
      durationInHours: 1
```

Next, you need to apply this manifest in your K8s cluster. This will create a custom resource instance inside (more info about an apply command you can find [here](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#apply)):

```console
kubectl apply -f awsvpcpeering.yaml
```

Now you can get and describe the instance:

```console
kubectl get maintenanceevents.clusterresources.instaclustr.com maintenanceevents-sample
```
```console
kubectl describe maintenanceevents.clusterresources.instaclustr.com maintenanceevents-sample
```

After you have applied the entity, the Instaclustr operator will create it on your K8s cluster and send a request to the Instaclustr API. You can be sure, that the resource creation call was sent if the instance has an id field filled in the status section.

If you want to reschedule the maintenance event, you need to prepare or update the existent yaml manifest. Here is an example:
```yaml
# maintenance-event.yaml
apiVersion: clusterresources.instaclustr.com/v1beta1
kind: MaintenanceEvents
metadata:
  name: maintenanceevents-sample
spec:
  maintenanceEventsReschedule:
    - scheduledStartTime: "2023-03-08T00:00:00Z"
      scheduleId: "0585598f-e672-4bd2-806a-673ba521665c"
```

You can find all info about maintenance events of your cluster by checking the cluster status.

## Resource update flow

To update a Maintenance Event resource you can apply an updated resource manifest or edit the custom resource instance in the K8s cluster:
* Apply an updated resource manifest:
```console
    kubectl apply -f maintenance-event.yaml
```
* Edit the custom resource instance:
```console
    kubectl edit maintenanceevents.clusterresources.instaclustr.com maintenanceevents-sample
```
You can only update fields that are **mutable**. These fields are marked in the “Available spec fields” table.

## Resource delete flow

To delete the AWS VPC Peering run:
```console
kubectl delete maintenanceevents.clusterresources.instaclustr.com maintenanceevents-sample
```

It can take some time to delete the resource.