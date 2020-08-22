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

package resources

import (
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"

	"knative.dev/eventing-redis/source/pkg/apis/sources/v1alpha1"
	sourcesv1alpha1 "knative.dev/eventing-redis/source/pkg/apis/sources/v1alpha1"
)

func RoleBindingName(source *sourcesv1alpha1.RedisStreamSource) string {
	return kmeta.ChildName(fmt.Sprintf("redistreamsource-%s-", source.Name), string(source.UID))
}

// MakeRoleBinding creates a RoleBinding object for the single-tenant receive adapter
// service account 'sa' in the Namespace 'ns'.
func MakeRoleBinding(source *v1alpha1.RedisStreamSource, clusterRoleName string) *rbacv1.RoleBinding {
	name := RoleBindingName(source)
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: source.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(source),
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     clusterRoleName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Namespace: source.Namespace,
				Name:      name,
			},
		},
	}
}
