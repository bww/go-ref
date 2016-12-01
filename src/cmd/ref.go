package main

import (
  "os"
  "io"
  "fmt"
  "flag"
  "strings"
  "reflect"
  "go/ast"
  "go/token"
  "go/parser"
  "go/printer"
)

var (
  DEBUG bool
  VERBOSE bool
  CMD string
)

/**
 * You know what it does
 */
func main() {
  
  if x := strings.LastIndex(os.Args[0], "/"); x > -1 {
    CMD = os.Args[0][x+1:]
  }else{
    CMD = os.Args[0]
  }
  
  cmdline     := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
  fDebug      := cmdline.Bool     ("debug",       false,      "Enable debugging mode.")
  fVerbose    := cmdline.Bool     ("verbose",     false,      "Be more verbose.")
  cmdline.Parse(os.Args[1:])
  
  DEBUG = *fDebug
  VERBOSE = *fVerbose
  
  for _, f := range cmdline.Args() {
    err := proc(os.Stdout, f)
    if err != nil {
      fmt.Printf("%v: could not process: %v\n", CMD, err)
      continue
    }
  }
  
}

func proc(w io.Writer, p string) error {
  fset := token.NewFileSet()
  
  // Parse the file containing this very example
  f, err := parser.ParseFile(fset, p, nil, 0)
  if err != nil {
    return err
  }
  
  // Print the imports from the file's AST.
  ast.Inspect(f, func(n ast.Node) bool {
    switch t := n.(type) {
      case *ast.GenDecl:
        typeSpecs(w, fset, t.Specs)
    }
    return true
  })
  
  printer.Fprint(w, fset, f)
  return nil
}

func typeSpecs(w io.Writer, fset *token.FileSet, s []ast.Spec) error {
  for _, e := range s {
    switch v := e.(type) {
      case *ast.TypeSpec:
        typeExpr(w, fset, v.Type)
    }
  }
  return nil
}

func typeExpr(w io.Writer, fset *token.FileSet, e ast.Expr) error {
  switch v := e.(type) {
    case *ast.StructType:
      structType(w, fset, v)
  }
  return nil
}

func structType(w io.Writer, fset *token.FileSet, s *ast.StructType) error {
  if s.Fields != nil {
    for i, e := range s.Fields.List {
      if e.Tag != nil  && e.Tag.Kind == token.STRING {
        t := reflect.StructTag(e.Tag.Value)
        if ref := t.Get("ref"); ref != "" {
          fmt.Println("REF!", i, ref)
          s.Fields.List[i] = &ast.Field{Names:e.Names, Type:&ast.InterfaceType{Methods:&ast.FieldList{}}}
        }
      }
    }
  }
  return nil
}
