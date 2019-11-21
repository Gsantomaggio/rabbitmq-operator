package controllers

import (
	"context"
	"time"

	scalingv1 "github.com/gsantomaggio/rabbitmq-operator/api/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("WorkspaceItem Controller", func() {

	const timeout = time.Second * 30
	const interval = time.Second * 1

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	// Add Tests for OpenAPI validation (or additonal CRD features) specified in
	// your API definition.
	// Avoid adding tests for vanilla CRUD operations because they would
	// test Kubernetes API server, which isn't the goal here.
	Context("Workspace Item", func() {
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
					Replicas: 4,
				},
			}

			// Create
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())

			By("Expecting submitted")
			Eventually(func() bool {
				f := &scalingv1.RabbitMQ{}
				k8sClient.Get(context.Background(), key, f)
				return true
			}, timeout, interval).Should(BeTrue())

			// Update
			updated := &scalingv1.RabbitMQ{}
			Expect(k8sClient.Get(context.Background(), key, updated)).Should(Succeed())

			updated.Spec.Replicas = 3
			Expect(k8sClient.Update(context.Background(), updated)).Should(Succeed())

			// Delete
			By("Expecting to delete successfully")
			Eventually(func() error {
				f := &scalingv1.RabbitMQ{}
				k8sClient.Get(context.Background(), key, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			By("Expecting to delete finish")
			Eventually(func() error {
				f := &scalingv1.RabbitMQ{}
				return k8sClient.Get(context.Background(), key, f)
			}, timeout, interval).ShouldNot(Succeed())
		})
	})
})
