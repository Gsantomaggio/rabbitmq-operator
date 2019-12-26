# RabbitMQ Kubernetes Operator

[![Build Status](https://travis-ci.org/Gsantomaggio/rabbitmq-operator.svg?branch=master)](https://travis-ci.org/Gsantomaggio/rabbitmq-operator) [![Coverage Status](https://coveralls.io/repos/github/Gsantomaggio/rabbitmq-operator/badge.svg?branch=master)](https://coveralls.io/github/Gsantomaggio/rabbitmq-operator?branch=master)


Kubernetes Operator to handle the RabbitMQ deploy.

### Install the RabbitMQ Operator 

* Install the operator from Docker-Hub:
```
kubectl apply -f https://github.com/Gsantomaggio/rabbitmq-operator/releases/download/v0.4-alpha/rabbitmq-operator_latest.yaml
```

See the [Check the Installation](#check-the-installation) section to test it


### Deploy RabbitMQ with the Operator

Inside the directory `config/samples` you can find the deploy examples.
The examples are built with [`kustomize`](https://github.com/kubernetes-sigs/kustomize), but you don't have to install anything, the command `kubectl apply -k`  already uses kustomize.

#### Localhost developing

For developing purpose you can use `config/samples/overlays/developing`, so:

```bash
kubectl apply -k config/samples/overlays/developing
```
It creates a custom Service with `nodePort` configuration, so it can be used in local configuration without load-balancers


#### Standard deploy

For the standard purpose you can use `config/samples/overlays/testing`, so:

```bash
kubectl apply -k config/samples/overlays/testing
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
kubectl describe rabbitmq
```

Running Pods:
```
$ kubectl get pods
NAME            READY   STATUS    RESTARTS   AGE
rabbitmq-op-0   1/1     Running   0          4m51s
rabbitmq-op-1   1/1     Running   0          3m45s
rabbitmq-op-2   1/1     Running   0          2m32s
```


## Project status

**The project is still experimental, not ready for production yet.**
