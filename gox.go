package main

import (
	"bytes"
	stdstrings "strings"
	"unicode"

	"github.com/anton2920/gofa/bools"
	"github.com/anton2920/gofa/strings"
)

type GeneratorGOX struct{}

func (g GeneratorGOX) Decl(r *Result, ctx GenerationContext, t *Type)          {}
func (g GeneratorGOX) Body(r *Result, ctx GenerationContext, t *Type)          {}
func (g GeneratorGOX) NamedType(r *Result, ctx GenerationContext, t *Type)     {}
func (g GeneratorGOX) Primitive(r *Result, ctx GenerationContext, lit TypeLit) {}
func (g GeneratorGOX) Struct(r *Result, ctx GenerationContext, s *Struct)      {}
func (g GeneratorGOX) StructField(r *Result, ctx GenerationContext, field *StructField, lit TypeLit) {
}
func (g GeneratorGOX) StructFieldSkip(field *StructField) bool {
	return false
}
func (g GeneratorGOX) Array(r *Result, ctx GenerationContext, a *Array) {}
func (g GeneratorGOX) Slice(r *Result, ctx GenerationContext, s *Slice) {}
func (g GeneratorGOX) Union(r *Result, ctx GenerationContext, u *Union) {}

type QuotedString struct {
	Key     string
	Value   string
	Quoted  bool
	Present bool
}

type Attributes map[string]QuotedString

func (attrs Attributes) Get(key string) QuotedString {
	ret := QuotedString{Key: key}

	v, ok := attrs[key]
	if ok {
		ret = v
	}

	delete(attrs, key)
	return ret
}

const (
	HandleComments = true
	Optimize       = false
	Inline         = false
)

var IntAttributes = []string{"minlength", "maxlength", "width", "height", "x", "y", "fontSize", "fontWeight", "strokeWidth", "cx", "cy", "r", "rx", "x1", "x2", "y1", "y2"}

func (qv QuotedString) String() string {
	if SliceContains(IntAttributes, qv.Key) {
		qv.Quoted = false
		if !qv.Present {
			return "0"
		}
	}
	if (qv.Quoted) || (!qv.Present) {
		return `"` + qv.Value + `"`
	}
	return qv.Value
}

func SliceContains(xs []string, s string) bool {
	for _, x := range xs {
		if s == x {
			return true
		}
	}
	return false
}

func StartsWithOneOf(xs []string, s string) bool {
	for _, x := range xs {
		if strings.StartsWith(s, x) {
			return true
		}
	}
	return false
}

func FindTagBegin(s string) int {
	var begin int

	for {
		pos := strings.FindChar(s[begin:], '<')
		if (pos == -1) || (begin+pos+1 >= len(s)) {
			return -1
		}
		begin += pos

		if (unicode.IsLetter(rune(s[begin+1]))) || (s[begin+1] == '-') || (s[begin+1] == '!') || (s[begin+1] == '/') || (s[begin+1] == '\n') {
			break
		}

		begin += 1
	}

	return begin
}

func FixAttr(s string) string {
	var buf bytes.Buffer

	pos := strings.FindChar(s, '-')
	if pos == -1 {
		return s
	}

	for pos >= 0 {
		buf.WriteString(s[:pos])
		s = stdstrings.Title(s[pos+1:])
		pos = strings.FindChar(s, '-')
	}
	buf.WriteString(s)

	return buf.String()
}

func UnfixAttr(s string) string {
	var buf bytes.Buffer

	if s == stdstrings.ToLower(s) {
		return s
	}

	var pos int
	for i := 0; i < len(s); i++ {
		if unicode.IsUpper(rune(s[i])) {
			buf.WriteString(s[pos:i])
			buf.WriteRune('-')
			buf.WriteRune(unicode.ToLower(rune(s[i])))
			pos = i + 1
		}
	}
	buf.WriteString(s[pos:])

	return buf.String()
}

func BeginStringBlock(r *Result) *Result {
	return r.Line("h.String(`").Backspace()
}

func EndStringBlock(r *Result) *Result {
	return r.WithoutTabs().Line("`)")
}

func GenerateGOXBody(r *Result, body string) {
	var codeBlock string
	var nblocks int
	var in bool

	for len(body) > 0 {
		var i int

		for i = 0; i < len(body); i++ {
			if !unicode.IsSpace(rune(body[i])) {
				break
			}
		}
		body = body[i:]

		if len(codeBlock) > 0 {
			needle := "</" + codeBlock + ">"
			end := strings.FindSubstring(body, needle)
			if end == -1 {
				Warnf("unterminated code block %q", codeBlock)
				break
			}

			r.Printf("h.String(`%s`)", strings.TrimSpace(body[:end])).Line("}").Printf("h.%sEnd2()", stdstrings.Title(codeBlock)).Line("")

			codeBlock = ""
			body = body[end+len(needle):]
			continue
		}

		begin := FindTagBegin(body)
		if (begin == -1) || (begin > 0) {
			const cutset = "\t\n"

			stmts := []string{"if", "else", "for", "switch"}
			cases := []string{"case", "default"}

			var otext string
			if begin == -1 {
				otext = stdstrings.Trim(body, cutset)
				body = ""
			} else {
				otext = stdstrings.Trim(body[:begin], cutset)
				body = body[begin:]
			}

			if in {
				EndStringBlock(r)
				in = false
			}

			for len(otext) > 0 {
				text := strings.TrimSpace(otext)

				if StartsWithOneOf(stmts, text) {
					end := strings.FindChar(text, '{')
					if end >= 0 {
						tabs := r.Tabs
						if strings.StartsWith(text, "else") {
							r.Backspace(1).Buffer.WriteRune(' ')
							r.Tabs = 0
						}
						r.Line(text[:end+1])
						r.Tabs = tabs + 1
						nblocks++

						otext = text[end+1:]
						continue
					}
				} else if StartsWithOneOf(cases, text) {
					end := strings.FindChar(text, ':')
					if end >= 0 {
						r.Tabs--
						r.RemoveLastNewline().Line(strings.TrimSpace(text[:end+1]))
						otext = text[end+1:]
						continue
					}
				}

				begin := strings.FindSubstring(otext, "{")
				end := strings.FindSubstring(otext, "}")
				if end == -1 {
					r.Printf("h.LString(`%s`)", otext)
					break
				} else if (end >= 0) && ((begin == -1) || (end < begin)) {
					if len(strings.TrimSpace(otext[:end])) > 0 {
						r.Printf("h.LString(`%s`)", stdstrings.Trim(otext[:end], cutset))
					}
					text = otext[end:]

					if nblocks == 0 {
						r.Line("h.LString(`}`)")
					} else {
						r.RemoveLastNewline().Line("}")
						nblocks--
					}
					otext = text[1:]
					continue
				} else {
					if strings.StartsWith(otext[begin:], "{{") {
						for (end+1 < len(otext)) && (otext[end+1] == '}') {
							end += 1
						}
					}
					value := otext[begin : end+1]

					if len(strings.TrimSpace(otext[:begin])) > 0 {
						r.Printf("h.LString(`%s`)", stdstrings.Trim(otext[:begin], cutset))
					}
					if value, ok := StripIfFound(value, "{{", "}}"); ok {
						r.RemoveLastNewline().Line(value)
					} else if value, ok := StripIfFound(value, "{", "}"); ok {
						r.Printf("h.LString(%s)", strings.TrimSpace(value))
					}

					otext = otext[end+1:]
				}
			}
		} else {
			end := strings.FindChar(body, '>')
			if end == -1 {
				break
			}

			if (Optimize) && (!in) {
				BeginStringBlock(r)
				in = true
			}

			if strings.StartsWith(body, "<!--") {
				end = strings.FindSubstring(body, "-->")

				if (!Optimize) && (HandleComments) {
					comment := body[begin+len("<!--") : end]
					lines := stdstrings.Split(comment, "\n")
					if len(lines) == 1 {
						r.Printf("/* %s */", comment)
					} else {
						r.RemoveLastNewline().Line("").Printf("/*").Backspace()
						r.Tabs++
						for i := 0; i < len(lines); i++ {
							r.Line(strings.TrimSpace(lines[i]))
						}
						r.Tabs--
						r.RemoveLastNewline().Printf("*/").Line("")
					}
				}

				body = body[end+len("-->"):]
				continue
			}

			s := body[1:end]
			body = body[end+1:]

			codeBlocks := []string{"style", "script"}

			noAttributesBlock := []string{"head", "body", "div", "select", "ol", "ul"}
			noAttributesNoBlock := []string{"h1", "h2", "h3", "h4", "h5", "h6", "p", "b", "i", "span", "option", "textarea", "title", "li", "label"}
			attributesBlock := []string{"form"}
			attributesNoBlock := []string{"a"}

			if s, ok := StripIfFound(s, "/", ""); ok {
				otag := strings.TrimSpace(s)
				tag := stdstrings.ToLower(otag)

				if !Optimize {
					switch {
					case (SliceContains(noAttributesBlock, tag)) || (SliceContains(attributesBlock, tag)) || (SliceContains([]string{"svg"}, tag)):
						r.RemoveLastNewline().Line("}")
						fallthrough
					case (SliceContains(noAttributesNoBlock, tag)) || (SliceContains(attributesNoBlock, tag)) || (SliceContains([]string{"text"}, tag)):
						r.RemoveLastNewline().Printf("h.%sEnd2()", stdstrings.Title(tag)).Line("")
					default:
						switch tag {
						case "html":
							r.Line("").Line("h.End2()")
						default:
							if (otag == stdstrings.ToUpper(otag)) || (otag != stdstrings.Title(otag)) {
								Warnf("unhandled %q", otag)
							} else {
								if strings.EndsWith(otag, "s") {
									r.RemoveLastNewline().Line("}")
								}
								r.RemoveLastNewline().Printf(`Display%sEnd(h)`, otag).Line("")
							}
						}
					}
				} else {
					switch tag {
					case "html":
						if in {
							EndStringBlock(r)
							in = false
						}
						r.Line("").Line("h.End2()")
					case "head":
						if in {
							EndStringBlock(r)
							in = false
						}
						r.RemoveEmptyStringBlock().Line("}").Line("h.HeadEnd2()").Line("")
					default:
						if (otag == stdstrings.ToUpper(otag)) || (otag != stdstrings.Title(otag)) {
							r.WithoutTabs().Printf("</%s>", tag).Backspace()
						} else {
							if in {
								EndStringBlock(r)
								in = false
							}
							if strings.EndsWith(otag, "s") {
								r.RemoveEmptyStringBlock().RemoveLastNewline().Line("}")
							}
							r.RemoveEmptyStringBlock().Printf(`Display%sEnd(h)`, otag).Line("")
						}
					}

				}
			} else {
				var createBlock, extraNewline, customTag, selfClosed bool

				otag, rest, ok := strings.Cut(s, " ")
				tag := stdstrings.ToLower(strings.TrimSpace(otag))

				if !Optimize {
					switch {
					case SliceContains(codeBlocks, tag):
						codeBlock = tag
						fallthrough
					case SliceContains(noAttributesBlock, tag):
						createBlock = true
						fallthrough
					case SliceContains(noAttributesNoBlock, tag):
						r.Printf("h.%sBegin2()", stdstrings.Title(tag))
					case SliceContains(attributesBlock, tag):
						createBlock = true
						fallthrough
					case SliceContains(attributesNoBlock, tag):
						r.Printf(`h.%sBegin2("")`, stdstrings.Title(tag))
					default:
						switch tag {
						case "!doctype":
							r.Line("h.Begin2()\n")
							ok = false
						case "br", "hr":
							r.Printf("h.%s2()", stdstrings.Title(tag))
						case "img":
							r.Printf(`h.%s2("", "")`, stdstrings.Title(tag))
							extraNewline = true
						case "input":
							extraNewline = true
							fallthrough
						case "link", "path":
							r.Printf(`h.%s2("")`, stdstrings.Title(tag))
						case "svg":
							createBlock = true
							fallthrough
						case "text":
							r.Printf(`h.%sBegin2(0, 0)`, stdstrings.Title(tag))
						case "circle":
							r.Printf(`h.%s2(0, 0, 0)`, stdstrings.Title(tag))
						case "line":
							r.Printf(`h.%s2(0, 0, 0, 0)`, stdstrings.Title(tag))
						case "rect":
							r.Printf(`h.%s2(0, 0, 0, 0, 0)`, stdstrings.Title(tag))
						default:
							if (otag == stdstrings.ToUpper(otag)) || (otag != stdstrings.Title(otag)) {
								Warnf("unhandled %q", tag)
								continue
							} else {
								if (strings.EndsWith(otag, "/")) || (strings.EndsWith(rest, "/")) {
									otag, _ = StripIfFound(otag, "", "/")
									r.Printf(`Display%s(h)`, otag)
								} else {
									r.Printf(`Display%sBegin(h)`, otag)
									if strings.EndsWith(otag, "s") {
										createBlock = true
									}
								}
								customTag = true
							}
						}
					}
				} else {
					switch tag {
					case "!doctype":
						if in {
							EndStringBlock(r)
							in = false
						}
						r.RemoveEmptyStringBlock().Line("h.Begin2()\n")
						ok = false
					case "head":
						/* TODO(anton2920): remove later in favor of //gpp:gox: Theme. */
						if in {
							EndStringBlock(r)
							in = false
						}
						r.RemoveEmptyStringBlock().Line("h.HeadBegin2()")
						createBlock = true
					default:
						if (otag == stdstrings.ToUpper(otag)) || (otag != stdstrings.Title(otag)) {
							r.WithoutTabs().Printf("<%s>", tag).Backspace()
						} else {
							if in {
								EndStringBlock(r)
								in = false
							}
							if (strings.EndsWith(otag, "/")) || (strings.EndsWith(rest, "/")) {
								otag, _ = StripIfFound(otag, "", "/")
								r.RemoveEmptyStringBlock().Printf(`Display%s(h)`, otag)
							} else {
								r.RemoveEmptyStringBlock().Printf(`Display%sBegin(h)`, otag)
								if strings.EndsWith(otag, "s") {
									createBlock = true
								}
							}
							customTag = true
						}
					}
				}
				tabs := r.Tabs
				r.Tabs = 0

				s, selfClosed = StripIfFound(rest, "", "/")
				if ok {
					var keys []string
					var done bool

					attrs := make(Attributes)
					if !Optimize {
						r.Backspace()
					}
					for !done {
						/* TODO(anton2920): attributes may be '\n'-separated. */
						attr, rest, ok := ProperCut(s, " ", "\"", "\"", "{{", "}}", "{", "}")
						if !ok {
							done = true
						}

						lval, rval, ok := strings.Cut(attr, "=")
						lval = FixAttr(stdstrings.ToLower(strings.TrimSpace(lval)))
						if !ok {
							attrs[lval] = QuotedString{lval, "true", false, true}
						} else {
							rval, quoted := StripIfFound(strings.TrimSpace(rval), "\"", "\"")
							if !quoted {
								var ok bool
								rval, ok = StripIfFound(rval, "{", "}")
								if !ok {
									Warnf("missing {} around %q", rval)
								}
							}

							switch lval {
							case "classname":
								lval = "class"
							}

							attrs[lval] = QuotedString{lval, rval, quoted, true}
						}

						if len(lval) > 0 {
							keys = append(keys, lval)
						}
						s = rest
					}

					/* Replace empty mandatory attributes with actual values. */
					if !Optimize {
						switch tag {
						case "a", "link":
							r.Backspace(len(`"")`)).Printf(`%s)`, attrs.Get("href")).Backspace()
						case "circle":
							r.Backspace(len(`0, 0, 0)`)).Printf(`%s, %s, %s)`, attrs.Get("cx"), attrs.Get("cy"), attrs.Get("r")).Backspace()
						case "form":
							r.Backspace(len(`"")`)).Printf(`%s)`, attrs.Get("method")).Backspace()
							delete(attrs, "method")
						case "img":
							r.Backspace(len(`"", "")`)).Printf(`%s, %s)`, attrs.Get("alt"), attrs.Get("src")).Backspace()
						case "line":
							r.Backspace(len(`0, 0, 0, 0)`)).Printf(`%s, %s, %s, %s)`, attrs.Get("x1"), attrs.Get("y1"), attrs.Get("x2"), attrs.Get("y2")).Backspace()
						case "svg":
							r.Backspace(len(`0, 0)`)).Printf(`%s, %s)`, attrs.Get("width"), attrs.Get("height")).Backspace()
							delete(attrs, "xmlns")
						case "text":
							r.Backspace(len(`0, 0)`)).Printf(`%s, %s)`, attrs.Get("x"), attrs.Get("y")).Backspace()
						case "path":
							r.Backspace(len(`"")`)).Printf(`%s)`, attrs.Get("d")).Backspace()
						case "rect":
							r.Backspace(len(`0, 0, 0, 0, 0)`)).Printf(`%s, %s, %s, %s, %s)`, attrs.Get("x"), attrs.Get("y"), attrs.Get("width"), attrs.Get("height"), attrs.Get("rx")).Backspace()
						case "input":
							switch attrs.Get("type").Value {
							case "":
								/* Do nothing. */
							default:
								r.Backspace(len(`"")`)).Printf(`%s)`, attrs.Get("type")).Backspace()
							case "checkbox":
								r.Backspace(len(`h.Input2("")`)).Line("h.Checkbox2()").Backspace()
							case "submit":
								r.Backspace(len(`h.Input2("")`)).Printf(`h.Button2(%s)`, attrs.Get("value")).Backspace()
								delete(attrs, "value")
							}
						}
					}

					if !customTag {
						needTranslation := []string{"alt", "placeholder", "value"}
						for _, k := range keys {
							if v, ok := attrs[k]; ok {
								if !Optimize {
									r.Printf(".%s(%s)", stdstrings.Title(k), v).Backspace()
								} else {
									if (SliceContains(needTranslation, k)) || ((!v.Quoted) && (v.Value != "true")) {
										if in {
											EndStringBlock(r).Backspace()
											in = false
										}
										r.Printf(".%s(%s)", stdstrings.Title(k), v).Backspace()
									} else {
										if !in {
											r.Line("")
											r.Tabs = tabs
											r.Line("h.Backspace().String(` ").Backspace()
											r.Tabs = 0
											in = true
										}
										if v.Value == "true" {
											r.WithoutTabs().Backspace().Printf(` %s>`, k).Backspace()
										} else {
											r.WithoutTabs().Backspace().Printf(` %s="%s">`, UnfixAttr(k), v.Value).Backspace()
										}
									}
								}
							}
						}
					} else {
						var appendAfter []string

						r.Backspace(1 + bools.ToInt(Optimize))
						for _, k := range keys {
							switch k {
							case "class", "style":
								appendAfter = append(appendAfter, k)
								continue
							}
							if v, ok := attrs[k]; ok {
								r.Line(", ").Backspace()
								r.Line(v.String()).Backspace()
							}
						}
						r.Line(")").Backspace()

						if len(appendAfter) > 0 {
							for _, k := range appendAfter {
								if v, ok := attrs[k]; ok {
									r.Printf(".%s(%s)", stdstrings.Title(k), v).Backspace()
								}
							}
						}
					}

					if (!Optimize) || (!in) {
						r.Line("")
					}
				}

				if (selfClosed) && (strings.EndsWith(r.Buffer.String(), ">")) {
					r.Backspace().Line("/>").Backspace()
				}
				if extraNewline {
					r.Line("")
				}
				r.Tabs = tabs
				if createBlock {
					r.Line("{")
				}
			}
		}
	}
}

func GenerateGOX(r *Result, p *Parser, fn *Func) {
	GenerateGOXBody(r, r.File.Source[fn.BodyBeginOffset+1:fn.BodyEndOffset-1])
}
