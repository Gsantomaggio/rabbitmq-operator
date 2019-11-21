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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	ownerKey = ".metadata.controller"
	//apiGVStr                      = susev1beta1.GroupVersion.String()
	terminationGracePeriodSeconds = int64(10)
)

func labelsForRabbitMQ(name string) map[string]string {
	return map[string]string{"app": "rabbitmq-operator", "rabbitmq_cr": name}
}

func labelSelector(labels map[string]string) *metav1.LabelSelector {
	return &metav1.LabelSelector{MatchLabels: labels}
}