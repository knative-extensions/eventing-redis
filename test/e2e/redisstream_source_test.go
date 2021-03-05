//+build e2e mtsource
//+build source

/*
Copyright 2021 The Knative Authors

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

package e2e

import (
	"context"
	"testing"

	//. "github.com/cloudevents/sdk-go/v2/test"
	testlib "knative.dev/eventing/test/lib"
	"knative.dev/eventing/test/lib/resources"

	sourcesv1alpha1 "knative.dev/eventing-redis/source/pkg/apis/sources/v1alpha1"
	sourcestestlib "knative.dev/eventing-redis/test/lib"
	sourcesresources "knative.dev/eventing-redis/test/lib/resources"
)

const (
	redisAddress                = "rediss://redis.redis.svc.cluster.local:6379"
	redisstreamClusterName      = "my-cluster"
	redisstreamClusterNamespace = "redisstream"
)

func TestRedisStreamSource(t *testing.T) {
	var (
		recordEventPodName    = "e2e-redisstream-source-pod-v1alpha1"
		redisstreamSourceName = "e2e-redisstream-source"
	)

	client := testlib.Setup(t, true)
	defer testlib.TearDown(client)

	if len(recordEventPodName) > 63 {
		recordEventPodName = recordEventPodName[:63]
	}
	//eventTracker, _ := recordevents.StartEventRecordOrFail(context.Background(), client, recordEventPodName)

	var (
		cloudEventsSourceName string
		cloudEventsEventType  string
	)

	t.Logf("Creating RedisStreamSource")
	sourcestestlib.CreateRedisStreamSourceV1Alpha1OrFail(client, sourcesresources.NewRedisStreamSourceV1Alpha1(
		redisstreamSourceName,
		client.Namespace,
		redisAddress,
		resources.ServiceRef(recordEventPodName),
	))
	cloudEventsSourceName = sourcesv1alpha1.RedisStreamEventSource(client.Namespace, redisstreamSourceName, redisAddress)
	cloudEventsEventType = sourcesv1alpha1.RedisStreamEventType

	t.Logf(cloudEventsSourceName)
	t.Logf(cloudEventsEventType)

	// matcherGen := func(cloudEventsSourceName, cloudEventsEventType string) EventMatcher {
	// 	return AnyOf(
	// 		HasSource(cloudEventsSourceName),
	// 		HasType(cloudEventsEventType),
	// 	)
	// }

	client.WaitForAllTestResourcesReadyOrFail(context.Background())

	//eventTracker.AssertExact(1, recordevents.MatchEvent(matcherGen(cloudEventsSourceName, cloudEventsEventType)))
}
