# RabbitMQ Kubernetes Operator

[![Build Status](https://travis-ci.org/Gsantomaggio/rabbitmq-operator.svg?branch=master)](https://travis-ci.org/Gsantomaggio/rabbitmq-operator)

Kubernetes Operator to handle the RabbitMQ deploy.

**The project is still experimental, not ready for production yet.**


### Install from DockerHUB

Install the operator:
```
kubectl apply -f https://raw.githubusercontent.com/Gsantomaggio/rabbitmq-operator/master/deploy/rabbitmq-operator_latest.
```
 
Install the confing map example file:

```
kubectl apply -f  https://raw.githubusercontent.com/Gsantomaggio/rabbitmq-operator/master/deploy/scaling_configmap.yaml
```

Deploy RabbitMQ using the Operator:
```
kubectl apply -f  https://raw.githubusercontent.com/Gsantomaggio/rabbitmq-operator/master/deploy/scaling_v1alpha_rabbitmq.yam
```

See the `Check the Installation` section to test it

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

Deploy the YAML
```
kubectl apply -f config/samples/
```
See the `Check the Installation` section to test it

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

you can can check the cluster using the script: `utils/export_rabbitmq_ports` 
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