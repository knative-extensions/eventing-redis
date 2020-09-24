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
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"knative.dev/pkg/kmeta"

	sourcesv1alpha1 "knative.dev/eventing-redis/source/pkg/apis/sources/v1alpha1"
)

func AdapterName(source *sourcesv1alpha1.RedisStreamSource) string {
	return kmeta.ChildName(fmt.Sprintf("redistreamsource-%s-", source.Name), string(source.UID))
}

// MakeReceiveAdapter generates (but does not insert into K8s) the Receive Adapter Deployment for
// RedisStream Sources.
func MakeReceiveAdapter(source *sourcesv1alpha1.RedisStreamSource, image string, sinkURI string) *appsv1.Deployment {
	replicas := int32(1)
	labels := Labels(source.Name)
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: source.Namespace,
			Name:      AdapterName(source),
			Labels:    labels,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(source),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: ServiceAccountName(source),
					Containers: []corev1.Container{
						{
							Name:  "receive-adapter",
							Image: image,
							Env: []corev1.EnvVar{{
								Name:  "STREAM",
								Value: source.Spec.Stream,
							}, {
								Name:  "ADDRESS",
								Value: source.Spec.Address,
							}, {
								Name:  "GROUP",
								Value: source.Spec.Group,
							}, {
								Name:  "K_SINK",
								Value: sinkURI,
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
							}},
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
}
