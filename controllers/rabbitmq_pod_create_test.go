package controllers

import (
	scalingv1alpha "github.com/gsantomaggio/rabbitmq-operator/api/v1alpha"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RabbitMQ Pod Create ", func() {

	var ConfigureVolumesMap = configureVolumesMap

	BeforeEach(func() {

	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

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
	var ConfigureLivenessProbe = configurelivenessProbe
	Context("Test Readiness Liveness ", func() {
		It("Liveness Should Be Equal", func() {
			crd := scalingv1alpha.NewRabbitMQStruct()
			crd.Spec.Template.Spec.Contaniers.LivenessProbe.PeriodSeconds = 13
			crd.Spec.Template.Spec.Contaniers.LivenessProbe.InitialDelaySeconds = 14
			crd.Spec.Template.Spec.Contaniers.LivenessProbe.TimeoutSeconds = 15
			res := &v1.Probe{
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
