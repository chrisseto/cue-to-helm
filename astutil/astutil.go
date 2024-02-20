package astutil

import (
	"fmt"
	"io"
	"strings"
	"text/template/parse"
)

type WalkFunc[N any] func(N, func(N) bool, func(N))

func PathDescriber[N any](describe func(io.Writer, N)) func(io.Writer, []N) bool {
	return func(b io.Writer, stack []N) bool {
		for i, n := range stack {
			if i > 0 {
				fmt.Fprintf(b, " > ")
			}
			fmt.Fprintf(b, "%T", n)
			describe(b, n)
		}
		fmt.Fprintf(b, "\n")
		return true
	}
}

func PrintTree[N any](walk WalkFunc[N], describe func(io.Writer, []N) bool, root N) string {
	var stack []N
	var b strings.Builder

	walk(root, func(n N) bool {
		stack = append(stack, n)
		return describe(&b, stack)
	}, func(n N) {
		stack = stack[:len(stack)-1]
	})

	return b.String()
}

func Walk(n parse.Node, before func(parse.Node) bool, after func(parse.Node)) {
	if !before(n) {
		return
	}
	defer after(n)

	switch node := n.(type) {
	case *parse.ListNode:
		for _, child := range node.Nodes {
			Walk(child, before, after)
		}

	case *parse.CommandNode:
		for _, a := range node.Args {
			Walk(a, before, after)
		}

	case *parse.PipeNode:
		for _, d := range node.Decl {
			Walk(d, before, after)
		}
		for _, c := range node.Cmds {
			Walk(c, before, after)
		}

	case *parse.BranchNode:
		Walk(node.Pipe, before, after)
		Walk(node.List, before, after)
		if node.ElseList != nil {
			Walk(node.ElseList, before, after)
		}

	case *parse.IfNode:
		Walk(&node.BranchNode, before, after)

	case *parse.ActionNode:
		Walk(node.Pipe, before, after)

	case *parse.WithNode:
		Walk(&node.BranchNode, before, after)

	case *DefineNode:
		Walk(node.Body, before, after)

	case *parse.RangeNode:
		Walk(node.Pipe, before, after)
		Walk(node.List, before, after)

	case *parse.IdentifierNode:
	case *parse.FieldNode:
	case *parse.TextNode:
	case *parse.StringNode:
	case *parse.DotNode:
	case *parse.NumberNode:
	case *parse.VariableNode:

	default:
		panic(fmt.Sprintf("unhandled: %T", n))
	}
}

type DefineNode struct {
	parse.NilNode

	Name string
	Body parse.Node
}
