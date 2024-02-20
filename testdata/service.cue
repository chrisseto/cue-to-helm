import corev1 "k8s.io/api/core/v1"

#_def: {
	in: {name: string, ports: [...int]} @template("context")

	out: corev1.#Service & {
		metadata: name: "\(in.name)"

		spec: ports: [
			for p in in.ports
			{port: p}
		]
	} @template("toYaml")
}

inputs: [
	{name: "hello"},
	{name: "world", ports: []},
	{name: "world", ports: [123]},
]
outputs: [
	for i in inputs
	{(#_def & {in: i}).out }
]
