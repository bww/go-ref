package main

type Int int

type Complex struct {
  A string  `json:"a,omitempty"`
  B int     `json:"b"`
}

type Hello struct {
  A, B string
  C Int `json:"hello" ref:"hello_id"`
}

type Example struct {
  A   *Int      `json:"a" ref:"a_id"`
  B   *Complex  `json:"b" ref:"b_id"`
}

type Another struct {
  A   Int     `json:"a" ref:"a_id"`
}
