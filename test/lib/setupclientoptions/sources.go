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

package setupclientoptions

import (
	"context"

	contribtestlib "knative.dev/eventing-redis/test/lib"
	contribresources "knative.dev/eventing-redis/test/lib/resources"

	testlib "knative.dev/eventing/test/lib"
	"knative.dev/eventing/test/lib/recordevents"
	"knative.dev/eventing/test/lib/resources"
)

// RedisStreamSourceV1Alpha1ClientSetupOption returns a ClientSetupOption that can be used
// to create a new RedisStreamSource. It creates a ServiceAccount, a Role, a
// RoleBinding, a RecordEvents pod and an RedisStreamSource object with the event
// mode and RecordEvent pod as its sink.
func RedisStreamSourceV1Alpha1ClientSetupOption(name string, address string, recordEventsPodName string) testlib.SetupClientOption {
	return func(client *testlib.Client) {

		recordevents.StartEventRecordOrFail(context.Background(), client, recordEventsPodName)

		contribtestlib.CreateRedisStreamSourceV1Alpha1OrFail(client, contribresources.NewRedisStreamSourceV1Alpha1(
			name,
			client.Namespace,
			address,
			resources.ServiceRef(recordEventsPodName),
		))

		client.WaitForAllTestResourcesReadyOrFail(context.Background())
	}
}
