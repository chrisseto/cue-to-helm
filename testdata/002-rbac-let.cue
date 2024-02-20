import rbacv1 "k8s.io/api/rbac/v1"

#_def: {

	in: { name: string } @template("context")

	let release_name = "\(in.name)-release"

	out: rbacv1.#ClusterRole @template("toYaml")
	out: kind: "ClusterRole"
	out: apiVersion: "rbac.authorization.k8s.io/v1"
	out: metadata: name: release_name
}

inputs: [{name: ""}, {name: "foo"}]
outputs: [
	for i in inputs
	{(#_def & {in: i}).out }
]
