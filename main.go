package main

import (
	"flag"
	"fmt"
	"go/scanner"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/anton2920/gofa/go/lexer"
	"github.com/anton2920/gofa/strings"
)

type ParsedFile struct {
	Filename string
	Imports  Imports
	Specs    []TypeSpec
}

const GOFA = "github.com/anton2920/gofa/"

var (
	ReferencedPackages = make(map[string]struct{})
	ParsedPackages     = make(map[string][]ParsedFile)
	FileSet            = token.NewFileSet()
	Sources            = make(map[int][]byte)
)

func ReadEntireFile(filename string) ([]byte, error) {
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
		paths = append(paths, ".")
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
		var l lexer.Lexer

		parsedFile.Filename = f.Name()
		l.FileSet = FileSet
		l.Scanner.Init(f, Sources[f.Base()], nil, scanner.ScanComments)

		done := false
		for !done {
			switch l.Curr().GoToken {
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
				if l.Prev().GoToken != token.LPAREN {
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
				}
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
					for k := 0; k < len(spec.Comment.Encodings); k++ {
						format := spec.Comment.Encodings[k]
						format.Serialize(&g, &spec)
						format.Deserialize(&g, &spec)
					}
				}
			}

			if g.ShouldDump() {
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
