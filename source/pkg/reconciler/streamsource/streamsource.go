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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	duckv1 "knative.dev/pkg/apis/duck/v1"
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

func newWarningSinkNotFound(sink *duckv1.Destination) pkgreconciler.Event {
	b, _ := json.Marshal(sink)
	return pkgreconciler.NewEvent(corev1.EventTypeWarning, "SinkNotFound", "Sink not found: %s", string(b))
}

// Reconciler reconciles a streamsource object
type Reconciler struct {
	kubeClientSet       kubernetes.Interface
	dr                  *reconciler.StatefulSetReconciler
	rbr                 *reconciler.RoleBindingReconciler
	sar                 *reconciler.ServiceAccountReconciler
	receiveAdapterImage string
	ceSource            string
	sinkResolver        *resolver.URIResolver
	configs             reconcilersource.ConfigAccessor
}

var _ streamsourcereconciler.Interface = (*Reconciler)(nil)

func (r *Reconciler) ReconcileKind(ctx context.Context, source *sourcesv1alpha1.RedisStreamSource) pkgreconciler.Event {
	source.Annotations = nil

	dest := source.Spec.Sink.DeepCopy()
	if dest.Ref != nil {
		if dest.Ref.Namespace == "" {
			dest.Ref.Namespace = source.GetNamespace()
		}
	}

	sinkURI, err := r.sinkResolver.URIFromDestinationV1(*dest, source)
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

	expectedStatefulSet := resources.MakeReceiveAdapter(source, r.receiveAdapterImage, sinkURI.String())
	ra, event := r.dr.ReconcileStatefulSet(ctx, source, expectedStatefulSet)
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
