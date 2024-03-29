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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	v1alpha1 "knative.dev/eventing-redis/pkg/source/apis/sources/v1alpha1"
)

// RedisStreamSourceLister helps list RedisStreamSources.
// All objects returned here must be treated as read-only.
type RedisStreamSourceLister interface {
	// List lists all RedisStreamSources in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.RedisStreamSource, err error)
	// RedisStreamSources returns an object that can list and get RedisStreamSources.
	RedisStreamSources(namespace string) RedisStreamSourceNamespaceLister
	RedisStreamSourceListerExpansion
}

// redisStreamSourceLister implements the RedisStreamSourceLister interface.
type redisStreamSourceLister struct {
	indexer cache.Indexer
}

// NewRedisStreamSourceLister returns a new RedisStreamSourceLister.
func NewRedisStreamSourceLister(indexer cache.Indexer) RedisStreamSourceLister {
	return &redisStreamSourceLister{indexer: indexer}
}

// List lists all RedisStreamSources in the indexer.
func (s *redisStreamSourceLister) List(selector labels.Selector) (ret []*v1alpha1.RedisStreamSource, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.RedisStreamSource))
	})
	return ret, err
}

// RedisStreamSources returns an object that can list and get RedisStreamSources.
func (s *redisStreamSourceLister) RedisStreamSources(namespace string) RedisStreamSourceNamespaceLister {
	return redisStreamSourceNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// RedisStreamSourceNamespaceLister helps list and get RedisStreamSources.
// All objects returned here must be treated as read-only.
type RedisStreamSourceNamespaceLister interface {
	// List lists all RedisStreamSources in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.RedisStreamSource, err error)
	// Get retrieves the RedisStreamSource from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.RedisStreamSource, error)
	RedisStreamSourceNamespaceListerExpansion
}

// redisStreamSourceNamespaceLister implements the RedisStreamSourceNamespaceLister
// interface.
type redisStreamSourceNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all RedisStreamSources in the indexer for a given namespace.
func (s redisStreamSourceNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.RedisStreamSource, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.RedisStreamSource))
	})
	return ret, err
}

// Get retrieves the RedisStreamSource from the indexer for a given namespace and name.
func (s redisStreamSourceNamespaceLister) Get(name string) (*v1alpha1.RedisStreamSource, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("redisstreamsource"), name)
	}
	return obj.(*v1alpha1.RedisStreamSource), nil
}
