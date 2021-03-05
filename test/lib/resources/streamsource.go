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
	"knative.dev/eventing-redis/source/pkg/apis/sources/v1alpha1"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

// RedisStreamSourceOptionV1Alpha1 enables further configuration of a RedisStreamSource.
type RedisStreamSourceOptionV1Alpha1 func(source *v1alpha1.RedisStreamSource)

// NewRedisStreamSourceV1Alpha1 creates a RedisStreamSource with RedisStreamSourceOption.
func NewRedisStreamSourceV1Alpha1(name, namespace string, address string, ref *corev1.ObjectReference, options ...RedisStreamSourceOptionV1Alpha1) *v1alpha1.RedisStreamSource {
	source := &v1alpha1.RedisStreamSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.RedisStreamSourceSpec{
			RedisConnection: v1alpha1.RedisConnection{
				Address: address,
			},
			Stream: "test-stream",
			Group:  "test-consumer-group",
			SourceSpec: duckv1.SourceSpec{
				Sink: duckv1.Destination{
					Ref: &duckv1.KReference{
						APIVersion: ref.APIVersion,
						Kind:       ref.Kind,
						Name:       ref.Name,
						Namespace:  ref.Namespace,
					},
				},
			},
		},
	}

	for _, opt := range options {
		opt(source)
	}

	return source
}

func WithRedisStreamSourceSpecV1Alpha1(spec v1alpha1.RedisStreamSourceSpec) RedisStreamSourceOptionV1Alpha1 {
	return func(c *v1alpha1.RedisStreamSource) {
		c.Spec = spec
	}
}

func WithRedisStreamSourceConditionsV1Alpha1(s *v1alpha1.RedisStreamSource) {
	s.Status.InitializeConditions()
}

func WithRedisStreamSourceSinkV1Alpha1(url string) RedisStreamSourceOptionV1Alpha1 {
	return func(s *v1alpha1.RedisStreamSource) {
		s.Status.MarkSink(url)
	}
}

func WithNameV1Alpha1(name string) RedisStreamSourceOptionV1Alpha1 {
	return func(source *v1alpha1.RedisStreamSource) {
		source.Name = name
	}
}
