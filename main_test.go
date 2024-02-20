package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/format"
	"cuelang.org/go/cue/token"
	cuejson "cuelang.org/go/pkg/encoding/json"
	"github.com/Masterminds/sprig/v3"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
)

func TestInterpolation(t *testing.T) {
	tcs := []struct {
		In  []ast.Expr
		Out string
	}{
		{
			In: []ast.Expr{
				&ast.BasicLit{Kind: token.STRING, Value: `"\(`},
				&ast.Ident{Name: "foo"},
				&ast.BasicLit{Kind: token.STRING, Value: `)"`},
			},
			Out: `printf "%s" $foo`,
		},
		{
			In: []ast.Expr{
				&ast.BasicLit{Kind: token.STRING, Value: `"pre-\(`},
				&ast.Ident{Name: "first"},
				&ast.BasicLit{Kind: token.STRING, Value: `)-inter-\(`},
				&ast.Ident{Name: "second"},
				&ast.BasicLit{Kind: token.STRING, Value: `)-post"`},
			},
			Out: `printf "pre-%s-inter-%s-post" $first $second`,
		},
	}

	for _, tc := range tcs {
		require.Equal(
			t,
			tc.Out,
			ToPrintf(&ast.Interpolation{Elts: tc.In}).String(),
		)
	}
}

type TestCase struct {
	Inputs    []map[string]any
	Outputs   []map[string]any
	Templates []ast.Node
}

func cueToAny(t *testing.T, c cue.Value) []map[string]any {
	out, err := cuejson.Marshal(c)
	require.NoError(t, err)

	var input []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &input))
	return input
}

func LoadTestCase(t *testing.T, path string) TestCase {
	root := LoadCueValue(path)
	require.NoError(t, root.Err())

	inputField, err := root.FieldByName("inputs", false)
	require.NoError(t, err, "missing inputs")

	outputField, err := root.FieldByName("outputs", false)
	require.NoError(t, err, "missing outputs")

	var templates []ast.Node
	it, err := root.Fields(cue.Definitions(true), cue.Hidden(true))
	for it.Next() {
		if !it.IsDefinition() {
			continue
		}

		// TODO
		// if attr := it.Value().Attribute("template"); attr.Err() != nil {
		// 	continue
		// }

		// tpl := it.Value().Lookup("out")

		templates = append(templates, it.Value().Syntax(cue.All()))
	}

	return TestCase{
		Inputs:    cueToAny(t, inputField.Value),
		Outputs:   cueToAny(t, outputField.Value),
		Templates: templates,
	}
}

func TestTranspile(t *testing.T) {
	cases, err := filepath.Glob("testdata/*.cue")
	require.NoError(t, err)

	// https://github.com/helm/helm/blob/e81f6140ddb22bc99a08f7409522a8dbe5338ee3/pkg/engine/funcs.go#L43-L76
	funcs := sprig.TxtFuncMap()
	funcs["toYaml"] = func(v any) any {
		data, err := yaml.Marshal(v)
		if err != nil {
			// Swallow errors inside of a template.
			return ""
		}
		return strings.TrimSuffix(string(data), "\n")
	}

	for _, c := range cases {
		c := c
		t.Run(c[len("testdata/"):], func(t *testing.T) {
			tc := LoadTestCase(t, c)

			t.Logf("Inputs: %#v", tc.Inputs)
			t.Logf("Outputs: %#v", tc.Outputs)

			t.Log("Templates:")

			for _, tpl := range tc.Templates {
				src, err := format.Node(tpl)
				require.NoError(t, err)

				t.Logf("Transpiling:\n%s", src)
				PrintCueAst(tpl)
				transpiled := Transpile(tpl)

				// TODO make a pretty printer.
				t.Logf("Into:\n%s", transpiled.String())

				compiled, err := template.New("").Funcs(funcs).Parse(transpiled.String())
				require.NoError(t, err)

				for i := range tc.Inputs {
					var buf bytes.Buffer
					require.NoError(t, compiled.Execute(&buf, tc.Inputs[i]))

					expected, err := yaml.Marshal(tc.Outputs[i])
					require.NoError(t, err)

					// Special case. Values can be entirely elided
					// within templates. Cue however will default to
					// an empty object. Cue is actually incorrect in
					// this case.
					if string(expected) == "{}\n" && buf.String() == "" {
						continue
					}

					require.Equal(t, string(expected), buf.String()+"\n")
				}
			}
		})
	}
}
