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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"knative.dev/eventing-redis/pkg/source/apis/sources/v1alpha1"
)

// RedisStreamSourceOptionV1Alpha1 enables further configuration of a RedisStreamSource.
type RedisStreamSourceOptionV1Alpha1 func(source *v1alpha1.RedisStreamSource)

// NewRedisStreamSourceV1Alpha1 creates a RedisStreamSource with RedisStreamSourceOption.
func NewRedisStreamSourceV1Alpha1(name, namespace string, o ...RedisStreamSourceOptionV1Alpha1) *v1alpha1.RedisStreamSource {
	c := &v1alpha1.RedisStreamSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	for _, opt := range o {
		opt(c)
	}
	//c.SetDefaults(context.Background()) // TODO: We should add defaults and validation.
	return c
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
