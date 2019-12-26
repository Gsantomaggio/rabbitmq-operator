package controllers

import (
	"context"
	"time"

	scalingv1 "github.com/gsantomaggio/rabbitmq-operator/api/v1alpha"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
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
		name      = "rabbitmq-operator"
		namespace = "rabbitmq-system"
		replicas  = int32(5)
	)

	BeforeEach(func() {

	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	Context("Reconcile", func() {

		It("Check the internal/external service ", func() {
			rabbitmqInternalService := &scalingv1.RabbitMQ{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: scalingv1.RabbitMQSpec{
					ServiceDefinition: scalingv1.Internal,
				}}

			s := scheme.Scheme
			s.AddKnownTypes(scalingv1.GroupVersion, rabbitmqInternalService)

			// Objects to track in the fake client.
			objs := []runtime.Object{rabbitmqInternalService}
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

			service := &corev1.Service{}
			err = r.Get(context.TODO(),
				types.NamespacedName{Name: req.Name, Namespace: req.Namespace},
				service)

			Ω(err).Should(BeNil())

			dep := &appsv1.StatefulSet{}
			err = r.Get(context.TODO(), req.NamespacedName, dep)
			err = r.Delete(context.TODO(), dep)
			Ω(err).Should(BeNil())

			rabbitmqExternalService := &scalingv1.RabbitMQ{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: scalingv1.RabbitMQSpec{
					ServiceDefinition: scalingv1.External,
				}}

			s.AddKnownTypes(scalingv1.GroupVersion, rabbitmqExternalService)

			// Objects to track in the fake client.
			objs = []runtime.Object{rabbitmqExternalService}
			cl = fake.NewFakeClient(objs...)

			r = &RabbitMQReconcilerCreate{
				Client:   cl,
				Log:      ctrl.Log.WithName("controllers").WithName("RabbitMQ"),
				Scheme:   s,
				Recorder: nil}
			req = reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      name,
					Namespace: namespace,
				},
			}
			_, err = r.Reconcile(req)

			Ω(err).Should(BeNil())

			service = &corev1.Service{}
			err = r.Get(context.TODO(),
				types.NamespacedName{Name: req.Name, Namespace: req.Namespace},
				service)

			Ω(err).ShouldNot(BeNil())

		})

		It("Check the Initial Scaling Value", func() {
			rabbitmqInternalService := &scalingv1.RabbitMQ{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: scalingv1.RabbitMQSpec{
					Replicas: replicas,
				}}

			s := scheme.Scheme
			s.AddKnownTypes(scalingv1.GroupVersion, rabbitmqInternalService)

			// Objects to track in the fake client.
			objs := []runtime.Object{rabbitmqInternalService}
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

			err = r.Delete(context.TODO(), dep)
			Ω(err).Should(BeNil())

		})
	})
	// Add Tests for OpenAPI validation (or additonal CRD features) specified in
	// your API definition.
	// Avoid adding tests for vanilla CRUD operations because they would
	// test Kubernetes API server, which isn't the goal here.
	Context("RabbitMQ Item", func() {
		It("Should create successfully", func() {

			key := types.NamespacedName{
				Name:      "rabbitmq",
				Namespace: "default",
			}

			created := &scalingv1.RabbitMQ{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: scalingv1.RabbitMQSpec{
					Replicas:          3,
					ConfigMap:         "TEST",
					ServiceDefinition: "Internal",
					PersistentVolume: scalingv1.PersistentVolumeClaimSpec{
						StorageClass: "standard",
						AccessModes:  []v1.PersistentVolumeAccessMode{"ReadWriteOnce"},
					},
					Template: scalingv1.TemplateSpec{
						Spec: scalingv1.ContainerSpec{
							Contaniers: scalingv1.ContainerDetailsSpec{
								Name:            "rabbtimq",
								Image:           "rabbitmq",
								ImagePullPolicy: "ifNotPreset",
								LivenessProbe: scalingv1.CheckProbe{
									InitialDelaySeconds: 60,
									TimeoutSeconds:      10,
									PeriodSeconds:       30,
								},
								ReadinessProbe: scalingv1.CheckProbe{
									InitialDelaySeconds: 60,
									TimeoutSeconds:      10,
									PeriodSeconds:       30,
								},
							},
						},
					},
				},
				Status: scalingv1.RabbitMQStatus{},
			}

			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())
			By("Expecting submitted")
			failed := scalingv1.NewRabbitMQStruct()

			Eventually(func() bool {
				k8sClient.Get(context.Background(), key, failed)
				return failed.ObjectMeta.Name == "invalid"
			}, timeout, interval).Should(BeFalse())

			success := scalingv1.NewRabbitMQStruct()
			Eventually(func() bool {
				k8sClient.Get(context.Background(), key, success)
				return success.ObjectMeta.Name == "rabbitmq"
			}, timeout, interval).Should(BeTrue())

			// Update
			updated := scalingv1.NewRabbitMQStruct()
			Expect(k8sClient.Get(context.Background(), key, updated)).Should(Succeed())

			Expect(k8sClient.Update(context.Background(), updated)).Should(Succeed())

			// Delete
			By("Expecting to delete successfully")
			Eventually(func() error {
				f := scalingv1.NewRabbitMQStruct()
				k8sClient.Get(context.Background(), key, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			By("Expecting to delete finish")
			Eventually(func() error {
				f := scalingv1.NewRabbitMQStruct()
				return k8sClient.Get(context.Background(), key, f)
			}, timeout, interval).ShouldNot(Succeed())
		})
	})
})
