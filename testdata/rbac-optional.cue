import rbacv1 "k8s.io/api/rbac/v1"

#_def: {
	in: {enabled: bool} @template("context")

	if in.enabled {
		out: rbacv1.#ClusterRole @template("toYaml")
		out: metadata: name: "hello world"
	}
}

// {{if $in.enabled}}
//  {{ (dict "metadata" (dict "name" "hello world")) | toYaml }}
// {{end}}

inputs: [{enabled: true}, {enabled: false}]
outputs: [
	for i in inputs
	{(#_def & {in: i}).out }
]
