import rbacv1 "k8s.io/api/rbac/v1"

#_def: {
	in: {} @template("context")

	out: rbacv1.#ClusterRole & {
		kind:       "ClusterRole"
		apiVersion: "rbac.authorization.k8s.io/v1"
		metadata: name: "hello world"
		rules: [{
			apiGroups: [""]
			resources: ["nodes"]
			verbs: ["get", "list"]
		}]
	} @template("toYaml")
}

inputs: [{}]
outputs: [
	for i in inputs
	{(#_def & {in: i}).out }
]
