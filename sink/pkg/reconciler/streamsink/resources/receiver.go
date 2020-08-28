/*
Copyright 2020 The Knative Authors

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

package resources

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"knative.dev/pkg/kmeta"

	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	sinksv1alpha1 "knative.dev/eventing-redis/sink/pkg/apis/sinks/v1alpha1"
)

func ReceiverName(source *sinksv1alpha1.RedisStreamSink) string {
	return kmeta.ChildName("redistreamsink", source.Name)
}

// MakeReceiver generates (but does not insert into K8s) the Receiver Knative Service for
// RedisStreamSinks
func MakeReceiver(sink *sinksv1alpha1.RedisStreamSink, image string) *servingv1.Service {
	labels := Labels(sink.Name)
	return &servingv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: sink.Namespace,
			Name:      ReceiverName(sink),
			Labels:    labels,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(sink),
			},
		},
		Spec: servingv1.ServiceSpec{
			ConfigurationSpec: servingv1.ConfigurationSpec{
				Template: servingv1.RevisionTemplateSpec{
					Spec: servingv1.RevisionSpec{
						PodSpec: corev1.PodSpec{
							ServiceAccountName: ServiceAccountName(sink),
							Containers: []corev1.Container{
								{
									Name:  "receiver",
									Image: image,
									Env: []corev1.EnvVar{{
										Name:  "STREAM",
										Value: sink.Spec.Stream,
									}, {
										Name:  "ADDRESS",
										Value: sink.Spec.Address,
									}, {
										Name:  "METRICS_DOMAIN",
										Value: "knative.dev/eventing",
									}},
								},
							},
						},
					},
				},
			},
		},
	}
}
