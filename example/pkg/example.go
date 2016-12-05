// +build ignore

package main

import (
  "fmt"
)

type Int int

type Complex struct {
  A string  `json:"a,omitempty"`
  B int     `json:"b"`
}

type Hello struct {
  A, B string
  C Int `json:"hello,omitempty" ref:"hello_id"`
}

type Example struct {
  A   *Int      `json:"a" ref:"a_id,value"`
  B   *Complex  `json:"b" ref:"b_id,id"`
}

type Another struct {
  A   Int     `json:"a" ref:"a_id"`
}

func main() {
  fmt.Println("OK!")
}
