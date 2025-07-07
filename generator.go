package main

import "io"

type Generator func(io.Writer, *Struct)

func SerialGenerator(w io.Writer, s *Struct) {
	w.Write([]byte("TODO(anton2920): serial generator\n"))
}

func JSONGenerator(w io.Writer, s *Struct) {
	w.Write([]byte("TODO(anton2920): JSON generator\n"))
}
