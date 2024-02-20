package main

import (
	"text/template/parse"
)

func asListNode(nodes ...parse.Node) *parse.ListNode {
	return &parse.ListNode{Nodes: nodes}
}

func wrapInParens(children ...parse.Node) parse.Node {
	return &parse.CommandNode{
		Args: []parse.Node{
			&parse.PipeNode{
				Cmds: []*parse.CommandNode{
					{Args: children},
				},
			},
		},
	}
}

func funcCall(fn string, arguments ...parse.Node) parse.Node {
	return wrapInParens(append([]parse.Node{
		parse.NewIdentifier(fn),
	}, arguments...)...)
}

func declare(x *parse.VariableNode, expr parse.Node) parse.Node {
	return &parse.ActionNode{
		Pipe: &parse.PipeNode{
			IsAssign: false,
			Decl:     []*parse.VariableNode{x},
			Cmds: []*parse.CommandNode{
				{Args: []parse.Node{expr}},
			},
		},
	}
}

func assign(x *parse.VariableNode, expr parse.Node) parse.Node {
	// Action / pipe nodes are kinda broken. IsAssign doesn't work for round
	// tripping.
	return &parse.ListNode{
		Nodes: []parse.Node{
			&parse.TextNode{Text: []byte("\n{{- ")},
			x,
			&parse.TextNode{Text: []byte(" = ")},
			expr,
			&parse.TextNode{Text: []byte(" -}}\n")},
		},
	}
}

func mkRange(idx, el *parse.VariableNode, over parse.Node, body ...parse.Node) parse.Node {
	return &parse.RangeNode{
		BranchNode: parse.BranchNode{
			NodeType: parse.NodeRange,
			Pipe: &parse.PipeNode{
				Decl: []*parse.VariableNode{idx, el},
				Cmds: []*parse.CommandNode{
					{Args: []parse.Node{over}},
				},
			},
			List: asListNode(body...),
		},
	}
}
