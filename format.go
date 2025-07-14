package main

type Format interface {
	Generate(*Generator, *Struct)
}
