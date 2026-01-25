package main

import (
	"bytes"
	stdstrings "strings"
	"unicode"

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
	Value   string
	Quoted  bool
	Present bool
}

const HandleComments = true

func (qv QuotedString) String() string {
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

func GenerateGOXBody(r *Result, body string) {
	var nblocks int

	for len(body) > 0 {
		var i int

		for i = 0; i < len(body); i++ {
			if !unicode.IsSpace(rune(body[i])) {
				break
			}
		}
		body = body[i:]

		begin := FindTagBegin(body)
		if (begin == -1) || (begin > 0) {
			stmts := []string{"if", "else", "for", "switch"}
			cases := []string{"case", "default"}

			var otext string
			if begin == -1 {
				otext = strings.TrimSpace(body)
				body = ""
			} else {
				otext = strings.TrimSpace(body[:begin])
				body = body[begin:]
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

				begin := strings.FindSubstring(text, "{")
				end := strings.FindSubstring(text, "}")
				if end == -1 {
					r.Printf("h.LString(`%s`)", text)
					break
				} else if (end >= 0) && ((begin == -1) || (end < begin)) {
					if len(strings.TrimSpace(text[:end])) > 0 {
						r.Printf("h.LString(`%s`)", strings.TrimSpace(text[:end]))
					}
					text = text[end:]

					if nblocks == 0 {
						r.Line("h.LString(`}`)")
					} else {
						r.RemoveLastNewline().Line("}")
						nblocks--
					}
					otext = text[1:]
					continue
				} else {
					if strings.StartsWith(text[begin:], "{{") {
						for (end+1 < len(text)) && (text[end+1] == '}') {
							end += 1
						}
					}
					value := text[begin : end+1]

					if len(strings.TrimSpace(text[:begin])) > 0 {
						r.Printf("h.LString(`%s`)", text[:begin])
					}
					if value, ok := StripIfFound(value, "{{", "}}"); ok {
						r.RemoveLastNewline().Line(value)
					} else if value, ok := StripIfFound(value, "{", "}"); ok {
						r.Printf("h.LString(%s)", strings.TrimSpace(value))
					}

					otext = text[end+1:]
				}
			}
		} else {
			end := strings.FindChar(body, '>')
			if end == -1 {
				break
			}

			if strings.StartsWith(body, "<!--") {
				end = strings.FindSubstring(body, "-->")

				if HandleComments {
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

			noAttributesBlock := []string{"head", "body", "div", "select", "ol", "ul"}
			noAttributesNoBlock := []string{"h1", "h2", "h3", "h4", "h5", "h6", "p", "b", "i", "span", "option", "textarea", "title", "li", "label"}
			attributesBlock := []string{"form"}
			attributesNoBlock := []string{"a"}

			if s, ok := StripIfFound(s, "/", ""); ok {
				otag := strings.TrimSpace(s)
				tag := stdstrings.ToLower(otag)
				switch {
				case (SliceContains(noAttributesBlock, tag)) || (SliceContains(attributesBlock, tag)) || (SliceContains([]string{"svg"}, tag)):
					r.RemoveLastNewline().Line("}")
					fallthrough
				case (SliceContains(noAttributesNoBlock, tag)) || (SliceContains(attributesNoBlock, tag)) || (SliceContains([]string{"text"}, tag)):
					r.RemoveLastNewline().Printf("h.%sEnd2()", stdstrings.Title(tag)).Line("")
				default:
					switch tag {
					case "html":
						r.Line("h.End()")
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
				var createBlock, extraNewline, customTag bool

				otag, rest, ok := strings.Cut(s, " ")
				tag := stdstrings.ToLower(strings.TrimSpace(otag))

				switch {
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
						r.Line("h.Begin2()").Line("")
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
				tabs := r.Tabs
				r.Tabs = 0

				s, _ = StripIfFound(rest, "", "/")
				if ok {
					var keys []string
					var newr Result
					var done bool

					attrs := make(map[string]QuotedString)
					r.Backspace()
					for !done {
						attr, rest, ok := ProperCut(s, " ", "\"", "\"", "{{", "}}", "{", "}")
						if !ok {
							done = true
						}

						lval, rval, ok := strings.Cut(attr, "=")
						lval = FixAttr(stdstrings.ToLower(strings.TrimSpace(lval)))
						if !ok {
							attrs[lval] = QuotedString{"true", false, true}
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
							case "minlength", "maxlength", "width", "height", "x", "y", "fontSize", "fontWeight", "strokeWidth", "cx", "cy", "r", "rx", "x1", "x2", "y1", "y2":
								quoted = false
							}

							attrs[lval] = QuotedString{rval, quoted, true}
						}

						if len(lval) > 0 {
							keys = append(keys, lval)
						}
						s = rest
					}
					delete(attrs, "xmlns")

					/* Replace empty mandatory attribute with actual value. */
					switch tag {
					case "a", "link":
						r.Backspace(len(`"")`)).Printf(`%s)`, attrs["href"]).Backspace()
						delete(attrs, "href")
					case "circle":
						if cx, ok1 := attrs["cx"]; ok1 {
							if cy, ok2 := attrs["cy"]; ok2 {
								if cr, ok3 := attrs["r"]; ok3 {
									r.Backspace(len(`0, 0, 0)`)).Printf(`%s, %s, %s)`, cx, cy, cr).Backspace()
									delete(attrs, "cx")
									delete(attrs, "cy")
									delete(attrs, "r")
								}
							}
						}
					case "form":
						r.Backspace(len(`"")`)).Printf(`%s)`, attrs["method"]).Backspace()
						delete(attrs, "method")
					case "img":
						r.Backspace(len(`"", "")`)).Printf(`%s, %s)`, attrs["alt"], attrs["src"]).Backspace()
						delete(attrs, "alt")
						delete(attrs, "src")
					case "line":
						if x1, ok1 := attrs["x1"]; ok1 {
							if y1, ok2 := attrs["y1"]; ok2 {
								if x2, ok3 := attrs["x2"]; ok3 {
									if y2, ok4 := attrs["y2"]; ok4 {
										r.Backspace(len(`0, 0, 0, 0)`)).Printf(`%s, %s, %s, %s)`, x1, y1, x2, y2).Backspace()
										delete(attrs, "x1")
										delete(attrs, "y1")
										delete(attrs, "x2")
										delete(attrs, "y2")
									}
								}
							}
						}
					case "svg":
						if width, ok1 := attrs["width"]; ok1 {
							if height, ok2 := attrs["height"]; ok2 {
								r.Backspace(len(`0, 0)`)).Printf(`%s, %s)`, width, height).Backspace()
								delete(attrs, "width")
								delete(attrs, "height")
							}
						}
					case "text":
						if x, ok1 := attrs["x"]; ok1 {
							if y, ok2 := attrs["y"]; ok2 {
								r.Backspace(len(`0, 0)`)).Printf(`%s, %s)`, x, y).Backspace()
								delete(attrs, "x")
								delete(attrs, "y")
							}
						}
					case "path":
						r.Backspace(len(`"")`)).Printf(`%s)`, attrs["d"]).Backspace()
						delete(attrs, "d")
					case "rect":
						if x, ok1 := attrs["x"]; ok1 {
							if y, ok2 := attrs["y"]; ok2 {
								if width, ok3 := attrs["width"]; ok3 {
									if height, ok4 := attrs["height"]; ok4 {
										if rx, ok5 := attrs["rx"]; ok5 {
											r.Backspace(len(`0, 0, 0, 0, 0)`)).Printf(`%s, %s, %s, %s, %s)`, x, y, width, height, rx).Backspace()
											delete(attrs, "x")
											delete(attrs, "y")
											delete(attrs, "width")
											delete(attrs, "height")
											delete(attrs, "rx")
										}
									}
								}
							}
						}
					case "input":
						switch attrs["type"].Value {
						case "":
							/* Do nothing. */
						default:
							r.Backspace(len(`"")`)).Printf(`%s)`, attrs["type"]).Backspace()
						case "checkbox":
							r.Backspace(len(`h.Input2("")`)).Line("h.Checkbox2()").Backspace()
						case "submit":
							r.Backspace(len(`h.Input2("")`)).Printf(`h.Button2(%s)`, attrs["value"]).Backspace()
							delete(attrs, "value")
						}
						delete(attrs, "type")
					}

					if !customTag {
						for _, k := range keys {
							if v, ok := attrs[k]; ok {
								newr.Printf(".%s(%s)", stdstrings.Title(k), v).Backspace()
							}
						}
					} else {
						class := -1

						r.Backspace()
						for i, k := range keys {
							if k == "class" {
								class = i
								continue
							}
							if v, ok := attrs[k]; ok {
								newr.Line(", ").Backspace()
								newr.Line(v.String()).Backspace()
							}
						}
						newr.Line(")").Backspace()

						if class >= 0 {
							newr.Printf(".Class(%s)", attrs[keys[class]]).Backspace()
						}
					}

					r.Line(newr.Buffer.String())
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
