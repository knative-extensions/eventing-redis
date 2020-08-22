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

	"k8s.io/apimachinery/pkg/api/equality"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/logging"
	pkgreconciler "knative.dev/pkg/reconciler"
)

// newDeploymentCreated makes a new reconciler event with event type Normal, and
// reason DeploymentCreated.
func newDeploymentCreated(namespace, name string) pkgreconciler.Event {
	return pkgreconciler.NewEvent(corev1.EventTypeNormal, "DeploymentCreated", "created deployment: \"%s/%s\"", namespace, name)
}

// newDeploymentFailed makes a new reconciler event with event type Warning, and
// reason DeploymentFailed.
func newDeploymentFailed(namespace, name string, err error) pkgreconciler.Event {
	return pkgreconciler.NewEvent(corev1.EventTypeWarning, "DeploymentFailed", "failed to create deployment: \"%s/%s\", %w", namespace, name, err)
}

// newDeploymentUpdated makes a new reconciler event with event type Normal, and
// reason DeploymentUpdated.
func newDeploymentUpdated(namespace, name string) pkgreconciler.Event {
	return pkgreconciler.NewEvent(corev1.EventTypeNormal, "DeploymentUpdated", "updated deployment: \"%s/%s\"", namespace, name)
}

type DeploymentReconciler struct {
	KubeClientSet kubernetes.Interface
}

func (r *DeploymentReconciler) ReconcileDeployment(ctx context.Context, owner kmeta.OwnerRefable, expected *appsv1.Deployment) (*appsv1.Deployment, pkgreconciler.Event) {
	namespace := owner.GetObjectMeta().GetNamespace()
	ra, err := r.KubeClientSet.AppsV1().Deployments(namespace).Get(expected.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		ra, err = r.KubeClientSet.AppsV1().Deployments(namespace).Create(expected)
		if err != nil {
			return nil, newDeploymentFailed(expected.Namespace, expected.Name, err)
		}
		return ra, newDeploymentCreated(ra.Namespace, ra.Name)
	} else if err != nil {
		return nil, fmt.Errorf("error getting deployment %q: %v", expected.Name, err)
	} else if !metav1.IsControlledBy(ra, owner.GetObjectMeta()) {
		return nil, fmt.Errorf("deployment %q is not owned by %s %q",
			ra.Name, owner.GetGroupVersionKind().Kind, owner.GetObjectMeta().GetName())
	} else if r.podSpecChanged(expected.Spec.Template.Spec, ra.Spec.Template.Spec) {
		if ra, err = r.KubeClientSet.AppsV1().Deployments(namespace).Update(ra); err != nil {
			return ra, err
		}
		return ra, newDeploymentUpdated(ra.Namespace, ra.Name)
	} else {
		logging.FromContext(ctx).Debugw("Reusing existing receive adapter", zap.Any("receiveAdapter", ra))
	}
	return ra, nil
}

// Returns false if an update is needed.
func (r *DeploymentReconciler) podSpecChanged(expected corev1.PodSpec, now corev1.PodSpec) bool {
	return !equality.Semantic.DeepDerivative(expected, now)
}
