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

type P struct {
  A int                           `json:"a"`
  B map[string]*json.RawMessage   `json:"b" ref:"b_id,value"`
}

type Q struct {
  A int                 `json:"a"`
  B *X                  `json:"b"`
}

type R struct {
  A int                 `json:"a"`
  B *X                  `json:"b" ref:"b_id,value"`
}

func TestMarshalRoundtrip(t *testing.T) {
  var s []byte
  var err error
  
  m := json.RawMessage(`{"a":123}`)
  x := &X{123, NewRawMessageRef(&m)}
  
  s, err = json.Marshal(x)
  if assert.Nil(t, err, fmt.Sprintf("%v", err)) {
    assert.Equal(t, `{"a":123,"b":{"a":123}}`, string(s))
  }
  
  var x1 X
  err = json.Unmarshal(s, &x1)
  if assert.Nil(t, err, fmt.Sprintf("%v", err)) {
    assert.Equal(t, x.A, x1.A)
    assert.Equal(t, x.B, x1.B)
  }
  
  y := &Y{123, NewRawMessageRef(&m)}
  
  s, err = json.Marshal(y)
  if assert.Nil(t, err, fmt.Sprintf("%v", err)) {
    assert.Equal(t, `{"a":123}`, string(s))
  }
  
  var y1 Y
  err = json.Unmarshal(s, &y1)
  if assert.Nil(t, err, fmt.Sprintf("%v", err)) {
    assert.Equal(t, y.A, y1.A)
    assert.Equal(t, (*RawMessageRef)(nil), y1.B) // B is not marshaled
  }
  
  z := &Z{123, NewArrayOfRawMessageRef([]json.RawMessage{m, m})}
  
  s, err = json.Marshal(z)
  if assert.Nil(t, err, fmt.Sprintf("%v", err)) {
    assert.Equal(t, `{"a":123,"b":[{"a":123},{"a":123}]}`, string(s))
  }
  
  var z1 Z
  err = json.Unmarshal(s, &z1)
  if assert.Nil(t, err, fmt.Sprintf("%v", err)) {
    assert.Equal(t, z.A, z1.A)
    assert.Equal(t, z.B, z1.B)
  }
  
  w := &W{123, NewArrayOfPtrToRawMessageRef([]*json.RawMessage{&m, &m})}
  
  s, err = json.Marshal(w)
  if assert.Nil(t, err, fmt.Sprintf("%v", err)) {
    assert.Equal(t, `{"a":123,"b":[{"a":123},{"a":123}]}`, string(s))
  }
  
  var w1 W
  err = json.Unmarshal(s, &w1)
  if assert.Nil(t, err, fmt.Sprintf("%v", err)) {
    assert.Equal(t, w.A, w1.A)
    assert.Equal(t, w.B, w1.B)
  }
  
  p := &P{123, NewMapOfStringToRawMessageRef(map[string]json.RawMessage{"yo": json.RawMessage(`{"a":false}`)})}
  
  s, err = json.Marshal(p)
  if assert.Nil(t, err, fmt.Sprintf("%v", err)) {
    assert.Equal(t, `{"a":123,"b":{"yo":{"a":false}}`, string(s))
  }
  
  var p1 P
  err = json.Unmarshal(s, &p1)
  if assert.Nil(t, err, fmt.Sprintf("%v", err)) {
    assert.Equal(t, p.A, p1.A)
    assert.Equal(t, p.B, p1.B)
  }
  
  q := &Q{123, x}
  
  s, err = json.Marshal(q)
  if assert.Nil(t, err, fmt.Sprintf("%v", err)) {
    assert.Equal(t, `{"a":123,"b":{"a":123,"b":{"a":123}}}`, string(s))
  }
  
  var q1 Q
  err = json.Unmarshal(s, &q1)
  if assert.Nil(t, err, fmt.Sprintf("%v", err)) {
    assert.Equal(t, q.A, q1.A)
    assert.Equal(t, q.B, q1.B)
  }
  
  r := &R{123, NewXRef(x)}
  
  s, err = json.Marshal(r)
  if assert.Nil(t, err, fmt.Sprintf("%v", err)) {
    assert.Equal(t, `{"a":123,"b":{"a":123,"b":{"a":123}}}`, string(s))
  }
  
  var r1 R
  err = json.Unmarshal(s, &r1)
  if assert.Nil(t, err, fmt.Sprintf("%v", err)) {
    assert.Equal(t, r.A, r1.A)
    assert.Equal(t, r.B, r1.B)
  }
  
}
