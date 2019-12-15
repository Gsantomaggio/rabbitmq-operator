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

	"github.com/go-logr/logr"
	scalingv1 "github.com/gsantomaggio/rabbitmq-operator/api/v1alpha"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// RabbitMQReconcilerCreate reconciles a RabbitMQ object
type RabbitMQReconcilerCreate struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=scaling.queues,resources=rabbitmqs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=scaling.queues,resources=rabbitmqs/status,verbs=get;update;patch
// Reconcile handles the reconcile
func (r *RabbitMQReconcilerCreate) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	logRmq := r.Log.WithValues("rabbitmq_w", req.NamespacedName)
	logRmq.Info("Handling Create RabbitMQReconciler ")

	instance, err := getRabbitMQInstanceResource(r.Recorder, r.Client, req)
	if err != nil {
		return ctrl.Result{Requeue: false}, nil
	}

	statefulset := &appsv1.StatefulSet{}

	err = r.Get(context.TODO(),
		types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace},
		statefulset)

	if err != nil && errors.IsNotFound(err) {
		service := &corev1.Service{}

		if instance.Spec.ServiceDefinition == scalingv1.Internal {
			serv, errSrv := newService(instance, r)
			errSrv = r.Create(context.TODO(), serv.DeepCopy())
			if errSrv != nil && errors.IsAlreadyExists(errSrv) == false {
				logRmq.Error(errSrv, "Failed to create new Service.", "Service Namespace", serv.Namespace,
					"Service Name:", serv.Name)
				return ctrl.Result{}, errSrv
			}
			service = serv
		} else {

			err = r.Get(context.TODO(),
				types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace},
				service)
		}

		// Define a new Statefulset
		dep, err := createStatefulSet(instance, r, service)

		err = r.Create(context.TODO(), dep.DeepCopy())
		if err != nil && errors.IsAlreadyExists(err) == false {
			logRmq.Error(err, "Failed to create new statefulset.", "statefulset.Namespace", dep.Namespace, "statefulset.Name", dep.Name)
			return ctrl.Result{}, err
		}

		r.Recorder.Event(instance, "Normal", "Creating",
			fmt.Sprintf("Creating Statefulset %s/%s", dep.Namespace, dep.Name))
		// Deployment created successfully - return and requeue
		// NOTE: that the requeue is made with the purpose to provide the deployment object for the next step to ensure the deployment size is the same as the spec.
		// Also, you could GET the deployment object again instead of requeue if you wish. See more over it here: https://godoc.org/sigs.k8s.io/controller-runtime/pkg/reconcile#Reconciler
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		logRmq.Error(err, "Failed to get Deployment.")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *RabbitMQReconcilerCreate) SetupWithManager(mgr ctrl.Manager) error {
	p := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return false
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},

		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
	}

	c := ctrl.NewControllerManagedBy(mgr)
	return c.For(&scalingv1.RabbitMQ{}).
		Named("RabbitMQCreate").
		WithEventFilter(p).
		Complete(r)

}
