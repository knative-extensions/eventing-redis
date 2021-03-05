// +build e2e

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

package conformance

import (
	"log"
	"os"
	"testing"

	"knative.dev/eventing-redis/test"
	"knative.dev/eventing-redis/test/lib/setupclientoptions"
	eventingTest "knative.dev/eventing/test"
	testlib "knative.dev/eventing/test/lib"
	"knative.dev/pkg/system"
	"knative.dev/pkg/test/zipkin"
)

const (
	redisAddress        = "rediss://redis.redis.svc.cluster.local:6379"
	recordEventsPodName = "redis-source-logger-pod"
)

var sourcesTestRunner testlib.ComponentsTestRunner

func TestMain(m *testing.M) {
	os.Exit(func() int {
		eventingTest.InitializeEventingFlags()
		sourcesTestRunner = testlib.ComponentsTestRunner{
			ComponentFeatureMap: test.SourcesFeatureMap,
			ComponentsToTest:    eventingTest.EventingFlags.Sources,
		}

		addSourcesInitializers()

		// Any tests may SetupZipkinTracing, it will only actually be done once. This should be the ONLY
		// place that cleans it up. If an individual test calls this instead, then it will break other
		// tests that need the tracing in place.
		defer zipkin.CleanupZipkinTracingSetup(log.Printf)
		defer testlib.ExportLogs(testlib.SystemLogsDir, system.Namespace())

		return m.Run()
	}())
}

func addSourcesInitializers() {
	sourcesTestRunner.AddComponentSetupClientOption(test.RedisStreamSourceTypeMeta,
		setupclientoptions.RedisStreamSourceV1Alpha1ClientSetupOption("redisstreamsource",
			redisAddress,
			recordEventsPodName,
		))
}
