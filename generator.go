package main

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"unicode"
)

type Generator struct {
	Buffer     bytes.Buffer
	ShouldDump bool
}

const (
	JSONGeneratorPrefix = "JSONSerialize"
)

func (g *Generator) Dump(w io.Writer) (int64, error) {
	if g.ShouldDump {
		return io.Copy(w, bytes.NewReader(g.Buffer.Bytes()))
	}
	return 0, nil
}

func (g *Generator) Printf(format string, args ...interface{}) (int, error) {
	g.ShouldDump = true
	return fmt.Fprintf(&g.Buffer, format, args...)
}

func (g *Generator) Write(b []byte) (int, error) {
	g.ShouldDump = true
	return g.Buffer.Write(b)
}

func (g *Generator) WriteRune(r rune) (int, error) {
	g.ShouldDump = true
	return g.Buffer.WriteRune(r)
}

func (g *Generator) WriteString(s string) (int, error) {
	g.ShouldDump = true
	return g.Buffer.WriteString(s)
}

func (g *Generator) PackageBegin(name string) {
	g.Printf("package %s\n", name)
	g.ShouldDump = false
}

func GeneratedName(filename string) string {
	ext := filepath.Ext(filename)
	base := filename[:len(filename)-len(ext)]
	return base + "_gpp" + ext
}

/* NOTE(anton2920): this supports only ASCII. */
func VariableName(typeName string, array bool) string {
	var lastUpper int
	for i := 0; i < len(typeName); i++ {
		if unicode.IsUpper(rune(typeName[i])) {
			lastUpper = i
		}
	}

	var suffix string
	if array {
		suffix = "s"
	}

	return fmt.Sprintf("%c%s%s", unicode.ToLower(rune(typeName[lastUpper])), typeName[lastUpper+1:], suffix)
}
