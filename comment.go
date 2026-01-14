package main

import (
	"fmt"
	"go/token"
	stdstrings "strings"

	"github.com/anton2920/gofa/net/url"
	"github.com/anton2920/gofa/strings"
)

type Comment interface {
	Comment()
}

type NOPComment struct{}

type InlineComment struct{}

type GenerateComment struct {
	Generators []Generator
}

type FillComment struct {
	InsertAfter  string
	InsertBefore string
	Func         string
	Enum         bool
	NOP          bool
}

type VerifyComment struct {
	Min       string
	Max       string
	MinLength string
	MaxLength string
	Funcs     []string
	Prefix    string
	Required  bool
	Optional  bool
}

type UnionComment struct {
	Types []string
}

func (NOPComment) Comment()      {}
func (InlineComment) Comment()   {}
func (GenerateComment) Comment() {}
func (FillComment) Comment()     {}
func (VerifyComment) Comment()   {}
func (UnionComment) Comment()    {}

func FixMyCut(s *string, rest *string, c1 byte, c2 byte) {
	l := strings.FindChar(*s, c1)
	if l >= 0 {
		r := strings.FindChar(*s, c2)
		if r == -1 {
			r := strings.FindChar(*rest, c2)
			if r == -1 {
				return
			}

			/* Be greedy: if there are multiple 'c2's close by, find the rightmost one. */
			r++
			for (r < len(*rest)) && ((*rest)[r] == c2) {
				r++
			}
			r--

			*s = fmt.Sprintf("%s,%s", *s, (*rest)[:r+1])
			*rest = (*rest)[r+1:]
		}
	}
}

func AppendComments(cs1 []Comment, cs2 []Comment) []Comment {
	for _, comment := range cs1 {
		if _, ok := comment.(NOPComment); ok {
			return []Comment{NOPComment{}}
		}
	}
	for _, comment := range cs2 {
		if _, ok := comment.(NOPComment); ok {
			return []Comment{NOPComment{}}
		}
	}
	return append(cs1, cs2...)
}

func (p *Parser) Comments(comments *[]Comment) bool {
	const prefix = "gpp:"

	/* NOTE(anton2920): since '(*Parser).Token()' skips comments unless they are requested, we probably need to backtrack a bit .*/
	pos := p.Position
	for p.Prev().GoToken == token.COMMENT {
		p.Position--
	}
	for p.Token(token.COMMENT) {
		tok := p.Prev()
		if !strings.StartsWith(tok.Literal[2:], prefix) {
			continue
		}
		lit := tok.Literal[2+len(prefix):]
		if strings.EndsWith(lit, "*/") {
			lit = lit[:len(lit)-2]
		}

		var done bool
		for !done {
			s, rest, ok := strings.Cut(lit, ";")
			if !ok {
				done = true
			}
			s = strings.TrimSpace(s)

			fn := url.Path(s)
			switch {
			case fn.Match("nop"):
				*comments = []Comment{NOPComment{}}
				return true
			case fn.Match("inline"):
				*comments = []Comment{InlineComment{}}
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
						FixMyCut(&s, &rest, '(', ')')

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

							/* TODO(anton2920): fix case when list is 'json, wire' (with space). */
							case gen.Match("(%s)", &list):
								var done bool

								lit := string(list)
								for !done {
									s, rest, ok := strings.Cut(lit, ",")
									if !ok {
										done = true
									}
									s = strings.TrimSpace(s)

									/* TODO(anton2920): add logic for processing 'json(serialize, deserialize)' */
									switch s {
									case "json":
										gc.Generators = append(gc.Generators, GeneratorsEncodingJSONAll()...)
									case "wire":
										gc.Generators = append(gc.Generators, GeneratorsEncodingWireAll()...)
									}

									lit = rest
								}
							}
						}

						lit = rest
					}
				}

				*comments = append(*comments, gc)
			case fn.Match("fill:..."):
				var fc FillComment
				var done bool

				lit := string(fn)
				for !done {
					s, rest, ok := strings.Cut(lit, ",")
					if !ok {
						done = true
					}

					s = strings.TrimSpace(s)
					FixMyCut(&s, &rest, '{', '}')

					switch stdstrings.ToLower(s) {
					case "enum":
						fc.Enum = true
					case "nop":
						fc.NOP = true
					default:
						lval, rval, ok := strings.Cut(s, "=")
						if ok {
							lval = stdstrings.ToLower(strings.TrimSpace(lval))
							rval = strings.TrimSpace(rval)

							switch lval {
							case "insertafter":
								fc.InsertAfter = rval
							case "insertbefore":
								fc.InsertBefore = rval
							case "func":
								fc.Func = rval
							}
						}
					}

					lit = rest
				}

				*comments = append(*comments, fc)
			case fn.Match("verify:..."):
				var vc VerifyComment
				var done bool

				lit := string(fn)
				for !done {
					s, rest, ok := strings.Cut(lit, ",")
					if !ok {
						done = true
					}

					s = strings.TrimSpace(s)
					FixMyCut(&s, &rest, '{', '}')

					lval, rval, ok := strings.Cut(s, "=")
					if !ok {
						lval = stdstrings.ToLower(strings.TrimSpace(lval))

						switch lval {
						case "optional":
							vc.Optional = true
						case "required":
							vc.Required = true
						}
					} else {
						lval = stdstrings.ToLower(strings.TrimSpace(lval))
						rval = strings.TrimSpace(rval)

						switch lval {
						case "min":
							vc.Min = rval
						case "max":
							vc.Max = rval
						case "minlength":
							vc.MinLength = rval
						case "maxlength":
							vc.MaxLength = rval
						case "func":
							vc.Funcs = append(vc.Funcs, rval)
						case "prefix":
							vc.Prefix = rval
						}
					}

					lit = rest
				}

				*comments = append(*comments, vc)
			case fn.Match("union:..."):
				var uc UnionComment
				var done bool

				lit := string(fn)
				for !done {
					s, rest, ok := strings.Cut(lit, ",")
					if !ok {
						done = true
					}
					s = strings.TrimSpace(s)
					uc.Types = append(uc.Types, s)
					lit = rest
				}

				*comments = append(*comments, uc)
			}

			lit = rest
		}
	}

	p.Position = pos
	return len(*comments) > 0
}
