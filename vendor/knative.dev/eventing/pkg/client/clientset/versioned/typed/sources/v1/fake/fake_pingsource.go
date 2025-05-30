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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	gentype "k8s.io/client-go/gentype"
	v1 "knative.dev/eventing/pkg/apis/sources/v1"
	sourcesv1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/sources/v1"
)

// fakePingSources implements PingSourceInterface
type fakePingSources struct {
	*gentype.FakeClientWithList[*v1.PingSource, *v1.PingSourceList]
	Fake *FakeSourcesV1
}

func newFakePingSources(fake *FakeSourcesV1, namespace string) sourcesv1.PingSourceInterface {
	return &fakePingSources{
		gentype.NewFakeClientWithList[*v1.PingSource, *v1.PingSourceList](
			fake.Fake,
			namespace,
			v1.SchemeGroupVersion.WithResource("pingsources"),
			v1.SchemeGroupVersion.WithKind("PingSource"),
			func() *v1.PingSource { return &v1.PingSource{} },
			func() *v1.PingSourceList { return &v1.PingSourceList{} },
			func(dst, src *v1.PingSourceList) { dst.ListMeta = src.ListMeta },
			func(list *v1.PingSourceList) []*v1.PingSource { return gentype.ToPointerSlice(list.Items) },
			func(list *v1.PingSourceList, items []*v1.PingSource) { list.Items = gentype.FromPointerSlice(items) },
		),
		fake,
	}
}
