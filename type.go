package main

type Type struct {
	Name string
	Kind TypeKind
}

type TypeKind int

const (
	TypeKindNone = TypeKind(iota)
	TypeKindInt
	TypeKindInt64
	TypeKindString
)

func ParseType(l *Lexer, t *Type) bool {
	if ParseIdent(l, &t.Name) {
		switch t.Name {
		case "int":
			t.Kind = TypeKindInt
		case "int64":
			t.Kind = TypeKindInt64
		case "string":
			t.Kind = TypeKindString
		}
	}

	return t.Kind != TypeKindNone
}
