/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"log"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	susev1beta1 "github.com/gsantomaggio/rabbitmq-operator/api/v1beta1"
)

// RabbitMQInstanceReconciler reconciles a RabbitMQInstance object
type RabbitMQInstanceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

var (
	ownerKey                      = ".metadata.controller"
	apiGVStr                      = susev1beta1.GroupVersion.String()
	terminationGracePeriodSeconds = int64(10)
)

// +kubebuilder:rbac:groups=suse.suse.rabbitmq-operator,resources=rabbitmqinstances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=suse.suse.rabbitmq-operator,resources=rabbitmqinstances/status,verbs=get;update;patch

func (r *RabbitMQInstanceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	instance := &susev1beta1.RabbitMQInstance{}
	err := r.Get(context.TODO(), req.NamespacedName, instance)
	//ctx := context.Background()
	log1 := r.Log.WithValues("rabbitmqinstance", req.NamespacedName)
	if err != nil {
		log.Printf("error %s", err)
		return ctrl.Result{}, err
	}
	log.Printf("Checking status of res: %d req: %s", instance.Spec.Replicas, req.NamespacedName.Name)
	statefulset := &appsv1.StatefulSet{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, statefulset)

	if err != nil && errors.IsNotFound(err) {
		// Define a new Deployment
		dep, err := r.statefulsetForRabbitMQ(instance)
		log1.Info("Creating a new statefulset.", "statefulset.Namespace", dep.Namespace, "statefulset.Name", dep.Name)
		err = r.Create(context.TODO(), dep)
		if err != nil {
			log1.Error(err, "Failed to create new statefulset.", "statefulset.Namespace", dep.Namespace, "statefulset.Name", dep.Name)
			return reconcile.Result{}, err
		}
		// Deployment created successfully - return and requeue
		// NOTE: that the requeue is made with the purpose to provide the deployment object for the next step to ensure the deployment size is the same as the spec.
		// Also, you could GET the deployment object again instead of requeue if you wish. See more over it here: https://godoc.org/sigs.k8s.io/controller-runtime/pkg/reconcile#Reconciler
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		log1.Error(err, "Failed to get Deployment.")
		return reconcile.Result{}, err
	}

	service := &corev1.Service{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, service)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Service object
		ser := r.serviceForRabbitMQ(instance)
		log1.Info("Creating a new Service.", "Service.Namespace", ser.Namespace, "Service.Name", ser.Name)
		err = r.Create(context.TODO(), ser)
		if err != nil {
			log1.Error(err, "Failed to create new Service.", "Service.Namespace", ser.Namespace, "Service.Name", ser.Name)
			return reconcile.Result{}, err
		}
	} else if err != nil {
		log1.Error(err, "Failed to get Service.")
		return reconcile.Result{}, err
	}

	// Ensure the StatefulSet size is the same as the spec
	size := instance.Spec.Replicas
	if *statefulset.Spec.Replicas != size {
		statefulset.Spec.Replicas = &size
		err = r.Update(context.TODO(), statefulset)
		if err != nil {
			log1.Error(err, "Failed to update StatefulSet.", "StatefulSet.Namespace", instance.Namespace, "StatefulSet.Name", instance.Name)
			return reconcile.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *RabbitMQInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {

	if err := mgr.GetFieldIndexer().IndexField(&core.Pod{}, ownerKey, func(rawObj runtime.Object) []string {
		// grab the job object, extract the owner...
		job := rawObj.(*core.Pod)
		owner := metav1.GetControllerOf(job)
		if owner == nil {
			return nil
		}
		// ...make sure it's a Pod...
		if owner.APIVersion != apiGVStr || owner.Kind != "RabbitMQInstance" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&susev1beta1.RabbitMQInstance{}).
		Complete(r)
}

func (r *RabbitMQInstanceReconciler) constructPod(s *susev1beta1.RabbitMQInstance) (*core.Pod, error) {
	namePrefix := fmt.Sprintf("rmq-%s-", s.Name)
	log.Printf("creating pod...")
	pod := &core.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels:       make(map[string]string),
			Annotations:  make(map[string]string),
			GenerateName: namePrefix,
			Namespace:    s.Namespace,
		},
		Spec: core.PodSpec{
			Containers: []core.Container{
				core.Container{
					Image: "rabbitmq:3.8-management",
					Name:  "rabbitmq",
					Env:   []core.EnvVar{},
					Ports: []core.ContainerPort{
						core.ContainerPort{
							ContainerPort: 15672,
							Name:          "rabbitmq-mgm",
							Protocol:      "TCP",
						},
						core.ContainerPort{
							ContainerPort: 5672,
							Name:          "rabbitmq",
							Protocol:      "TCP",
						},
					},
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(s, pod, r.Scheme); err != nil {
		return nil, err
	}
	return pod, nil
}
func labelsForRabbitMQ(name string) map[string]string {
	return map[string]string{"app": "rabbitmq", "rabbitmq_cr": name}
}

func labelSelector(labels map[string]string) *metav1.LabelSelector {
	return &metav1.LabelSelector{MatchLabels: labels}
}

// statefulsetForRabbitMQ returns a rabbitmq statefulset object
func (r *RabbitMQInstanceReconciler) statefulsetForRabbitMQ(m *susev1beta1.RabbitMQInstance) (*appsv1.StatefulSet, error) {
	labels := labelsForRabbitMQ(m.Name)
	replicas := m.Spec.Replicas
	statefulset := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.StatefulSetSpec{
			Selector:    labelSelector(labels),
			ServiceName: m.Name,
			Replicas:    &replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
					Containers: []corev1.Container{
						corev1.Container{
							Name:  "rabbitmq",
							Image: "rabbitmq:3.8-management",
						},
					},
				},
			},
		},
	}

	controllerutil.SetControllerReference(m, statefulset, r.Scheme)
	return statefulset, nil
}

// serviceForRabbitMQ function takes in a RabbitMQ object and returns a Service for that object.
func (r *RabbitMQInstanceReconciler) serviceForRabbitMQ(m *susev1beta1.RabbitMQInstance) *corev1.Service {
	ls := labelsForRabbitMQ(m.Name)
	ser := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: ls,
			Ports: []corev1.ServicePort{
				{
					Port: 15672,
					Name: m.Name,
				},
			},
		},
	}
	// Set Memcached instance as the owner of the Service.
	controllerutil.SetControllerReference(m, ser, r.Scheme)
	return ser
}
