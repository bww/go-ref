// +build ignore

package main

import (
  "fmt"
  "testing"
  "go/parser"
  "encoding/json"
  "github.com/stretchr/testify/assert"
)

type E struct {
  A int                 `json:"a"`
  B json.RawMessage     `json:"b" ref:"b_id"`
}

// type SimpleExample struct {
//   A []*json.RawMessage  `json:"as" ref:"a_ids,id"`
// }

func TestPkg(t *testing.T) {
  var s []byte
  var err error
  
  m := json.RawMessage(`{}`)
  v := &E{123, NewRawMessageRef(&m)}
  
  s, err = json.Marshal(v)
  if assert.Nil(t, err, fmt.Sprintf("%v", err)) {
    assert.Equal(t, `{"a":123,"b":{}}`, string(s))
  }
  
  fmt.Println("OKOK")
}
