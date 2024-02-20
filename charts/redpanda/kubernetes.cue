package redpanda

import rbacv1 "k8s.io/api/rbac/v1"

#ClusterRole: rbacv1.#ClusterRole & {
		apiVersion: "rbac.authorization.k8s.io/v1"
		kind:       "ClusterRole"
}
