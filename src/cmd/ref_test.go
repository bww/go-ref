package main

import (
  "fmt"
  "testing"
)

func TestThis(t *testing.T) {
  fmt.Println(indent(4, "Hello"))
  fmt.Println(indent(4, `
Hello {
  if {
    then...
  }
}
`))
}
