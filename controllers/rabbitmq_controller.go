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
	"log"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1 "k8s.io/api/apps/v1"

	scalingv1 "github.com/gsantomaggio/rabbitmq-operator/api/v1alpha"
)

// RabbitMQReconciler reconciles a RabbitMQ object
type RabbitMQReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=scaling.queues,resources=rabbitmqs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=scaling.queues,resources=rabbitmqs/status,verbs=get;update;patch
// Reconcile handle the reconcile
func (r *RabbitMQReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	logRmq := r.Log.WithValues("rabbitmq", req.NamespacedName)
	instance := scalingv1.NewRabbitMQStruct()
	err := r.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil && errors.IsNotFound(err) {
		log.Printf("Checking err is not nil: %s ", err)
		return reconcile.Result{Requeue: false}, nil
	}

	statefulset := &appsv1.StatefulSet{}
	err = r.Get(context.TODO(),
		types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace},
		statefulset)

	if err != nil && errors.IsNotFound(err) {
		// Define a new Statefulset
		dep, err := createStatefulSet(instance, r)
		logRmq.Info("Creating a new statefulset.", "statefulset.Namespace", dep.Namespace, "statefulset.Name", dep.Name)
		err = r.Create(context.TODO(), dep)
		if err != nil {
			logRmq.Error(err, "Failed to create new statefulset.", "statefulset.Namespace", dep.Namespace, "statefulset.Name", dep.Name)
			return reconcile.Result{}, err
		}
		// Deployment created successfully - return and requeue
		// NOTE: that the requeue is made with the purpose to provide the deployment object for the next step to ensure the deployment size is the same as the spec.
		// Also, you could GET the deployment object again instead of requeue if you wish. See more over it here: https://godoc.org/sigs.k8s.io/controller-runtime/pkg/reconcile#Reconciler
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		logRmq.Error(err, "Failed to get Deployment.")
		return reconcile.Result{}, err
	}

	// Update the definition, in case there is a scaling up or down the statefulset
	// has to change.
	size := instance.Spec.Replicas
	if *statefulset.Spec.Replicas != size {
		statefulset.Spec.Replicas = &size
		err = r.Update(context.TODO(), statefulset)
		if err != nil {
			logRmq.Error(err, "Failed to update StatefulSet.", "StatefulSet.Namespace", instance.Namespace, "StatefulSet.Name", instance.Name)
			return reconcile.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *RabbitMQReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&scalingv1.RabbitMQ{}).
		Complete(r)
}
