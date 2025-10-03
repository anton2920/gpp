package main

import (
	"go/token"

	"github.com/anton2920/gofa/strings"
)

type Comment struct {
	Encodings []Encoding
}

func (p *Parser) GofaComment(comment *Comment) bool {
	const prefix = "//gpp:generate"

	tok := p.Curr()
	if (p.Error != nil) || (tok.GoToken != token.COMMENT) || (!strings.StartsWith(tok.Literal, prefix)) {
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
			comment.Encodings = append(comment.Encodings, &EncodingJSON{Parser: p})
		}

		lit = rest
	}

	p.Next()
	return true
}
