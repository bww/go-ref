// +build ignore

package main

import (
  "fmt"
  "encoding/json"
)

type Hello struct {
  A json.RawMessage   `json:"a" ref:"a_id"`
}

type Example struct {
  A *json.RawMessage  `json:"a" ref:"a_id,id"`
}

func main() {
  fmt.Println("OK!")
}
