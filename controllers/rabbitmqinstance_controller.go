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
	core "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

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
	ctx := context.Background()
	log1 := r.Log.WithValues("rabbitmqinstance", req.NamespacedName)
	if err != nil {
		log.Printf("error %s", err)
		return ctrl.Result{}, err
	}
	log.Printf("Checking status of res: %d req: %s", instance.Spec.Replicas, req.NamespacedName.Name)
	var childPods core.PodList
	if err := r.List(ctx, &childPods, client.InNamespace(req.Namespace), client.MatchingField(ownerKey, req.Name)); err != nil {
		log1.Error(err, "unable to list child Pods")
		return ctrl.Result{}, err
	}
	for len(childPods.Items) <= instance.Spec.Replicas {
		log.Printf("starting RabbitMQ: %d", len(childPods.Items))

		var pod *core.Pod
		var err error
		pod, err = r.constructPod(instance)

		if err != nil {
			log.Printf("error: %s", err)
			return ctrl.Result{}, err
		}

		if err := r.Create(ctx, pod); err != nil {
			log1.Error(err, "unable to create Pod for RabbitMQ", "pod", pod)
			return ctrl.Result{}, err

		}
		if err := r.List(ctx, &childPods, client.InNamespace(req.Namespace), client.MatchingField(ownerKey, req.Name)); err != nil {
			log1.Error(err, "unable to list child Pods")
			return ctrl.Result{}, err
		}
		if err := r.List(ctx, &childPods, client.InNamespace(req.Namespace), client.MatchingField(ownerKey, req.Name)); err != nil {
			log1.Error(err, "unable to list child Pods")
			return ctrl.Result{}, err
		}
		log.Printf("Finished RabbitMQ: %d", len(childPods.Items))

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

	///	addEnv := func(key string, value string) {
	///		pod.Spec.Containers[0].Env = append(pod.Spec.Containers[0].Env,
	///			core.EnvVar{Name: key, Value: value})
	///	}
	///	bool2string := func(b bool) string {
	///		if b {
	///			return "TRUE"
	///		} else {
	///			return "FALSE"
	///		}
	///	}
	///
	// TODO: If these values are blank we should just not set the env variable.
	//addEnv("EULA", bool2string(s.Spec.EULA))
	//addEnv("TYPE", s.Spec.ServerType)
	//addEnv("SERVER_NAME", s.Spec.ServerName)
	//addEnv("OPS", strings.Join(s.Spec.Ops, ","))
	//addEnv("WHITELIST", strings.Join(s.Spec.Allowlist, ","))
	if err := ctrl.SetControllerReference(s, pod, r.Scheme); err != nil {
		return nil, err
	}
	return pod, nil
}
