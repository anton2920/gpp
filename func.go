package main

import (
	"go/token"
)

func (p *Parser) FuncArg(arg *FuncArg) bool {
	return false
}

func (p *Parser) Func(fn *Func) bool {
	var comments []Comment
	p.Comments(&comments)
	p.Error = nil

	fn.BeginOffset = p.Curr().Offset
	if p.Token(token.FUNC) {
		fn.Comments = comments

		/* TODO(anton2920): do a proper function parsing. */
		for p.Curr().GoToken != token.LBRACE {
			p.Next()
		}
		fn.BodyBeginOffset = p.Curr().Offset

		count := 1
		for (p.Next().GoToken != token.EOF) && (count > 0) {
			tok := p.Curr()
			switch tok.GoToken {
			case token.LBRACE:
				count++
			case token.RBRACE:
				count--
			}
		}

		fn.BodyEndOffset = p.Curr().Offset
		return true
	}

	return false
}
