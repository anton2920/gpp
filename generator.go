package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	stdstrings "strings"

	"github.com/anton2920/gofa/strings"
)

type Generator struct {
	Buffer bytes.Buffer
	Tabs   int

	Package   string
	Imports   Imports
	DoImports Imports
}

const GeneratedSuffix = "_gpp"

func (g *Generator) AddImport(s string) {
	var found *Import
	for i := 0; i < len(g.Imports); i++ {
		if (g.Imports[i].QualifiedName == s) || (strings.EndsWith(g.Imports[i].Path, s)) {
			found = &g.Imports[i]
			break
		}
	}

	var insert Import
	if found == nil {
		insert.Path = s
	} else {
		insert = *found
	}

	for _, di := range g.DoImports {
		if ((len(insert.QualifiedName) > 0) && (insert.QualifiedName == di.QualifiedName)) || (insert.Path == di.Path) {
			return
		}
	}
	g.DoImports = append(g.DoImports, insert)
}

func (g *Generator) Dump(w io.Writer) (int64, error) {
	var buf bytes.Buffer
	var total int64

	if len(g.Package) == 0 {
		g.Package = "main"
	}

	fmt.Fprintf(&buf, "/* File generated by \"gpp %s\"; DO NOT EDIT. */\n\n", stdstrings.Join(os.Args[1:], " "))
	fmt.Fprintf(&buf, "package %s\n", g.Package)

	sort.Sort(g.DoImports)

	var newline bool
	if len(g.DoImports) > 0 {
		fmt.Fprintf(&buf, "\nimport ")
		if len(g.DoImports) > 1 {
			buf.WriteString("(\n")
		}
		for i := 0; i < len(g.DoImports); i++ {
			imp := &g.DoImports[i]

			if len(g.DoImports) > 1 {
				if (!newline) && (strings.FindChar(imp.Path, '/') != -1) {
					buf.WriteRune('\n')
					newline = true
				}
				buf.WriteRune('\t')
			}
			if len(imp.QualifiedName) > 0 {
				fmt.Fprintf(&buf, "%s ", imp.QualifiedName)
			}
			fmt.Fprintf(&buf, "\"%s\"\n", imp.Path)
		}
		if len(g.DoImports) > 1 {
			buf.WriteString(")\n")
		}
	}

	n, err := io.Copy(w, bytes.NewReader(buf.Bytes()))
	if err != nil {
		return 0, err
	}
	total += n

	n, err = io.Copy(w, bytes.NewReader(g.Buffer.Bytes()))
	if err != nil {
		return 0, err
	}
	total += n

	return total, nil
}

func (g *Generator) ShouldDump() bool {
	return g.Buffer.Len() != 0
}

func (g *Generator) WriteTabs() {
	for i := 0; i < g.Tabs; i++ {
		g.Buffer.WriteRune('\t')
	}
}

func (g *Generator) Printf(format string, args ...interface{}) (int, error) {
	g.WriteTabs()
	return fmt.Fprintf(&g.Buffer, format, args...)
}

func (g *Generator) Write(b []byte) (int, error) {
	g.WriteTabs()
	return g.Buffer.Write(b)
}

func (g *Generator) WriteRune(r rune) (int, error) {
	g.WriteTabs()
	return g.Buffer.WriteRune(r)
}

func (g *Generator) WriteString(s string) (int, error) {
	g.WriteTabs()
	return g.Buffer.WriteString(s)
}

func GeneratedName(filename string) string {
	ext := filepath.Ext(filename)
	base := filename[:len(filename)-len(ext)]
	return base + GeneratedSuffix + ext
}
