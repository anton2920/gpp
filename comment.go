package main

import (
	"fmt"
	"go/token"

	"github.com/anton2920/gofa/net/url"
	"github.com/anton2920/gofa/strings"
)

type Comment interface{}

type GenerateComment struct {
	Generators []Generator
}

type NOPComment struct{}

type InlineComment struct{}

type VerifyComment struct {
	Min       int
	Max       int
	MinLength int
	MaxLength int
}

func (p *Parser) Comment(comment *Comment) bool {
	const prefix = "gpp:"

	if p.Token(token.COMMENT) {
		tok := p.Prev()
		if !strings.StartsWith(tok.Literal[2:], prefix) {
			p.Error = fmt.Errorf("expected prefix %q, got %q", prefix, tok.Literal)
			return false
		}
		lit := tok.Literal[2+len(prefix):]

		fn := url.Path(lit)
		switch {
		case fn.Match("nop"):
			*comment = NOPComment{}
			return true
		case fn.Match("inline"):
			*comment = InlineComment{}
			return true
		case fn.Match("generate..."):
			var gc GenerateComment

			switch {
			case fn == "":
				gc.Generators = append(gc.Generators, GeneratorsAll()...)
			case fn.Match(":..."):
				var done bool

				lit := string(fn)
				for !done {
					s, rest, ok := strings.Cut(lit, ",")
					if !ok {
						done = true
					}
					s = strings.TrimSpace(s)

					gen := url.Path(s)
					switch {
					case gen.Match("fill"):
						gc.Generators = append(gc.Generators, GeneratorFill{})
					case gen.Match("verify"):
						gc.Generators = append(gc.Generators, GeneratorVerify{})
					case gen.Match("encoding..."):
						var list string
						switch {
						case gen == "":
							gc.Generators = append(gc.Generators, GeneratorsEncodingAll()...)
						case gen.Match("(%s)", &list):
							var done bool

							lit := string(list)
							for !done {
								s, rest, ok := strings.Cut(lit, ",")
								if !ok {
									done = true
								}
								s = strings.TrimSpace(s)

								switch s {
								case "json":
									gc.Generators = append(gc.Generators, GeneratorJSON{})
								case "wire":
									gc.Generators = append(gc.Generators, GeneratorWire{})
								}

								lit = rest
							}
						}
					}

					lit = rest
				}
			}

			*comment = gc
			return true
		case strings.StartsWith(lit, "verify:"):
			var vc VerifyComment
			*comment = vc
			return true
		}
	}
	return false
}
