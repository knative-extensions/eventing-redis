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

package lib

import (
	"k8s.io/apimachinery/pkg/runtime"
	fakekubeclientset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
	fakeeventingclientset "knative.dev/eventing/pkg/client/clientset/versioned/fake"
	"knative.dev/pkg/reconciler/testing"

	redissourcev1alpha1 "knative.dev/eventing-redis/pkg/source/apis/sources/v1alpha1"
	fakeredissourceclientset "knative.dev/eventing-redis/pkg/source/client/clientset/versioned/fake"
	redissourcev1alpha1listers "knative.dev/eventing-redis/pkg/source/client/listers/sources/v1alpha1"
)

var clientSetSchemes = []func(*runtime.Scheme) error{
	fakekubeclientset.AddToScheme,
	fakeeventingclientset.AddToScheme,
	fakeredissourceclientset.AddToScheme,
}

type Listers struct {
	sorter testing.ObjectSorter
}

func NewScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()

	for _, addTo := range clientSetSchemes {
		addTo(scheme)
	}
	return scheme
}

func NewListers(objs []runtime.Object) Listers {
	scheme := runtime.NewScheme()

	for _, addTo := range clientSetSchemes {
		addTo(scheme)
	}

	ls := Listers{
		sorter: testing.NewObjectSorter(scheme),
	}

	ls.sorter.AddObjects(objs...)

	return ls
}

func (l Listers) indexerFor(obj runtime.Object) cache.Indexer {
	return l.sorter.IndexerForObjectType(obj)
}

func (l Listers) GetKubeObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakekubeclientset.AddToScheme)
}

func (l Listers) GetEventingObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakeeventingclientset.AddToScheme)
}

func (l Listers) GetAllObjects() []runtime.Object {
	all := l.GetEventingObjects()
	all = append(all, l.GetKubeObjects()...)
	return all
}

func (l Listers) GetRedisSourceLister() redissourcev1alpha1listers.RedisStreamSourceLister {
	return redissourcev1alpha1listers.NewRedisStreamSourceLister(l.indexerFor(&redissourcev1alpha1.RedisStreamSource{}))
}
