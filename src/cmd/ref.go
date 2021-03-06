package main

import (
  "os"
  "io"
  "fmt"
  "flag"
  "path"
  "strings"
  "strconv"
  "reflect"
  "go/ast"
  "go/token"
  "go/parser"
)

var (
  DEBUG bool
  TRACE bool
  VERBOSE bool
  FORCE bool
  CMD string
)

/**
 * Imports
 */
type importSet map[string]*ast.ImportSpec

/**
 * Add an import to the set
 */
func (s importSet) Add(t *ast.ImportSpec) {
  s[importPackage(t)] = t
}

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
type identSet map[string]*ident

/**
 * Add an ident to the set
 */
func (s identSet) Add(id *ident) {
  s[id.Name] = id
}

/**
 * Reference type suffix
 */
const (
  refSuffix     = "Ref"
  idSuffix      = "Id"
)

/**
 * Macros
 */
const (
  macro         = "+goref"
  macroIgnore   = "ignore"
)

var (
  idType        = "string"
  buildTag      = ""
  fileSuffix    = "_ref"
  pkgSrc        = "pkg"
)

var (
  stripComments = true
  extraImports importSet
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
 * Source
 */
type source struct {
  Generate  int
}

/**
 * Context
 */
type context struct {
  Package   string
  Options   options
  Imports   importSet
  Deps      importSet
  Types     typeSet
  Generate  identSet
  Marshal   identSet
  Lookup    map[string]*ident
}

/**
 * Create a new context
 */
func newContext(pkg string, extra importSet, opts options) *context {
  if extra == nil {
    extra = make(importSet)
  }
  return &context{pkg, opts, make(importSet), extra, make(typeSet), make(identSet), make(identSet), make(map[string]*ident)}
}

/**
 * You know what it does
 */
func main() {
  var imports flagList
  
  if x := strings.LastIndex(os.Args[0], "/"); x > -1 {
    CMD = os.Args[0][x+1:]
  }else{
    CMD = os.Args[0]
  }
  
  cmdline         := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
  fIdent          := cmdline.String   ("ident",           "string",   "The type to use for generated identifiers.")
  fBuildTag       := cmdline.String   ("build-tag",       "",         "Specify a Go build tag to be emitted in generated files.")
  fFileSuffix     := cmdline.String   ("file-suffix",     "_ref",     "Specify the suffix to append to generated filenames.")
  fStripComments  := cmdline.Bool     ("strip-comments",  true,       "Strip out build tags (and anything else in leading/doc comments).")
  fForce          := cmdline.Bool     ("force",           false,      "Generate all files, including those which are not out-of-date.")
  fDebug          := cmdline.Bool     ("debug",           false,      "Enable debugging mode.")
  fTrace          := cmdline.Bool     ("trace",           false,      "Trace out (un)marshaled data.")
  fVerbose        := cmdline.Bool     ("verbose",         false,      "Be more verbose.")
  cmdline.Var      (&imports,          "import",                      "Consider the provided package for import.")
  cmdline.Parse(os.Args[1:])
  
  DEBUG           = *fDebug
  TRACE           = *fTrace
  VERBOSE         = *fVerbose
  FORCE           = *fForce
  idType          = *fIdent
  buildTag        = *fBuildTag
  fileSuffix      = *fFileSuffix
  stripComments   = *fStripComments
  
  if len(imports) > 0 {
    extraImports = make(importSet)
    for _, e := range imports {
      extraImports.Add(&ast.ImportSpec{Path:&ast.BasicLit{Kind:token.STRING, Value:strconv.Quote(e)}})
    }
  }
  
  opts := optionNone
  for _, f := range cmdline.Args() {
    
    info, err := os.Stat(f)
    if err != nil {
      fmt.Printf("%v: %v\n", CMD, err)
      return
    }
    
    if !info.IsDir() {
      fmt.Printf("%v: not a package; skipping: %v\n", CMD, f)
      continue
    }
    
    err = procDir(f, opts)
    if err != nil {
      fmt.Printf("%v: %v\n", CMD, err)
      return
    }
    
  }
  
}

func procDir(dir string, opts options) error {
  fset := token.NewFileSet()
  
  excludeGenerated := func(info os.FileInfo) bool {
    return !strings.HasSuffix(info.Name(), fileSuffix +".go")
  }
  
  pkgs, err := parser.ParseDir(fset, dir, excludeGenerated, parser.ParseComments)
  if err != nil {
    return err
  }
  
  for pname, pkg := range pkgs {
    err := procPackage(newContext(pname, extraImports, opts), fset, dir, pkg)
    if err != nil {
      return err
    }
  }
  
  return nil
}

func procPackage(cxt *context, fset *token.FileSet, dir string, pkg *ast.Package) error {
  var err error
  
  for fname, file := range pkg.Files {
    src, dst := fname, refFile(fname)
    var ood bool
    if DEBUG || FORCE {
      ood = true // always out of date for debug or force-generate
    }else{
      ood, err = isFileOutOfDate(dst, src)
      if err != nil {
        return err
      }
    }
    if ood {
      err := procAST(cxt, fset, pkg.Name, src, dst, file)
      if err != nil {
        return err
      }
    }
  }
  
  if len(cxt.Generate) > 0 || len(cxt.Marshal) > 0 {
    outpkg := path.Join(dir, pkgSrc + fileSuffix +".go")
    out, err := refWriter(outpkg)
    if err != nil {
      return err
    }
    
    if buildTag != "" {
      fmt.Fprintf(out, "// %s\n\n", buildTag)
    }
    
    fmt.Fprintf(out, `// This file was generated by Go-Ref. Changes will be overwritten.
// %v
package %v

import (
  ref_fmt "fmt"
  ref_reflect "reflect"
  ref_json "encoding/json"
)
`, outpkg, cxt.Package)
    
    if len(cxt.Deps) > 0 {
      fmt.Fprintf(out, "\n// Dependency imports\nimport (\n")
      for _, e := range cxt.Deps {
        fmt.Fprintf(out, "  ")
        printSource(out, fset, e)
        fmt.Fprintf(out, "\n")
      }
      fmt.Fprintf(out, ")\n")
    }
    
    for _, v := range cxt.Generate {
      err := genType(cxt, out, fset, v)
      if err != nil {
        return err
      }
    }
    
    for _, v := range cxt.Marshal {
      err := genMarshal(cxt, out, fset, v)
      if err != nil {
        return err
      }
      err = genUnmarshal(cxt, out, fset, v)
      if err != nil {
        return err
      }
    }
    
    routines := `
func isEmptyValue(v ref_reflect.Value) bool {
  switch v.Kind() {
  case ref_reflect.Array, ref_reflect.Map, ref_reflect.Slice, ref_reflect.String:
    return v.Len() == 0
  case ref_reflect.Bool:
    return !v.Bool()
  case ref_reflect.Int, ref_reflect.Int8, ref_reflect.Int16, ref_reflect.Int32, ref_reflect.Int64:
    return v.Int() == 0
  case ref_reflect.Uint, ref_reflect.Uint8, ref_reflect.Uint16, ref_reflect.Uint32, ref_reflect.Uint64, ref_reflect.Uintptr:
    return v.Uint() == 0
  case ref_reflect.Float32, ref_reflect.Float64:
    return v.Float() == 0
  case ref_reflect.Interface, ref_reflect.Ptr:
    return v.IsNil()
  }
  return false
}
`
    fmt.Fprint(out, routines)
  }
  
  return nil
}

func procAST(cxt *context, fset *token.FileSet, pkg, src, dst string, file *ast.File) error {
  pkgrefs := make(map[string]int)
  fcxt := &source{}
  nerr := 0
  
  // check the first line comment group for macro directives
  if len(file.Comments) > 0 {
    for _, e := range file.Comments[0].List {
      c, t := args(commentText(e))
      if c == macro {
        c, t = args(t)
        if c == macroIgnore {
          if VERBOSE {
            fmt.Printf("%v: skipping ignored source: %v\n", CMD, src)
          }
          return nil
        }
      }
    }
  }
  
  // traverse the source first to handle types
  ast.Inspect(file, func(n ast.Node) bool {
    switch t := n.(type) {
      case *ast.GenDecl:
        err := typeSpecs(cxt, fcxt, fset, t.Specs)
        if err != nil {
          fmt.Printf("%v: %v: %v\n", CMD, src, err)
          nerr++
        }
    }
    return true
  })
  
  if nerr < 1 && fcxt.Generate > 0 {
    
    // traverse the source a second time to compile a set of package references
    ast.Inspect(file, func(n ast.Node) bool {
      switch t := n.(type) {
        case *ast.Ident:
          c := pkgrefs[t.Name]
          c++
          pkgrefs[t.Name] = c
          return false
        // case *ast.SelectorExpr:
        //   e := leftmost(t)
        //   if id, ok := e.(*ast.Ident); ok {
        //     n := pkgrefs[id.Name]
        //     n++
        //     pkgrefs[id.Name] = n
        //   }
        //   return false
      }
      return true
    })
    
    // trim unused imports from decls (this is what's actually used by the printer)
    // file.Imports seems to just be a higher-level convenience, which we don't bother with
    for _, e := range file.Decls {
      if g, ok := e.(*ast.GenDecl); ok {
        if g.Tok == token.IMPORT {
          for i, s := range g.Specs {
            if m, ok := s.(*ast.ImportSpec); ok {
              p := importPackage(m)
              if _, ok := pkgrefs[p]; !ok {
                g.Specs[i] = &ast.ImportSpec{
                  Path: &ast.BasicLit{
                    ValuePos: m.Path.ValuePos,
                    Kind: m.Path.Kind,
                    Value: "// "+ m.Path.Value,
                  },
                }
              }
            }
          }
        }
      }
    }
    
    w, err := refWriter(dst)
    if err != nil {
      return err
    }
    
    if buildTag != "" {
      fmt.Fprintf(w, "// %s\n\n", buildTag)
    }
    
    fmt.Fprintf(w, strings.TrimSpace(`
// This file was generated by Go-Ref from the source file:
// > %v
// Changes will be overwritten.
    `) +"\n", src)
    
    // strip comments if necessary
    if stripComments {
      file.Comments = nil
    }
    
    printSource(w, fset, file)
  }
  return nil
}

func typeSpecs(cxt *context, src *source, fset *token.FileSet, s []ast.Spec) error {
  for _, e := range s {
    switch v := e.(type) {
      case *ast.ImportSpec:
        cxt.Imports.Add(v)
      case *ast.TypeSpec:
        cxt.Types.Add(v)
        gen, err := typeExpr(cxt, src, fset, v.Type)
        if err != nil {
          return err
        }
        if gen {
          cxt.Marshal.Add(astIdent(v.Name))
        }
    }
  }
  return nil
}

func typeExpr(cxt *context, src *source, fset *token.FileSet, e ast.Expr) (bool, error) {
  var err error
  var gen bool
  switch v := e.(type) {
    case *ast.StructType:
      gen, err = structType(cxt, src, fset, v)
      if err != nil {
        return false, err
      }
  }
  return gen, nil
}

func structType(cxt *context, src *source, fset *token.FileSet, s *ast.StructType) (bool, error) {
  deps := make(importSet)
  var gen bool
  
  if s.Fields != nil {
    for i, e := range s.Fields.List {
      
      var x ast.Expr
      if c, ok := e.Type.(*ast.StarExpr); ok {
        x = deref(c, 0)
      }else{
        x = e.Type
      }
      
      if c, ok := x.(*ast.SelectorExpr); ok {
        v := leftmost(c)
        if p, ok := v.(*ast.Ident); ok {
          m, ok := cxt.Imports[p.Name]
          if !ok {
            return false, fmt.Errorf("Referenced package has no corresponding import: %v", p)
          }
          deps.Add(m)
        }
      }
      
      if e.Tag != nil  && e.Tag.Kind == token.STRING {
        tag, err := strconv.Unquote(e.Tag.Value)
        if err != nil {
          return false, err
        }
        t := reflect.StructTag(tag)
        if ref := t.Get(refTag); ref != "" {
          
          id, err := parseIdent(e.Type)
          if err != nil {
            return false, err
          }
          if !ast.IsExported(id.Base) {
            return false, fmt.Errorf("Field must be exported: %v", id.Base)
          }
          
          genId := ast.NewIdent(id.Base + refSuffix)
          s.Fields.List[i] = &ast.Field{
            Names:e.Names,
            Type:indirect(genId, 1),
            Comment:e.Comment,
            Tag:e.Tag,
          }
          
          cxt.Generate.Add(id)
          cxt.Lookup[genId.Name] = id
          src.Generate++
          gen = true
        }
      }
      
    }
  }
  
  if gen {
    for _, v := range deps {
      cxt.Deps.Add(v)
    }
  }
  
  return gen, nil
}

func genType(cxt *context, w io.Writer, fset *token.FileSet, id *ident) error {
  
  var inds int
  if !id.Nullable() {
    inds++
  }
  
  refId := id.Base + refSuffix
  tspec := fmt.Sprintf(`
type %v struct {
  Id    %v
  Value %v
}

func New%v(v %v) *%v {
  return &%v{Value:v}
}

func New%vId(v %v) *%v {
  return &%v{Id:v}
}

func (v %v) HasValue() bool {
  return v.Value != nil
}`,
  refId, idType, repeat(inds, '*') + id.Name,
  refId, repeat(inds, '*') + id.Name, refId,
  refId,
  refId, idType, refId,
  refId,
  refId)
  
  fmt.Fprint(w, "\n"+ strings.TrimSpace(tspec) +"\n")
  return nil
}

func genMarshal(cxt *context, w io.Writer, fset *token.FileSet, id *ident) error {
  
  spec, ok := cxt.Types[id.Name]
  if !ok {
    return fmt.Errorf("No type found for: %s", id.Name)
  }
  
  base, ok := spec.Type.(*ast.StructType)
  if !ok {
    return fmt.Errorf("Base type must be a struct: %s", id.Name)
  }
  
  decl := fmt.Sprintf(`func (v %s) MarshalJSON() ([]byte, error) {`, id.Name)
  var defX, defErr int
  
  marshal := `  fc := 0` +"\n"+ `  s := "{"` +"\n\n"
  if base.Fields != nil {
    fields:
    for _, e := range base.Fields.List {
      for _, v := range e.Names {
        
        id, err := parseIdent(v)
        if err != nil {
          return err
        }
          if !ast.IsExported(id.Base) {
          continue // ignore unexported fields
        }
        
        policy, err := fieldMarshalPolicy(e, id)
        if err != nil {
          return err
        }
        if policy.Omit {
          continue fields
        }
        
        marshal += fmt.Sprintf(`  // %s`, id.Base) +"\n"
        if policy.Ref {
          defX++; defErr++
          if policy.Marshal == marshalValue {
            marshal += indent(1, fmt.Sprintf(strings.TrimSpace(`
if v.%s != nil {
  if v.%s.HasValue() {
    if fc > 0 { s += "," }; fc++
    x, err = ref_json.Marshal(%q)
    if err != nil {
      return nil, err
    }
    s += ref_fmt.Sprintf("%%s:", x)
    x, err = ref_json.Marshal(v.%s.Value)
    if err != nil {
      return nil, err
    }
    s += string(x)
  }
}
`),         id.Base, id.Base, policy.Names.Value, id.Base)) +"\n"
          }else if policy.Marshal == marshalId {
            marshal += indent(1, fmt.Sprintf(strings.TrimSpace(`
if v.%s != nil {
  if v.%s.Id != "" {
    if fc > 0 { s += "," }; fc++
    x, err = ref_json.Marshal(%q)
    if err != nil {
      return nil, err
    }
    s += ref_fmt.Sprintf("%%s:", x)
    x, err = ref_json.Marshal(v.%s.Id)
    if err != nil {
      return nil, err
    }
    s += string(x)
  }
}
`),         id.Base, id.Base, policy.Names.Id, id.Base)) +"\n"
          }else{
            return fmt.Errorf("Invalid marshaling variant: %v", policy.Marshal)
          }
        }else{
          defX++; defErr++
          iv := 1
          if policy.OmitEmpty {
            marshal += fmt.Sprintf(`  if !isEmptyValue(ref_reflect.ValueOf(v.%s)) {`, id.Base) + "\n"
            iv++
          }
          marshal += indent(iv, fmt.Sprintf(strings.TrimSpace(`
if fc > 0 { s += "," }; fc++
x, err = ref_json.Marshal(%q)
if err != nil {
  return nil, err
}
s += ref_fmt.Sprintf("%%s:", string(x))
x, err = ref_json.Marshal(v.%s)
if err != nil {
  return nil, err
}
s += string(x)
`),         policy.Names.Value, id.Base)) +"\n"
          if policy.OmitEmpty {
            marshal += `  }` +"\n"
          }
        }
        marshal += "\n"
        
      }
      
    }
  }
  
  marshal += `  s += "}"` + "\n"
  if TRACE {
    marshal += fmt.Sprintf(`  ref_fmt.Println(">>>", %q, s)`, id.Name) + "\n"
  }
  marshal += `  return []byte(s), nil
}`
  
  fmt.Fprint(w, "\n"+ decl +"\n")
  if defErr > 0 {
    fmt.Fprint(w, "  var err error\n")
  }
  if defX > 0 {
    fmt.Fprint(w, "  var x []byte\n")
  }
  fmt.Fprint(w, marshal +"\n")
  
  return nil
}

func genUnmarshal(cxt *context, w io.Writer, fset *token.FileSet, id *ident) error {
  
  spec, ok := cxt.Types[id.Name]
  if !ok {
    return fmt.Errorf("No type found for: %s", id.Name)
  }
  
  base, ok := spec.Type.(*ast.StructType)
  if !ok {
    return fmt.Errorf("Base type must be a struct: %s", id.Name)
  }
  
  decl := fmt.Sprintf(`func (v *%s) UnmarshalJSON(data []byte) error {`, id.Name)
  var defX, defErr int
  
  marshal := indent(1, strings.TrimSpace(fmt.Sprintf(`
fields := make(map[string]ref_json.RawMessage)
var x %s

err := ref_json.Unmarshal(data, &fields)
if err != nil {
  return err
}
`, id.Name))) +"\n"
  
  if base.Fields != nil {
    fields:
    for _, e := range base.Fields.List {
      ftype, err := parseIdent(e.Type)
      if err != nil {
        return err
      }
      for _, v := range e.Names {
        
        id, err := parseIdent(v)
        if err != nil {
          return err
        }
        if !ast.IsExported(id.Base) {
          continue // ignore unexported fields
        }
        
        policy, err := fieldMarshalPolicy(e, id)
        if err != nil {
          return err
        }
        if policy.Omit {
          continue fields
        }
        
        rev, ok := cxt.Lookup[ftype.Base]
        if !ok {
          rev = ftype
        }
        
        var vassign string
        if policy.Ref {
          vassign = fmt.Sprintf(`New%v(e)`, ftype.Base)
        }else{
          vassign = `e`
        }
        
        var inds int
        if id.Dims < 1 {
          inds = id.Inds
        }
        if policy.Ref && !rev.Nullable() {
          inds++
        }
        
        marshal += "\n"
        marshal += fmt.Sprintf(`  // %s`, id.Name) +"\n"
        marshal += indent(1, strings.TrimSpace(fmt.Sprintf(`
if f, ok := fields[%q]; ok {
  var e %s
  err := ref_json.Unmarshal(f, &e)
  if err != nil {
    return err
  }
  if !isEmptyValue(ref_reflect.ValueOf(e)) {
    x.%s = %s
  }
}
`,      policy.Names.Value, repeat(inds, '*') + rev.Name, id.Name, vassign)))
        
        if policy.Ref {
          marshal += strings.TrimSpace(fmt.Sprintf(`
else if f, ok = fields[%q]; ok {
`,        policy.Names.Id))
          marshal += "\n  "
          marshal += indent(1, strings.TrimSpace(fmt.Sprintf(`
  var e %s
  err := ref_json.Unmarshal(f, &e)
  if err != nil {
    return err
  }
  if !isEmptyValue(ref_reflect.ValueOf(e)) {
    x.%v = New%vId(e)
  }
}
`,        idType, id.Name, ftype.Base)))
        }
        
        marshal += "\n"
      }
      
    }
  }
  
  if TRACE {
    marshal += "\n"
    marshal += fmt.Sprintf(`  ref_fmt.Printf("<<< %s %%+v\n", fields)`, id.Name)
  }
  marshal += "\n"
  marshal += "  *v = x\n"
  marshal += "  return nil\n"
  marshal += `}`
  
  fmt.Fprint(w, "\n"+ decl +"\n")
  if defErr > 0 {
    fmt.Fprint(w, "  var err error\n")
  }
  if defX > 0 {
    fmt.Fprint(w, "  var x []byte\n")
  }
  fmt.Fprint(w, marshal +"\n")
  
  return nil
}

func refWriter(f string) (io.Writer, error) {
  if DEBUG {
    return os.Stdout, nil
  }else{
    w, err := os.OpenFile(f, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0644)
    if err != nil {
      return nil, err
    }
    return w, nil
  }
}

func refFile(src string) string {
  base := path.Base(src)
  ext  := path.Ext(src)
  return path.Join(path.Dir(src), base[:len(base) - len(ext)] + fileSuffix + ext)
}

/**
 * Flag string list
 */
type flagList []string

/**
 * Set a flag
 */
func (s *flagList) Set(v string) error {
  *s = append(*s, v)
  return nil
}

/**
 * Describe
 */
func (s *flagList) String() string {
  return fmt.Sprintf("%+v", *s)
}
