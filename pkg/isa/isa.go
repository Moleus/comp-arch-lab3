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
