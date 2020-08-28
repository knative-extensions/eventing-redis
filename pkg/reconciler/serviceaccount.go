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
	"k8s.io/client-go/kubernetes"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/logging"
	pkgreconciler "knative.dev/pkg/reconciler"
)

// newServiceAccountCreated makes a new reconciler event with event type Normal, and
// reason ServiceAccountCreated.
func newServiceAccountCreated(namespace, name string) pkgreconciler.Event {
	return pkgreconciler.NewEvent(corev1.EventTypeNormal, "ServiceAccountCreated", "created service account: \"%s/%s\"", namespace, name)
}

// newServiceAccountFailed makes a new reconciler event with event type Warning, and
// reason ServiceAccountFailed.
func newServiceAccountFailed(namespace, name string, err error) pkgreconciler.Event {
	return pkgreconciler.NewEvent(corev1.EventTypeWarning, "ServiceAccountFailed", "failed to create service account: \"%s/%s\", %w", namespace, name, err)
}

type ServiceAccountReconciler struct {
	KubeClientSet kubernetes.Interface
}

func (r *ServiceAccountReconciler) ReconcileServiceAccount(ctx context.Context, owner kmeta.OwnerRefable, expected *corev1.ServiceAccount) (*corev1.ServiceAccount, error) {
	rb, err := r.KubeClientSet.CoreV1().ServiceAccounts(expected.Namespace).Get(expected.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		rb, err := r.KubeClientSet.CoreV1().ServiceAccounts(expected.Namespace).Create(expected)
		if err != nil {
			return nil, newServiceAccountFailed(expected.Namespace, expected.Name, err)
		}
		return rb, newServiceAccountCreated(expected.Namespace, expected.Name)
	} else if err != nil {
		return nil, fmt.Errorf("error getting service account  %q: %v", expected.Name, err)
	} else if !metav1.IsControlledBy(rb, owner.GetObjectMeta()) {
		return nil, fmt.Errorf("service account %q is not owned by %s %q",
			rb.Name, owner.GetGroupVersionKind().Kind, owner.GetObjectMeta().GetName())
	} else {
		logging.FromContext(ctx).Debugw("Reusing existing service account", zap.Any("serviceAccount", rb))
	}
	return rb, nil
}
