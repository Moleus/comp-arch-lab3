/*
DataPath - принимает поток данных. Работает с вводом-выводом
Управляется ControlUnit-ом через сигналы.

Реализуем аккумуляторную архитектуру.

состояние аккумулятора может понядобиться выгружать и загружать обратно из памяти.

DataPath - работа через аккумулятор

шагает по памяти по одному слову за такт

данные + команды хранятся в одной памяти

Регистры:
AC - аккумулятор
IP - счетчик команд
SP - указатель стека
DR - регистр данных
BR - буферный регистр
PS - регистр флагов (состояния)

//TODO: написать набор функций, который позволят удобно читать из памяти в нужные регистры и обратно

подумать про адресацию...

# Реализация
Есть набор функций с префиксом signal, которые вызываются из ControlUnit-а

Должна быть реализовано строковое предствление состояние процессора.
*/

package machine

import (
	"bytes"
	"github.com/Moleus/comp-arch-lab3/pkg/isa"
)

type SignalDriven interface {
	SignalLatchAccumulator()
	// TODO: decide on other signals
}

type Memory struct {
	values []int
}

type Registers struct {
	Accumulator        int
	InstructionPointer int
}

type DataPath struct {
	InstructionCounter int
	CurrentTick        int
	input              bytes.Buffer
	output             bytes.Buffer
	registers          Registers
	memory             Memory
}

func NewDataPath(dataInput bytes.Buffer) *DataPath {
	return &DataPath{input: dataInput}
}

func (dp *DataPath) ReadOutput() string {
	return dp.output.String()
}

func (dp *DataPath) SignalLatchAccumulator() {
	dp.registers.Accumulator = dp.memory.values[dp.registers.InstructionPointer]
}

type BinaryOperationExec func(left int, right int) int

type AluOperation int

const (
	AluOperationAdd AluOperation = iota
	AluOperationSub
	AluOperationMul
	AluOperationDiv
	AluOperationMod
	AluOperationRight
	AluOperationLeft
)

type Alu struct {
	bitFlags       int
	operation2func map[AluOperation]BinaryOperationExec
}

func add(left int, right int) int {
	return left + right
}

func sub(left int, right int) int {
	return left - right
}

func mul(left int, right int) int {
	return left * right
}

func div(left int, right int) int {
	return left / right
}

func mod(left int, right int) int {
	return left % right
}

func takeRight(left int, right int) int {
	return right
}

func takeLeft(left int, right int) int {
	return left
}

func NewAlu() *Alu {
	return &Alu{
		operation2func: map[AluOperation]BinaryOperationExec{
			AluOperationAdd:   add,
			AluOperationSub:   sub,
			AluOperationMul:   mul,
			AluOperationDiv:   div,
			AluOperationMod:   mod,
			AluOperationRight: takeRight,
			AluOperationLeft:  takeLeft,
		},
	}
}

func wrapOverflow(value int) int {
	if value > isa.WORD_MAX_VALUE || value < isa.WORD_MIN_VALUE {
		return (value+(isa.WORD_MAX_VALUE+1))%(2*(isa.WORD_MAX_VALUE+1)) - isa.WORD_MAX_VALUE - 1
	}
	return value
}

type FlagBit int

const (
	ZERO FlagBit = iota
	NEGATIVE
)

func (a *Alu) getBit(bit FlagBit) bool {
	return (a.bitFlags >> bit) & 1
}

func (a *Alu) setBit(bit FlagBit, value bool) {
	if value {
		a.bitFlags |= 1 << bit
	} else {
		a.bitFlags &= ^(1 << bit)
	}
}

func (a *Alu) setFlags(value int) {
	a.setBit(ZERO, value == 0)
	a.setBit(NEGATIVE, value < 0)
}

func (a *Alu) Execute(operation AluOperation, left int, right int) int {
	if a.operation2func[operation] == nil {
		panic("unknown operation")
	}
	output := a.operation2func[operation](left, right)
	wrapOverflow(output)
	a.setFlags(output)
	return output
}
