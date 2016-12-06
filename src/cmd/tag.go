package main

import (
  "fmt"
  "strings"
  "strconv"
  "reflect"
  "go/ast"
  "go/token"
)

const (
  refTag          = "ref"
  jsonTag         = "json"
  omitEmpty       = "omitempty"
)

const (
  marshalIdTag    = "id"
  marshalValueTag = "value"
)

type marshalVariant int
const (
  marshalId       = marshalVariant(iota)
  marshalValue    = marshalVariant(iota)
)

type fieldNames struct {
  Id, Value string
}

type marshalPolicy struct {
  Names   fieldNames
  Marshal marshalVariant
  Ref, Omit, OmitEmpty bool
}

func fieldMarshalPolicy(field *ast.Field, id *ast.Ident) (marshalPolicy, error) {
  var jtag, rtag, name, flags string
  
  if field.Tag != nil  && field.Tag.Kind == token.STRING {
    tag, err := strconv.Unquote(field.Tag.Value)
    if err != nil {
      return marshalPolicy{}, err
    }
    t := reflect.StructTag(tag)
    jtag = t.Get(jsonTag)
    rtag = t.Get(refTag)
  }
  
  if jtag == "-" || rtag == "-" {
    return marshalPolicy{Omit:true}, nil
  }else if jtag != "" && len(field.Names) > 1 {
    return marshalPolicy{}, fmt.Errorf("Field list has %d identifiers for one tag", len(field.Names))
  }
  
  policy := marshalPolicy{}
  
  if jtag != "" {
    name, flags = parseTag(jtag)
  }else{
    name = id.Name
  }
  
  policy.Names.Value = name
  policy.OmitEmpty = policy.OmitEmpty || flags == omitEmpty
  
  if rtag != "" {
    name, flags = parseTag(rtag)
    policy.Ref  = true
  }
  
  policy.Names.Id = name
  if flags == marshalValueTag {
    policy.Marshal = marshalValue
  }else{
    policy.Marshal = marshalId
  }
  
  return policy, nil
}

func parseTag(t string) (string, string) {
  if x := strings.Index(t, ","); x > 0 {
    return t[:x], t[x+1:]
  }else{
    return t, ""
  }
}
