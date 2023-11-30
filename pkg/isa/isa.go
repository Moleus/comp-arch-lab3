/*
Package isa: Instruction Set Architecture (Система команд)
Фон-Неймановская архитектура

Задачи:
- читает машинный код из файла
- записывает машинный код в файл

# По сути занимается сериализацией и десериализаций программы в JSON

Используется в machine.go, controlunit.go и datapath.go
*/
package isa

import (
	"encoding/json"
	"github.com/Moleus/comp-arch-lab3/cmd/translator"
	"io"
)

/* accumulator based ISA */

// Instruction represents all supported instructions for our architecture
type Instruction int

const (
	InstructionAnd Instruction = iota
	InstructionOr
	InstructionAdd
	InstructionSub
	InstructionCmp
)

// AddressingType like relative direct
type AddressingType int

const (
	DirectAbsolute AddressingType = iota
	Indirect
)

const (
	WORD_WIDTH     = 16
	WORD_MAX_VALUE = 1<<(WORD_WIDTH-1) - 1
	WORD_MIN_VALUE = -1 << (WORD_WIDTH - 1)
)

type MemoryWord struct {
	address int
	label   string
	value   int
}

type AddressingMode struct {
	targetAddress  int
	addressingType AddressingType
}

type InstructionWord struct {
	MemoryWord
	instruction    Instruction
	addressingMode AddressingMode
}

// TODO: think about dependencies and move MachineCodeTerm in ISA
func ReadCode(input io.Reader) ([]translator.MachineCodeTerm, error) {
	var machineCode []translator.MachineCodeTerm
	decoder := json.NewDecoder(input)
	err := decoder.Decode(&machineCode)
	if err != nil {
		return []translator.MachineCodeTerm{}, err
	}
	return machineCode, nil
}

func WriteCode(target io.Writer, machineCode []translator.MachineCodeTerm) error {
	encodedCode, err := json.MarshalIndent(machineCode, "", "  ")
	if err != nil {
		return err
	}
	_, err = target.Write(encodedCode)
	if err != nil {
		return err
	}
	return nil
}
