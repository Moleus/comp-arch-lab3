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

package datapath

import (
	"bytes"
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
