package main

import (
  "fmt"
  "go/ast"
)

func ident(e ast.Expr, r int) (*ast.Ident, *ast.Ident, int, error) {
  switch v := e.(type) {
    case *ast.Ident:
      return v, v, r, nil
    case *ast.StarExpr:
      return ident(v.X, r + 1)
    case *ast.SelectorExpr:
      p, s, n, err := stringifyIdent(e, r + 1)
      if err != nil {
        return nil, nil, -1, err
      }
      fmt.Println(">>>", p, s)
      return ast.NewIdent(p +"."+ s), ast.NewIdent(s), n, nil
    default:
      return nil, nil, -1, fmt.Errorf("Not an identifier: %T (%v)", e, e)
  }
}

func stringifyIdent(e ast.Expr, r int) (string, string, int, error) {
  switch v := e.(type) {
    case *ast.Ident:
      return "", v.Name, r, nil
    case *ast.StarExpr:
      return stringifyIdent(v.X, r + 1)
    case *ast.SelectorExpr:
      p, s, n, err := stringifyIdent(v.X, r + 1)
      if err != nil {
        return "", "", -1, err
      }
      if p != "" {
        p = p +"."+ s
      }
      return p, v.Sel.Name, n, nil
    default:
      return "", "", -1, fmt.Errorf("Unsupported type: %T (%v)", e, e)
  }
}

func indirect(e *ast.Ident, r int) ast.Expr {
  v := ast.Expr(e)
  for i := 0; i < r; i++ {
    v = &ast.StarExpr{X:v}
  }
  return v
}

