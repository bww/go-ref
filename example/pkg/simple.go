// +build ignore

package main

import (
  "fmt"
  "encoding/json"
)

type SimpleHello struct {
  A json.RawMessage   `json:"a" ref:"a_id"`
}

type SimpleExample struct {
  A *json.RawMessage  `json:"a" ref:"a_id,id"`
}
