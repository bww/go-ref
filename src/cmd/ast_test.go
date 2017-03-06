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
      assert.Equal(t, c.Expect.Inds, id.Inds)
      assert.Equal(t, c.Expect.Dims, id.Dims)
    }
  }
}

func TestParseIdent(t *testing.T) {
  testIdent(t, testCase{`Example`, newIdent(`Example`, `Example`, 0, 0)})
  testIdent(t, testCase{`example`, newIdent(`example`, `example`, 0, 0)})
  testIdent(t, testCase{`foo.bar.Car`, newIdent(`foo.bar.Car`, `Car`, 0, 0)})
  testIdent(t, testCase{`*Example`, newIdent(`*Example`, `Example`, 1, 0)})
  testIdent(t, testCase{`[]Example`, newIdent(`[]Example`, `ArrayOfExample`, 0, 1)})
  testIdent(t, testCase{`[]*json.RawMessage`, newIdent(`[]*json.RawMessage`, `ArrayOfPtrToRawMessage`, 1, 1)})
  testIdent(t, testCase{`[][]*json.RawMessage`, newIdent(`[][]*json.RawMessage`, `ArrayOfArrayOfPtrToRawMessage`, 1, 2)})
  testIdent(t, testCase{`[][]**json.RawMessage`, newIdent(`[][]**json.RawMessage`, `ArrayOfArrayOfPtrToPtrToRawMessage`, 2, 2)})
  testIdent(t, testCase{`map[string]Example`, &ident{`map[string]Example`, `MapOfStringToExample`, 0, 0, newIdent(`string`, `string`, 0, 0)}})
  testIdent(t, testCase{`map[string]*Example`, &ident{`map[string]*Example`, `MapOfStringToPtrToExample`, 0, 0, newIdent(`string`, `string`, 0, 0)}})
  testIdent(t, testCase{`map[*string]*Example`, &ident{`map[*string]*Example`, `MapOfPtrToStringToPtrToExample`, 0, 0, newIdent(`*string`, `string`, 1, 0)}})
  testIdent(t, testCase{`map[string]*json.RawMessage`, &ident{`map[string]*json.RawMessage`, `MapOfStringToPtrToRawMessage`, 0, 0, newIdent(`string`, `string`, 0, 0)}})
}
