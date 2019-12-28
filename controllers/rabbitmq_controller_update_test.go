package controllers

import (
	"context"
	"time"

	scalingv1alpha "github.com/gsantomaggio/rabbitmq-operator/api/v1alpha"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("RabbitMQ Controller", func() {

	const timeout = time.Second * 30
	const interval = time.Second * 4

	var (
		name           = "rabbitmq-operator"
		namespace      = "rabbitmq-system"
		replicas       = int32(5)
		replicasUpdate = int32(8)
	)

	BeforeEach(func() {

	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	Context("Reconcile", func() {
		It("Check the Initial Scaling Value", func() {
			rabbitmqScaling := &scalingv1alpha.RabbitMQ{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: scalingv1alpha.RabbitMQSpec{
					Replicas: replicas,
				}}

			s := scheme.Scheme
			s.AddKnownTypes(scalingv1alpha.GroupVersion, rabbitmqScaling)

			// Objects to track in the fake client.
			objs := []runtime.Object{rabbitmqScaling}
			cl := fake.NewFakeClient(objs...)

			r := &RabbitMQReconcilerCreate{
				Client:   cl,
				Log:      ctrl.Log.WithName("controllers").WithName("RabbitMQ"),
				Scheme:   s,
				Recorder: nil}

			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      name,
					Namespace: namespace,
				},
			}
			_, err := r.Reconcile(req)

			Ω(err).Should(BeNil())

			// Check if deployment has been created and has the correct size.
			dep := &appsv1.StatefulSet{}
			err = r.Get(context.TODO(), req.NamespacedName, dep)
			Ω(err).Should(BeNil())
			// Check if the quantity of Replicas for this StatefulSet is equals the specification
			dsize := *dep.Spec.Replicas
			Ω(dsize).Should(Equal(replicas))

			rabbitmqScalingUp := &scalingv1alpha.RabbitMQ{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: scalingv1alpha.RabbitMQSpec{
					Replicas: replicasUpdate,
				}}

			cl.Update(context.TODO(), rabbitmqScalingUp)
			s.AddKnownTypes(scalingv1alpha.GroupVersion, rabbitmqScalingUp)

			rUpdate := &RabbitMQReconcilerUpdate{
				Client:   cl,
				Log:      ctrl.Log.WithName("controllers").WithName("RabbitMQ"),
				Scheme:   s,
				Recorder: nil}
			_, err = rUpdate.Reconcile(req)

			depUpdate := &appsv1.StatefulSet{}
			err = rUpdate.Get(context.TODO(), req.NamespacedName, depUpdate)
			Ω(err).Should(BeNil())

			dsize = *depUpdate.Spec.Replicas
			Ω(dsize).Should(Equal(replicasUpdate))

			err = r.Delete(context.TODO(), depUpdate)
			Ω(err).Should(BeNil())

		})
	})
})
