package main

import (
	"go/token"

	"github.com/anton2920/gofa/strings"
)

type Comment struct {
	Targets [TargetCount]bool
}

func ParseGofaComment(l *Lexer, comment *Comment) bool {
	const prefix = "//gofa:generate"

	/* TODO(anton2920): make it in line with other parsers. */
	if (l.Error != nil) || (l.Token != token.COMMENT) || (!strings.StartsWith(l.Literal, prefix)) {
		return false
	}
	lit := l.Literal[len(prefix):]

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

	return true
}
