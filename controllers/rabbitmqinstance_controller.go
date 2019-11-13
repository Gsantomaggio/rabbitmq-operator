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
	ownerKey = ".metadata.controller"
	apiGVStr = susev1beta1.GroupVersion.String()
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
	deployment := &appsv1.Deployment{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, deployment)

	if err != nil && errors.IsNotFound(err) {
		// Define a new Deployment
		dep := r.deploymentForRabbitMQ(instance)
		log1.Info("Creating a new Deployment.", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(context.TODO(), dep)
		if err != nil {
			log1.Error(err, "Failed to create new Deployment.", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
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

// deploymentForRabbitMQ returns a rabbitmq Deployment object
func (r *RabbitMQInstanceReconciler) deploymentForRabbitMQ(m *susev1beta1.RabbitMQInstance) *appsv1.Deployment {
	ls := labelsForRabbitMQ(m.Name)
	instance := &susev1beta1.RabbitMQInstance{}

	replicas := instance.Spec.Replicas

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "rabbitmq:3.8-management",
						Name:  "rabbitmq",
						//						Command: []string{"memcached", "-m=64", "-o", "modern", "-v"},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 15672,
							Name:          "rabbitmqui",
						}},
					}},
				},
			},
		},
	}
	// Set Memcached instance as the owner of the Deployment.
	controllerutil.SetControllerReference(m, dep, r.Scheme)
	return dep
}
