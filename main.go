package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"text/template/parse"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/format"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/cue/token"
	"github.com/chrisseto/cue-to-helm/astutil"
)

func ToPrintf(intpl *ast.Interpolation) *parse.CommandNode {
	fmtStr := ""
	var args []parse.Node
	for i, el := range intpl.Elts {
		switch el := el.(type) {
		case *ast.BasicLit:
			v := el.Value

			if i == 0 {
				v = v[1:]
			} else if i == len(intpl.Elts)-1 {
				v = v[:len(v)-1]
			}
			if strings.HasPrefix(v, `)`) {
				v = v[1:]
			}
			if strings.HasSuffix(v, `\(`) {
				v = v[:len(v)-2]
			}
			fmtStr += v
		default:
			fmtStr += "%s"
			args = append(args, cueToTemplate(nil, el))
		}
	}

	return &parse.CommandNode{
		Args: append([]parse.Node{
			&parse.IdentifierNode{Ident: "printf"},
			&parse.StringNode{Text: fmtStr, Quoted: strconv.Quote(fmtStr)},
		}, args...),
	}
}

type templateContext struct {
	idx     int
	globals []parse.Node
	prepend []parse.Node
	append  []parse.Node
}

func (c *templateContext) AsRoot() *parse.ListNode {
	var out []parse.Node
	out = append(out, c.globals...)
	out = append(out, c.prepend...)
	out = append(out, c.append...)
	return &parse.ListNode{Nodes: out}
}

func (c *templateContext) GlobalBinding(n parse.Node) {
	c.globals = append(c.globals, n)
}

func (c *templateContext) Push(n parse.Node) {
	c.prepend = append(c.prepend, n)
}

func (c *templateContext) Append(n parse.Node) {
	c.append = append(c.append, n)
}

func (c *templateContext) NewIdentifier() *parse.IdentifierNode {
	x := c.idx
	c.idx += 1
	return &parse.IdentifierNode{
		Ident: fmt.Sprintf("$tmp_%d", x),
	}
}

func (c *templateContext) NewVariable() *parse.VariableNode {
	x := c.idx
	c.idx += 1
	return &parse.VariableNode{
		Ident: []string{fmt.Sprintf("$tmp_%d", x)},
	}
}

func (c *templateContext) PushVariable(declaration parse.Node) *parse.IdentifierNode {
	ident := c.NewIdentifier()
	c.Push(&parse.ActionNode{
		Pipe: &parse.PipeNode{
			Decl: []*parse.VariableNode{
				{Ident: []string{ident.Ident}},
			},
			Cmds: []*parse.CommandNode{
				{Args: []parse.Node{declaration}},
			},
		},
	})
	return ident
}

func cueToTemplate(ctx *templateContext, n ast.Node) parse.Node {
	switch node := n.(type) {
	case *ast.ImportDecl:
		return &parse.ListNode{}

	case *ast.Attribute: // TODO probably want to process these??
		return &parse.ListNode{}

	case *ast.EmbedDecl: // TODO
		return &parse.ListNode{}

	case *ast.LetClause:
		ctx.Push(&parse.ActionNode{
			Pipe: &parse.PipeNode{
				Decl: []*parse.VariableNode{{Ident: []string{
					fmt.Sprintf("$%s", node.Ident),
				}}},
				Cmds: []*parse.CommandNode{
					{Args: []parse.Node{cueToTemplate(ctx, node.Expr)}},
				},
			},
		})
		return &parse.ListNode{}

	case *ast.File:
		for _, decl := range node.Decls {
			ctx.Append(cueToTemplate(ctx, decl))
		}
		return &parse.ListNode{}

	case *ast.BinaryExpr:
		// if ident, ok := node.X.(*ast.Ident); !ok || ident.Name[0] != '#' {
		// 	panic(fmt.Sprintf("Uhandled: %T", node))
		// }
		// TODO this is presumputous
		return cueToTemplate(ctx, node.Y)

	case *ast.SelectorExpr:
		out, _ := format.Node(node)
		return &parse.IdentifierNode{Ident: "$" + string(out)}

	case *ast.Interpolation:
		return &parse.PipeNode{
			Cmds: []*parse.CommandNode{
				ToPrintf(node),
			},
		}

	case *ast.ListLit:
		if len(node.Elts) == 1 {
			if _, ok := node.Elts[0].(*ast.Comprehension); ok {
				return cueToTemplate(ctx, node.Elts[0])
			}
		}

		var nodes []parse.Node
		for _, el := range node.Elts {
			nodes = append(nodes, cueToTemplate(ctx, el))
		}
		return funcCall("list", nodes...)

	case *ast.StructLit:
		var nodes []parse.Node
		for _, el := range node.Elts {
			n := cueToTemplate(ctx, el)
			if l, ok := n.(*parse.ListNode); ok {
				nodes = append(nodes, l.Nodes...)
			} else {
				nodes = append(nodes, n)
			}
		}
		return funcCall("dict", nodes...)

	case *ast.BasicLit:
		switch node.Kind {
		case token.STRING:
			return &parse.StringNode{Quoted: node.Value}
		default:
			panic("unknown kind")
		}

	case *ast.Ident:
		return &parse.IdentifierNode{Ident: "$" + node.Name}

	case *ast.Field:
		if len(node.Attrs) == 1 && node.Attrs[0].Text == `@template("context")` {
			ctx.GlobalBinding(&parse.ActionNode{
				Pipe: &parse.PipeNode{
					Decl: []*parse.VariableNode{
						{Ident: []string{fmt.Sprintf("$%s", node.Label)}},
					},
					Cmds: []*parse.CommandNode{
						{Args: []parse.Node{&parse.DotNode{}}},
					},
				},
			})
			return &parse.ListNode{}
		}

		if len(node.Attrs) == 1 && node.Attrs[0].Text == `@template("toYaml")` {
			ident := &parse.VariableNode{Ident: []string{fmt.Sprintf("$%s", node.Label)}}

			return &parse.ListNode{
				Nodes: []parse.Node{
					&parse.ActionNode{
						Pipe: &parse.PipeNode{
							Decl: []*parse.VariableNode{ident},
							Cmds: []*parse.CommandNode{
								{Args: []parse.Node{cueToTemplate(ctx, node.Value)}},
							},
						},
					},
					&parse.ActionNode{
						Pipe: &parse.PipeNode{
							Cmds: []*parse.CommandNode{
								{Args: []parse.Node{ident}},
								{Args: []parse.Node{
									&parse.IdentifierNode{Ident: "toYaml"},
								}},
							},
						},
					},
				},
			}
		}

		// If we have a schema definition, just recurse into it.
		if strings.Contains(node.Label.(*ast.Ident).Name, "#") {
			// Unwrapp the "dict" call.
			// TODO this is nasty and hacky.
			children := cueToTemplate(ctx, node.Value).(*parse.CommandNode).Args[0].(*parse.PipeNode).Cmds[0].Args[1:]
			ctx.Append(&parse.ListNode{Nodes: children})
			return &parse.ListNode{}
		}

		return &parse.ListNode{
			Nodes: []parse.Node{
				&parse.StringNode{Quoted: strconv.Quote(fmt.Sprintf("%s", node.Label))},
				cueToTemplate(ctx, node.Value),
			},
		}

	case *ast.Comprehension:
		result := node.Value.(*ast.StructLit)

		if len(node.Clauses) == 1 {
			if clause, ok := node.Clauses[0].(*ast.IfClause); ok {
				children := cueToTemplate(ctx, result).(*parse.CommandNode).Args[0].(*parse.PipeNode).Cmds[0].Args[1:]
				return &parse.ListNode{
					Nodes: []parse.Node{
						&parse.IfNode{
							BranchNode: parse.BranchNode{
								NodeType: parse.NodeIf,
								Pipe: &parse.PipeNode{
									Cmds: []*parse.CommandNode{
										{Args: []parse.Node{cueToTemplate(ctx, clause.Condition)}},
									},
								},
								List: &parse.ListNode{Nodes: children},
							},
						},
					},
				}
			}

			if clause, ok := node.Clauses[0].(*ast.ForClause); ok {
				return transpileForClause(ctx, clause, node.Value)
			}

			panic(fmt.Sprintf("Unhandled: %T\n", node.Clauses[0]))

		}
		panic(fmt.Sprintf("Unhandled: %T\n", n))

	default:
		// out, _ := format.Node(n)
		panic(fmt.Sprintf("Unhandled: %T\n", n))
	}
}

// Transform ForClauses into something like:
// {{$result := list}}
// {{range ....}}
//
//	{{ $result = append($result, ....) }}
//
// {{end}}
// And returns a handle to the result variable
// TODO could consider using defines for cases like this.
func transpileForClause(ctx *templateContext, clause *ast.ForClause, value ast.Node) *parse.VariableNode {
	idxIdent := &parse.VariableNode{Ident: []string{"$_"}}
	elIdent := &parse.VariableNode{Ident: []string{fmt.Sprintf("$%s", clause.Value)}}
	resultIdent := ctx.NewVariable()

	over := cueToTemplate(ctx, clause.Source)
	mkResult := cueToTemplate(ctx, value)

	// Start with an empty list.
	declare(resultIdent, funcCall("list"))

	// Push item into it.
	body := assign(resultIdent, funcCall("append", resultIdent, mkResult))

	ctx.Push(asListNode(
		declare(resultIdent, funcCall("list")),
		mkRange(idxIdent, elIdent, over, body),
	))

	return resultIdent
}

func Transpile(input ast.Node) parse.Node {
	// Using the ctx is kinda working for me.
	// Need to get everything working again though.
	// The lack of pattern matching in go is pretty rough for this type of work.
	// So instead of relying on patterns, we'll fallback to explicit templates
	// attributes.
	// That way we can walk the entire tree and push definitions into the
	// context as need be.
	// Currently we've got two: @template(context) and @template(toYaml)
	// Otherwise we'll just build out a big ol' nested tree.
	// Everything will get bound to variables that we then reference.
	ctx := &templateContext{}
	cueToTemplate(ctx, input)
	return ctx.AsRoot()
}

func loadTpl(filename string) *parse.ListNode {
	tpl, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	trees, err := parse.Parse("", string(tpl), "{{", "}}", map[string]any{
		"include": func(args ...any) {},
		"nindent": func(args ...any) {},
		"toYaml":  func(args ...any) {},
		"quote":   func(args ...any) {},
		"and":     func(args ...any) {},
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v\n", trees)

	root := trees[""].Root.Nodes
	for key, tree := range trees {
		if key == "" {
			continue
		}

		root = append([]parse.Node{
			&parse.ActionNode{Pipe: &parse.PipeNode{
				Cmds: []*parse.CommandNode{
					{Args: []parse.Node{
						&parse.IdentifierNode{Ident: "define"},
						&parse.StringNode{Text: key, Quoted: fmt.Sprintf("%q", key)},
					}},
				},
			}},
			tree.Root,
			&parse.ActionNode{Pipe: &parse.PipeNode{
				Cmds: []*parse.CommandNode{
					{Args: []parse.Node{
						&parse.IdentifierNode{Ident: "end"},
					}},
				},
			}},
		}, root...)

	}

	return &parse.ListNode{Nodes: root}
}

func PrintCueAst(n ast.Node) {
	fmt.Printf(astutil.PrintTree[ast.Node](ast.Walk, astutil.PathDescriber(func(b io.Writer, n ast.Node) {
		switch node := n.(type) {
		case *ast.File:
			fmt.Fprintf(b, "{%q}", path.Base(node.Filename))
		case *ast.Field:
			fmt.Fprintf(b, "{%q}", node.Label)
		case *ast.BasicLit:
			fmt.Fprintf(b, "{%q}", node.Value)
		case *ast.LetClause:
			fmt.Fprintf(b, "{%q}", node.Ident)
		case *ast.SelectorExpr:
			fmt.Fprintf(b, "{%q}", node.Sel)
		case *ast.Ident:
			fmt.Fprintf(b, "{%q}", node.String())
		}
	}), n))
}

func PrintTplAst(root parse.Node) {
	fmt.Printf(astutil.PrintTree[parse.Node](astutil.Walk, astutil.PathDescriber(func(b io.Writer, n parse.Node) {
		switch node := n.(type) {
		case *parse.IdentifierNode:
			fmt.Fprintf(b, "{%q}", node.Ident)
		case *parse.TextNode:
			fmt.Fprintf(b, "{%q}", node.Text)
		}
	}), root))
}

func Walk(root cue.Value, before func(cue.Value) bool, after func(cue.Value)) {
	it, err := root.Fields(cue.All())
	if err != nil {
		root.Walk(before, after)
		return
	}

	for it.Next() {
		if !before(it.Value()) {
			continue
		}
		Walk(it.Value(), before, after)
		if after != nil {
			after(it.Value())
		}
	}
}

func LoadCueValue(filepath string) cue.Value {
	instances := load.Instances([]string{filepath}, &load.Config{
		ModuleRoot: ".",
	})
	c := cuecontext.New()
	return c.BuildInstance(instances[0])
}

func LoadCue(filepath string) (ast.Node, ast.Node, error) {
	instances := load.Instances([]string{filepath}, &load.Config{
		ModuleRoot: ".",
	})

	if instances[0].Err != nil {
		return nil, nil, instances[0].Err
	}

	// Round trip the syntax to simplify syntactic sugar and all that jazz.
	c := cuecontext.New()
	value := c.BuildInstance(instances[0])

	return instances[0].Files[0], value.Syntax(cue.All()), nil
}

func main() {
	if len(os.Args) == 1 {
		fmt.Printf("Usage: [print-cue <file> | print-tpl <file>]")
	}

	switch os.Args[1] {
	case "print-cue":
		raw, rt, err := LoadCue(os.Args[2])
		if err != nil {
			panic(err)
		}
		PrintCueAst(raw)
		fmt.Printf("\n===========\n")
		PrintCueAst(rt)

	case "print-tpl":
		ast := loadTpl(os.Args[2])
		PrintTplAst(ast)

	case "transpile":
		raw, in, err := LoadCue(os.Args[2])
		if err != nil {
			panic(err)
		}

		fmt.Printf("\n===========\n")
		PrintCueAst(raw)
		fmt.Printf("\n===========\n")
		PrintCueAst(in)
		fmt.Printf("\n===========\n")

		transpiled := Transpile(in)
		PrintTplAst(transpiled)

		fmt.Printf("\n===========\n")
		fmt.Printf("%s", transpiled.String())
		fmt.Printf("\n===========\n")

	default:
		panic(fmt.Sprintf("Unhandled: %q", os.Args[1]))
	}
}
