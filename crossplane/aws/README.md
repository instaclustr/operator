# Connect Crossplane to AWS

Crossplane installs into an existing Kubernetes cluster.
If you don’t have a Kubernetes cluster create one locally with Kind.

### Install the Crossplane Helm chart
Helm enables Crossplane to install all its Kubernetes components through a Helm Chart.
Enable the Crossplane Helm Chart repository:

```
helm repo add \
crossplane-stable https://charts.crossplane.io/stable
helm repo update
```

Install the Crossplane components:

```
helm install crossplane \
crossplane-stable/crossplane \
--namespace crossplane-system \
--create-namespace
```

Verify Crossplane installed with kubectl get pods ```kubectl get pods -n crossplane-system```.

# Install the AWS provider

Install the provider into the Kubernetes cluster with a Kubernetes configuration file.
```
kubectl apply -f crossplane-provider.yaml
```

Verify the provider installed with ```kubectl get providers```.

It may take up to five minutes for the provider to list HEALTHY as True.

```
NAME                   INSTALLED   HEALTHY   PACKAGE                                        AGE
upbound-provider-aws   True        True      xpkg.upbound.io/upbound/provider-aws:v0.34.0   12m
```

A provider installs their own Kubernetes Custom Resource Definitions (CRDs). These CRDs allow you to create AWS resources directly inside Kubernetes.

You can view the new CRDs with ```kubectl get crds```

## Generate an AWS key-pair file
For basic user authentication, use an AWS Access keys key-pair file.

In [aws-credentials.txt](aws-credentials.txt) replace ```<aws_access_key>``` and  ```<aws_secret_key>``` with your data.

Use kubectl create secret to generate the secret object named aws-secret in the crossplane-system namespace.
Use the --from-file= argument to set the value to the contents of the aws-credentials.txt file.
```
kubectl create secret \
generic aws-secret \
-n crossplane-system \
--from-file=creds=./aws-credentials.txt
```

View the secret with ```kubectl describe secret aws-secret -n crossplane-system```

## Create a ProviderConfig
A ProviderConfig customizes the settings of the AWS Provider.
Apply the ProviderConfig with the command 
```
kubectl apply -f provider-config.yaml
```

## Cassandra spin up with Crossplane

### Create Cassandra On-premise cluster
In the Instaclustr Console create a Cassandra On-Premise cluster.

### Create AWS resources

VPC 
```
kubectl apply -f vpc.yaml
```

Internet Gateway
```
kubectl apply -f internet-gateway.yaml
```

Subnets
```
kubectl apply -f subnet.yaml
```

Security Group 
```
kubectl apply -f security-group.yaml
```

Roles for Security Group
```
kubectl apply -f security-group-role.yaml
```

Paste your ssh into ```SSH_PUB_KEY``` in [startup-script.sh](startup-script.sh) and generate binary of this file:
```
cat startup-script.sh | base64 -w0
```
Paste generated binary into each ```userDataBase64``` field in [instance.yaml](instance.yaml) manifests.

Create instances 
```
kubectl apply -f instance.yaml
```

In the AWS Console -> VPC -> Route tables -> Edit routes -> add your [internet-gateway](internet-gateway.yaml) (Destination - 0.0.0.0/0).

Get Private and Public IPs from instances and put into [ignition commands](ignition-command.txt).

Run [ignition-commands](ignition-command.txt) in icadmnin. Take care to ensure that the right IPs are assigned to the right nodes and importantly are in the right rack
For each node, get the ignition scripts from Zendesk, add them to each instance accordingly, and make them executable.

Run the ignition scripts as a root and after that we should have a functioning Cassandra cluster. 

