package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/scanner"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/anton2920/gofa/strings"
)

type ParsedFile struct {
	Filename string
	Imports  Imports
	Specs    []TypeSpec
}

const GOFA = "github.com/anton2920/gofa/"

const Stdin = "<stdin>"

var (
	ReferencedPackages = make(map[string]struct{})
	ParsedPackages     = make(map[string][]ParsedFile)
	FileSet            = token.NewFileSet()
	Sources            = make(map[int][]byte)
)

func ReadEntireStdin() ([]byte, error) {
	var buf bytes.Buffer

	if _, err := io.Copy(&buf, os.Stdin); err != nil {
		return nil, fmt.Errorf("failed to read from stdin: %v", err)
	}

	return buf.Bytes(), nil
}

func ReadEntireFile(filename string) ([]byte, error) {
	if filename == Stdin {
		return ReadEntireStdin()
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open %q: %v", filename, err)
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file stat: %v", err)
	}

	contents := make([]byte, int(st.Size()))
	if _, err := io.ReadFull(f, contents); err != nil {
		return nil, fmt.Errorf("failed to read all data from file: %v", err)
	}

	return contents, nil
}

func ParseToken(l *Lexer, expectedTok token.Token) bool {
	if l.Error != nil {
		return false
	}

	if expectedTok != token.COMMENT {
		for l.Peek().GoToken == token.COMMENT {
			l.Next()
		}
	}

	tok := l.Peek()
	if tok.GoToken == expectedTok {
		l.Next()
		return true
	}

	l.Error = fmt.Errorf("%s:%d:%d: expected %q, got %q (%q)", tok.Position.Filename, tok.Position.Line, tok.Position.Column, expectedTok, tok.GoToken, tok.Literal)
	return false
}

func ParseIdentList(l *Lexer, idents *[]string) bool {
	var ident string

	for ParseIdent(l, &ident) {
		*idents = append(*idents, ident)
		if !ParseToken(l, token.COMMA) {
			l.Error = nil
			return true
		}
	}

	return len(*idents) != 0
}

func ParseIdent(l *Lexer, ident *string) bool {
	if ParseToken(l, token.IDENT) {
		*ident = l.Prev().Literal
		return true
	}
	return false
}

func ParseIntLit(l *Lexer, n *int) bool {
	if ParseToken(l, token.INT) {
		var err error
		*n, err = strconv.Atoi(l.Prev().Literal)
		if err != nil {
			l.Error = fmt.Errorf("failed to parse int value: %v", err)
		}
		return err == nil
	}
	return false
}

func ParseStringLit(l *Lexer, s *string) bool {
	if ParseToken(l, token.STRING) {
		*s = l.Prev().Literal
		if ((*s)[0] == '"') || ((*s)[0] == '`') {
			*s = (*s)[1:]
		}
		if ((*s)[len(*s)-1] == '"') || ((*s)[len(*s)-1] == '`') {
			*s = (*s)[:len(*s)-1]
		}
		return true
	}
	return false
}

func Usage() {
	fmt.Fprintf(os.Stderr, "usage: gpp [flags] [path ...]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func Errorf(format string, args ...interface{}) {
	if format[len(format)-1] != '\n' {
		format += "\n"
	}
	fmt.Fprintf(os.Stderr, format, args...)
}

func Fatalf(format string, args ...interface{}) {
	Errorf(format, args...)
	os.Exit(1)
}

func PopulateFileSet(paths []string) error {
	var files []string

	for i := 0; i < len(paths); i++ {
		path := paths[i]
		st, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("failed to stat %q: %v", path, err)
		}

		if !st.IsDir() {
			files = append(files, path)
		} else {
			dir, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open directory %q: %v", path, err)
			}
			defer dir.Close()

			names, err := dir.Readdirnames(-1)
			if err != nil {
				return fmt.Errorf("failed to read names of directory %q entries: %v", path, err)
			}
			for j := 0; j < len(names); j++ {
				name := names[j]
				if (strings.EndsWith(name, ".go")) && (strings.FindSubstring(name, GeneratedSuffix) == -1) {
					files = append(files, filepath.Join(path, name))
				}
			}
		}
	}

	for i := 0; i < len(files); i++ {
		file := files[i]

		src, err := ReadEntireFile(file)
		if err != nil {
			return fmt.Errorf("failed to read entire file %q: %v", file, err)
		}

		base := FileSet.Base()
		FileSet.AddFile(file, base, len(src))
		Sources[base] = src
	}

	return nil
}

func FindPackageName(is Imports, name string) string {
	packageName := name
	for _, i := range is {
		if i.QualifiedName == packageName {
			packageName = filepath.Base(i.Path)
			break
		}
	}
	return packageName
}

func FindPackagePath(is Imports, packageName string) string {
	for _, i := range is {
		if (i.QualifiedName == packageName) || (strings.EndsWith(i.Path, packageName)) {
			return i.Path
		}
	}
	return ""
}

func ResolvePackagePath(path string) string {
	/* Resolve order:
	1. $GOROOT/src
	2. $GOPATH/src
	3. $GOPATH/pkg/mod -- not implemented
	*/
	test := filepath.Join(runtime.GOROOT(), "src", path)
	f, err := os.Open(test)
	if err == nil {
		f.Close()
		return test
	}

	gopath := strings.Or(os.Getenv("GOPATH"), filepath.Join(os.Getenv("HOME"), "go"))
	for {
		part, rest, ok := strings.Cut(gopath, ":")
		test := filepath.Join(part, "src", path)
		f, err := os.Open(test)
		if err == nil {
			f.Close()
			return test
		}
		if !ok {
			break
		}
		gopath = rest
	}

	return ""
}

func main() {
	listFiles := flag.Bool("l", false, "list files which gpp processed")
	flag.Usage = Usage
	flag.Parse()

	paths := flag.Args()
	if len(paths) == 0 {
		paths = append(paths, Stdin)
	}
	if err := PopulateFileSet(paths); err != nil {
		Fatalf("Failed to process arguments: %v", err)
	}

	processedPackages := make(map[string]struct{})
	FileSet.Iterate(func(f *token.File) bool {
		var comment *Comment
		var parsedFile ParsedFile
		var packageName string
		var paths []string
		var l Lexer

		parsedFile.Filename = f.Name()
		l.Scanner.Init(f, Sources[f.Base()], nil, scanner.ScanComments)

		done := false
		for !done {
			switch l.Peek().GoToken {
			case token.PACKAGE:
				if !ParsePackage(&l, &packageName) {
					Errorf("Failed to parse package: %v", l.Error)
					return false
				}
			case token.IMPORT:
				if !ParseImports(&l, &parsedFile.Imports) {
					Errorf("Failed to parse imports: %v", l.Error)
					return false
				}
				continue
			case token.COMMENT:
				var c Comment
				if ParseGofaComment(&l, &c) {
					comment = &c
					continue
				}
				l.Error = nil
			case token.TYPE:
				var specs []TypeSpec
				if !ParseTypeDecl(&l, &specs) {
					Errorf("Failed to parse type declarations: %v", l.Error)
					return false
				}
				if comment != nil {
					for i := 0; i < len(specs); i++ {
						spec := &specs[i]
						spec.Comment = comment
					}
					comment = nil
				}
				parsedFile.Specs = append(parsedFile.Specs, specs...)
				continue
			case token.EOF:
				done = true
			}
			comment = nil
			l.Next()
		}

		for packageName := range processedPackages {
			delete(ReferencedPackages, packageName)
		}
		for name := range ReferencedPackages {
			packageName := FindPackageName(parsedFile.Imports, name)
			if _, ok := processedPackages[packageName]; ok {
				continue
			}
			paths = append(paths, ResolvePackagePath(FindPackagePath(parsedFile.Imports, packageName)))
			processedPackages[packageName] = struct{}{}
		}
		if err := PopulateFileSet(paths); err != nil {
			Errorf("Failed to process additional packages: %v", err)
			return false
		}

		ParsedPackages[packageName] = append(ParsedPackages[packageName], parsedFile)
		return true
	})

	for packageName, parsedFiles := range ParsedPackages {
		for _, parsedFile := range parsedFiles {
			var g Generator
			g.Package = packageName
			g.Imports = parsedFile.Imports

			if _, ok := processedPackages[packageName]; ok {
				continue
			}

			filename := parsedFile.Filename
			specs := parsedFile.Specs

			for _, spec := range specs {
				if spec.Comment != nil {
					for k := 0; k < len(spec.Comment.Formats); k++ {
						format := spec.Comment.Formats[k]
						format.Serialize(&g, &spec)
					}
				}
			}

			if g.ShouldDump() {
				if filename == Stdin {
					g.Dump(os.Stdout)
				} else {
					name := GeneratedName(filename)
					file, err := os.Create(name)
					if err != nil {
						Errorf("Failed to create generated file %q: %v", name, err)
						continue
					}
					g.Dump(file)
					file.Close()

					if *listFiles {
						fmt.Println(filename)
					}
				}
			}
		}
	}

}
