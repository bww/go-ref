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
 * Types
 */
type typeSet map[string]*ast.TypeSpec

/**
 * Add a type to the set
 */
func (s typeSet) Add(t *ast.TypeSpec) {
  id := t.Name
  if id == nil {
    panic(fmt.Errorf("Field type must be an identifier: %v", t.Pos()))
  }
  s[id.Name] = t
}

/**
 * Idents
 */
type identSet map[string]*ast.Ident

/**
 * Add an ident to the set
 */
func (s identSet) Add(v *ast.Ident) {
  s[v.Name] = v
}

/**
 * Reference type suffix
 */
const (
  refTag    = "ref"
  refSuffix = "Ref"
  idType    = "string"
)

/**
 * Context
 */
type context struct {
  Types     typeSet
  Generate  identSet
}

/**
 * Create a new context
 */
func newContext() *context {
  return &context{make(typeSet), make(identSet)}
}

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
  
  DEBUG   = *fDebug
  VERBOSE = *fVerbose
  
  cxt := newContext()
  for _, f := range cmdline.Args() {
    err := proc(cxt, os.Stdout, f)
    if err != nil {
      fmt.Printf("%v: could not process: %v\n", CMD, err)
      return
    }
  }
  
  for _, v := range cxt.Generate {
    err := genType(cxt, os.Stdout, v)
    if err != nil {
      fmt.Printf("%v: could not generate: %v\n", CMD, err)
      return
    }
  }
  
}

func proc(cxt *context, w io.Writer, p string) error {
  fset := token.NewFileSet()
  nerr := 0
  
  f, err := parser.ParseFile(fset, p, nil, 0)
  if err != nil {
    return err
  }
  
  ast.Inspect(f, func(n ast.Node) bool {
    switch t := n.(type) {
      case *ast.GenDecl:
        err = typeSpecs(cxt, w, fset, t.Specs)
        if err != nil {
          fmt.Printf("%v: %v: %v\n", CMD, p, err)
          nerr++
        }
    }
    return true
  })
  
  if nerr < 1 {
    printer.Fprint(w, fset, f)
  }
  return nil
}

func typeSpecs(cxt *context, w io.Writer, fset *token.FileSet, s []ast.Spec) error {
  for _, e := range s {
    switch v := e.(type) {
      case *ast.TypeSpec:
        cxt.Types.Add(v)
        err := typeExpr(cxt, w, fset, v.Type)
        if err != nil {
          return err
        }
    }
  }
  return nil
}

func typeExpr(cxt *context, w io.Writer, fset *token.FileSet, e ast.Expr) error {
  switch v := e.(type) {
    case *ast.StructType:
      err := structType(cxt, w, fset, v)
      if err != nil {
        return err
      }
  }
  return nil
}

func structType(cxt *context, w io.Writer, fset *token.FileSet, s *ast.StructType) error {
  if s.Fields != nil {
    for i, e := range s.Fields.List {
      if e.Tag != nil  && e.Tag.Kind == token.STRING {
        t := reflect.StructTag(e.Tag.Value)
        if ref := t.Get(refTag); ref != "" {
          id, n, err := ident(e.Type, 0)
          if err != nil {
            return err
          }
          if !id.IsExported() {
            return fmt.Errorf("Field must be exported: %v", id.Name)
          }
          s.Fields.List[i] = &ast.Field{Names:e.Names, Type:indirect(ast.NewIdent(id.Name + refSuffix), n)}
          cxt.Generate.Add(id)
        }
      }
    }
  }
  return nil
}

func ident(e ast.Expr, r int) (*ast.Ident, int, error) {
  switch v := e.(type) {
   case *ast.Ident:
    return v, r, nil
   case *ast.StarExpr:
    return ident(v.X, r + 1)
   default:
    return nil, -1, fmt.Errorf("Not an identifier: %T", e)
  }
}

func indirect(e *ast.Ident, r int) ast.Expr {
  v := ast.Expr(e)
  for i := 0; i < r; i++ {
    v = &ast.StarExpr{X:v}
  }
  return v
}

func genType(cxt *context, w io.Writer, id *ast.Ident) error {
  
  _, ok := cxt.Types[id.Name]
  if !ok {
    return fmt.Errorf("No base type found for: %v", id.Name)
  }
  
  tspec := fmt.Sprintf(`type %v struct {
  Id    %v
  Value *%v
}`, id.Name + refSuffix, idType, id.Name)
  
  fmt.Fprint(w, tspec +"\n\n")
  return nil
}
