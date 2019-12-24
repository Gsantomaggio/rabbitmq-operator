package controllers

import (
	"fmt"

	scalingv1alpha "github.com/gsantomaggio/rabbitmq-operator/api/v1alpha"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RabbitMQ Pod Create ", func() {

	BeforeEach(func() {

	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	var NewService = newService

	Context("Test New Service ", func() {
		It("Should Be Equal ", func() {
			crd := scalingv1alpha.NewRabbitMQStruct()
			m := make(map[string]string)
			m["label1"] = "test_label"
			crd.ObjectMeta.Labels = m
			labels := crd.ObjectMeta.Labels
			res := &corev1.Service{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Service",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      crd.ObjectMeta.Name,
					Namespace: crd.Namespace,
					Labels:    labels,
				},
				Spec: corev1.ServiceSpec{
					ClusterIP: corev1.ClusterIPNone,
					Selector:  labels,
				},
			}
			Ω(NewService(crd, nil)).Should(Equal(res))
		})
	})

	var CreateStatefulSet = createStatefulSet

	Context("Test Stateful creation  ", func() {
		It("Should Be Equal ", func() {
			crd := scalingv1alpha.NewRabbitMQStruct()
			labels := make(map[string]string)
			labels["label1"] = "test_label"
			crd.ObjectMeta.Labels = labels

			service := &corev1.Service{}
			service.Name = "TEST_SERVICE"
			crd.Spec.Template.Spec.Contaniers.Name = "TEST_NAME"
			crd.Spec.Template.Spec.Contaniers.Image = "TEST_IMAGE"
			crd.Spec.Template.Spec.Contaniers.ImagePullPolicy = "Always"
			res := &appsv1.StatefulSet{
				TypeMeta: metav1.TypeMeta{
					Kind:       "StatefulSet",
					APIVersion: "apps/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      crd.Name,
					Namespace: crd.Namespace,
					Labels:    labels,
				},
				Spec: appsv1.StatefulSetSpec{
					Selector:             labelSelector(labels),
					ServiceName:          crd.Name,
					Replicas:             &crd.Spec.Replicas,
					VolumeClaimTemplates: ConfigurePersistentVolumes(crd),
					Template:             configurePodTemplateSpec(crd, service),
				},
			}
			Ω(CreateStatefulSet(crd, nil, service)).Should(Equal(res))
		})
	})

	var ConfigurePodTemplateSpec = configurePodTemplateSpec
	Context("Test Configur ePod Template Spec ", func() {
		It("Should Be Equal ", func() {
			crd := scalingv1alpha.NewRabbitMQStruct()
			service := &corev1.Service{}
			service.Name = "TEST_SERVICE"
			crd.Spec.Template.Spec.Contaniers.Name = "TEST_NAME"
			crd.Spec.Template.Spec.Contaniers.Image = "TEST_IMAGE"
			crd.Spec.Template.Spec.Contaniers.ImagePullPolicy = "Always"
			res := corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: crd.ObjectMeta.Labels,
				},
				Spec: corev1.PodSpec{
					TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
					Volumes:                       configureVolumes(crd),
					Containers:                    configureContaniers(crd, service),
				},
			}
			Ω(ConfigurePodTemplateSpec(crd, service)).Should(Equal(res))
		})
	})

	var ConfigureContaniers = configureContaniers
	Context("Test Configure Containers ", func() {
		It("Should Be Equal ", func() {
			crd := scalingv1alpha.NewRabbitMQStruct()
			service := &corev1.Service{}
			service.Name = "TEST_SERVICE"
			crd.Spec.Template.Spec.Contaniers.Name = "TEST_NAME"
			crd.Spec.Template.Spec.Contaniers.Image = "TEST_IMAGE"
			crd.Spec.Template.Spec.Contaniers.ImagePullPolicy = "Always"
			res := []corev1.Container{
				corev1.Container{
					Name:            crd.Spec.Template.Spec.Contaniers.Name,
					Image:           crd.Spec.Template.Spec.Contaniers.Image,
					ReadinessProbe:  configureReadinessProbe(crd),
					LivenessProbe:   configureLivenessProbe(crd),
					ImagePullPolicy: crd.Spec.Template.Spec.Contaniers.ImagePullPolicy,
					VolumeMounts:    configureVolumesMap(crd),
					Env:             configureEnvVariables(crd, service),
				},
			}
			Ω(ConfigureContaniers(crd, service)).Should(Equal(res))
		})
	})

	var ConfigureEnvVariables = configureEnvVariables
	Context("Test Enviroment Variables", func() {
		It("Should Be Equal ", func() {
			crd := scalingv1alpha.NewRabbitMQStruct()
			service := &corev1.Service{}
			service.Name = "TEST_SERVICE"
			res := []v1.EnvVar{
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
					Value: service.Name,
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

			Ω(ConfigureEnvVariables(crd, service)).Should(Equal(res))
		})
	})

	var ConfigureVolumesMap = configureVolumesMap
	Context("Test Configure Map without storage class ", func() {
		It("Should Be Equal ", func() {
			crd := scalingv1alpha.NewRabbitMQStruct()
			res := []v1.VolumeMount{
				v1.VolumeMount{
					Name:      "config-volume",
					MountPath: "/etc/rabbitmq/",
				},
			}
			Ω(ConfigureVolumesMap(crd)).Should(Equal(res))
		})
	})

	Context("Test Configure Map with storage class ", func() {
		It("Should create successfully", func() {
			crd := scalingv1alpha.NewRabbitMQStruct()
			storageClass := "standard"
			crd.Spec.PersistentVolume.StorageClass = storageClass
			crd.Spec.PersistentVolume.Name = "pvname"
			res := []v1.VolumeMount{
				v1.VolumeMount{
					Name:             "config-volume",
					ReadOnly:         false,
					MountPath:        "/etc/rabbitmq/",
					SubPath:          "",
					MountPropagation: nil,
					SubPathExpr:      "",
				},
				v1.VolumeMount{
					Name:             storageClass,
					ReadOnly:         false,
					MountPath:        "/var/lib/rabbitmq/",
					SubPath:          "",
					MountPropagation: nil,
					SubPathExpr:      "",
				},
			}
			Ω(ConfigureVolumesMap(crd)).Should(Equal(res))
		})
	})
	var ConfigureReadinessProbe = configureReadinessProbe
	var ConfigureLivenessProbe = configureLivenessProbe
	var ConfigureNessHandler = configureNessHandler

	Context("Test Readiness Liveness ", func() {
		It("Liveness Should Be Equal", func() {
			crd := scalingv1alpha.NewRabbitMQStruct()
			crd.Spec.Template.Spec.Contaniers.LivenessProbe.PeriodSeconds = 13
			crd.Spec.Template.Spec.Contaniers.LivenessProbe.InitialDelaySeconds = 14
			crd.Spec.Template.Spec.Contaniers.LivenessProbe.TimeoutSeconds = 15
			res := &v1.Probe{
				Handler:             ConfigureNessHandler(),
				PeriodSeconds:       crd.Spec.Template.Spec.Contaniers.LivenessProbe.PeriodSeconds,
				TimeoutSeconds:      crd.Spec.Template.Spec.Contaniers.LivenessProbe.TimeoutSeconds,
				FailureThreshold:    6,
				InitialDelaySeconds: crd.Spec.Template.Spec.Contaniers.LivenessProbe.InitialDelaySeconds,
			}
			Ω(ConfigureLivenessProbe(crd)).Should(Equal(res))
		})
		It("Readiness Should Be Equal", func() {
			crd := scalingv1alpha.NewRabbitMQStruct()
			crd.Spec.Template.Spec.Contaniers.ReadinessProbe.PeriodSeconds = 10
			crd.Spec.Template.Spec.Contaniers.ReadinessProbe.InitialDelaySeconds = 11
			crd.Spec.Template.Spec.Contaniers.ReadinessProbe.TimeoutSeconds = 12
			res := &v1.Probe{
				Handler:             ConfigureNessHandler(),
				PeriodSeconds:       crd.Spec.Template.Spec.Contaniers.ReadinessProbe.PeriodSeconds,
				TimeoutSeconds:      crd.Spec.Template.Spec.Contaniers.ReadinessProbe.TimeoutSeconds,
				FailureThreshold:    6,
				InitialDelaySeconds: crd.Spec.Template.Spec.Contaniers.ReadinessProbe.InitialDelaySeconds,
			}
			Ω(ConfigureReadinessProbe(crd)).Should(Equal(res))
		})
	})

	var ConfigureVolumes = configureVolumes
	Context("Test Configure Volumes", func() {
		It("Should Be Equal", func() {
			var mode int32 = 0777
			crd := scalingv1alpha.NewRabbitMQStruct()
			res := []v1.Volume{
				v1.Volume{
					Name: "-volume",
					VolumeSource: v1.VolumeSource{
						ConfigMap: &v1.ConfigMapVolumeSource{
							DefaultMode: &mode,
							LocalObjectReference: v1.LocalObjectReference{
								Name: crd.Spec.ConfigMap,
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
			Ω(ConfigureVolumes(crd)).Should(Equal(res))
		})
	})

	Context("Test Persistent Volume without storage class", func() {
		It("Should Be Equal", func() {
			res := []v1.PersistentVolumeClaim{}
			Ω(ConfigurePersistentVolumes(scalingv1alpha.NewRabbitMQStruct())).Should(Equal(res))
		})
	})

	Context("Test Persistent Volume with storage class", func() {
		It("Should create successfully", func() {
			crd := scalingv1alpha.NewRabbitMQStruct()
			storageClass := "standard"
			crd.Spec.PersistentVolume.StorageClass = storageClass
			crd.Spec.PersistentVolume.Name = "pvname"
			crd.Spec.PersistentVolume.AccessModes = []v1.PersistentVolumeAccessMode{"ReadWriteOnce"}
			crd.Spec.PersistentVolume.Resources = v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceEphemeralStorage: resource.MustParse("1Gi"),
				},
			}
			res := []v1.PersistentVolumeClaim{
				v1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pvname",
					},
					Spec: v1.PersistentVolumeClaimSpec{
						StorageClassName: &storageClass,
						AccessModes:      []v1.PersistentVolumeAccessMode{"ReadWriteOnce"},

						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceEphemeralStorage: resource.MustParse("1Gi"),
							},
						},
					},
				},
			}
			Ω(ConfigurePersistentVolumes(crd)).Should(Equal(res))
		})
	})
})
