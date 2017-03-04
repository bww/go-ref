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
  Name      string
  Base      string
  Indirects int
  Dims      int
}

func (d ident) Nullable() bool {
  return d.Indirects > 0 || d.Dims > 0
}

func newIdent(name, base string, inds, dims int) *ident {
  return &ident{name, base, inds, dims}
}

func astIdent(id *ast.Ident) *ident {
  return &ident{id.Name, id.Name, 0, 0}
}

func parseIdent(e ast.Expr) (*ident, error) {
  d, err := parseIdentR(e, 0, 0)
  if err != nil {
    return nil, err
  }
  var s string
  for i := 0; i < d.Dims; i++ {
    s += "[]"
  }
  for i := 0; i < d.Indirects; i++ {
    s += "*"
  }
  d.Name = s + d.Name
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
