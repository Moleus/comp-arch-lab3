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
	"fmt"
	"io"

	"github.com/Moleus/comp-arch-lab3/pkg/isa"
)

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

const (
	StatusRegisterCarryBit            = 0x0
	StatusRegisterZeroBit             = 0x2
	StatusRegisterNegativeBit         = 0x3
	StatusRegisterEnableInterruptBit  = 0x5
	StatusRegisterRequestInterruptBit = 0x6
)

const (
	AccumulatorIOReadyBit = 0x6
)

type BitFlags struct {
	ZERO     bool
	NEGATIVE bool
	CARRY    bool
}

func (r Register) String() string {
	switch r {
	case AC:
		return "AC"
	case IP:
		return "IP"
	case CR:
		return "CR"
	case PS:
		return "PS"
	case SP:
		return "SP"
	case DR:
		return "DR"
	case AR:
		return "AR"
	default:
		panic(fmt.Sprintf("unknown register: %d", r))
	}
}

type DataPath struct {
	// TODO: maybe move out from isa
	inputBuffer []isa.IoData
	// TODO: handle outputBuffer
	outputBuffer io.Writer
	registers    map[Register]isa.MachineWord
	memory       []isa.MachineWord

	clock TickProvider

	Alu *Alu
}

func NewDataPath(dataInput []isa.IoData, output io.Writer, clock TickProvider) *DataPath {
	registers := make(map[Register]isa.MachineWord)
	for _, register := range []Register{AC, IP, CR, PS, SP, DR, AR} {
		registers[register] = isa.NewConstantNumber(0)
	}
	registers[SP] = isa.NewConstantNumber(isa.AddrMaxValue + 1)
	memory := make([]isa.MachineWord, isa.AddrMaxValue+1)
	alu := NewAlu()
	return &DataPath{inputBuffer: dataInput, outputBuffer: output, memory: memory, registers: registers, Alu: alu, clock: clock}
}

func (dp *DataPath) GetFlags() BitFlags {
	return BitFlags{
		ZERO:     dp.registers[PS].Value&StatusRegisterZeroBit == 1,
		NEGATIVE: dp.registers[PS].Value&StatusRegisterNegativeBit == 1,
		CARRY:    dp.registers[PS].Value&StatusRegisterCarryBit == 1,
	}
}

func (dp *DataPath) IsInterruptRequired() bool {
	return dp.registers[PS].Value&StatusRegisterEnableInterruptBit == 1 && dp.registers[PS].Value&StatusRegisterRequestInterruptBit == 1
}

func (dp *DataPath) WriteOutput(character rune) {
	_, err := dp.outputBuffer.Write([]byte(string(character)))
	if err != nil {
		panic(err)
	}
}

func (dp *DataPath) SigLatchRegister(register Register, value isa.MachineWord) {
	dp.registers[register] = value
}

func (dp *DataPath) SigCheckIrq() {
	if dp.isInputReady() {
		dp.setIoState(true)
	}
}

func (dp *DataPath) SigReadPortIn() {
	inReady := dp.isInputReady()
	if !inReady {
		dp.setIoState(inReady)
		return
	}
	ioData := dp.inputBuffer[0]
	dp.registers[AC] = isa.NewMemoryWordFromIO(ioData)
	dp.inputBuffer = dp.inputBuffer[1:]
	// don't override AC but only set IO ready bit
	dp.setIoState(inReady)
}

func (dp *DataPath) isInputReady() bool {
	return len(dp.inputBuffer) > 0 && dp.inputBuffer[0].ArrivesAt <= dp.clock.GetCurrentTick()
}

func (dp *DataPath) setIoState(isReady bool) {
	reg := dp.registers[AC]
	if isReady {
		reg.Value |= AccumulatorIOReadyBit
	} else {
		reg.Value &= ^(AccumulatorIOReadyBit)
	}
	dp.registers[AC] = reg
}

func (dp *DataPath) SigWritePortOut() {
	if _, err := dp.outputBuffer.Write([]byte{byte(dp.registers[AC].Value)}); err != nil {
		panic(err)
	}
}

func (dp *DataPath) GetRegister(register Register) isa.MachineWord {
	return dp.registers[register]
}

func (dp *DataPath) ReadMemory(address int) isa.MachineWord {
	return dp.memory[address]
}

func (dp *DataPath) WriteMemory() {
	// we need to store u8 bytes in memory
	// we need to store instructions and their parameters in memory
	dp.memory[dp.GetRegister(AR).Value] = dp.GetRegister(DR)
}
