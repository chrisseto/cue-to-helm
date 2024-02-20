# Cue to Helm

Experiments in transpiling [cue](http://cuelang.org/) into functioning helm charts.

Run `go test . -v` to see the example inputs and outputs.

`charts/` contains attempts of translating helm charts into cue files.


## Notes

Transpilation is done by walking the Cue AST that's output from
`cue.Value.Syntax`. This ensures that all of cue's syntax is resolved to an
easily parsable/walkable form.

`cue.Value`'s can't be walked because they don't appear to provide a way
discover `if`'s or `lets` within the struct.

Rather than trying to output a valid YAML template, the strategy is to build
out the Kubernetes Objects using nested calls to the `dict` helper and end by
piping the output to `toYaml`. Dealing with the whitespace of YAML (especially
dealing with lists that might be empty) is too much of a headache.

Transpilation is done from cue AST to template AST. The template ASTs should
eventually be replaced with something bespoke; there's little to no benefit to
using the go template AST as they don't round trip.


## Example

Input Cue:
```cue
import corev1 "k8s.io/api/core/v1"

_#def
_#def: {
        in: {
                name: string
                ports: [...int]
        } @template("context")
        out: corev1.#Service & {
                metadata: {
                        name: "\(in.name)"
                }
                spec: {
                        ports: [
                                for p in in.ports {
                                        port: p
                                }]
                }
        } @template("toYaml")
}
```

Output Template:
```gotemplate
{{$in := .}}

{{$tmp_0 := (list)}}
{{range $_, $p := $in.ports}}
    {{- $tmp_0 = (append $tmp_0 (dict "port" $p)) -}}
{{end}}

{{$out := (dict "metadata" (dict "name" (printf "%s" $in.name)) "spec" (dict "ports" $tmp_0))}}

{{$out | toYaml}}
```
