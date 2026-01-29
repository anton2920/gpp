package main

import (
	"fmt"
	"go/scanner"
	"go/token"
	"path/filepath"
)

type File struct {
	Name    string
	Package string

	Imports Imports
	Specs   []TypeSpec
	Funcs   []Func

	Source string
}

func (p *Parser) File(f *token.File, file *File, processedPackages map[string]struct{}, recursive bool) bool {
	src, err := ReadEntireFile(f.Name())
	if err != nil {
		p.Error = fmt.Errorf("failed to read file %q: %v", f.Name(), err)
		return false
	}
	file.Name = f.Name()
	file.Source = string(src)

	p.Scanner.Init(f, src, nil, scanner.ScanComments)
	p.Lexer.Tokens = p.Lexer.Tokens[:0]
	p.Lexer.Position = 0

	done := false
	for !done {
		switch p.Curr().GoToken {
		case token.PACKAGE:
			if !p.Package(&file.Package) {
				p.Error = fmt.Errorf("failed to parse package declaration: %v", p.Error)
				return false
			}
		case token.IMPORT:
			if !p.Imports(&file.Imports) {
				p.Error = fmt.Errorf("failed to parse imports: %v", p.Error)
				return false
			}
			continue
		case token.TYPE:
			/* If not type assertion (i.e. variable.(type)). */
			if p.Prev().GoToken != token.LPAREN {
				var specs []TypeSpec
				if !p.TypeDecl(&specs) {
					p.Error = fmt.Errorf("failed to parse type declarations: %v", p.Error)
					return false
				}
				file.Specs = append(file.Specs, specs...)
				continue
			}
		case token.FUNC:
			if filepath.Ext(file.Name) == ".gox" {
				var fn Func
				if !p.Func(&fn) {
					p.Error = fmt.Errorf("failed to parse func declaration: %v", p.Error)
				}
				fn.Body = file.Source[fn.BodyBeginOffset+1 : fn.BodyEndOffset-1]
				file.Funcs = append(file.Funcs, fn)
			}
		case token.EOF:
			done = true
		}
		p.Next()
	}

	if recursive {
		var paths []string
		for pkg := range p.ReferencedPackages {
			packagePath := file.Imports.PackagePath(pkg)
			if _, ok := processedPackages[packagePath]; !ok {
				paths = append(paths, ResolvePackagePath(packagePath))
				processedPackages[packagePath] = struct{}{}
			}
		}
		if err := PopulateFileSet(p.FileSet, paths); err != nil {
			p.Error = err
			return false
		}
		p.ReferencedPackages = make(map[string]struct{})
	}

	return true
}
