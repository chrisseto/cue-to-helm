import rbacv1 "k8s.io/api/rbac/v1"

#_def: {
	in: { name: string } @template("context")

	out: rbacv1.#ClusterRole @template("toYaml")
	out: kind: "ClusterRole"
	out: apiVersion: "rbac.authorization.k8s.io/v1"
	out: metadata: name: "\(in.name)"
}

inputs: [{name: "hello"}, {name: "world"}]
outputs: [
	for i in inputs
	{(#_def & {in: i}).out }
]
