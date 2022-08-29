# VerneMQ Kubernetes Operator

Project status: **alpha**

The main goal of the VerneMQ Kubernetes Operator is to simplify the deployment of a VerneMQ cluster on Kubernetes. While the operator isn't the silver bullet for every VerneMQ deployment we hope to cover most cases, where scalability and high availability are required. 

See: [Getting Started Guide][getting_started]


## Development

Note: the following sections are mostly copy pasted from https://github.com/operator-framework/operator-sdk/edit/master/doc/user-guide.md

### Prerequisites

- [git][git_tool]
- [go][go_tool] version v1.10+.
- [docker][docker_tool] version 17.03+.
- [kubectl][kubectl_tool] version v1.11.0+.
- Access to a kubernetes v.1.11.0+ cluster (use Minikube locally).

### Quick Start

First, checkout and install the operator-sdk CLI, see https://sdk.operatorframework.io/docs/installation/ for more information.

### Build and run the operator

Before running the operator, the CRD must be registered with the Kubernetes apiserver:

```sh
$ kubectl create -f deploy/crds/vernemq_v1alpha1_vernemq_crd.yaml
```

Once this is done, there are two ways to run the operator:

- As a Deployment inside a Kubernetes cluster
- As Go program outside a cluster

#### 1. Run as a Deployment inside the cluster

Build the vmq-operator image and push it to a registry [not required for minicube testing]:
```
$ operator-sdk build vernemq/vmq-operator:latest
$ sed -i 's|REPLACE_IMAGE|vernemq/vmq-operator:latest|g' deploy/operator.yaml
$ docker push vernemq/vmq-operator:latest
```

The Deployment manifest is generated at `deploy/operator.yaml`. Be sure to update the deployment image as shown above since the default is just a placeholder.

Setup RBAC and deploy the vmq-operator:

```sh
$ kubectl create -f deploy/service_account.yaml
$ kubectl create -f deploy/role.yaml
$ kubectl create -f deploy/role_binding.yaml
$ kubectl create -f deploy/operator.yaml
```

Verify that the vmq-operator is up and running:

```sh
$ kubectl get deployment
NAME                     DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
vmq-operator       1         1         1            1           1m
```

#### 2. Run locally outside the cluster

This method is preferred during development cycle to deploy and test faster.

Set the name of the operator in an environment variable:

```sh
export OPERATOR_NAME=vmq-operator
```

Run the operator locally with the default kubernetes config file present at `$HOME/.kube/config`:

```sh
$ operator-sdk up local --namespace=default
2018/09/30 23:10:11 Go Version: go1.10.2
2018/09/30 23:10:11 Go OS/Arch: darwin/amd64
2018/09/30 23:10:11 operator-sdk Version: 0.0.6+git
2018/09/30 23:10:12 Registering Components.
2018/09/30 23:10:12 Starting the Cmd.
```

You can use a specific kubeconfig via the flag `--kubeconfig=<path/to/kubeconfig>`.

### Create a VerneMQ CR

Create the example `VerneMQ` CR that was generated at `deploy/crds/vernemq_v1alpha1_vernemq_cr.yaml`:

```sh
$ cat deploy/crds/vernemq_v1alpha1_vernemq_cr.yaml
apiVersion: "vernrmq.com/v1alpha1"
kind: "VerneMQ"
metadata:
  name: "example-vernemq"
spec:
  size: 3

$ kubectl apply -f deploy/crds/vernemq_v1alpha1_vernemq_cr.yaml
```
Check the pods and CR status to confirm the status is updated with the vernemq pod names:

```sh
$ kubectl get pods
NAME                            READY   STATUS             RESTARTS   AGE
example-vernemq-0               1/1     Running            0          6m31s
example-vernemq-1               1/1     Running            0          6m19s
example-vernemq-2               1/1     Running            0          6m18s
vmq-operator-7fbfd5bfbc-9cbjc   0/1     ImagePullBackOff   0          11m
```

## License

The Operator SDK and the VerneMQ Operator are under Apache 2.0 license. See the [LICENSE][license_file] file for details.

[getting_started]: ./docs/getting-started.md
[license_file]:./LICENSE
[git_tool]:https://git-scm.com/downloads
[go_tool]:https://golang.org/dl/
[docker_tool]:https://docs.docker.com/install/
[kubectl_tool]:https://kubernetes.io/docs/tasks/tools/install-kubectl/
