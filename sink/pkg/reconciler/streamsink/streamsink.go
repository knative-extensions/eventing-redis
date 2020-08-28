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
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	pkgreconciler "knative.dev/pkg/reconciler"

	eventingresources "knative.dev/eventing/pkg/reconciler/resources"
	reconcilersource "knative.dev/eventing/pkg/reconciler/source"

	"knative.dev/eventing-redis/pkg/reconciler"
	sinksv1alpha1 "knative.dev/eventing-redis/sink/pkg/apis/sinks/v1alpha1"
	streamsinkreconciler "knative.dev/eventing-redis/sink/pkg/client/injection/reconciler/sinks/v1alpha1/redisstreamsink"
	"knative.dev/eventing-redis/sink/pkg/reconciler/streamsink/resources"
)

const (
	component               = "redisstreamsink"
	receiverClusterRoleName = "knative-sinks-redisstream-receiver"
)

func newWarningSinkNotFound(sink *duckv1.Destination) pkgreconciler.Event {
	b, _ := json.Marshal(sink)
	return pkgreconciler.NewEvent(corev1.EventTypeWarning, "SinkNotFound", "Sink not found: %s", string(b))
}

// Reconciler reconciles a streamsink object
type Reconciler struct {
	kubeClientSet kubernetes.Interface

	ksr           *reconciler.KnativeServiceReconciler
	rbr           *reconciler.RoleBindingReconciler
	sar           *reconciler.ServiceAccountReconciler
	receiverImage string
	configs       reconcilersource.ConfigAccessor
}

var _ streamsinkreconciler.Interface = (*Reconciler)(nil)

func (r *Reconciler) ReconcileKind(ctx context.Context, sink *sinksv1alpha1.RedisStreamSink) pkgreconciler.Event {
	sink.Annotations = nil

	expectedServiceAccount := eventingresources.MakeServiceAccount(sink, resources.ServiceAccountName(sink))
	sa, event := r.sar.ReconcileServiceAccount(ctx, sink, expectedServiceAccount)
	if sa == nil {
		sink.Status.MarkNoServiceAccount(event.Error())
		return event
	}

	expectedRoleBinding := resources.MakeRoleBinding(sink, resources.RoleBindingName(sink), receiverClusterRoleName)
	rb, event := r.rbr.ReconcileRoleBinding(ctx, sink, expectedRoleBinding)
	if rb == nil {
		sink.Status.MarkNoRoleBinding(event.Error())
		return event
	}

	expectedKService := resources.MakeReceiver(sink, r.receiverImage)
	ra, event := r.ksr.ReconcileService(ctx, sink, expectedKService)
	if ra == nil {
		sink.Status.MarkNoKnativeService(event.Error())
		return event
	}

	if !sink.Status.PropagateKnativeServiceAddress(ra) {
		return nil // no need to retry since the controller tracks it.
	}

	return nil
}
