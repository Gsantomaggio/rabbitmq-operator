# RabbitMQ Kubernetes Operator

[![Build Status](https://travis-ci.org/Gsantomaggio/rabbitmq-operator.svg?branch=master)](https://travis-ci.org/Gsantomaggio/rabbitmq-operator) [![Coverage Status](https://coveralls.io/repos/github/Gsantomaggio/rabbitmq-operator/badge.svg?branch=add_persistence)](https://coveralls.io/github/Gsantomaggio/rabbitmq-operator?branch=add_persistence)


Kubernetes Operator to handle the RabbitMQ deploy.

**The project is still experimental, not ready for production yet.**


### Install the RabbitMQ Operator 

* Install the operator from Docker-Hub:
```
kubectl apply -f https://github.com/Gsantomaggio/rabbitmq-operator/releases/download/v0.3-alpha/rabbitmq-operator_latest.yaml
```

* Install the confing map example file:

```
kubectl apply -f  https://github.com/Gsantomaggio/rabbitmq-operator/releases/download/v0.3-alpha/scaling_configmap.yaml
```

* Deploy RabbitMQ using the Operator:
```
kubectl apply -f  https://github.com/Gsantomaggio/rabbitmq-operator/releases/download/v0.3-alpha/scaling_v1alpha_rabbitmq.yaml
```

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



## Test it using Kind
### Requirements:

 - [kind](https://github.com/kubernetes-sigs/kind)
 - [docker](https://www.docker.com/)
 - [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

Create the Kind cluster:
```
kind create cluster
export KUBECONFIG="$(kind get kubeconfig-path --name="kind")" 
make && make install && make run
```

Deploy the YAML files
```
kubectl apply -f config/samples/
```
See the [Check the Installation](#check-the-installation) section to test it


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

You can check the cluster locally using the script: `utils/export_rabbitmq_ports` 
```
$ utils/export_rabbitmq_ports
Forwarding from 127.0.0.1:5672 -> 5672
Forwarding from [::1]:5672 -> 5672
Forwarding from 127.0.0.1:15672 -> 15672
Forwarding from [::1]:15672 -> 15672
Forwarding from 127.0.0.1:15692 -> 15692
Forwarding from [::1]:15692 -> 15692
Handling connection for 15672
```

Then http://localhost:15672 (guest guest)
