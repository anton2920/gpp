package main

type Import struct {
	QualifiedName string
	Path          string
}

type Imports []Import

func (is Imports) Len() int               { return len(is) }
func (is Imports) Less(i int, j int) bool { return is[i].Path < is[j].Path }
func (is Imports) Swap(i int, j int)      { is[i], is[j] = is[j], is[i] }
