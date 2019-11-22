module github.com/gsantomaggio/rabbitmq-operator

go 1.12

require (
	github.com/go-logr/logr v0.1.0
	github.com/microsoft/azure-databricks-operator v0.0.0-20191114060425-d3dbafff0a49 // indirect
	github.com/onsi/ginkgo v1.10.3
	github.com/onsi/gomega v1.7.0
	k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apiextensions-apiserver v0.0.0-20190409022649-727a075fdec8
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	sigs.k8s.io/controller-runtime v0.2.2
)
