/*
ControlUnit.
Интерпретирует команды.
Управляющие потоки идут в ControlUnit
hw - hardwired. Реализуется как часть модели. microcode не нужен.

На вход получает информацию, на выходе выставляет сигналы. Возможно state register и не нужен...

У ControlUnit должно быть состояние, которое описывает текущее состояние исполнения команды (методичка)

Потактовое исполнение команд.
Цикл команды (стр. 53):
1. Цикл выборки команды (Instruction Fetch)
2. Цикл выборки адреса (Address Fetch)
3. Цикл выборки операнда (Operand Fetch)
4. Цикл исполнения (Execution)
5. Цикл прерывания (Interruption) - нужен для ввода-вывода
*/
package machine

import (
	"github.com/Moleus/comp-arch-lab3/pkg/isa"
	"golang.org/x/text/cases"
)

type ControlUnitError struct {
  message string
}

func NewControlUnitError(message string) *ControlUnitError {
  return &ControlUnitError{message: message}
}

func (e *ControlUnitError) Error() string {
  return e.message
}

// TODO: implement interfacte Debuggable to dump current CPU state
type ControlUnit struct {
  program []isa.MachineCodeTerm
  dataPath *DataPath
  instructionCounter int
  currentTick int

  logger *Logger
}

func (cu *ControlUnit) SigLatchReg(register Register, value int) {
  cu.dataPath.SigLatchRegister(register, value)
}

func (cu *ControlUnit) GetReg(register Register) int {
  return cu.dataPath.GetRegister(register)
}

func (cu *ControlUnit) Tick() {
  cu.currentTick++
}

func (cu *ControlUnit) DecodeAndExecuteInstruction() error {
  instruction := cu.program[cu.instructionCounter]
  instructionType := instruction.Opcode.Type()

  switch instructionType {
  case isa.OpcodeTypeAddress:
    return cu.decodeAndExecuteAddressInstruction(instruction)
  case isa.OpcodeTypeAddressless:
    return cu.decodeAndExecuteAddresslessInstruction(instruction)
  case isa.OpcodeTypeBranch:
    return cu.decodeAndExecuteBranchInstruction(instruction)
  }
  return nil
}

func (cu *ControlUnit) decodeAndExecuteAddressInstruction(instruction isa.MachineCodeTerm) error {
  switch instruction.Opcode {
  // ...
  case isa.OpcodeAdd:
    // cu.dataPath.SigLatchAccumulator()
  }
  return nil
}

func (cu *ControlUnit) decodeAndExecuteAddresslessInstruction(instruction isa.MachineCodeTerm) error {
  switch instruction.Opcode {
    case isa.OpcodeHlt:
      return NewControlUnitError("Halt")
    case isa.OpcodeIret:
      return NewControlUnitError("Interrupt return")
  }
    return nil
  }

func (cu *ControlUnit) decodeAndExecuteBranchInstruction(instruction isa.MachineCodeTerm) error {
  switch instruction.Opcode {
    case isa.OpcodeJmp:
      // TODO: jump
  }
  return nil
}

func (cu *ControlUnit)
