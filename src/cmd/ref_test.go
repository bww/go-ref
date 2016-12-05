package main

type Hello struct {
  A, B string
  C Int `json:"hello" ref:"hello_id"`
}

type Int int

type Example struct {
  A   *Int    `json:"a" ref:"a_id"`
}

type Another struct {
  A   Int     `json:"a" ref:"a_id"`
}
