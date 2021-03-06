// +build ignore
// +goref ignore

package main

import (
  "fmt"
  "encoding/json"
)

type Int int

type Complex struct {
  A string              `json:"a,omitempty"`
  B int                 `json:"b"`
}

type Hello struct {
  A, B string
  C Int                 `json:"hello,omitempty" ref:"hello_id,id"`
  D int                 `json:"d,omitempty"`
  E json.RawMessage     `json:"c" ref:"c_id"`
}

type Example struct {
  A   *Int              `json:"a" ref:"a_id,value"`
  B   *Complex          `json:"b" ref:"b_id,id"`
  C   *json.RawMessage  `json:"c" ref:"c_id,id"`
}

type Another struct {
  A   Int     `json:"a" ref:"a_id"`
}

func main() {
  var s []byte
  var err error
  
  v := Int(123)
  c := Hello{}
  
  s, err = json.Marshal(c)
  if err != nil {
    panic(err)
  }else{
    fmt.Println(string(s))
  }
  
  c.C = &IntRef{"abc", &v}
  s, err = json.Marshal(c)
  if err != nil {
    panic(err)
  }else{
    fmt.Println(string(s))
  }
  
  d := &Hello{}
  err = json.Unmarshal(s, d)
  if err != nil {
    panic(err)
  }
  
  if d.C != nil {
    if d.C.HasValue() {
      fmt.Printf("HAS VALUE: %+v\n", d.C.Value)
    }else{
      fmt.Printf("HAS IDENT: %+v\n", d.C.Id)
    }
  }
  
  fmt.Println("OK!")
}
