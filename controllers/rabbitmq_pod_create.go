package controllers

import (
	"fmt"

	opv1alpha "github.com/gsantomaggio/rabbitmq-operator/api/v1alpha"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newService(cr *opv1alpha.RabbitMQ, r *RabbitMQReconcilerCreate) (*corev1.Service, error) {
	labels := cr.ObjectMeta.Labels
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
	if r != nil {
		controllerutil.SetControllerReference(cr, service, r.Scheme)
	}
	return service, nil
}

func getResourceList(storage string) v1.ResourceList {
	res := v1.ResourceList{}
	if storage != "" {
		res[v1.ResourceStorage] = resource.MustParse(storage)
	}
	return res
}

func ConfigurePersistentVolumes(cr *opv1alpha.RabbitMQ) []v1.PersistentVolumeClaim {
	if cr.Spec.PersistentVolume.StorageClass != "" {
		volumeClaimTemplates := []v1.PersistentVolumeClaim{
			v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: cr.Spec.PersistentVolume.Name,
				},
				Spec: v1.PersistentVolumeClaimSpec{
					StorageClassName: &cr.Spec.PersistentVolume.StorageClass,
					AccessModes:      cr.Spec.PersistentVolume.AccessModes,
					Resources:        cr.Spec.PersistentVolume.Resources,
				},
			},
		}
		return volumeClaimTemplates
	}
	return []v1.PersistentVolumeClaim{}
}

func configureVolumesMap(cr *opv1alpha.RabbitMQ) []v1.VolumeMount {
	volumeMounts := []v1.VolumeMount{
		v1.VolumeMount{
			Name:      "config-volume",
			MountPath: "/etc/rabbitmq/",
		},
	}
	if cr.Spec.PersistentVolume.StorageClass != "" {
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      cr.Spec.PersistentVolume.StorageClass,
			MountPath: "/var/lib/rabbitmq/",
		})
	}
	return volumeMounts
}

func configureReadinessProbe(cr *opv1alpha.RabbitMQ) *v1.Probe {
	return &v1.Probe{
		Handler:             configureNessHandler(),
		PeriodSeconds:       cr.Spec.Template.Spec.Contaniers.ReadinessProbe.PeriodSeconds,
		TimeoutSeconds:      cr.Spec.Template.Spec.Contaniers.ReadinessProbe.TimeoutSeconds,
		FailureThreshold:    6,
		InitialDelaySeconds: cr.Spec.Template.Spec.Contaniers.ReadinessProbe.InitialDelaySeconds,
	}
}

func configureLivenessProbe(cr *opv1alpha.RabbitMQ) *v1.Probe {
	return &v1.Probe{
		Handler:             configureNessHandler(),
		PeriodSeconds:       cr.Spec.Template.Spec.Contaniers.LivenessProbe.PeriodSeconds,
		TimeoutSeconds:      cr.Spec.Template.Spec.Contaniers.LivenessProbe.TimeoutSeconds,
		FailureThreshold:    6,
		InitialDelaySeconds: cr.Spec.Template.Spec.Contaniers.LivenessProbe.InitialDelaySeconds,
	}
}

func configureVolumes(cr *opv1alpha.RabbitMQ) []v1.Volume {
	var mode int32 = 0777
	Volumes := []v1.Volume{
		v1.Volume{
			Name: "-volume",
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					DefaultMode: &mode,
					LocalObjectReference: v1.LocalObjectReference{
						Name: cr.Spec.ConfigMap,
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
	}
	return Volumes
}

func configureEnvVariables(crd *opv1alpha.RabbitMQ, s *corev1.Service) []v1.EnvVar {
	Env := []v1.EnvVar{
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
			Value: s.Name,
		},
		v1.EnvVar{
			Name:  "RABBITMQ_NODENAME",
			Value: fmt.Sprintf("rabbit@%s.%s.%s.svc.cluster.local", "$(MY_POD_NAME)", "$(K8S_SERVICE_NAME)", "$(MY_POD_NAMESPACE)"),
		},
		v1.EnvVar{
			Name:  "K8S_HOSTNAME_SUFFIX",
			Value: fmt.Sprintf(".%s.%s.svc.cluster.local", crd.Name, crd.Namespace),
		},
		v1.EnvVar{
			Name:  "RABBITMQ_ERLANG_COOKIE",
			Value: "here_need_a_secret",
		},
	}

	return Env

}

func configureNessHandler() v1.Handler {
	commandRMQ := []string{"rabbitmq-diagnostics", "status"}
	nessProbeHandler := v1.Handler{
		Exec: &v1.ExecAction{
			Command: commandRMQ,
		},
	}
	return nessProbeHandler
}

func configureContaniers(crd *opv1alpha.RabbitMQ, s *corev1.Service) []corev1.Container {
	containers := []corev1.Container{
		corev1.Container{
			Name:            crd.Spec.Template.Spec.Contaniers.Name,
			Image:           crd.Spec.Template.Spec.Contaniers.Image,
			ReadinessProbe:  configureReadinessProbe(crd),
			LivenessProbe:   configureLivenessProbe(crd),
			ImagePullPolicy: crd.Spec.Template.Spec.Contaniers.ImagePullPolicy,
			VolumeMounts:    configureVolumesMap(crd),
			Env:             configureEnvVariables(crd, s),
		},
	}
	return containers
}

func configurePodTemplateSpec(crd *opv1alpha.RabbitMQ, s *corev1.Service) corev1.PodTemplateSpec {

	template := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: crd.ObjectMeta.Labels,
		},
		Spec: corev1.PodSpec{
			TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
			Volumes:                       configureVolumes(crd),
			Containers:                    configureContaniers(crd, s),
		},
	}
	return template
}

func createStatefulSet(cr *opv1alpha.RabbitMQ, r *RabbitMQReconcilerCreate, s *corev1.Service) (*appsv1.StatefulSet, error) {
	labels := cr.ObjectMeta.Labels

	statefulset := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.StatefulSetSpec{
			Selector:             labelSelector(labels),
			ServiceName:          cr.Name,
			Replicas:             &cr.Spec.Replicas,
			VolumeClaimTemplates: ConfigurePersistentVolumes(cr),
			Template:             configurePodTemplateSpec(cr, s),
		},
	}
	if r != nil {
		controllerutil.SetControllerReference(cr, statefulset, r.Scheme)
	}
	return statefulset, nil
}
