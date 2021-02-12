/*
Copyright 2019 The Knative Authors

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
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1alpha1 "knative.dev/eventing-redis/source/pkg/apis/sources/v1alpha1"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/kmp"
)

func TestMakeReceiveAdapter(t *testing.T) {
	src := &v1alpha1.RedisStreamSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "source-name",
			Namespace: "source-namespace",
		},
		Spec: v1alpha1.RedisStreamSourceSpec{
			RedisConnection: v1alpha1.RedisConnection{
				Address: "redis.redis.svc.cluster.local:6379",
			},
			Stream: "mystream",
		},
	}

	got := MakeReceiveAdapter(src, "test-image", "sink-uri", "5", "")

	one := int32(1)
	labels := Labels(src.Name)
	want := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    "source-namespace",
			GenerateName: "source-name-",
			Labels:       labels,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(src),
			},
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Replicas: &one,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: ServiceAccountName(src),
					Containers: []corev1.Container{
						{
							Name:  "receive-adapter",
							Image: "test-image",
							Env: []corev1.EnvVar{
								{
									Name:  "STREAM",
									Value: src.Spec.Stream,
								}, {
									Name:  "ADDRESS",
									Value: src.Spec.Address,
								}, {
									Name:  "K_SINK",
									Value: "sink-uri",
								}, {
									Name:  "NUM_CONSUMERS",
									Value: "5",
								}, {
									Name:  "TLS_CERTIFICATE",
									Value: "",
								}, {
									Name: "NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								}, {
									Name: "NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								}, {
									Name:  "METRICS_DOMAIN",
									Value: "knative.dev/eventing",
								},
							},
							Ports: []corev1.ContainerPort{{
								Name:          "metrics",
								ContainerPort: 9090,
							}},
						},
					},
				},
			},
		},
	}

	if diff, err := kmp.SafeDiff(want, got); err != nil {
		t.Errorf("unexpected deploy (-want, +got) = %v", diff)
	}
}
