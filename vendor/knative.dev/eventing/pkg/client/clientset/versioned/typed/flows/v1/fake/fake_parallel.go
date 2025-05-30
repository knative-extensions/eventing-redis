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
	v1 "knative.dev/eventing/pkg/apis/flows/v1"
	flowsv1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/flows/v1"
)

// fakeParallels implements ParallelInterface
type fakeParallels struct {
	*gentype.FakeClientWithList[*v1.Parallel, *v1.ParallelList]
	Fake *FakeFlowsV1
}

func newFakeParallels(fake *FakeFlowsV1, namespace string) flowsv1.ParallelInterface {
	return &fakeParallels{
		gentype.NewFakeClientWithList[*v1.Parallel, *v1.ParallelList](
			fake.Fake,
			namespace,
			v1.SchemeGroupVersion.WithResource("parallels"),
			v1.SchemeGroupVersion.WithKind("Parallel"),
			func() *v1.Parallel { return &v1.Parallel{} },
			func() *v1.ParallelList { return &v1.ParallelList{} },
			func(dst, src *v1.ParallelList) { dst.ListMeta = src.ListMeta },
			func(list *v1.ParallelList) []*v1.Parallel { return gentype.ToPointerSlice(list.Items) },
			func(list *v1.ParallelList, items []*v1.Parallel) { list.Items = gentype.FromPointerSlice(items) },
		),
		fake,
	}
}
