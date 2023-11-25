/*
Как выглядит разрабатываемый язык программирования:

program:
  line |
  line program

line
  : label
  | instruction
  | comment

instruction
   : addr operand
   | nonaddr
   | branch label
   | io dev

variable_declaration: <name> ':' <value>

addr: AND | OR | ADD | SUB | CMP | LOOP | LD | JUMP | CALL | ST;
nonaddr: NOP | HLT | CLA | NOT | CLC | CMC | ROL | ROR | ASL | ASR | SXTB | SWAB |
         INC | DEC | NEG | POP | POPF | RET | IRET | PUSH | PUSHF | SWAP |
         EI  | DI;
branch: BEQ | BNE | BMI | BPL | BCS | BCC | BVS | BVC | BLT | BGE | BR;

io:  IN | OUT | INT;
dev: number;

Пример:
```asm
; задание переменных в 16-ричной системе счисления:
; <имя переменной>: <значение>
X: 0x2

; Начало прогрммы - точка входа с метки START
; все операции работают с аккумулятором
START:
  CLA ; очистить аккумулятор
  LD 42 ; загрузить в аккумулятор значение 42
  ADD X ; прибавить к аккумулятору значение переменной X
  NOP ; ничего не делать
  HLT ; остановить выполнение программы
```

*/
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
)

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

