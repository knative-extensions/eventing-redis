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

package test

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testlib "knative.dev/eventing/test/lib"
)

var RedisStreamSourceTypeMeta = metav1.TypeMeta{
	APIVersion: "sources.knative.dev/v1alpha1",
	Kind:       RedisStreamSourceKind,
}

var RedisStreamSinkTypeMeta = metav1.TypeMeta{
	APIVersion: "sinks.knative.dev/v1alpha1",
	Kind:       RedisStreamSinkKind,
}

var SourcesFeatureMap = map[metav1.TypeMeta][]testlib.Feature{
	RedisStreamSourceTypeMeta: {testlib.FeatureBasic, testlib.FeatureLongLiving},
	RedisStreamSinkTypeMeta:   {testlib.FeatureBasic, testlib.FeatureLongLiving},
}
