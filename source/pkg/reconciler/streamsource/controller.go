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

package streamsource

import (
	"context"

	"github.com/kelseyhightower/envconfig"
	"k8s.io/client-go/tools/cache"

	"go.uber.org/zap"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	reconcilersource "knative.dev/eventing/pkg/reconciler/source"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	statefulsetinformer "knative.dev/pkg/client/injection/kube/informers/apps/v1/statefulset"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/resolver"
	"knative.dev/pkg/system"

	"knative.dev/eventing-redis/source/pkg/apis/sources/v1alpha1"
	redisstreamsourceinformer "knative.dev/eventing-redis/source/pkg/client/injection/informers/sources/v1alpha1/redisstreamsource"
	redisstreamsourcereconciler "knative.dev/eventing-redis/source/pkg/client/injection/reconciler/sources/v1alpha1/redisstreamsource"
	"knative.dev/eventing-redis/source/pkg/reconciler"
)

// envConfig will be used to extract the required environment variables using
// github.com/kelseyhightower/envconfig. If this configuration cannot be extracted, then
// NewController will panic.
type envConfig struct {
	Image string `envconfig:"STREAMSOURCE_RA_IMAGE" required:"true"`
}

//Removing watch on Redis config and not reloading numConsumers from CM dynamically since we don't automatically rollout new adapters on watch change. Will scale adapters via replicas

// NewController initializes the controller and is called by the generated code
// Registers event handlers to enqueue events
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	env := &envConfig{}
	if err := envconfig.Process("", env); err != nil {
		logging.FromContext(ctx).Panicf("unable to processRedisStreamSource's required environment variables: %v", err)
	}

	statefulsetInformer := statefulsetinformer.Get(ctx)
	redisstreamSourceInformer := redisstreamsourceinformer.Get(ctx)

	r := &Reconciler{
		kubeClientSet:       kubeclient.Get(ctx),
		ssr:                 &reconciler.StatefulSetReconciler{KubeClientSet: kubeclient.Get(ctx)},
		rbr:                 &reconciler.RoleBindingReconciler{KubeClientSet: kubeclient.Get(ctx)},
		sar:                 &reconciler.ServiceAccountReconciler{KubeClientSet: kubeclient.Get(ctx)},
		configs:             reconcilersource.WatchConfigurations(ctx, component, cmw),
		receiveAdapterImage: env.Image,
	}

	impl := redisstreamsourcereconciler.NewImpl(ctx, r)

	r.sinkResolver = resolver.NewURIResolver(ctx, impl.EnqueueKey)

	// Get Redis config map and set Redis configuration, to pass data to receive adapter.
	// Not rolling out new adapters on watch change. Will scale adapters via replicas at a later time.
	if _, err := kubeclient.Get(ctx).CoreV1().ConfigMaps(system.Namespace()).Get(ctx, ConfigMapName(), metav1.GetOptions{}); err == nil {
		cmw.Watch(ConfigMapName(), func(configMap *v1.ConfigMap) {
			r.updateRedisConfig(ctx, configMap)
		})
	} else if !apierrors.IsNotFound(err) {
		logging.FromContext(ctx).With(zap.Error(err)).Info("Error reading Redis ConfigMap'")
	}

	// Get TLS config map and set TLS certificate, to pass data to receive adapter.
	// Not rolling out new adapters on watch change.
	if _, err := kubeclient.Get(ctx).CoreV1().ConfigMaps(system.Namespace()).Get(ctx, TLSConfigMapName(), metav1.GetOptions{}); err == nil {
		cmw.Watch(TLSConfigMapName(), func(configMap *v1.ConfigMap) {
			r.updateTLSConfig(ctx, configMap)
		})
	} else if !apierrors.IsNotFound(err) {
		logging.FromContext(ctx).With(zap.Error(err)).Info("Error reading TLS ConfigMap'")
	}

	logging.FromContext(ctx).Info("Setting up event handlers")

	redisstreamSourceInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	statefulsetInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterControllerGK(v1alpha1.Kind("RedisStreamSource")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	return impl
}
