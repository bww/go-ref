package main

import (
  "os"
  "io"
  "fmt"
  "flag"
  "strings"
  "strconv"
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
  jsonTag   = "json"
)

/**
 * Options
 */
type options uint32
const (
  optionNone        = options(0)
  optionPreferIdent = options(1 << 0)
)

/**
 * Context
 */
type context struct {
  Options   options
  Types     typeSet
  Generate  identSet
  Marshal   identSet
}

/**
 * Create a new context
 */
func newContext(opts options) *context {
  return &context{opts, make(typeSet), make(identSet), make(identSet)}
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
  fPreferId   := cmdline.Bool     ("use-id",    false,      "Marshal the identifier instead of the value when both are present.")
  fDebug      := cmdline.Bool     ("debug",     false,      "Enable debugging mode.")
  fVerbose    := cmdline.Bool     ("verbose",   false,      "Be more verbose.")
  cmdline.Parse(os.Args[1:])
  
  DEBUG   = *fDebug
  VERBOSE = *fVerbose
  
  opts := optionNone
  if *fPreferId {
    opts |= optionPreferIdent
  }
  
  cxt := newContext(opts)
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
  
  for _, v := range cxt.Marshal {
    err := genMarshal(cxt, os.Stdout, v)
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
        gen, err := typeExpr(cxt, w, fset, v.Type)
        if err != nil {
          return err
        }
        if gen {
          cxt.Marshal.Add(v.Name)
        }
    }
  }
  return nil
}

func typeExpr(cxt *context, w io.Writer, fset *token.FileSet, e ast.Expr) (bool, error) {
  var err error
  var gen bool
  switch v := e.(type) {
    case *ast.StructType:
      gen, err = structType(cxt, w, fset, v)
      if err != nil {
        return false, err
      }
  }
  return gen, nil
}

func structType(cxt *context, w io.Writer, fset *token.FileSet, s *ast.StructType) (bool, error) {
  var gen bool
  if s.Fields != nil {
    for i, e := range s.Fields.List {
      if e.Tag != nil  && e.Tag.Kind == token.STRING {
        tag, err := strconv.Unquote(e.Tag.Value)
        if err != nil {
          return false, err
        }
        t := reflect.StructTag(tag)
        if ref := t.Get(refTag); ref != "" {
          id, n, err := ident(e.Type, 0)
          if err != nil {
            return false, err
          }
          if !id.IsExported() {
            return false, fmt.Errorf("Field must be exported: %v", id.Name)
          }
          s.Fields.List[i] = &ast.Field{
            Names:e.Names,
            Type:indirect(ast.NewIdent(id.Name + refSuffix), n),
            Comment:e.Comment,
            Tag:e.Tag,
          }
          cxt.Generate.Add(id)
          gen = true
        }
      }
    }
  }
  return gen, nil
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
  
  refId := id.Name + refSuffix
  tspec := fmt.Sprintf(`type %v struct {
  Id    %v
  Value *%v
}

func New%v(v *%v) *%v {
  return &%v{Value:v}
}

func New%vId(v %v) *%v {
  return &%v{Id:v}
}

func (v %v) HasValue() bool {
  return v.Value != nil
}`,
  refId, idType, id.Name,
  refId, id.Name, refId,
  refId,
  refId, idType, refId,
  refId,
  refId)
  
  fmt.Fprint(w, "\n"+ tspec +"\n")
  return nil
}

func genMarshal(cxt *context, w io.Writer, id *ast.Ident) error {
  
  spec, ok := cxt.Types[id.Name]
  if !ok {
    return fmt.Errorf("No type found for: %v", id.Name)
  }
  
  base, ok := spec.Type.(*ast.StructType)
  if !ok {
    return fmt.Errorf("Base type must be a struct: %v", id.Name)
  }
  
  marshal := fmt.Sprintf(`func (v %v) MarshalJSON() ([]byte, error) {
  preferValue := %v // this is a generation-time option
  s := "{"
`,
  id.Name,
  (cxt.Options & optionPreferIdent) != optionPreferIdent)
  
  if base.Fields != nil {
    for _, e := range base.Fields.List {
      
      var jtag, rtag string
      if e.Tag != nil  && e.Tag.Kind == token.STRING {
        tag, err := strconv.Unquote(e.Tag.Value)
        if err != nil {
          return err
        }
        t := reflect.StructTag(tag)
        jtag = t.Get(jsonTag)
        rtag = t.Get(refTag)
      }
      
      if jtag == "-" {
        continue // explicitly ignored
      }else if jtag != "" && len(e.Names) > 1 {
        return fmt.Errorf("Field list has %d identifiers for one tag", len(e.Names))
      }
      
      for _, v := range e.Names {
        id, _, err := ident(v, 0)
        if err != nil {
          return err
        }
        if !id.IsExported() {
          continue // ignore unexported fields
        }
        var f string
        if jtag != "" {
          f = jtag
        }else{
          f = id.Name
        }
        if rtag != "" {
          // fname, vtype := parseTag(rtag)
          marshal += fmt.Sprintf(`  if v.%v != nil {
    if v.%v.HasValue() {
      s += fmt.Sprintf("%%s:", json.Marshal(%q))
      s += json.Marshal(v.%v.Value)
    }else if v.%v.Id != "" {
      s += fmt.Sprintf("%%s:", json.Marshal(%q))
      s += json.Marshal(v.%v.Id)
    }
  }`, id.Name, id.Name, f, id.Name, id.Name, rtag, id.Name) +"\n"
        }else{
          marshal += fmt.Sprintf(`  s += fmt.Sprintf("%%s:", json.Marshal(%q))`, f) +"\n"
        }
      }
      
    }
  }
  
  marshal += `  s += "}"` + "\n"
  marshal += `  return []byte(s), nil` + "\n"
  marshal += `}`
  
  fmt.Fprint(w, "\n"+ marshal +"\n")
  return nil
}

func parseTag(t string) (string, string) {
  if x := strings.Index(t, ","); x > 0 {
    return t[:x], t[x+1:]
  }else{
    return t, ""
  }
}
