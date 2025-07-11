package main

import (
	"go/token"

	"github.com/anton2920/gofa/strings"
)

type Comment struct {
	Targets [TargetCount]bool
}

func ParseGofaComment(l *Lexer, comment *Comment) bool {
	const prefix = "//gofa:gpp"

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

		for t := TargetNone + 1; t < TargetCount; t++ {
			if s == Target2String[t] {
				comment.Targets[t] = true
			}
		}

		lit = rest
	}

	l.Next()
	return true
}
