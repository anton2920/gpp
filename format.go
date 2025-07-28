package main

import (
	"fmt"
	"unicode"
)

type Format interface {
	Serialize(*Generator, *TypeSpec)
	Deserialize(*Generator, *TypeSpec)
}

/* NOTE(anton2920): this supports only ASCII. */
func VariableName(typeName string, array bool) string {
	var lastUpper int
	for i := 0; i < len(typeName); i++ {
		if unicode.IsUpper(rune(typeName[i])) {
			lastUpper = i
		}
	}

	var suffix string
	if array {
		suffix = "s"
	}

	return fmt.Sprintf("%c%s%s", unicode.ToLower(rune(typeName[lastUpper])), typeName[lastUpper+1:], suffix)
}
