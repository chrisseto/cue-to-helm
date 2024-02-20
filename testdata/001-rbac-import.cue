import rbacv1 "k8s.io/api/rbac/v1"

#_def: {
	in: {name: string} @template("context")

	out: rbacv1.#ClusterRole & {
		apiVersion: "rbac.authorization.k8s.io/v1"
		kind:       "ClusterRole"
		metadata: {
			name: "pre-\(in.name)-inter-\(in.name)-post"
		}
	} @template("toYaml")
}

inputs: [{name: ""}, {name: "foo"}]
outputs: [
	for i in inputs
	{(#_def & {in: i}).out }
]
