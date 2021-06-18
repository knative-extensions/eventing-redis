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

package streamsink

import (
	"context"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/system"

	serviceclient "knative.dev/serving/pkg/client/injection/client"
	kserviceinformer "knative.dev/serving/pkg/client/injection/informers/serving/v1/service"

	reconcilersource "knative.dev/eventing/pkg/reconciler/source"

	"knative.dev/eventing-redis/pkg/reconciler"
	"knative.dev/eventing-redis/sink/pkg/apis/sinks/v1alpha1"
	redisstreamsinkinformer "knative.dev/eventing-redis/sink/pkg/client/injection/informers/sinks/v1alpha1/redisstreamsink"
	redisstreamssinkreconciler "knative.dev/eventing-redis/sink/pkg/client/injection/reconciler/sinks/v1alpha1/redisstreamsink"
)

// envConfig will be used to extract the required environment variables using
// github.com/kelseyhightower/envconfig. If this configuration cannot be extracted, then
// NewController will panic.
type envConfig struct {
	Image string `envconfig:"STREAMSINK_RA_IMAGE" required:"true"`
}

// NewController initializes the controller and is called by the generated code
// Registers event handlers to enqueue events
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	env := &envConfig{}
	if err := envconfig.Process("", env); err != nil {
		logging.FromContext(ctx).Panicf("unable to processRedisStreamSink's required environment variables: %v", err)
	}

	kserviceInformer := kserviceinformer.Get(ctx)
	redisstreamSinkInformer := redisstreamsinkinformer.Get(ctx)

	r := &Reconciler{
		kubeClientSet: kubeclient.Get(ctx),
		ksr:           &reconciler.KnativeServiceReconciler{ServingClientSet: serviceclient.Get(ctx)},
		rbr:           &reconciler.RoleBindingReconciler{KubeClientSet: kubeclient.Get(ctx)},
		sar:           &reconciler.ServiceAccountReconciler{KubeClientSet: kubeclient.Get(ctx)},
		configs:       reconcilersource.WatchConfigurations(ctx, component, cmw),
		receiverImage: env.Image,
	}

	impl := redisstreamssinkreconciler.NewImpl(ctx, r)

	// Get TLS secret and set TLS certificate, to pass data to receiver.
	// Not rolling out new adapters on watch change.
	if secret, err := kubeclient.Get(ctx).CoreV1().Secrets(system.Namespace()).Get(ctx, TLSSecretName(), metav1.GetOptions{}); err == nil {

		r.updateTLSSecret(ctx, secret)

	} else if !apierrors.IsNotFound(err) {
		logging.FromContext(ctx).With(zap.Error(err)).Info("Error reading TLS Secret'")
	}

	logging.FromContext(ctx).Info("Setting up event handlers")

	redisstreamSinkInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	kserviceInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterControllerGK(v1alpha1.Kind("RedisStreamSink")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	return impl
}
