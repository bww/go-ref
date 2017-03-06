// +build ignore

package main

import (
  "fmt"
  "testing"
  "go/parser"
  "encoding/json"
  "github.com/stretchr/testify/assert"
)

type SimpleHello struct {
  A json.RawMessage     `json:"as" ref:"a_ids"`
}

// type SimpleExample struct {
//   A []*json.RawMessage  `json:"as" ref:"a_ids,id"`
// }

func TestPkg(t *testing.T) {
  fmt.Println("OKOK")
}
