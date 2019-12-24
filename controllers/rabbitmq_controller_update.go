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
	opv1alpha "github.com/gsantomaggio/rabbitmq-operator/api/v1alpha"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// RabbitMQReconciler reconciles a RabbitMQ object
type RabbitMQReconcilerUpdate struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=scaling.queues,resources=rabbitmqs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=scaling.queues,resources=rabbitmqs/status,verbs=get;update;patch
// Reconcile handles the reconcile
func (r *RabbitMQReconcilerUpdate) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()

	logRmq := r.Log.WithValues("rabbitmq", req.NamespacedName)
	logRmq.Info("Handling .... RabbitMQReconciler Update")
	instance, err := getRabbitMQInstanceResource(r.Recorder, r.Client, req)
	if err != nil {
		return ctrl.Result{Requeue: false}, nil
	}

	statefulset := &appsv1.StatefulSet{}
	err = r.Get(context.TODO(),
		types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace},
		statefulset)

	size := instance.Spec.Replicas
	if size != *statefulset.Spec.Replicas {
		r.Recorder.Event(instance, "Normal", "Scaling",
			fmt.Sprintf("Scaling Statefulset  %s/%s, from %d to %d",
				req.Namespace, req.Name, *statefulset.Spec.Replicas, size))

		statefulset.Spec.Replicas = &size
		err = r.Update(context.TODO(), statefulset.DeepCopy())
		if err != nil {
			logRmq.Error(err, "Failed to scale statefulset.", "statefulset.Namespace",
				req.Namespace, "statefulset.Name", req.Name)
			return ctrl.Result{}, err
		}
	}

	if err != nil && errors.IsAlreadyExists(err) {
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *RabbitMQReconcilerUpdate) SetupWithManager(mgr ctrl.Manager) error {
	p := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return true
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return false
		},

		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
	}

	c := ctrl.NewControllerManagedBy(mgr)
	return c.For(&opv1alpha.RabbitMQ{}).
		Named("RabbitMQUpdate").
		WithEventFilter(p).
		Complete(r)

}
