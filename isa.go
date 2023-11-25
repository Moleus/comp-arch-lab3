package isa

import (
  "fmt"
  "io"
  "log"
)

/* accumulator based ISA */

type AddrInstr int

const (
  AddrInstrHalt AddrInstr = iota



