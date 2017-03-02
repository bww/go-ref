package main

import (
  "fmt"
  "strconv"
  "go/ast"
  "go/token"
)

func ident(e ast.Expr, r int) (*ast.Ident, *ast.Ident, int, error) {
  switch v := e.(type) {
    case *ast.Ident:
      return v, v, r, nil
    case *ast.StarExpr:
      return ident(v.X, r + 1)
    case *ast.SelectorExpr:
      p, s, n, err := joinIdent(e, r + 1)
      if err != nil {
        return nil, nil, -1, err
      }
      return ast.NewIdent(p +"."+ s), ast.NewIdent(s), n, nil
    default:
      return nil, nil, -1, fmt.Errorf("Not an identifier: %T (%v)", e, e)
  }
}

func joinIdent(e ast.Expr, r int) (string, string, int, error) {
  switch v := e.(type) {
    case *ast.Ident:
      return "", v.Name, r, nil
    case *ast.StarExpr:
      return joinIdent(v.X, r + 1)
    case *ast.SelectorExpr:
      p, s, n, err := joinIdent(v.X, r + 1)
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
