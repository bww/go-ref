package main

type Hello struct {
  A, B string
  C int `json:"hello" ref:"hello_id"`
}
