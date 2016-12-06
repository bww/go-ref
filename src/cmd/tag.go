package main

import (
  "fmt"
  "go/ast"
  "strconv"
  "reflect"
)

const (
  refTag  = "ref"
  jsonTag = "json"
)

type fieldNames struct {
  Id, Value string
}

type marshalPolicy struct {
  Names fieldNames
  Omit, OmitEmpty bool
}

func fieldMarshalPolicy(f *ast.Field) (marshalPolicy, error) {
  var jtag, rtag string
  
  if f.Tag != nil  && f.Tag.Kind == token.STRING {
    tag, err := strconv.Unquote(f.Tag.Value)
    if err != nil {
      return marshalPolicy{}, err
    }
    t := reflect.StructTag(tag)
    jtag = t.Get(jsonTag)
    rtag = t.Get(refTag)
  }
  
  if jtag == "-" || rtag == "-" {
    return marshalPolicy{Omit:true}, nil
  }else if jtag != "" && len(e.Names) > 1 {
    return marshalPolicy{}, fmt.Errorf("Field list has %d identifiers for one tag", len(e.Names))
  }
  
}

func parseTag(t string) (string, string) {
  if x := strings.Index(t, ","); x > 0 {
    return t[:x], t[x+1:]
  }else{
    return t, ""
  }
}
