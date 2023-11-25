package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

type OpCode int

const (
  OpHalt OpCode = iota
  OpSet
  OpPush
  OpPop
  OpEq
  OpGt
  OpJmp


func symbolToOpCode

var (
  input_file = flag.String("input", "", "input file")
  target_file = flag.String("target", "", "target file")
  // flags
)

type Translator interface {
  Translate(input io.Reader, output io.Writer) error
}

type translator struct {
  labels map[string]int
  // TODO
}

/*
<!-- TODO: описание CLI -->
Консольное приложение cli
input: `translator.bin <input_file> <target_file> [flags]`
*/
// takes input file, target file and flags
// if input is not provided then read from stdin
// if target is not provided write to stdout
func main() {
  var input io.Reader
  var output io.Writer
  flag.Parse()

  if *input_file == "" {
    input = os.Stdin
  } else {
    f, err := os.Open(*input_file)
    if err != nil {
      log.Fatal(err)
    }
    defer f.Close()
    input = f
  }

  if *target_file == "" {
    output = os.Stdout
  } else {
    f, err := os.Create(*target_file)
    if err != nil {
      log.Fatal(err)
    }
    defer f.Close()
    output = f
  }

  translator
}

