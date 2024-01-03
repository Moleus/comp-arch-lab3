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
	"log/slog"
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

  logger *slog.Logger
}

type SingleTickOperation func()

func (cu *ControlUnit) SigLatchReg(register Register, value int) {
  cu.dataPath.SigLatchRegister(register, value)
}

func (cu *ControlUnit) SigWriteMemory() {
  cu.dataPath.WriteMemory()
}

func (cu *ControlUnit) GetReg(register Register) int {
  return cu.dataPath.GetRegister(register)
}

func (cu *ControlUnit) Tick() {
  cu.currentTick++
}

func (cu *ControlUnit) calculate(aluParams ExecutionParams) int {
  return cu.dataPath.Alu.Execute(aluParams)
}

func (cu *ControlUnit) DecodeAndExecuteInstruction() error {
  instruction := cu.program[cu.instructionCounter]
  instructionType := instruction.Opcode.Type()

  // цикл выборки команды
  cu.doInOneTick(func() {
    // IP -> AR
    cu.SigLatchReg(AR, cu.calculate(cu.aluRegisterPassthrough(IP)))
  })
  cu.doInOneTick(func() {
    // AR -> IP, IP + 1 -> IP
    cu.SigLatchReg(IP, cu.calculate(cu.aluIncrement(IP)))
    cu.SigLatchReg(DR, cu.dataPath.ReadMemory(cu.GetReg(AR)))
  })
  cu.doInOneTick(func() {
    // DR -> CR
    cu.SigLatchReg(AC, cu.calculate(cu.aluRegisterPassthrough(DR)))
  })

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

func (cu *ControlUnit) aluIncrement(register Register) ExecutionParams {
  return *NewAluOp(AluOperationAdd).SetLeft(1).SetRight(cu.GetReg(register))
}

func (cu *ControlUnit) aluRegisterPassthrough(register Register) ExecutionParams {
  // called for DR for address commands
  return *NewAluOp(AluOperationAdd).SetRight(cu.GetReg(register))
}

func (cu *ControlUnit) toAluOp(left Register, right Register, operation isa.Opcode) ExecutionParams {
  aluOp := opcodeToAluOperation[operation]
  return *NewAluOp(aluOp).SetLeft(cu.GetReg(left)).SetRight(cu.GetReg(right))
}

func (cu *ControlUnit) doInOneTick(singleTickOperation SingleTickOperation) {
  singleTickOperation()
  cu.Tick()
}

// Пробуем реализовать без косвенной адресации. Только прямая абсолютная.
func (cu *ControlUnit) decodeAndExecuteAddressInstruction(instruction isa.MachineCodeTerm) error {
  // цикл выборки операнда
  cu.doInOneTick(func() {
    // DR -> AR
    cu.SigLatchReg(AR, cu.calculate(cu.aluRegisterPassthrough(DR)))
  })
  cu.doInOneTick(func() {
    // memory[AR] -> DR
    cu.SigLatchReg(DR, cu.dataPath.ReadMemory(cu.GetReg(AR)))
  })
  // значение лежит в DR

  opcode := instruction.Opcode
  switch {
  case opcode == isa.OpcodeLoad:
    // DR -> AC
    cu.doInOneTick(func() {
      cu.SigLatchReg(AC, cu.calculate(cu.aluRegisterPassthrough(DR)))
    })
  case opcode == isa.OpcodeStore:
    // AC -> DR
    cu.doInOneTick(func() {
      cu.SigLatchReg(DR, cu.calculate(cu.aluRegisterPassthrough(AC)))
    })
    // DR -> memory[AR]
    cu.doInOneTick(func() {
      cu.SigWriteMemory()
    })
  case opcode.Type() == isa.OpcodeTypeIO:
  default:
    // арифметическая
    cu.doInOneTick(func() {
      cu.SigLatchReg(AC, cu.calculate(cu.toAluOp(AC, DR, instruction.Opcode)))
    })
  }
  return nil
}

func (cu *ControlUnit) decodeAndExecuteAddresslessInstruction(instruction isa.MachineCodeTerm) error {
  switch instruction.Opcode {
  case isa.OpcodeHlt:
    cu.doInOneTick(func() {})
    return NewControlUnitError("Halt")
  case isa.OpcodeIret:
    cu.doInOneTick(func() {})
    return NewControlUnitError("Interrupt return")
  case isa.OpcodePush:
    cu.doInOneTick(func() {

    })
  case isa.OpcodePop:
    cu.doInOneTick(func() {
    })
  case isa.OpcodeEi:
    cu.doInOneTick(func() {
      cu.SigLatchReg(PS, cu.calculate(*NewAluOp(AluOperationOr).SetLeft(cu.GetReg(PS)).SetRight(1 << 4)))
    })
  case isa.OpcodeDi:
    cu.doInOneTick(func() {
      cu.SigLatchReg(PS, cu.calculate(*NewAluOp(AluOperationAnd).SetLeft(cu.GetReg(PS)).SetRight(^(1 << 4))))
    })
  case isa.OpcodeCla:
    cu.doInOneTick(func() {
      // 0 -> AC
      cu.SigLatchReg(AC, cu.calculate(cu.toAluOp(AC, 0, instruction.Opcode)))
    })
  case isa.OpcodeNop:
  default:
    // unary arithmetic operation
    cu.doInOneTick(func() {
      // TODO: think about better way to pass left/right if unnecessary
      cu.SigLatchReg(AC, cu.calculate(cu.toAluOp(AC, AC, instruction.Opcode)))
    })
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

