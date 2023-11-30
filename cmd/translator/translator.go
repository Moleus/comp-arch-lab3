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
	"flag"
	"fmt"
	"github.com/Moleus/comp-arch-lab3/pkg/isa"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	inputFile  = flag.String("input", "", "input file")
	targetFile = flag.String("target", "", "target file")
	// flags
)

type Translator interface {
	Translate(input string) ([]MachineCodeTerm, error)
}

type translator struct {
	// TODO
}

func NewTranslator() Translator {
	return &translator{}
}

// ParsedInstruction
// Каждая инструкция находится на новой строке.
// Метки должны находиться на той же строке, что и инструкции
// возможный вариант строки с инструкцией:
// <label>: <instruction> <operand>
// <instruction> <operand>
type ParsedInstruction struct {
	label       string
	instruction string
	operand     string
	metaInfo    TermMetaInfo
}

// ParsedConstant - объявленная константа в исходном коде
type ParsedConstant struct {
	label string
	value int
}

type ParseError struct {
	content string
	line    int
}

func (e ParseError) Error() string {
	return fmt.Sprintf("invalid instruction at %d: '%s'", e.line, e.content)
}

func NewParseError(content string, line int) error {
	return ParseError{content, line}
}

func ParseInstruction(line string, lineNumber int) (ParsedInstruction, error) {
	// group2: label, group3: instruction, group4: operand
	// TODO: parse single opcode HLT
	// parse labels
	instructionRegexTmpl := `^((\w+)?\s*:)?\s*(\w+)(\s+(\w+))?$`
	instructionRegex := regexp.MustCompile(instructionRegexTmpl)
	matches := instructionRegex.FindStringSubmatch(line)
	if len(matches) == 0 {
		return ParsedInstruction{}, NewParseError(line, lineNumber)
	}
	// TODO: debug matches
	instruction := ParsedInstruction{
		label:       matches[2],
		instruction: matches[3],
		operand:     matches[5],
		metaInfo: TermMetaInfo{
			LineNum:         lineNumber,
			OriginalContent: line,
		},
	}
	return instruction, nil
}

// ParseConstant takes a line and returns a constant if any
// Constant input is represented as
// CONST <label>: <value>
func ParseConstant(line string, lineNum int) (ParsedConstant, error) {
	constantRegexTmpl := `^CONST\s+(\w+)\s*:\s*(\d+)$`
	constantRegex := regexp.MustCompile(constantRegexTmpl)
	matches := constantRegex.FindStringSubmatch(line)
	if len(matches) == 0 {
		return ParsedConstant{}, NewParseError(line, lineNum)
	}
	value, err := strconv.Atoi(matches[2])
	if err != nil {
		return ParsedConstant{}, NewParseError(line, lineNum)
	}
	return ParsedConstant{
		label: matches[1],
		value: value,
	}, nil
}

func isConstant(line string) bool {
	return strings.HasPrefix(line, "CONST")
}

func isEmpty(line string) bool {
	// spaces and tabs are empty
	emptyRegexTmpl := `^\s*$`
	emptyRegex := regexp.MustCompile(emptyRegexTmpl)
	return emptyRegex.MatchString(line)
}

func prepareLine(line string) string {
	// remove comments
	commentRegexTmpl := `;.*$`
	commentRegex := regexp.MustCompile(commentRegexTmpl)
	withoutComments := commentRegex.ReplaceAllString(line, "")
	// remove trailing spaces
	withoutSpaces := strings.TrimSpace(withoutComments)
	return withoutSpaces
}

func (t *translator) ParseConstants(input string) ([]ParsedConstant, error) {
	var constants []ParsedConstant

	lines := strings.Split(input, "\n")
	for i, line := range lines {
		line = prepareLine(line)
		if !isConstant(line) || isEmpty(line) {
			continue
		}
		constant, err := ParseConstant(line, i+1)
		if err != nil {
			return nil, err
		}
		constants = append(constants, constant)
	}
	return constants, nil
}

func (t *translator) ParseInstructions(input string) ([]ParsedInstruction, error) {
	var instructions []ParsedInstruction

	lines := strings.Split(input, "\n")
	for i, line := range lines {
		line = prepareLine(line)
		if isConstant(line) || isEmpty(line) {
			continue
		}
		instruction, err := ParseInstruction(line, i+1)
		if err != nil {
			return nil, err
		}
		instructions = append(instructions, instruction)
	}

	return instructions, nil
}

type TermMetaInfo struct {
	LineNum         int    `json:"line_num"`
	OriginalContent string `json:"original_content"`
}

type MachineCodeTerm struct {
	Index    int          `json:"index"`
	Label    string       `json:"label,omitempty"`
	Opcode   string       `json:"opcode"`
	Operand  string       `json:"operand,omitempty"`
	TermInfo TermMetaInfo `json:"term_info"`
}

func (t *translator) ConvertTermsToMachineCode(instructions []ParsedInstruction) ([]MachineCodeTerm, error) {
	var machineCode []MachineCodeTerm
	// TODO: check instruction correctness
	for i, instruction := range instructions {
		// TODO: it can be variable, not instruction
		newMachineCodeTerm := MachineCodeTerm{
			Index:    i,
			Label:    instruction.label,
			Opcode:   instruction.instruction,
			Operand:  instruction.operand,
			TermInfo: instruction.metaInfo,
		}
		machineCode = append(machineCode, newMachineCodeTerm)
	}
	return machineCode, nil
}

func (t *translator) Translate(input string) ([]MachineCodeTerm, error) {
	parsedInstructions, err := t.ParseInstructions(input)
	if err != nil {
		return []MachineCodeTerm{}, err
	}
	machineCode, err := t.ConvertTermsToMachineCode(parsedInstructions)
	if err != nil {
		return []MachineCodeTerm{}, err
	}
	return machineCode, nil
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

	if *inputFile == "" {
		input = os.Stdin
	} else {
		f, err := os.Open(*inputFile)
		if err != nil {
			log.Fatal(err)
		}
		defer func(f *os.File) {
			err := f.Close()
			if err != nil {
				log.Fatal(err)
			}
		}(f)
		input = f
	}

	if *targetFile == "" {
		output = os.Stdout
	} else {
		f, err := os.Create(*targetFile)
		if err != nil {
			log.Fatal(err)
		}
		defer func(f *os.File) {
			err := f.Close()
			if err != nil {
				log.Fatal(err)
			}
		}(f)
		output = f
	}

	translator := NewTranslator()
	inputStr, err := io.ReadAll(input)
	if err != nil {
		log.Fatal(err)
	}
	code, err := translator.Translate(string(inputStr))
	if err != nil {
		log.Fatal(err)
	}

	err = isa.WriteCode(output, code)
	if err != nil {
		log.Fatal(err)
	}
}
