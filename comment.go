package main

import (
	"go/token"

	"github.com/anton2920/gofa/strings"
)

type Comment struct {
	Formats []Format
}

func ParseGofaComment(l *Lexer, comment *Comment) bool {
	const prefix = "//gpp:generate"

	tok := l.Peek()
	if (l.Error != nil) || (tok.GoToken != token.COMMENT) || (!strings.StartsWith(tok.Literal, prefix)) {
		return false
	}
	lit := tok.Literal[len(prefix):]

	done := false
	for !done {
		s, rest, ok := strings.Cut(lit, ",")
		if !ok {
			done = true
		}
		s = strings.TrimSpace(s)

		switch {
		case strings.StartsWith(s, "json"):
			comment.Formats = append(comment.Formats, &FormatJSON{})
		}

		lit = rest
	}

	l.Next()
	return true
}
