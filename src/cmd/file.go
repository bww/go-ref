package main

import (
  "os"
)

/**
 * Determine if the destination file is out of date relative to the source file
 */
func isFileOutOfDate(dst, src string) (bool, error) {
  
  df, err := os.Stat(dst)
  if err != nil && os.IsNotExist(err) {
    return true, nil  // doesn't exist; out of date
  }else if err != nil{
    return false, err // other errors are also errors
  }
  
  sf, err := os.Stat(src)
  if err != nil && os.IsNotExist(err) {
    return false, err // doesn't exist; must have a source
  }else if err != nil{
    return false, err // other errors are also errors
  }
  
  return sf.ModTime().After(df.ModTime()), nil
}
