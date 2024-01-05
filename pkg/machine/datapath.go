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


// TODO: как будет выглядеть DataPath на Golang?

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

type Register int

const (
  AC Register = iota
  IP
  CR
  PS
  SP
  DR
  AR
)

type RegisterValue struct {
  value int
}

type BitFlags struct {
  ZERO bool
  NEGATIVE bool
  CARRY bool
}

func (rv *RegisterValue) GetValue() int {
  return rv.value
}

func (rv *RegisterValue) GetAddress() int {
  return rv.value
}

type DataPath struct {
	InstructionCounter int
	CurrentTick        int
	input              bytes.Buffer
  // TODO: handle output
	output             bytes.Buffer
	registers          map[Register]int
	memory             []int

  Alu *Alu
}

func NewDataPath(dataInput bytes.Buffer) *DataPath {
  registers := make(map[Register]int)
  memory := make([]int, isa.ADDR_MAX_VALUE + 1)
  alu := NewAlu()
  registers[AC] = 0
  registers[IP] = 0
  registers[CR] = 0
  registers[PS] = 0
  registers[SP] = 0
  registers[DR] = 0
  registers[AR] = 0
	return &DataPath{input: dataInput, memory: memory, registers: registers, Alu: alu}
}

func (dp *DataPath) GetFlags() BitFlags {
  return BitFlags{
    ZERO: dp.registers[PS] & 0x1 == 1,
    NEGATIVE: dp.registers[PS] & 0x2 == 1,
    CARRY: dp.registers[PS] & 0x4 == 1,
  }
}

func (dp *DataPath) IsInterruptRequired() bool {
  // TODO: check binary logic
  return dp.registers[PS] & 0x8 == 1 && dp.registers[PS] & 0x10 == 1
}

func (dp *DataPath) ReadOutput() string {
	return dp.output.String()
}

func (dp *DataPath) SigLatchRegister(register Register, value int) {
  dp.registers[register] = value
}

func (dp *DataPath) GetRegister(register Register) int {
  return dp.registers[register]
}

func (dp *DataPath) ReadMemory(address int) int {
  return dp.memory[address]
}

func (dp *DataPath) WriteMemory() {
  dp.memory[dp.GetRegister(AR)] = dp.GetRegister(DR)
}
