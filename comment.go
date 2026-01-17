package main

import (
	"fmt"
	"go/token"
	"os"
	stdstrings "strings"

	"github.com/anton2920/gofa/bools"
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
	InsertAfter  []string
	InsertBefore []string
	Func         string
	Enum         bool
	NOP          bool
	Optional     bool
}

type VerifyComment struct {
	InsertAfter  []string
	InsertBefore []string
	Min          string
	Max          string
	MinLength    string
	MaxLength    string
	Funcs        []string
	Prefix       string
	Optional     bool
	Required     bool
	SOA          bool
	SOAPrefix    string
}

type UnionComment struct {
	Types []string
}

const (
	LCompound = "{{"
	RCompound = "}}"
)

func (NOPComment) Comment()      {}
func (InlineComment) Comment()   {}
func (GenerateComment) Comment() {}
func (FillComment) Comment()     {}
func (VerifyComment) Comment()   {}
func (UnionComment) Comment()    {}

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

func ProperCut(s string, sep string, s1 string, s2 string) (string, string, bool) {
	lval, rval, ok := strings.Cut(s, sep)
	if !ok {
		return lval, rval, ok
	}

	s1pos := strings.FindSubstring(lval, s1)
	if s1pos == -1 {
		return lval, rval, ok
	}

	s2pos := strings.FindSubstring(lval, s2)
	if s2pos >= 0 {
		return lval, rval, ok
	}

	s2pos = strings.FindSubstring(rval, s2)
	if s2pos == -1 {
		return lval, rval, ok
	}

	seppos := strings.FindSubstring(rval[s2pos:], sep)
	seppos += s2pos
	if (seppos == s2pos-1) || (seppos == len(rval)-1) {
		return lval + sep + rval[:len(rval)-bools.ToInt(seppos == len(rval)-1)], "", seppos == len(rval)-1
	}

	return lval + sep + rval[:seppos], rval[seppos+1:], true
}

func StripIfFound(s string, prefix string, suffix string) (string, bool) {
	if strings.StartsEndsWith(s, prefix, suffix) {
		return s[len(prefix) : len(s)-len(suffix)], true
	}
	return s, false
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
			s, rest, ok := ProperCut(lit, ";", LCompound, RCompound)
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
						s, rest, ok := ProperCut(lit, ",", "(", ")")
						if !ok {
							done = true
						}
						s = strings.TrimSpace(s)

						gen := url.Path(s)
						switch {
						case gen.Match("fill..."):
							if gen == "" {
								gc.Generators = append(gc.Generators, GeneratorsFillAll()...)
							} else if lit, ok := StripIfFound(string(gen), "(", ")"); ok {
								var done bool
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
							if gen == "" {
								gc.Generators = append(gc.Generators, GeneratorsEncodingAll()...)
							} else if lit, ok := StripIfFound(string(gen), "(", ")"); ok {
								var done bool
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
					s, rest, ok := ProperCut(lit, ",", LCompound, RCompound)
					if !ok {
						done = true
					}
					s = strings.TrimSpace(s)

					switch stdstrings.ToLower(s) {
					case "enum":
						fc.Enum = true
					case "nop":
						fc.NOP = true
					case "optional":
						fc.Optional = true
					default:
						lval, rval, ok := strings.Cut(s, "=")
						if ok {
							lval = stdstrings.ToLower(strings.TrimSpace(lval))
							rval = strings.TrimSpace(rval)

							switch lval {
							case "insertafter":
								fc.InsertAfter = append(fc.InsertAfter, rval)
							case "insertbefore":
								fc.InsertBefore = append(fc.InsertBefore, rval)
							case "func":
								fc.Func = rval
							}
						}
					}

					lit = rest
				}

				if (fc.Optional) && (!fc.Enum) {
					fmt.Fprintf(os.Stderr, "WARNING: ignoring 'Optional' without 'Enum'")
					fc.Optional = false
				}

				*comments = append(*comments, fc)
			case fn.Match("verify:..."):
				var vc VerifyComment
				var done bool

				lit := string(fn)
				for !done {
					s, rest, ok := ProperCut(lit, ",", LCompound, RCompound)
					if !ok {
						done = true
					}
					s = strings.TrimSpace(s)

					lval, rval, ok := strings.Cut(s, "=")
					if !ok {
						lval = stdstrings.ToLower(strings.TrimSpace(lval))

						switch lval {
						case "optional":
							vc.Optional = true
						case "required":
							vc.Required = true
						case "soa":
							vc.SOA = true
						}
					} else {
						lval = stdstrings.ToLower(strings.TrimSpace(lval))
						rval = strings.TrimSpace(rval)

						switch lval {
						case "insertafter":
							vc.InsertAfter = append(vc.InsertAfter, rval)
						case "insertbefore":
							vc.InsertBefore = append(vc.InsertBefore, rval)
						case "min":
							vc.Min = rval
						case "max":
							vc.Max = rval
						case "minlen", "minlength":
							vc.MinLength = rval
						case "maxlen", "maxlength":
							vc.MaxLength = rval
						case "func":
							vc.Funcs = append(vc.Funcs, rval)
						case "prefix":
							vc.Prefix = rval
						case "soa":
							vc.SOA = true
							vc.SOAPrefix = rval
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
