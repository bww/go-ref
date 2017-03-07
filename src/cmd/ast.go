package main

import (
  "fmt"
  "path"
  "strings"
  "strconv"
  "go/ast"
  "go/token"
)

type ident struct {
  Name  string
  Base  string
  Inds  int
  Dims  int
  Key   *ident
}

func (d ident) Nullable() bool {
  return d.Inds > 0 || d.Dims > 0 || d.Key != nil
}

func newIdent(name, base string, inds, dims int) *ident {
  return &ident{name, base, inds, dims, nil}
}

func astIdent(id *ast.Ident) *ident {
  return &ident{id.Name, id.Name, 0, 0, nil}
}

func parseIdent(e ast.Expr) (*ident, error) {
  return parseIdentE(e, false)
}

func parseIdentE(e ast.Expr, r bool) (*ident, error) {
  d, err := parseIdentR(e, 0, 0)
  if err != nil {
    return nil, err
  }
  var s, p string
  if d.Key != nil {
    p += "MapOf"+ strings.Title(d.Key.Base) +"To"
    s += "map["+ d.Key.Name +"]"
  }
  for i := 0; i < d.Dims; i++ {
    p += "ArrayOf"
    s += "[]"
  }
  for i := 0; i < d.Inds; i++ {
    s += "*"
    if r || d.Dims > 0 {
      p += "PtrTo"
    }
  }
  d.Name = s + d.Name
  base := d.Base
  if d.Key != nil || d.Dims > 0 || d.Inds > 0 {
    base = strings.Title(base)
  }
  d.Base = p + base
  return d, nil
}

func parseIdentR(e ast.Expr, r, d int) (*ident, error) {
  switch v := e.(type) {
    
    case *ast.Ident:
      return newIdent(v.Name, v.Name, r, d), nil
      
    case *ast.StarExpr:
      return parseIdentR(v.X, r + 1, d)
      
    case *ast.SelectorExpr:
      p, s, n, err := concatIdent(e, r)
      if err != nil {
        return nil, err
      }
      return newIdent(p +"."+ s, s, n, d), nil
      
    case *ast.ArrayType:
      if v.Len != nil {
        return nil, fmt.Errorf("Array types are not supported; only slice types.")
      }
      return parseIdentR(v.Elt, r, d + 1)
      
    case *ast.MapType:
      key, err := parseIdentE(v.Key, true)
      if err != nil {
        return nil, err
      }
      val, err := parseIdentE(v.Value, true)
      if err != nil {
        return nil, err
      }
      return &ident{val.Name, val.Base, 0, 0, key}, nil
      
    default:
      return nil, fmt.Errorf("Not a valid identifier: %T (%v)", e, e)
      
  }
}

func concatIdent(e ast.Expr, r int) (string, string, int, error) {
  switch v := e.(type) {
    
    case *ast.Ident:
      return "", v.Name, r, nil
      
    case *ast.StarExpr:
      return concatIdent(v.X, r)
      
    case *ast.SelectorExpr:
      p, s, n, err := concatIdent(v.X, r)
      if err != nil {
        return "", "", -1, err
      }
      if p != "" {
        p = p +"."+ s
      }else{
        p = s
      }
      return p, v.Sel.Name, n, nil
      
    default:
      return "", "", -1, fmt.Errorf("Unsupported type: %T (%v)", e, e)
      
  }
}

func leftmost(e *ast.SelectorExpr) ast.Expr {
  if v, ok := e.X.(*ast.SelectorExpr); ok {
    return leftmost(v)
  }else{
    return e.X
  }
}

func indirect(e *ast.Ident, r int) ast.Expr {
  v := ast.Expr(e)
  for i := 0; i < r; i++ {
    v = &ast.StarExpr{X:v}
  }
  return v
}

func importPackage(e *ast.ImportSpec) string {
  if id := e.Name; id != nil {
    return id.Name
  }else{
    return path.Base(stringLit(e.Path))
  }
}

func commentText(c *ast.Comment) string {
  t := c.Text
  if len(t) > 2 && t[:2] == "//" {
    return strings.TrimSpace(t[2:])
  }else{
    return t
  }
}

func stringLit(e *ast.BasicLit) string {
  if e.Kind != token.STRING {
    panic(fmt.Errorf("Literal is not a string: %v: %v", e.Kind, e.Value))
  }
  u, err := strconv.Unquote(e.Value)
  if err != nil {
    panic(err)
  }
  return u
}
