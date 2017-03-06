// +build ignore

package main

import (
  "fmt"
  "testing"
  "go/parser"
  "encoding/json"
  "github.com/stretchr/testify/assert"
)

type X struct {
  A int                 `json:"a"`
  B json.RawMessage     `json:"b" ref:"b_id,value"`
}

type Y struct {
  A int                 `json:"a"`
  B json.RawMessage     `json:"b" ref:"b_id"`
}

type Z struct {
  A int                 `json:"a"`
  B []json.RawMessage   `json:"b" ref:"b_id,value"`
}

type W struct {
  A int                 `json:"a"`
  B []*json.RawMessage  `json:"b" ref:"b_id,value"`
}

func TestMarshal(t *testing.T) {
  var s []byte
  var err error
  
  m := json.RawMessage(`{"a":123}`)
  x := &X{123, NewRawMessageRef(&m)}
  
  s, err = json.Marshal(x)
  if assert.Nil(t, err, fmt.Sprintf("%v", err)) {
    assert.Equal(t, `{"a":123,"b":{"a":123}}`, string(s))
  }
  
  y := &Y{123, NewRawMessageRef(&m)}
  
  s, err = json.Marshal(y)
  if assert.Nil(t, err, fmt.Sprintf("%v", err)) {
    assert.Equal(t, `{"a":123}`, string(s))
  }
  
  z := &Z{123, NewArrayOfRawMessageRef([]json.RawMessage{m, m})}
  
  s, err = json.Marshal(z)
  if assert.Nil(t, err, fmt.Sprintf("%v", err)) {
    assert.Equal(t, `{"a":123,"b":[{"a":123},{"a":123}]}`, string(s))
  }
  
  w := &Z{123, NewArrayOfRawMessageRef([]json.RawMessage{m, m})}
  
  s, err = json.Marshal(w)
  if assert.Nil(t, err, fmt.Sprintf("%v", err)) {
    assert.Equal(t, `{"a":123,"b":[{"a":123},{"a":123}]}`, string(s))
  }
  
}
