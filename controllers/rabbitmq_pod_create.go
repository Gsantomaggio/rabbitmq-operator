package controllers

import (
	scalingv1 "github.com/gsantomaggio/rabbitmq-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createStatefulSet(m *scalingv1.RabbitMQ, r *RabbitMQReconciler) (*appsv1.StatefulSet, error) {
	labels := labelsForRabbitMQ(m.Name)
	replicas := m.Spec.Replicas
	commandRMQ := []string{"rabbitmq-diagnostics", "status"}

	readinessProbeHandler := v1.Handler{
		Exec: &v1.ExecAction{
			Command: commandRMQ,
		},
	}

	readinessProbe := &v1.Probe{
		Handler:          readinessProbeHandler,
		PeriodSeconds:    50,
		TimeoutSeconds:   60,
		FailureThreshold: 6,
	}

	livenessProbeHandler := v1.Handler{
		Exec: &v1.ExecAction{
			Command: commandRMQ,
		},
	}

	livenessProbe := &v1.Probe{
		Handler:          livenessProbeHandler,
		PeriodSeconds:    50,
		TimeoutSeconds:   60,
		FailureThreshold: 6,
	}

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
							Name:           "rabbitmq",
							Image:          "rabbitmq:3.8-management",
							LivenessProbe:  livenessProbe,
							ReadinessProbe: readinessProbe,
						},
					},
				},
			},
		},
	}

	controllerutil.SetControllerReference(m, statefulset, r.Scheme)
	return statefulset, nil
}

func updateStatefulSet(m *scalingv1.RabbitMQ, r *RabbitMQReconciler) {

}
