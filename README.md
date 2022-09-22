# operator
The Instaclustr Operator allows users to deploy, scale and manage any one of
data layer technologies deployed on the Instaclustr platform from Kubernetes.

## Description
The Instaclustr Operator provides all the same functionality as the Instaclustr API,
Instaclustr Console and Instaclustr terraform provider. 
- Removes the need to write custom code integration directly with the Instaclustr API
- Use of existing Instaclustr authentication methodologies
- Declare Instaclustr infrastructure up front
- Scale, modify and change Instaclustr infrastructure with
  existing Kubernetes management tools / workflows
- Simplify Kubernetes native development workflow patterns
  (CI/CD integrations, production deployments etc).

## Visual Software Architecture

Context diagram                                                         |  Container diagram                                       |  Component diagram
:---------------------------------------------------------------------:|:--------------------------------------------------------:|:-------------------------:
![Context diagram](doc/diagrams/context_diagram.png "Context diagram") |  ![Container diagram](doc/diagrams/container_diagram.png) |  ![Component diagram](doc/diagrams/component_diagram.png)

CRDs Relations|
:---------------------------------------------------------------------:|
![CRD Relations](doc/diagrams/crd_relations.png "CRD Relations")|


## Getting Started
You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Running on the cluster
1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Build and push your image to the location specified by `IMG`:
	
```sh
make docker-build docker-push IMG=<some-registry>/operator:tag
```
	
3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/operator:tag
```

### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller
UnDeploy the controller to the cluster:

```sh
make undeploy
```

## Contributing

We welcome all contributors. Please see our public [contributing guidelines](CONTRIBUTING.md).


### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/) 
which provides a reconcile function responsible for synchronizing resources untile the desired state is reached on the cluster 

### Test It Out
1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

