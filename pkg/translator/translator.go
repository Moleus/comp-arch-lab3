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
	"fmt"
	"strings"

	"github.com/Moleus/comp-arch-lab3/pkg/isa"
)

type Translator interface {
	Translate(input string) ([]isa.MachineCodeTerm, error)
}

type AsmTranslator struct {
	instructions []ParsedInstruction
	currentIndex int
}

func NewTranslator() Translator {
	instructions := make([]ParsedInstruction, 0)
	return &AsmTranslator{instructions: instructions, currentIndex: 0}
}

// ParsedInstruction
// Каждая инструкция находится на новой строке.
// Метки должны находиться на той же строке, что и инструкции
// возможный вариант строки с инструкцией:
// <label>: <instruction> <label> ; comment
// <label>: <instruction> <operand>
// <instruction> <operand>
type ParsedInstruction struct {
	Index      int
	Label      string
	Opcode     string
	IsConstant bool
	Operand    string
	MetaInfo   isa.TermMetaInfo
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

func (t *AsmTranslator) Translate(input string) ([]isa.MachineCodeTerm, error) {
	if err := t.ParseInstructions(input); err != nil {
		return []isa.MachineCodeTerm{}, err
	}
	t.instructions = addIndicies(t.instructions)
	machineCode, err := t.convertTermsToMachineCode()
	if err != nil {
		return []isa.MachineCodeTerm{}, err
	}
	return machineCode, nil
}

func (t *AsmTranslator) ParseInstructions(input string) error {
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		err := t.parseLine(line)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *AsmTranslator) parseLine(line string) error {
	var instructions []ParsedInstruction

	parts := strings.Fields(line)

	if len(parts) == 0 {
		return nil
	}

	if parts[0] == "word:" {
		return NewParseError(line, 0)
	}

	if isConstantDeclaration(parts) {
		instructions = parseConstantDeclaration(parts)
		for _, instructions := range instructions {
			t.addConstant(instructions)
		}
	} else {
		instruction := t.parseInstructionDeclaration(parts)
		t.addInstruction(instruction)
	}

	return nil
}

func parseConstantDeclaration(parts []string) []ParsedInstruction {
	label := strings.Split(parts[0], ":")[0]
	operand := strings.Join(parts[2:], "")
	values := strings.Split(operand, ",")
	instructions := make([]ParsedInstruction, 0)
	for _, value := range values {
		instructions = append(instructions, parseConstValue(value)...)
	}
	instructions[0].Label = label
	return instructions
}

func (t *AsmTranslator) parseInstructionDeclaration(parts []string) ParsedInstruction {
	instruction := ParsedInstruction{}
	if hasLabel(parts) {
		label := strings.Split(parts[0], ":")[0]
		instruction.Label = label
		parts = parts[1:]
	}
	instruction.Opcode = parts[0]
	if len(parts) > 1 {
		instruction.Operand = parts[1]
	}
	return instruction
}

func parseConstValue(value string) []ParsedInstruction {
	isQuotedString := strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")
	if isQuotedString {
		return parseConstString(value)
	} else {
		return []ParsedInstruction{{Operand: value, IsConstant: true, Opcode: "nop"}}
	}
}

func parseConstString(value string) []ParsedInstruction {
	value = strings.Trim(value, "'")
	instructions := make([]ParsedInstruction, 0)
	for _, char := range value {
		instructions = append(instructions, ParsedInstruction{Operand: string(char), IsConstant: true, Opcode: "nop"})
	}
	return instructions
}

func (t *AsmTranslator) addConstant(instruction ParsedInstruction) {
	t.addInstruction(instruction)
}

func (t *AsmTranslator) addInstruction(instruction ParsedInstruction) {
	instruction.Index = t.currentIndex
	t.instructions = append(t.instructions, instruction)
	t.currentIndex++
}

func (t *AsmTranslator) labelToAddress(label string) (int, error) {
	for _, instruction := range t.instructions {
		if instruction.Label == label {
			return instruction.Index, nil
		}
	}
	return 0, fmt.Errorf("label '%s' not found", label)
}

func (t *AsmTranslator) convertTermsToMachineCode() ([]isa.MachineCodeTerm, error) {
	var machineCode = make([]isa.MachineCodeTerm, len(t.instructions))
	for i, instruction := range t.instructions {
		var label *string
		var operand *int
		var constant *string
		if instruction.Label != "" {
			label = new(string)
			*label = instruction.Label
		}
		if instruction.Operand != "" && !instruction.IsConstant {
			address, err := t.labelToAddress(instruction.Operand)
			if err != nil {
				return []isa.MachineCodeTerm{}, err
			}
			operand = new(int)
			*operand = address
		}
		if instruction.IsConstant {
			constant = new(string)
			*constant = instruction.Operand
		}
		opcode, err := isa.GetOpcodeFromString(instruction.Opcode)
		if err != nil {
			return []isa.MachineCodeTerm{}, err
		}
		newMachineCodeTerm := isa.MachineCodeTerm{
			Index:    instruction.Index,
			Label:    label,
			Opcode:   opcode,
			Constant: constant,
			Operand:  operand,
			TermInfo: instruction.MetaInfo,
		}
		machineCode[i] = newMachineCodeTerm
	}
	return machineCode, nil
}
