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

package lib

import (
	"context"

	testlib "knative.dev/eventing/test/lib"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alpha1 "knative.dev/eventing-redis/source/pkg/apis/sources/v1alpha1"
	clientset "knative.dev/eventing-redis/source/pkg/client/clientset/versioned"
)

func CreateRedisStreamSourceV1Alpha1OrFail(c *testlib.Client, redisStreamSource *v1alpha1.RedisStreamSource) {
	redisStreamSourceSourceClientSet, err := clientset.NewForConfig(c.Config)
	if err != nil {
		c.T.Fatalf("Failed to create v1alpha1 RedisStreamSource client: %v", err)
	}

	rsSources := redisStreamSourceSourceClientSet.SourcesV1alpha1().RedisStreamSources(c.Namespace)
	if createdRedisStreamSource, err := rsSources.Create(context.Background(), redisStreamSource, metav1.CreateOptions{}); err != nil {
		c.T.Fatalf("Failed to create v1alpha1 RedisStreamSource %q: %v", redisStreamSource.Name, err)
	} else {
		c.Tracker.AddObj(createdRedisStreamSource)
	}
}
