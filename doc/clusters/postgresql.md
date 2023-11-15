# PostgreSQL cluster management

## Available spec fields

| Field                 | Type                                                                        | Description                                                                                                                                                                                                                                                                                                                                  |
|-----------------------|-----------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| name                  | string <br /> **required**                                                  | Cluster name. Should have length from 3 to 32 symbols.                                                                                                                                                                                                                                                                                       |
| version               | string <br /> **required**                                                  | PostgreSQL instance version. <br />**Available versions**: `13.11.0`, `13.12.0`, `14.9.0`, `13.10.0`, `14.7.0`, `14.8.0`, `16.0.0`, `15.4.0`, `15.2.0`, `15.3.0`                                                                                                                                                                             |
| privateNetworkCluster | bool <br /> **required**                                                    | Creates the cluster with private network only, see [Private Network Clusters](https://www.instaclustr.com/support/documentation/useful-information/private-network-clusters/).                                                                                                                                                               |
| slaTier               | string  <br /> **required**                                                 | SLA Tier of the cluster. Non-production clusters may receive lower priority support and reduced SLAs. Production tier is not available when using Developer class nodes. See [SLA Tier](https://www.instaclustr.com/support/documentation/useful-information/sla-tier/) for more information. <br/>**Enum**: `PRODUCTION`, `NON_PRODUCTION`. |
| twoFactorDelete       | Array of objects ([TwoFactorDelete](#TwoFactorDeleteObject))<br />_mutable_ | Contacts that will be contacted when cluster request is sent.                                                                                                                                                                                                                                                                                |
| dataCentres           | Array of objects ([PgDataCentre](#PgDataCentreObject)) <br />_mutable_      | List of data centre settings.                                                                                                                                                                                                                                                                                                                |
| clusterConfigurations | map[string]string<br />_mutable_                                            | PostgreSQL cluster configurations. Cluster nodes will need to be manually reloaded to apply configuration changes. <br />**Format**:<br />clusterConfigurations:<br />- key: value                                                                                                                                                           |
| description           | string<br />_mutable_                                                       | A description of the cluster.                                                                                                                                                                                                                                                                                                                |
| synchronousModeStrict | bool  <br /> **required**                                                   | Create the PostgreSQL cluster with the selected replication mode, see [PostgreSQL replication mode](https://www.instaclustr.com/support/documentation/postgresql/options/replication-mode/).                                                                                                                                                 |
| pgRestoreFrom         | Object ([PgRestoreFrom](#PgRestoreFromObject))                              | Triggers a restore cluster operation.                                                                                                                                                                                                                                                                                                        |
| resizeSettings        | Array of objects ([ResizeSettings](#ResizeSettingsObject)) <br /> _mutable_ | Settings to determine how resize requests will be performed for the cluster.                                                                                                                                                                                                                                                                 |
| onPremisesSpec        | Object ([OnPremisesSpec](#OnPremisesSpecObject))                            | Specifies settings to provision on-premises cluster inside K8s cluster.                                                                                                                                                                                                                                                                      |

### ResizeSettingsObject
| Field                  | Type    | Description                                                                                                           |
|------------------------|---------|-----------------------------------------------------------------------------------------------------------------------|
| notifySupportContacts  | boolean | Setting this property to true will notify the Instaclustr Account's designated support contacts on resize completion. |
| concurrency            | integer | Number of concurrent nodes to resize during a resize operation.                                                       |

### TwoFactorDeleteObject

| Field                   | Type                       | Description                                                                            |
|-------------------------|----------------------------|----------------------------------------------------------------------------------------|
| confirmationPhoneNumber | string                     | The phone number which will be contacted when the cluster is requested to be deleted.  |
| confirmationEmail       | string <br /> **required** | The email address which will be contacted when the cluster is requested to be deleted. |

### PgDataCentreObject

| Field                      | Type                                                                               | Description                                                                                                                                                                                                                                                                                                                                                          |
|----------------------------|------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| name                       | string                                                                             | A logical name for the data centre within a cluster. These names must be unique in the cluster.                                                                                                                                                                                                                                                                      |
| region                     | string <br /> **required**                                                         | Region of the Data Centre. See the description for node size for a compatible Data Centre for a given node size.                                                                                                                                                                                                                                                     |
| cloudProvider              | string    <br /> **required**                                                      | Name of the cloud provider service in which the Data Centre will be provisioned. <br />**Enum**: `AWS_VPC` `GCP` `AZURE` `AZURE_AZ`                                                                                                                                                                                                                                  |
| cloudProviderSettings      | Array of objects ([CloudProviderSettings](#CloudProviderSettingsObject))           | Cloud provider specific settings for the Data Centre                                                                                                                                                                                                                                                                                                                 |
| network                    | string     <br /> **required**                                                     | The private network address block for the Data Centre specified using CIDR address notation. The network must have a prefix length between /12 and /22 and must be part of a private address space.                                                                                                                                                                  |
| nodeSize                   | string     <br /> **required**<br />_mutable_                                      | Size of the nodes provisioned in the Data Centre. Available node sizes, see [Instaclustr API docs NodeSize](https://instaclustr.redoc.ly/Current/tag/Postgresql-Cluster-V2?_ga=2.261654733.1169092465.1661786388-1184637702.1659440305#paths/~1cluster-management~1v2~1resources~1applications~1postgresql~1clusters~1v2~1/post!path=dataCentres/nodeSize&t=request) |
| nodesNumber                | int32  <br /> **required**                                                         | Total number of nodes in the Data Centre. <br/>Available values: [1…5]                                                                                                                                                                                                                                                                                               |
| tags                       | map[string]string                                                                  | List of tags to apply to the Data Centre. Tags are metadata labels which allow you to identify, categorise and filter clusters. This can be useful for grouping together clusters into applications, environments, or any category that you require.<br/>**Format**:<br/>tags:<br/>- key: value                                                                      |
| clientEncryption           | bool                                                                               | Enable client to cluster Encryption.                                                                                                                                                                                                                                                                                                                                 |
| interDataCentreReplication | Array of objects ([InterDataCentreReplication](#InterDataCentreReplicationObject)) |                                                                                                                                                                                                                                                                                                                                                                      |
| intraDataCentreReplication | Array of objects ([IntraDataCentreReplication](#IntraDataCentreReplicationObject)) |                                                                                                                                                                                                                                                                                                                                                                      |
| pgBouncer                  | Array of objects ([PgBouncerDetails](#PgBouncerDetailsObject))                     | Version of Pg Bouncer to run on the cluster. Required to enable Pg Bouncer. <br/>**Available versions**: `1.17.0`                                                                                                                                                                                                                                                    |
| poolMode                   | string                                                                             | Creates PgBouncer with the selected mode, see PgBouncer pool modes. Only available with `pgBouncerVersion` filled. <br/>**Enum**: `TRANSACTION` `SESSION` `STATEMENT`                                                                                                                                                                                                |

### CloudProviderSettingsObject

| Field                  | Type   | Description                                                                                                                                                                                                                                                                                                                                                                                                                                 |
|------------------------|--------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| customVirtualNetworkId | string | **AWS**: VPC ID into which the Data Centre will be provisioned. The Data Centre's network allocation must match the IPv4 CIDR block of the specified VPC.<br />**GCP**: Network name or a relative Network or Subnetwork URI e.g. projects/my-project/regions/us-central1/subnetworks/my-subnet. The Data Centre's network allocation must match the IPv4 CIDR block of the specified subnet. <br />Cannot be provided with `resourceGroup` |
| resourceGroup          | string | The name of the Azure Resource Group into which the Data Centre will be provisioned. <br />Cannot be provided with `customVirtualNetworkId` and `diskEncryptionKey`                                                                                                                                                                                                                                                                         |
| diskEncryptionKey      | string | ID of a KMS encryption key to encrypt data on nodes. KMS encryption key must be set in Cluster Resources through the Instaclustr Console before provisioning an encrypted Data Centre. <br />Cannot be provided with `customVirtualNetworkId`                                                                                                                                                                                               |

### #PgBouncerDetailsObject

| Field              | Type                        | Description                                                                                                                                                                                            |
|--------------------|-----------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| pgBouncerVersion   | string <br /> **required**  | Version of Pg Bouncer to run on the cluster. Available versions: `1.19.0`, `1.20.0`, `1.18.0`                                                                                                          |
| poolMode           | string <br /> **required**  | Creates PgBouncer with the selected mode, see [PgBouncer pool modes] (https://www.instaclustr.com/support/documentation/postgresql-add-ons/using-pgbouncer/) Enum: "TRANSACTION" "SESSION" "STATEMENT" |

### InterDataCentreReplicationObject

| Field               | Type                     | Description                                                                                        |
|---------------------|--------------------------|----------------------------------------------------------------------------------------------------|
| isPrimaryDataCentre | bool <br /> **required** | Is this Data centre considered to be the primary (only required if multiple data centres defined). |

### IntraDataCentreReplicationObject

| Field           | Type                     | Description                                                                                                                                                                                                                              |
|-----------------|--------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| replicationMode | bool <br /> **required** | Create the PostgreSQL cluster with the selected replication mode, see [PostgreSQL replication mode](https://www.instaclustr.com/support/documentation/postgresql/options/replication-mode/). <br/>**Enum**: `ASYNCHRONOUS` `SYNCHRONOUS` |

### PgRestoreFromObject

| Field               | Type                                                           | Description                                                                                                                                                                                         |
|---------------------|----------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| clusterId           | string `<uuid>`<br /> **required**                             | Original cluster ID. Backup from that cluster will be used for restore                                                                                                                              |
| clusterNameOverride | string                                                         | The display name of the restored cluster. <br/>By default, the restored cluster will be created with its current name appended with “restored” and the date & time it was requested to be restored. |
| cdcInfos            | Array of objects ([PgRestoreCDCInfo](#PgRestoreCDCInfoObject)) | An optional list of Cluster Data Centres for which custom VPC settings will be used. <br/>This property must be populated for all Cluster Data Centres or none at all.                              |
| pointInTime         | int64   `>= 1420070400000`                                     | Timestamp in milliseconds since epoch. All backed up data will be restored for this point in time.                                                                                                  |

### PgRestoreCDCInfoObject

| Field            | Type            | Description                                                                                                                             |
|------------------|-----------------|-----------------------------------------------------------------------------------------------------------------------------------------|
| cdcId            | string `<uuid>` | Cluster Data Centre ID                                                                                                                  |
| restoreToSameVpc | bool            | Determines if the restored cluster will be allocated to the existing VPC. <br/>Either restoreToSameVpc or customVpcId must be provided. |
| customVpcId      | string          | Custom VPC ID to which the restored cluster will be allocated. <br/>Either restoreToSameVpc or customVpcId must be provided.            |
| customVpcNetwork | string          | CIDR block in which the cluster will be allocated for a custom VPC.                                                                     |

### OnPremisesSpecObject

| Field              | Type                                                       | Description                                                                                                                                                                        |
|--------------------|------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| storageClassName   | string <br /> **required**                                 | Name of the storage class that will be used to provision disks for on-premises nodes.                                                                                              |
| osDiskSize         | string <br /> **required**                                 | Disk size on which OS will be installed.                                                                                                                                           |
| dataDiskSize       | string <br /> **required**                                 | Disk size on which on-premises cluster data will be stored.                                                                                                                        |
| sshGatewayCPU      | int64                                                      | Amount of CPU that will be dedicated to provision SSH Gateway node (only for private clusters).                                                                                    |
| sshGatewayMemory   | string                                                     | Amount of RAM that will be dedicated to provision SSH Gateway node (only for private clusters).                                                                                    |
| nodeCPU            | int64 <br /> **required**                                  | Amount of CPU that will be dedicated to provision on-premises worker node.                                                                                                         |
| nodeMemory         | string <br /> **required**                                 | Amount of RAM that will be dedicated to provision on-premises worker node                                                                                                          |
| osImageURL         | string <br /> **required**                                 | OS image URL that will be use to dynamically provision disks with preinstalled OS (more info can be found [here](https://kubevirt.io/2020/KubeVirt-VM-Image-Usage-Patterns.html)). |
| cloudInitScriptRef | Object ([Reference](#ReferenceObject)) <br /> **required** | Reference to the secret with cloud-init script (must be located in the same namespace with the cluster). Example can be found [here](#CloudInitScript).                            |

### ReferenceObject

| Field     | Type   | Description                                       |
|-----------|--------|---------------------------------------------------|
| name      | string | Name of the cloud-init secret.                    |
| namespace | string | Namespace in which the cloud-init secret located. |

## Cluster creation example

Notice:
- If you choose to omit the cluster name in the specification, the operator will use it from the metadata name instead. In this case make sure that metadata name matches the name pattern, you can check the pattern on instaclustr API specification

To create a cluster you need to prepare a cluster manifest. Here is an example:
```yaml
# postgresql.yaml
apiVersion: clusters.instaclustr.com/v1beta1
kind: PostgreSQL
metadata:
  name: postgresql-sample
spec:
  name: "postgresql-sample"
  version: "14.5.0"
  dataCentres:
    - region: "US_WEST_2"
      network: "10.1.0.0/16"
      cloudProvider: "AWS_VPC"
      nodeSize: "PGS-DEV-t4g.small-5"
      nodesNumber: 2
      clientEncryption: false
      name: "exampleDC1"
      intraDataCentreReplication:
        - replicationMode: "SYNCHRONOUS"
      interDataCentreReplication:
        - isPrimaryDataCentre: true
  description: "example"
  slaTier: "NON_PRODUCTION"
  privateNetworkCluster: false
  synchronousModeStrict: false
```


Or if you want to create an on-premises cluster:
```yaml
# postgresql.yaml
apiVersion: clusters.instaclustr.com/v1beta1
kind: PostgreSQL
metadata:
  name: postgresql-sample
spec:
  name: "postgresql-sample"
  version: "15.4.0"
  onPremisesSpec:
    storageClassName: managed-csi-premium
    osDiskSize: 20Gi
    dataDiskSize: 200Gi
    sshGatewayCPU: 2
    sshGatewayMemory: 4096Mi
    nodeCPU: 2
    nodeMemory: 8192Mi
    osImageURL: "https://s3.amazonaws.com/debian-bucket/debian-11-generic-amd64-20230601-1398.raw"
    cloudInitScriptRef:
      namespace: default
      name: cloud-init-secret
  dataCentres:
    - region: "CLIENT_DC" # Don't change if you want to run on-premises
      network: "10.1.0.0/16"
      cloudProvider: "ONPREMISES" # Don't change if you want to run on-premises
      nodeSize: "PGS-DEV-OP.4.8-200"
      nodesNumber: 2
      clientEncryption: false
      name: "onPremisesDC"
      intraDataCentreReplication:
        - replicationMode: "SYNCHRONOUS"
      interDataCentreReplication:
        - isPrimaryDataCentre: true
  slaTier: "NON_PRODUCTION"
  privateNetworkCluster: false
  synchronousModeStrict: false
```

Also, don't forget to create cloud-init script firstly.

### CloudInitScript

cloud-init.sh:
```shell
#!/bin/bash

export NEW_PASS="qwerty12345"
export SSH_PUB_KEY=""
export BOOTSTRAP_SSH_KEY="ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAEAQDgeaO3zkY5v1dww3fFONPzUnEgIqJ4kUK0Usu8iFdp+TWIulw9dDeQHa+PdWXP97l5Vv1mG9ipqShEIu7/2bp13KxSblWX4iV1MYZbtimhY3UDOsPn1G3E1Ipis6y+/tosDAm8LoWaGEMcLuE5UjP6gs6K57rCEjkVGjg7vjhIypAMC0J2N2+CxK9o/Y1+LZAec+FL5cmSJoajoo9y/VYJjz/a52jJ93wRafD2uu6ObAl5gkN/+gqY4IJFSMx20sPsIRXdbiBNDqiap56ibHHPKTeZhRbdXvZfjYkHtutXnSM2xn7BjnV8CguISxS3rXlzlzRVYwSUjnKUf5SKBbeyZbCokx271vCeUd5EXfHphvW6FIOna2AI5lpCSYkw5Kho3HaPi2NjXJ9i2dCr1zpcZpCiettDuaEjxR0Cl4Jd6PrAiAEZ0Ns0u2ysVhudshVzQrq6qdd7W9/MLjbDIHdTToNjFLZA6cbE0MQf18LXwJAl+G/QrXgcVaiopUmld+68JL89Xym55LzkMhI0NiDtWufawd/NiZ6jm13Z3+atvnOimdyuqBYeFWgbtcxjs0yN9v7h7PfPh6TbiQYFx37DCfQiIucqx1GWmMijmd7HMY6Nv3UvnoTUTSn4yz1NxhEpC61N+iAInZDpeJEtULSzlEMWlbzL4t5cF+Rm1dFdq3WpZt1mi8F4DgrsgZEuLGAw22RNW3++EWYFUNnJXaYyctPrMpWQktr4DB5nQGIHF92WR8uncxTNUXfWuT29O9e+bFYh1etmq8rsCoLcxN0zFHWvcECK55aE+47lfNAR+HEjuwYW10mGU/pFmO0F9FFmcQRSw4D4tnVUgl3XjKe3bBiTa4lUrzrKkLZ6n9/buW2e7c3jbjmXdPh2R+2Msr/vfuWs9glxQf+CYEbBW6Ye4pekIyI77SaB/bVhaHtXutKxm+QWdNle8aeqiA8Ji1Ml+s75vIg+n5v6viCnl5aV33xHRFpGQJzj2ktsXl9P9d5kgal9eXJYTywC2SnVbZVLb6FGN4kPZTVwX1f+u7v7JCm4YWlbQZtwwiXKjs99AVtQnBWqQvUH5sFUkVXlHA1Y9W6wlup0r+F6URL+7Yw+d0dHByfevrJg3pvmpLb3sEpjIAZodW3dIUReE7Ku3s/q/O9foFnfRBnCcZ2QsnxI5pqNrbrundD1ApOnNXEvICvPXHBBQ44cW0hzAO+WxY5VxyG8y/kXnb48G9efkIQFkNaITJrU9SiOk6bFP4QANdS/pmaSLjJIsHixa+7vmYjRy1SVoQ/39vDUnyCbqKtO56QMH32hQLRO3Vk7NVG6o4dYjFkiaMSaqVlHKMkJQHVzlK2PW9/fjVXfkAHmmhoD debian"
echo "debian:$NEW_PASS" | chpasswd
echo "root:$NEW_PASS" | sudo chpasswd root
sudo echo "$SSH_PUB_KEY" > /home/debian/.ssh/authorized_keys
sudo echo "$BOOTSTRAP_SSH_KEY" >> /home/debian/.ssh/authorized_keys
sudo chown -R debian: /home/debian/.ssh
sudo cp /usr/share/doc/apt/examples/sources.list /etc/apt/sources.list
data_device=$(lsblk -dfn -o NAME,SERIAL | awk '$2 == "DATADISK" {print $1}')
sudo mkfs -t ext4 /dev/"${data_device}"
```

Create the base64 encoded string with the script using `cat cloud-init.sh | base64 -w0` command and create a secret using this yaml manifest:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: cloud-init-secret
data:
  userdata: <base64 encoded string>
```

Use `kubectl apply -f cloud-init-secret.yaml` command to create the secret.

Next you need to apply this manifest. This will create PostgreSQL custom resource instance:
```console
kubectl apply -f postgresql.yaml
```

Now you can get and describe the instance:
```console
kubectl get postgresqls.clusters.instaclustr.com postgresql-sample
```
```console
kubectl describe postgresqls.clusters.instaclustr.com postgresql-sample
```

Cluster was created on Instaclustr premise if the instance has an id in the status section.

You can check access to the created cluster from your kubernetes cluster and run some simple command to check that it is working with a lot of tools.
All available tools you can find in the Instaclustr console -> Choose your cluster -> Connection Info -> Examples section.

When a cluster is provisioned, a new service will be created along with it that expose public IP addresses. You can use this service name (pattern **{k8s_cluster_name}-service**) instead of public IP addresses and ports to connect to and interact with your cluster.
To do this, the public IP address of your machine must be added to the Firewall Rules tab of your cluster in the Instaclustr console.
![Firewall Rules icon](../images/firewall_rules_screen.png "Firewall Rules icon")
Then, you can use the service name instead of public addresses and port.

## Cluster update example

To update a cluster you can apply an updated cluster manifest or edit the custom resource instance in kubernetes cluster:
* Apply an updated cluster manifest:
    ```console
    kubectl apply -f postgresql.yaml
    ```
* Edit the custom resource instance:
    ```console
    kubectl edit postgresqls.clusters.instaclustr.com postgresql-sample
    ```
You can only update fields that are **mutable**

## Cluster delete flow

### Cluster deletion
To delete cluster run:
```console
kubectl delete postgresqls.clusters.instaclustr.com postgresql-sample
```

### Cluster deletion with twoFactorDelete option enabled
To delete a cluster with the `twoFactorDelete` option enabled you need to do the simple [cluster deletion flow ](#cluster-deletion).
After that, a deletion email will be sent to the email defined in the `confirmationEmail` field of `twoFactorDelete`.
When deletion is confirmed via email, Instaclustr support will delete the cluster.
It can take some time to delete the resource.

If you cancel cluster deletion and want to put cluster on delete again, remove `triggered` from `clusterDeletionAnnotation` annotation like this:

```yaml
Annotations:
  "instaclustr.com/clusterDeletion": ""
```

## Update Default User Password

Default user password is stored in a kubernetes secret. It is created before cluster creation. It has name template `default-user-password-<resource-name>`<br />
To update default user password you need to edit `defaultUserPassword` field in the secret:
```console
kubectl edit default-user-password-postgresql-sample
```

## Cluster restore example

To restore a PostgreSQL cluster instance from an existing one you need to prepare the yaml manifest. Here is an example:
```yaml
# postgresql-restore.yaml
apiVersion: clusters.instaclustr.com/v1beta1
kind: PostgreSQL
metadata:
  name: postgresql-sample-restored
  annotations:
    testAnnotation: test
spec:
  pgRestoreFrom:
    clusterId: 826e00dc-eda5-426a-9185-48fc35f13dfc
    clusterNameOverride: "pgRestoredTest"
    pointInTime: 1676371087358
    clusterNetwork: "10.13.0.0/16"
```

Next, you need to apply this manifest in your K8s cluster. This will create a custom resource instance inside:
```console
kubectl apply -f postgresql-restore.yaml
```

New cluster will be created from the backup of the "restored-from" cluster. Spec will be updated automatically.