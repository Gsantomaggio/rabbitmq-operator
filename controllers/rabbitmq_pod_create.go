package controllers

import (
	"fmt"

	scalingv1 "github.com/gsantomaggio/rabbitmq-operator/api/v1alpha"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//func getStatefulSet(m *scalingv1.RabbitMQ, r *RabbitMQReconciler) (*appsv1.StatefulSet, error) {
//labels := labelsForRabbitMQ(m.Name)

// m.Spec.Template.ObjectMeta.Labels = labels
// m.Spec.Template.Spec.Selector = labelSelector(labels)
// m.Spec.Template.Spec.Template.ObjectMeta.Labels = labels

//	controllerutil.SetControllerReference(m, &m.Spec.Template, r.Scheme)
//	return &m.Spec.Template, nil
//}

func newService(cr *scalingv1.RabbitMQ, r *RabbitMQReconciler) (*corev1.Service, error) {
	labels := labelsForHelloStateful(cr.ObjectMeta.Name)
	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.ObjectMeta.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: corev1.ClusterIPNone,
			Selector:  labels,
		},
	}
	controllerutil.SetControllerReference(cr, service, r.Scheme)
	return service, nil
}

func createStatefulSet(m *scalingv1.RabbitMQ, r *RabbitMQReconciler) (*appsv1.StatefulSet, error) {
	labels := labelsForRabbitMQ(m.Name)
	replicas := &m.Spec.Replicas
	commandRMQ := []string{"rabbitmq-diagnostics", "status"}
	var mode int32 = 0777

	readinessProbeHandler := v1.Handler{
		Exec: &v1.ExecAction{
			Command: commandRMQ,
		},
	}

	readinessProbe := &v1.Probe{
		Handler:             readinessProbeHandler,
		PeriodSeconds:       m.Spec.Template.Spec.Contaniers.ReadinessProbe.PeriodSeconds,
		TimeoutSeconds:      m.Spec.Template.Spec.Contaniers.ReadinessProbe.TimeoutSeconds,
		FailureThreshold:    6,
		InitialDelaySeconds: m.Spec.Template.Spec.Contaniers.ReadinessProbe.InitialDelaySeconds,
	}

	livenessProbeHandler := v1.Handler{
		Exec: &v1.ExecAction{
			Command: commandRMQ,
		},
	}

	livenessProbe := &v1.Probe{
		Handler:             livenessProbeHandler,
		PeriodSeconds:       m.Spec.Template.Spec.Contaniers.LivenessProbe.PeriodSeconds,
		TimeoutSeconds:      m.Spec.Template.Spec.Contaniers.LivenessProbe.TimeoutSeconds,
		FailureThreshold:    6,
		InitialDelaySeconds: m.Spec.Template.Spec.Contaniers.LivenessProbe.InitialDelaySeconds,
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
			Replicas:    replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
					Volumes: []v1.Volume{
						v1.Volume{
							Name: "config-volume",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									DefaultMode: &mode,
									LocalObjectReference: v1.LocalObjectReference{
										Name: "rabbitmq-config",
									},
									Items: []v1.KeyToPath{
										v1.KeyToPath{
											Key:  "rabbitmq.conf",
											Path: "rabbitmq.conf",
										},
										v1.KeyToPath{
											Key:  "enabled_plugins",
											Path: "enabled_plugins",
										},
									},
								},
							},
						},
					},
					Containers: []corev1.Container{
						corev1.Container{
							Name:            m.Spec.Template.Spec.Contaniers.Name,
							Image:           m.Spec.Template.Spec.Contaniers.Image,
							LivenessProbe:   livenessProbe,
							ReadinessProbe:  readinessProbe,
							ImagePullPolicy: m.Spec.Template.Spec.Contaniers.ImagePullPolicy,
							VolumeMounts: []v1.VolumeMount{
								v1.VolumeMount{
									Name:      "config-volume",
									MountPath: "/etc/rabbitmq",
								},
							},
							Env: []v1.EnvVar{
								v1.EnvVar{
									Name: "MY_POD_NAME",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "metadata.name",
										},
									},
								},
								v1.EnvVar{
									Name: "MY_POD_NAMESPACE",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "metadata.namespace",
										},
									},
								},

								v1.EnvVar{
									Name:  "RABBITMQ_USE_LONGNAME",
									Value: "true",
								},
								v1.EnvVar{
									Name:  "K8S_SERVICE_NAME",
									Value: "rabbitmq-op",
								},
								v1.EnvVar{
									Name:  "RABBITMQ_NODENAME",
									Value: fmt.Sprintf("rabbit@%s.%s.%s.svc.cluster.local", "$(MY_POD_NAME)", "$(K8S_SERVICE_NAME)", "$(MY_POD_NAMESPACE)"),
								},
								v1.EnvVar{
									Name:  "K8S_HOSTNAME_SUFFIX",
									Value: fmt.Sprintf(".%s.%s.svc.cluster.local", m.Name, m.Namespace),
								},
								v1.EnvVar{
									Name:  "RABBITMQ_ERLANG_COOKIE",
									Value: "here_need_a_secret",
								},
							},
						},
					},
				},
			},
		},
	}

	controllerutil.SetControllerReference(m, statefulset, r.Scheme)
	return statefulset, nil
}
