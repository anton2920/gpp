package main

import (
	"bytes"
	"fmt"
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

type (
	Attributes map[string]QuotedString
	Theme      map[string]Attributes
)

type GOXComment struct {
	Theme Theme

	HandleComments bool
	DoNotOptimize  bool
	DoNotInline    bool
}

func (GOXComment) Comment() {}

func (attrs Attributes) Get(key string) QuotedString {
	ret := QuotedString{Key: key}

	v, ok := attrs[key]
	if ok {
		ret = v
	}

	delete(attrs, key)
	return ret
}

var GOXGlobalTheme *GOXComment

var IntAttributes = []string{"min", "max", "minlength", "maxlength", "width", "height", "x", "y", "fontSize", "fontWeight", "strokeWidth", "cx", "cy", "r", "rx", "x1", "x2", "y1", "y2"}

var AppendAttributes = []string{"class", "style"}

func (qv QuotedString) String() string {
	if SliceContains(IntAttributes, qv.Key) {
		qv.Quoted = false
		if !qv.Present {
			return "0"
		}
	}
	if (qv.Quoted) || (!qv.Present) {
		quote := "\""
		if strings.FindChar(qv.Value, '\n') >= 0 {
			quote = "`"
		}
		return quote + qv.Value + quote
	}
	return qv.Value
}

func CustomTag(tag string) bool {
	return tag == stdstrings.Title(tag)
}

func MergeGOXAttributes(a1 Attributes, a2 Attributes) Attributes {
	res := make(Attributes)
	for k, v := range a1 {
		res[k] = v
	}
	for k, v := range a2 {
		rk, ok := res[k]
		if (ok) && (SliceContains(AppendAttributes, k)) && (!SliceContains(stdstrings.Split(rk.Value, " "), v.Value)) {
			res[k] = QuotedString{k, stdstrings.Join([]string{rk.Value, v.Value}, " "), true, true}
		} else {
			res[k] = v
		}
	}
	return res
}

func MergeGOXComments(comments []Comment) GOXComment {
	var gc GOXComment
	for _, comment := range comments {
		if c, ok := comment.(GOXComment); ok {
			if c.Theme != nil {
				if gc.Theme == nil {
					gc.Theme = make(Theme)
				}
				for k := range c.Theme {
					gc.Theme[k] = MergeGOXAttributes(gc.Theme[k], c.Theme[k])
				}
			}
			gc.HandleComments = gc.HandleComments || c.HandleComments
			gc.DoNotOptimize = gc.DoNotOptimize || c.DoNotOptimize
			gc.DoNotInline = gc.DoNotInline || c.DoNotInline
		}
	}
	return gc
}

func ParseGOXComment(comment string, gc *GOXComment) bool {
	var done bool
	for !done {
		s, rest, ok := ProperCut(comment, ",", LBraces, RBraces)
		if !ok {
			done = true
		}

		lval, rval, ok := strings.Cut(s, "=")
		lval = stdstrings.ToLower(strings.TrimSpace(lval))
		if !ok {
			switch lval {
			case "global":
				GOXGlobalTheme = gc
			case "-c":
				gc.HandleComments = true
			case "-n":
				gc.DoNotOptimize = true
			case "-l":
				gc.DoNotInline = true
			}
		} else {
			switch lval {
			case "theme":
				rval, ok := StripIfFound(strings.TrimSpace(rval), LBraces, RBraces)
				if ok {
					var done bool
					for !done {
						entry, rest, ok := ProperCut(rval, ",", "{", "}")
						if !ok {
							done = true
						}
						entry = strings.TrimSpace(entry)

						tag, sattrs, ok := strings.Cut(entry, ":")
						if ok {
							tag = stdstrings.ToLower(tag)
							sattrs, ok := StripIfFound(strings.TrimSpace(sattrs), "{", "}")
							if ok {
								var done bool

								attrs := make(Attributes)
								for !done {
									entry, rest, ok := strings.Cut(sattrs, ",")
									if !ok {
										done = true
									}

									attr, value, ok := strings.Cut(entry, ":")
									if ok {
										attr := FixAttr(stdstrings.ToLower(strings.TrimSpace(attr)))
										value, quoted := StripIfFound(strings.TrimSpace(value), "\"", "\"")
										attrs[attr] = QuotedString{attr, value, quoted, true}
									}

									sattrs = rest
								}

								if gc.Theme == nil {
									gc.Theme = make(Theme)
								}
								gc.Theme[tag] = attrs
							}
						}

						rval = rest
					}
				}
			}
		}

		comment = rest
	}

	return true
}

func SliceContains(xs []string, s string) bool {
	for _, x := range xs {
		if s == x {
			return true
		}
	}
	return false
}

func StartsWithOneOf(s string, xs []string) bool {
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

func IsGOXFunction(fn *Func) bool {
	var foundGen, foundL bool

	if fn != nil {
		for _, comment := range fn.Comments {
			if c, ok := comment.(GenerateComment); ok {
				for _, gen := range c.Generators {
					if _, ok := gen.(GeneratorGOX); ok {
						foundGen = true
					}
				}
			} else if c, ok := comment.(GOXComment); ok {
				if c.DoNotInline {
					foundL = true
				}
			}
		}
	}

	return (foundGen) && (!foundL)
}

func BeginStringBlock(r *Result) *Result {
	return r.Line("h.String(`").Backspace()
}

func EndStringBlock(r *Result) *Result {
	return r.WithoutTabs().Line("`)")
}

func GenerateGOXBody(r *Result, p *Parser, body string, comments []Comment, in bool) {
	const withoutThemeTag = "notheme"
	const cutset = "\t\n"

	var nattrblocks, ncodeblocks int
	var withoutTheme bool
	var codeBlock string

	gc := MergeGOXComments(comments)
	wasIn := in

	for len(body) > 0 {
		body = stdstrings.Trim(body, cutset)

		if len(codeBlock) > 0 {
			needle := "</" + codeBlock + ">"
			end := strings.FindSubstring(body, needle)
			if end == -1 {
				Warnf("unterminated code block %q", codeBlock)
				break
			}

			nattrblocks--
			r.Printf("h.String(`%s`)", strings.TrimSpace(body[:end])).Line("}").Printf("h.%sEnd2()", stdstrings.Title(codeBlock)).Line("")

			codeBlock = ""
			body = body[end+len(needle):]
			continue
		}

		begin := FindTagBegin(body)
		if (begin == -1) || (begin > 0) {
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

				if StartsWithOneOf(text, stmts) {
					newline := strings.FindChar(text, '\n')
					if newline == -1 {
						newline = len(text)
					}
					end := strings.FindCharReverse(text[:newline], '{')
					if end >= 0 {
						tabs := r.Tabs
						if strings.StartsWith(text, "else") {
							r.Backspace(1).Buffer.WriteRune(' ')
							r.Tabs = 0
						}
						r.Line(text[:end+1])
						r.Tabs = tabs + 1
						ncodeblocks++

						otext = text[end+1:]
						continue
					}
				} else if StartsWithOneOf(text, cases) {
					end := strings.FindChar(text, ':')
					if end >= 0 {
						r.Tabs--
						r.RemoveLastNewline().Line(strings.TrimSpace(text[:end+1]))
						otext = text[end+1:]
						continue
					}
				}

				begin := strings.FindChar(otext, '{')
				end := strings.FindChar(otext, '}')
				if end == -1 {
					r.Printf("h.LString(`%s`)", otext)
					break
				} else if (end >= 0) && ((begin == -1) || (end < begin)) {
					trimmed := stdstrings.Trim(otext[:end], cutset)
					if len(trimmed) > 0 {
						r.Printf("h.LString(`%s`)", trimmed)
					}
					text = otext[end:]

					if ncodeblocks == 0 {
						r.Line("h.LString(`}`)")
					} else {
						r.RemoveLastNewline().Line("}")
						ncodeblocks--
					}
					otext = text[1:]
					continue
				} else {
					if strings.StartsWith(otext[begin:], "{{") {
						end = strings.FindSubstring(otext[begin+1:], "}}")
						if end >= 0 {
							end += begin + 1
							for (end+1 < len(otext)) && (otext[end+1] == '}') {
								end += 1
							}
						}
					}
					value := otext[begin : end+1]

					trimmed := stdstrings.Trim(otext[:begin], cutset)
					if len(trimmed) > 0 {
						r.Printf("h.LString(`%s`)", trimmed)
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
			body = strings.TrimSpace(body)

			end := strings.FindChar(body, '>')
			if end == -1 {
				break
			}

			/* Handling '<tag attr={value > 0}/>', where 'greater-than sign' is treated as 'closing bracket'. */
			for {
				lbrace := strings.FindCharReverse(body[:end], '{')
				if lbrace >= 0 {
					rbrace := strings.FindCharReverse(body[lbrace+1:end], '}')
					if rbrace == -1 {
						rbrace = strings.FindChar(body[lbrace+1:], '}')
						if rbrace >= 0 {
							rbrace += lbrace + 1
							end = strings.FindChar(body[rbrace+1:], '>')
							if end == -1 {
								break
							}
							end += rbrace + 1
							continue
						}
					}
				}
				break
			}
			if end == -1 {
				break
			}

			if (!gc.DoNotOptimize) && (!in) {
				BeginStringBlock(r)
				in = true
			}

			if strings.StartsWith(body, "<!--") {
				end = strings.FindSubstring(body, "-->")

				if (gc.DoNotOptimize) && (gc.HandleComments) {
					comment := body[begin+len("<!--") : end]
					lines := stdstrings.Split(comment, "\n")
					if len(lines) == 1 {
						r.Printf("/* %s */", comment)
					} else {
						r.RemoveLastNewline().Line("").Printf("/*")
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
			noAttributesNoBlock := []string{"h1", "h2", "h3", "h4", "h5", "h6", "p", "b", "i", "span", "option", "textarea", "title", "li", "label", "button"}
			attributesBlock := []string{"form"}
			attributesNoBlock := []string{"a"}

			if s, ok := StripIfFound(s, "/", ""); ok {
				otag := strings.TrimSpace(s)
				tag := stdstrings.ToLower(otag)

				if otag == withoutThemeTag {
					withoutTheme = false
					continue
				}

				if gc.DoNotOptimize {
					switch {
					case (SliceContains(noAttributesBlock, tag)) || (SliceContains(attributesBlock, tag)) || (SliceContains([]string{"svg"}, tag)):
						if nattrblocks > 0 {
							nattrblocks--
							r.RemoveLastNewline().Line("}")
						}
						fallthrough
					case (SliceContains(noAttributesNoBlock, tag)) || (SliceContains(attributesNoBlock, tag)) || (SliceContains([]string{"text"}, tag)):
						r.RemoveLastNewline().Printf("h.%sEnd2()", stdstrings.Title(tag)).Line("")
					default:
						switch tag {
						case "html":
							r.Line("").RemoveLastNewline().Line("h.End2()")
						default:
							if !CustomTag(otag) {
								Warnf("unhandled %q", otag)
							} else {
								if strings.EndsWith(otag, "s") {
									if nattrblocks > 0 {
										nattrblocks--
										r.RemoveLastNewline().Line("}")
									}
								}

								name := fmt.Sprintf("Display%sEnd", otag)
								fn := p.FindFunction(r.File.Imports, "main", name)
								if (!gc.DoNotInline) && (IsGOXFunction(fn)) {
									GenerateGOX(r.RemoveLastNewline(), p, fn)
								} else {
									r.RemoveLastNewline().Printf(`%s(h)`, name).Line("")
								}
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
					default:
						if !CustomTag(otag) {
							r.WithoutTabs().Printf("</%s>", tag).Backspace()
						} else {
							name := fmt.Sprintf("Display%sEnd", otag)
							fn := p.FindFunction(r.File.Imports, "main", name)
							if (!gc.DoNotInline) && (IsGOXFunction(fn)) {
								GenerateGOXEx(r.RemoveLastNewline(), p, fn, in)
							} else {
								if in {
									EndStringBlock(r)
									in = false
								}
								if strings.EndsWith(otag, "s") {
									r.RemoveEmptyStringBlock().RemoveLastNewline().Line("}")
								}
								r.RemoveEmptyStringBlock().Printf(`%s(h)`, name).Line("")
							}
						}
					}

				}
			} else {
				var createBlock, extraNewline, customTag, selfClosed bool

				sep := " "
				nl := strings.FindChar(s, '\n')
				if nl >= 0 {
					sep = "\n"
				}

				otag, rest, ok := strings.Cut(s, sep)
				otag = strings.TrimSpace(otag)
				tag := stdstrings.ToLower(otag)

				if otag == withoutThemeTag {
					withoutTheme = true
					continue
				}

				if gc.DoNotOptimize {
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
						tag, _ = StripIfFound(tag, "", "/")
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
						case "error":
							r.Printf(`h.%s2(nil)`, stdstrings.Title(tag))
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
							if !CustomTag(otag) {
								Warnf("unhandled %q", otag)
								continue
							} else {
								var name string

								if (strings.EndsWith(otag, "/")) || (strings.EndsWith(rest, "/")) {
									otag, _ = StripIfFound(otag, "", "/")
									name = fmt.Sprintf(`Display%s`, otag)
								} else {
									name = fmt.Sprintf(`Display%sBegin`, otag)
								}

								fn := p.FindFunction(r.File.Imports, "main", name)
								if (!gc.DoNotInline) && (IsGOXFunction(fn)) && (!ok) {
									GenerateGOX(r, p, fn)
								} else {
									r.Printf("%s(h)", name)
									if strings.EndsWith(otag, "s") {
										createBlock = true
									}
									customTag = true
								}
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
						r.WithoutTabs().Printf("<%s>", tag).Backspace().Line(`<meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0">`).Backspace()
						if (!withoutTheme) && (gc.Theme != nil) {
							if v, ok := gc.Theme["headlink"]; ok {
								r.WithoutTabs().Printf(`<link href=%s rel=%s>`, v.Get("href"), v.Get("rel")).Backspace()
							}
							if v, ok := gc.Theme["headscript"]; ok {
								r.WithoutTabs().Printf(`<script src=%s></script>`, v.Get("src")).Backspace()
							}
						}
					case "error":
						if in {
							EndStringBlock(r)
							in = false
						}
						r.RemoveEmptyStringBlock().Printf("h.%s2(nil)", stdstrings.Title(tag))
					default:
						if !CustomTag(otag) {
							r.WithoutTabs().Printf("<%s>", tag).Backspace()
						} else {
							var name string

							if (strings.EndsWith(otag, "/")) || (strings.EndsWith(rest, "/")) {
								otag, _ = StripIfFound(otag, "", "/")
								name = fmt.Sprintf(`Display%s`, otag)
							} else {
								name = fmt.Sprintf(`Display%sBegin`, otag)
							}

							fn := p.FindFunction(r.File.Imports, "main", name)
							if (!gc.DoNotInline) && (IsGOXFunction(fn)) && (!ok) {
								//runtime.Breakpoint()
								GenerateGOXEx(r, p, fn, in)
							} else {
								if in {
									EndStringBlock(r)
									in = false
								}
								r.RemoveEmptyStringBlock().Printf("%s(h)", name)
								if strings.EndsWith(otag, "s") {
									createBlock = true
								}
								customTag = true
							}
						}
					}
				}
				tabs := r.Tabs
				r.Tabs = 0

				s, selfClosed = StripIfFound(rest, "", "/")

				var keys []string
				attrs := make(Attributes)
				if ok {
					var done bool
					for !done {
						sep := " "
						nl := strings.FindChar(s, '\n')
						if nl >= 0 {
							sep = "\n"
						}

						attr, rest, ok := ProperCut(s, sep, "{{", "}}", "{", "}", "\"", "\"")
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
				}

				etag := tag
				if etag == "input" {
					switch attrs["type"].Value {
					case "checkbox":
						etag = "checkbox"
					case "submit":
						etag = "button"
					}
				}

				if (!withoutTheme) && (gc.Theme != nil) {
					if tattrs, ok := gc.Theme[etag]; ok {
						for k := range tattrs {
							if !SliceContains(keys, k) {
								keys = append(keys, k)
							}
						}
						attrs = MergeGOXAttributes(tattrs, attrs)
					}
				}

				if len(attrs) > 0 {
					if gc.DoNotOptimize {
						r.Backspace()

						/* Replace empty mandatory attributes with actual values. */
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
							typ := attrs.Get("type")
							switch typ.Value {
							case "":
								/* Do nothing. */
							default:
								r.Backspace(len(`"")`)).Printf(`%s)`, typ).Backspace()
							case "checkbox":
								r.Backspace(len(`h.Input2("")`)).Line("h.Checkbox2()").Backspace()
							case "submit":
								r.Backspace(len(`h.Input2("")`)).Printf(`h.Button2(%s)`, attrs.Get("value")).Backspace()
							}
						}
					}
					switch tag {
					case "error":
						r.Backspace(len(`nil)`)+bools.ToInt(!gc.DoNotOptimize)).Printf(`%s)`, attrs.Get("err")).Backspace()
					}

					if customTag {
						var appendAfter []string

						r.Backspace(1 + bools.ToInt(!gc.DoNotOptimize))
						for _, k := range keys {
							if SliceContains(AppendAttributes, k) {
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
					} else {
						needTranslation := []string{"alt", "placeholder", "value"}
						for _, k := range keys {
							if v, ok := attrs[k]; ok {
								if gc.DoNotOptimize {
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
					}

					if (gc.DoNotOptimize) || (!in) {
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
					nattrblocks++
				}
			}
		}
	}

	if (in) && (!wasIn) {
		EndStringBlock(r)
		in = false
	}

	for i := 0; i < nattrblocks; i++ {
		r.RemoveLastNewline().Line("}")
	}
}

func GenerateGOXEx(r *Result, p *Parser, fn *Func, in bool) {
	if fn != nil {
		comments := fn.Comments
		if GOXGlobalTheme != nil {
			comments = append([]Comment{*GOXGlobalTheme}, fn.Comments...)
		}
		GenerateGOXBody(r, p, fn.Body, comments, in)
	}
}

func GenerateGOX(r *Result, p *Parser, fn *Func) {
	GenerateGOXEx(r, p, fn, false)
}
