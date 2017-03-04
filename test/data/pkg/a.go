// +build ignore

package main

import (
  "fmt"
  "encoding/json"
)

type SimpleHello struct {
  A json.RawMessage     `json:"as" ref:"a_ids"`
}

// type SimpleExample struct {
//   A []*json.RawMessage  `json:"as" ref:"a_ids,id"`
// }
