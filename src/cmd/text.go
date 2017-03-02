package main

import (
  "io"
  "go/token"
  "go/printer"
)

func indent(t int, s string) string {
  r := spaces(t)
  for _, c := range s {
    r += string(c)
    if c == '\n' {
      r += spaces(t)
    }
  }
  return r
}

func spaces(t int) string {
  var s string
  for i := 0; i < t; i++ {
    s += "  "
  }
  return s
}

func repeat(n int, c rune) string {
  var s string
  for i := 0; i < n; i++ {
    s += string(c)
  }
  return s
}

func printSource(output io.Writer, fset *token.FileSet, node interface{}) error {
  return (&printer.Config{Tabwidth: 2, Mode: printer.UseSpaces}).Fprint(output, fset, node)
}
