package cue

#Kind: "Literal" | "Ident" | "List" | "Struct"

#CueNode: #CueBasicLit | #CueIdent | #CueListLit

#CueIdent: {
	kind:  "Ident"
	ident: string
}

#CueBasicLit: {
	kind:  "BasicLit"
	value: string
}

#CueListLit: {
	kind: "List"
	elts: [...#CueNode]
}
