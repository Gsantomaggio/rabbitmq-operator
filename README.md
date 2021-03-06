# RabbitMQ Kubernetes Operator

[![Build Status](https://travis-ci.org/Gsantomaggio/rabbitmq-operator.svg?branch=master)](https://travis-ci.org/Gsantomaggio/rabbitmq-operator) [![Coverage Status](https://coveralls.io/repos/github/Gsantomaggio/rabbitmq-operator/badge.png?branch=master)](https://coveralls.io/github/Gsantomaggio/rabbitmq-operator?branch=master)

This project is not mainted anymore. I joined to the Kubernetes RabbitMQ team, so consider to use the official RabbitMQ cluster Operator:
https://github.com/rabbitmq/cluster-operator
=

The easier way to deploy a RabbitMQ Cluster on Kubernetes.

Current Features:
 * Automatic RabbitMQ service creation 
 * Configure the right policies to  run the Cluster
 * Easy `storage-class` configuration for persistent the data

Work in progress features:
* Handle the node remove from the cluster, avoiding to lose the the messages

### Parameters

|  Name | Description | Values | Default 
| ------------ | ------------ | ------------ | ------------ |
|   `replicas`| Number of the nodes  |  Integer (1 to x)  | 1
|  `serviceDefinition` | Service creating   | `Internal` = Auto create the service `External` = it  does not create the service | `Internal`
|`persistentVolume.storageClass` | k8s [Storage Class ](https://kubernetes.io/docs/concepts/storage/storage-classes/ "Storage Class ")definition,  if `""` won't use the `persistentVolume` configuration | `string` | `""` |
|`configMap`| Config Map name |`string` |`rabbitmq-config` 

### Install the RabbitMQ Operator 

* Install the operator from Docker-Hub:
```
kubectl apply -f https://github.com/Gsantomaggio/rabbitmq-operator/releases/download/v0.7-alpha/rabbitmq-operator-lastest.yaml
```


### Deploy RabbitMQ with the Operator

Inside the directory `config/samples` you can find the deploy examples.
The examples are built with [`kustomize`](https://github.com/kubernetes-sigs/kustomize).

Use the command: `kubectl apply -k`  to deploy it.

#### Localhost developing

For developing purpose you can use `config/samples/overlays/developing`, so:

```bash
kubectl apply -k config/samples/overlays/developing
```
By setting `serviceDefinition = External` It creates a custom Service with `nodePort` configuration, so it can be used in local configuration without load-balancers


#### Standard deploy

For the standard purpose you can use `config/samples/overlays/testing`, so:

```bash
kubectl apply -k config/samples/overlays/testing
```

### Install the RabbitMQ Operator with [Helm3](https://helm.sh/)

Add the repo:
```
 helm repo add alpha https://github.com/Gsantomaggio/rabbitmq-operator/releases/download/v0.7-alpha/
```

Update the repo list:
```
helm repo update
```

Search the package (optional):
```
helm search repo rabbitmq
NAME                                    CHART VERSION   APP VERSION     DESCRIPTION
alpha/rabbitmq-operator-helm            0.2.7           0.7             Helm for the RabbitMQ Operator
```

Install the package:
```
helm install  rabbitmq alpha/rabbitmq-operator-helm
```


## Localhost developing using Kind

_[Kind](https://github.com/kubernetes-sigs/kind) is a tool for running local Kubernetes clusters using Docker container "nodes"._

Create the Kind cluster:
```
kind create cluster --config utils/kind/kind-cluster.yaml
```

The `kind-cluster.yaml` configuration creates a localhost binding:
```yaml
 extraPortMappings:
  - containerPort: 31672
    hostPort: 15672
  - containerPort: 30672
    hostPort: 5672
```

Then you can use the [localhost developing](https://github.com/Gsantomaggio/rabbitmq-operator/blob/master/README.md#localhost-developing), the `service.yaml` exposes the `AMQP` and `HTTP` ports

```yaml
spec:
  type: NodePort
  ports:
   - name: http
     protocol: TCP
     port: 15672
     targetPort: 15672
     nodePort: 31672
   - name: amqp
     protocol: TCP
     port: 5672
     targetPort: 5672
     nodePort: 30672
```

So you can easly use it in `http://localhost:15672` and `amqp://localhost`


See the [Check the Installation](#check-the-installation) section to test it

## Build for source
### Requirements:
 - [kubebuilder]( https://book.kubebuilder.io/quick-start.html#installation)
 - [Golang](https://golang.org/)

```
git clone https://github.com/Gsantomaggio/rabbitmq-operator.git
cd rabbitmq-operator
make
```



## Check the Installation

Describe:
```
kubectl describe rabbitmq --all-namespaces
```

Running Pods:
```
kubectl get pods --all-namespaces | grep rabbitmq
NAMESPACE                  NAME                                                   READY   STATUS    RESTARTS   AGE
ns-developing              rabbitmq-op-developing-0                               1/1     Running   0          22m
ns-developing              rabbitmq-op-developing-1                               1/1     Running   0          20m
rabbitmq-operator-system   rabbitmq-operator-controller-manager-6b695f98d-jfk7j   2/2     Running   0          23m
```


## Project status

**The project is still experimental, not ready for production yet.**
