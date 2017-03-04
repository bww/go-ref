package main

import (
  "fmt"
  "testing"
  "go/parser"
  "github.com/stretchr/testify/assert"
)

type testCase struct {
  Source  string
  Expect  *ident
}

func testIdent(t *testing.T, c testCase) {
  ex, err := parser.ParseExpr(c.Source)
  if assert.Nil(t, err, fmt.Sprintf("%v", err)) {
    id, err := parseIdent(ex)
    if assert.Nil(t, err, fmt.Sprintf("%v", err)) {
      fmt.Println(">>>", id.Base, id.Name)
      assert.Equal(t, c.Expect.Name, id.Name)
      assert.Equal(t, c.Expect.Base, id.Base)
      assert.Equal(t, c.Expect.Indirects, id.Indirects)
      assert.Equal(t, c.Expect.Dims, id.Dims)
    }
  }
}

func TestParseIdent(t *testing.T) {
  testIdent(t, testCase{`[]*json.RawMessage`, newIdent(`[]*json.RawMessage`, `RawMessage`, 1, 1)})
}
