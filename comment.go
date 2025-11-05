package main

import (
	"fmt"
	"go/token"
	"strconv"

	"github.com/anton2920/gofa/net/url"
	"github.com/anton2920/gofa/strings"
)

type Comment interface{}

type NOPComment struct{}

type InlineComment struct{}

type GenerateComment struct {
	Generators []Generator
}

type FillCommentNOP struct{}

type VerifyComment struct {
	Min       int
	Max       int
	MinLength int
	MaxLength int
	Func      string
	Enum      string
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
		if strings.EndsWith(lit, "*/") {
			lit = lit[:len(lit)-2]
		}

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
					case gen.Match("fill..."):
						var list string
						switch {
						case gen == "":
							gc.Generators = append(gc.Generators, GeneratorsFillAll()...)
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
								case "values":
									gc.Generators = append(gc.Generators, GeneratorFillValues{})
								}

								lit = rest
							}
						}
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
		case fn.Match("fill: nop"):
			*comment = FillCommentNOP{}
			return true
		case fn.Match("verify..."):
			var vc VerifyComment
			var done bool

			lit := string(fn)
			for !done {
				s, rest, ok := strings.Cut(lit, ",")
				if !ok {
					done = true
				}
				s = strings.TrimSpace(s)

				lval, rval, ok := strings.Cut(s, "=")
				if ok {
					lval = strings.TrimSpace(lval)
					rval = strings.TrimSpace(rval)

					switch lval {
					case "Min":
						vc.Min, _ = strconv.Atoi(rval)
					case "Max":
						vc.Max, _ = strconv.Atoi(rval)
					case "MinLength":
						vc.MinLength, _ = strconv.Atoi(rval)
					case "MaxLength":
						vc.MaxLength, _ = strconv.Atoi(rval)
					case "Func":
						vc.Func = rval
					case "Enum":
						vc.Enum = rval
					}
				}

				lit = rest
			}
			*comment = vc
			return true
		}
	}
	return false
}
