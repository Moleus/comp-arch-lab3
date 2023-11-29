/*
Package translator:
struct - Машинный код представляется в виде высокоуровневой структуры

Translator - полностью независимая программа.
Принимает текстовое представление программы и преобразует его в машинный код.
Содержит информацию про токены и символы, которые используются в языке.
Отображает символы на операции (OpCode). Операции описываются в файле ISA

## Как выглядит разрабатываемый язык программирования:
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

## Пример:
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

## Реализация
Трансляция проходит в 2 этапа:
1. Парсинг строки в термы
2. Трансляция термов в машинный код
*/
package translator

import (
	"bytes"
	"flag"
	"io"
	"log"
	"os"
)

var (
	input_file  = flag.String("input", "", "input file")
	target_file = flag.String("target", "", "target file")
	// flags
)

type Translator interface {
	Translate(input io.Reader, output io.Writer) error
}

type translator struct {
	// TODO
}

func NewTranslator() Translator {
	return &translator{}
}

func (t *translator) ParseTerms(input io.Reader) ([]string, error) {
	// TODO
	terms := []string{}
	return terms, nil
}

func (t *translator) ConvertTermsToMachineCode(terms []string) (string, error) {
	// TODO
	return "", nil
}

func (t *translator) Translate(input io.Reader, output io.Writer) error {
	terms, err := t.ParseTerms(input)
	if err != nil {
		return err
	}
	machineCode, err := t.ConvertTermsToMachineCode(terms)
	if err != nil {
		return err
	}
	output.Write(bytes.NewBufferString(machineCode).Bytes())
	return nil
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

	translator := NewTranslator()
	err := translator.Translate(input, output)
	if err != nil {
		log.Fatal(err)
	}
}
