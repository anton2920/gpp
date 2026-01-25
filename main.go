package main

import (
	"flag"
	"fmt"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/anton2920/gofa/bools"
	"github.com/anton2920/gofa/strings"
)

const GOFA = "github.com/anton2920/gofa/"

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
	Errorf("usage: gpp [flags] [path ...]")
	flag.PrintDefaults()
	os.Exit(2)
}

func Warnf(format string, args ...interface{}) {
	Errorf("WARNING: "+format, args...)
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

func PopulateFileSet(fs *token.FileSet, paths []string) error {
	for i := 0; i < len(paths); i++ {
		path := paths[i]

		file, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("failed to stat %q: %v", path, err)
		}

		if !file.IsDir() {
			fs.AddFile(path, fs.Base(), int(file.Size()))
		} else {
			dir, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open directory %q: %v", path, err)
			}
			defer dir.Close()

			files, err := dir.Readdir(-1)
			if err != nil {
				return fmt.Errorf("failed to read names of directory %q entries: %v", path, err)
			}
			for j := 0; j < len(files); j++ {
				file := files[j]
				name := file.Name()

				if ((strings.EndsWith(name, ".go")) || strings.EndsWith(name, ".gox")) && (strings.FindSubstring(name, GeneratedSuffix) == -1) && (strings.FindSubstring(name, "_gox") == -1) {
					fs.AddFile(filepath.Join(path, name), fs.Base(), int(file.Size()))
				}
			}
		}
	}

	return nil
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
	recursive := flag.Bool("r", false, "process referenced imported packages")
	flag.Usage = Usage
	flag.Parse()

	_ = listFiles

	p := NewParser(token.NewFileSet())

	paths := flag.Args()
	if len(paths) == 0 {
		paths = append(paths, ".")
	}
	if err := PopulateFileSet(p.FileSet, paths); err != nil {
		Fatalf("Failed to process arguments: %v", err)
	}

	processedPackages := make(map[string]struct{})
	p.FileSet.Iterate(func(f *token.File) bool {
		var file File

		p.File(f, &file, processedPackages, *recursive)
		if p.Error != nil {
			p.Error = fmt.Errorf("failed to parse file %q: %v", f.Name(), p.Error)
			return false
		}

		p.Packages[file.Package] = append(p.Packages[file.Package], file)
		return true
	})
	if p.Error != nil {
		Fatalf("%v", p.Error)
	}

	for _, files := range p.Packages {
		for i := 0; i < len(files); i++ {
			file := &files[i]
			fileName := file.Name

			r := Result{File: file}
			for _, spec := range file.Specs {
				if len(spec.Comments) > 0 {
					for _, comment := range spec.Comments {
						switch comment := comment.(type) {
						case ImportComment:
							r.AddImport(comment.Path)
						}
					}
					for _, comment := range spec.Comments {
						switch comment := comment.(type) {
						case GenerateComment:
							for _, g := range comment.Generators {
								Generate(g, &r, &p, &spec)
							}
						}
					}
				}
			}
			if filepath.Ext(fileName) == ".gox" {
				var offset int
				for _, fn := range file.Funcs {
					r.Tabs = 0
					r.Line(r.File.Source[offset:fn.BeginOffset]).Backspace()
					if len(fn.Comments) == 0 {
						r.Line(r.File.Source[fn.BeginOffset:fn.BodyEndOffset]).Backspace()
					} else {
						r.Line(r.File.Source[fn.BeginOffset : fn.BodyBeginOffset+1])
						for _, comment := range fn.Comments {
							switch comment.(type) {
							case GenerateComment:
								GenerateGOX(&r, &p, &fn)
							}
						}
						r.Line("}")
					}
					offset = fn.BodyEndOffset + bools.ToInt(len(fn.Comments) > 0)
				}
			}

			if r.ShouldDump() {
				name := GeneratedName(fileName)
				file, err := os.Create(name)
				if err != nil {
					Errorf("Failed to create generated file %q: %v", name, err)
					continue
				}
				r.Dump(file)
				file.Close()

				if *listFiles {
					fmt.Println(fileName)
				}
			}
		}
	}
}
