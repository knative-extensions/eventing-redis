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
	"encoding/json"

	"go.uber.org/zap"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/logging"
	pkgreconciler "knative.dev/pkg/reconciler"
	"knative.dev/pkg/resolver"

	eventingresources "knative.dev/eventing/pkg/reconciler/resources"
	reconcilersource "knative.dev/eventing/pkg/reconciler/source"

	sourcesv1alpha1 "knative.dev/eventing-redis/source/pkg/apis/sources/v1alpha1"
	streamsourcereconciler "knative.dev/eventing-redis/source/pkg/client/injection/reconciler/sources/v1alpha1/redisstreamsource"
	"knative.dev/eventing-redis/source/pkg/reconciler"
	"knative.dev/eventing-redis/source/pkg/reconciler/streamsource/resources"
)

const (
	component              = "redisstreamsource"
	adapterClusterRoleName = "knative-sources-redisstream-adapter"
)

func newFinalizedNormal(namespace, name string) pkgreconciler.Event {
	return pkgreconciler.NewEvent(corev1.EventTypeNormal, "RedisStreamSourceFinalized", "RedisStreamSource finalized: \"%s/%s\"", namespace, name)
}

func newWarningSinkNotFound(sink *duckv1.Destination) pkgreconciler.Event {
	b, _ := json.Marshal(sink)
	return pkgreconciler.NewEvent(corev1.EventTypeWarning, "SinkNotFound", "Sink not found: %s", string(b))
}

// Reconciler reconciles a streamsource object
type Reconciler struct {
	kubeClientSet       kubernetes.Interface
	ssr                 *reconciler.StatefulSetReconciler
	rbr                 *reconciler.RoleBindingReconciler
	sar                 *reconciler.ServiceAccountReconciler
	receiveAdapterImage string
	ceSource            string
	sinkResolver        *resolver.URIResolver
	configs             reconcilersource.ConfigAccessor
	numConsumers        string
	tlsCert             string
}

// Check that our Reconciler implements ReconcileKind.
var _ streamsourcereconciler.Interface = (*Reconciler)(nil)

// Check that our Reconciler implements FinalizeKind.
var _ streamsourcereconciler.Finalizer = (*Reconciler)(nil)

func (r *Reconciler) ReconcileKind(ctx context.Context, source *sourcesv1alpha1.RedisStreamSource) pkgreconciler.Event {
	source.Annotations = nil

	dest := source.Spec.Sink.DeepCopy()
	if dest.Ref != nil {
		if dest.Ref.Namespace == "" {
			dest.Ref.Namespace = source.GetNamespace()
		}
	}

	sinkURI, err := r.sinkResolver.URIFromDestinationV1(ctx, *dest, source)
	if err != nil {
		source.Status.MarkNoSink("NotFound", "")
		return newWarningSinkNotFound(dest)
	}
	source.Status.MarkSink(sinkURI.String())

	expectedServiceAccount := eventingresources.MakeServiceAccount(source, resources.ServiceAccountName(source))
	sa, event := r.sar.ReconcileServiceAccount(ctx, source, expectedServiceAccount)
	if sa == nil {
		source.Status.MarkNoServiceAccount(event.Error())
		return event
	}

	expectedRoleBinding := resources.MakeRoleBinding(source, adapterClusterRoleName)
	rb, event := r.rbr.ReconcileRoleBinding(ctx, source, expectedRoleBinding)
	if rb == nil {
		source.Status.MarkNoRoleBinding(event.Error())
		return event
	}

	expectedStatefulSet := resources.MakeReceiveAdapter(source, r.receiveAdapterImage, sinkURI.String(), r.numConsumers, r.tlsCert)
	ra, event := r.ssr.ReconcileStatefulSet(ctx, source, expectedStatefulSet)
	if ra == nil {
		if source.Status.Annotations == nil {
			source.Status.Annotations = make(map[string]string)
		}
		source.Status.Annotations["StatefulSet"] = event.Error()
		return event
	}
	source.Status.PropagateStatefulSetAvailability(ra)

	return nil
}

func (r *Reconciler) FinalizeKind(ctx context.Context, source *sourcesv1alpha1.RedisStreamSource) pkgreconciler.Event {
	//Nothing to do since adapter will gracefully shutdown the consumers
	return nil //ok to remove finalizer
}

func (r *Reconciler) updateRedisConfig(ctx context.Context, configMap *corev1.ConfigMap) {
	logging.FromContext(ctx).Info("Reloading Redis configuration")
	redisConfig, err := GetRedisConfig(configMap.Data)
	if err != nil {
		logging.FromContext(ctx).Errorw("Error reading Redis configuration", zap.Error(err))
	}
	// For now just override the previous config.
	r.numConsumers = redisConfig.NumConsumers
}

func (r *Reconciler) updateTLSConfig(ctx context.Context, configMap *corev1.ConfigMap) {
	tlsConfig, err := GetTLSConfig(configMap.Data)
	if err != nil {
		logging.FromContext(ctx).Errorw("Error reading TLS configuration", zap.Error(err))
	}
	// For now just override the previous config.
	r.tlsCert = tlsConfig.TLSCertificate
}
