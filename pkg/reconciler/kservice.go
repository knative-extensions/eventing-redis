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
package reconciler

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/logging"
	pkgreconciler "knative.dev/pkg/reconciler"

	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	servingclientset "knative.dev/serving/pkg/client/clientset/versioned"
)

// newKnativeServiceCreated makes a new reconciler event with event type Normal, and
// reason ServiceCreated.
func newKnativeServiceCreated(namespace, name string) pkgreconciler.Event {
	return pkgreconciler.NewEvent(corev1.EventTypeNormal, "KnativeServiceCreated", "created service: \"%s/%s\"", namespace, name)
}

// newKnativeServiceFailed makes a new reconciler event with event type Warning, and
// reason ServiceFailed.
func newKnativeServiceFailed(namespace, name string, err error) pkgreconciler.Event {
	return pkgreconciler.NewEvent(corev1.EventTypeWarning, "ServiceFailed", "failed to create service: \"%s/%s\", %w", namespace, name, err)
}

type KnativeServiceReconciler struct {
	ServingClientSet servingclientset.Interface
}

func (r *KnativeServiceReconciler) ReconcileService(ctx context.Context, owner kmeta.OwnerRefable, expected *servingv1.Service) (*servingv1.Service, error) {
	svc, err := r.ServingClientSet.ServingV1().Services(expected.Namespace).Get(expected.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		svc, err := r.ServingClientSet.ServingV1().Services(expected.Namespace).Create(expected)
		if err != nil {
			return nil, newKnativeServiceFailed(expected.Namespace, expected.Name, err)
		}
		return svc, newKnativeServiceCreated(expected.Namespace, expected.Name)
	} else if err != nil {
		return nil, fmt.Errorf("error getting Knative service %q: %v", expected.Name, err)
	} else if !metav1.IsControlledBy(svc, owner.GetObjectMeta()) {
		return nil, fmt.Errorf("Knative service %q is not owned by %s %q",
			svc.Name, owner.GetGroupVersionKind().Kind, owner.GetObjectMeta().GetName())
	} else {
		logging.FromContext(ctx).Debugw("reusing existing Knative service", zap.Any("knativeService", svc))
	}
	return svc, nil
}
