# RabbitMQ Kubernetes Operator

[![Build Status](https://travis-ci.org/Gsantomaggio/rabbitmq-operator.svg?branch=master)](https://travis-ci.org/Gsantomaggio/rabbitmq-operator)

Kubernetes Operator to handle the RabbitMQ deploy.

**The project is still experimental, not ready for production yet.**

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
kubectl apply -f config/samples/scaling_v1alpha_rabbitmq.yaml
```

Check it:
```
kubectl describe rabbitmq
```

Running Pods:
```
kubectl get pods
NAME            READY   STATUS    RESTARTS   AGE
rabbitmq-op-0   1/1     Running   0          2m44s
rabbitmq-op-1   1/1     Running   0          115s
```

Create the service (optional):
```
kubectl apply -f config/samples/scaling_rabbitmq_service.yaml
```


