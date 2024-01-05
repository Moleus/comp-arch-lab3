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
	"fmt"
	"log/slog"

	"github.com/Moleus/comp-arch-lab3/pkg/isa"
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

func NewControlUnit(program []isa.MachineCodeTerm, dataPath *DataPath, logger *slog.Logger) *ControlUnit {
  return &ControlUnit{program: program, dataPath: dataPath, logger: logger}
}

type SingleTickOperation func()

func (cu *ControlUnit) SigLatchRegFunc(register Register, value int) func() {
  return func() {
    cu.dataPath.SigLatchRegister(register, value)
  }
}

func (cu *ControlUnit) SigWriteMemoryFunc() func() {
  return func() {
    cu.dataPath.WriteMemory()
  }
}

func (cu *ControlUnit) SigReadMemoryFunc() func() {
  return cu.SigLatchRegFunc(DR, cu.dataPath.ReadMemory(cu.GetReg(AR)))
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

func (cu *ControlUnit) RunInstructionCycle() error {
  for {
    err := cu.DecodeAndExecuteInstruction()
    if err != nil {
      return err
    }
    for cu.dataPath.IsInterruptRequired() {
      cu.processInterrupt()
    }
    cu.instructionCounter++
  }
}

func (cu *ControlUnit) DecodeAndExecuteInstruction() error {
  instruction := cu.program[cu.instructionCounter]
  instructionType := instruction.Opcode.Type()

  cu.InstructionFetch()

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

func (cu *ControlUnit) processInterrupt() {
  // TODO: check PS bits!!! 5 - EI, 6 - IRQ
  // disable interrupts and save PS on stack
  cu.doInOneTick(cu.SigLatchRegFunc(PS, cu.calculate(*NewAluOp(AluOperationAnd).SetLeft(cu.GetReg(PS)).SetRight(^(1 << 3)))))

  cu.pushOnStack(IP)
  cu.pushOnStack(PS)

  if err:= cu.RunInstructionCycle(); err != nil {
    cu.logger.Debug(fmt.Sprintf("Interrupt error: %s", err.Error()))
  }

  cu.popFromStack(PS)
  cu.popFromStack(IP)

  // TODO: check PS bits and offset
  // enable interrupts
  cu.doInOneTick(cu.SigLatchRegFunc(PS, cu.calculate(*NewAluOp(AluOperationOr).SetLeft(cu.GetReg(PS)).SetRight(1 << 3))))
}

func (cu *ControlUnit) pushOnStack(register Register) {
  // ~0 + SP → SP, AR; reg → DR; DR → MEM(AR)
  cu.doInOneTick(
    cu.SigLatchRegFunc(SP, cu.calculate(cu.aluDecrement(SP))),
    cu.SigLatchRegFunc(AR, cu.calculate(cu.aluRegisterPassthrough(SP))),
    )
  cu.doInOneTick(cu.SigLatchRegFunc(DR, cu.calculate(cu.aluRegisterPassthrough(register))),)
  cu.doInOneTick(cu.SigWriteMemoryFunc(),)
}

func (cu *ControlUnit) popFromStack(target Register) {
  // SP + 1 -> SP, AR; mem[AR] -> DR; DR -> target
  cu.doInOneTick(
    cu.SigLatchRegFunc(SP, cu.calculate(cu.aluIncrement(SP))),
    cu.SigLatchRegFunc(AR, cu.calculate(cu.aluRegisterPassthrough(SP))),
    )
  cu.doInOneTick(
    cu.SigReadMemoryFunc(),)
  cu.doInOneTick(
    cu.SigLatchRegFunc(target, cu.calculate(cu.aluRegisterPassthrough(DR))),
    )
}

func (cu *ControlUnit) InstructionFetch() {
  // цикл выборки команды
  // IP -> AR
  cu.doInOneTick(cu.SigLatchRegFunc(AR, cu.calculate(cu.aluRegisterPassthrough(IP))))
  // AR -> IP, IP + 1 -> IP
  cu.doInOneTick(cu.SigLatchRegFunc(IP, cu.calculate(cu.aluIncrement(IP))), cu.SigLatchRegFunc(DR, cu.dataPath.ReadMemory(cu.GetReg(AR))))
  // DR -> CR
  cu.doInOneTick(cu.SigLatchRegFunc(AC, cu.calculate(cu.aluRegisterPassthrough(DR))))
}

func (cu *ControlUnit) AddressFetch() {
  // цикл выборки адреса
  // IP -> AR
  cu.doInOneTick(cu.SigLatchRegFunc(AR, cu.calculate(cu.aluRegisterPassthrough(IP))))
  // IP + 1 -> IP
  cu.doInOneTick(cu.SigLatchRegFunc(IP, cu.calculate(cu.aluIncrement(IP))))
}

func (cu *ControlUnit) OperandFetch() {
    // цикл выборки операнда
    // DR -> AR
    cu.doInOneTick(cu.SigLatchRegFunc(AR, cu.calculate(cu.aluRegisterPassthrough(DR))))
    // memory[AR] -> DR
    cu.doInOneTick(cu.SigLatchRegFunc(DR, cu.dataPath.ReadMemory(cu.GetReg(AR))))
    // значение лежит в DR
}

// Пробуем реализовать без косвенной адресации. Только прямая абсолютная.
func (cu *ControlUnit) decodeAndExecuteAddressInstruction(instruction isa.MachineCodeTerm) error {
  cu.AddressFetch()
  cu.OperandFetch()

  opcode := instruction.Opcode
  switch {
  case opcode == isa.OpcodeLoad:
    // DR -> AC
    cu.doInOneTick(cu.SigLatchRegFunc(AC, cu.calculate(cu.aluRegisterPassthrough(DR))))
  case opcode == isa.OpcodeStore:
    // AC -> DR
    cu.doInOneTick(cu.SigLatchRegFunc(DR, cu.calculate(cu.aluRegisterPassthrough(AC))))
    // DR -> memory[AR]
    cu.doInOneTick(cu.SigWriteMemoryFunc())
  case opcode.Type() == isa.OpcodeTypeIO:
  default:
    // арифметическая
    cu.doInOneTick(cu.SigLatchRegFunc(AC, cu.calculate(cu.toAluOp(AC, DR, instruction.Opcode))))
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
    //  AC -> DR, SP -> AR, SP - 1 -> SP, DR -> mem[AR]
    cu.doInOneTick(
      cu.SigLatchRegFunc(DR, cu.calculate(cu.aluRegisterPassthrough(AC))),
      )
    cu.doInOneTick(
      cu.SigLatchRegFunc(AR, cu.calculate(cu.aluRegisterPassthrough(SP))),
      )
    cu.doInOneTick(
      cu.SigLatchRegFunc(SP, cu.calculate(cu.aluDecrement(SP))),
      cu.SigWriteMemoryFunc(),
      )
  case isa.OpcodePop:
    // SP + 1 -> SP, SP -> AR, mem[SP] -> DR, DR -> AC
    cu.doInOneTick(
      cu.SigLatchRegFunc(SP, cu.calculate(cu.aluIncrement(SP))),
      )
    cu.doInOneTick(
      cu.SigLatchRegFunc(AR, cu.calculate(cu.aluRegisterPassthrough(SP))),
      cu.SigReadMemoryFunc(),
      )

  case isa.OpcodeEi:
    cu.doInOneTick(cu.SigLatchRegFunc(PS, cu.calculate(*NewAluOp(AluOperationOr).SetLeft(cu.GetReg(PS)).SetRight(1 << 4))))
  case isa.OpcodeDi:
    cu.doInOneTick(cu.SigLatchRegFunc(PS, cu.calculate(*NewAluOp(AluOperationAnd).SetLeft(cu.GetReg(PS)).SetRight(^(1 << 4)))))
  case isa.OpcodeCla:
    // 0 -> AC
    cu.doInOneTick( cu.SigLatchRegFunc(AC, cu.calculate(cu.toAluOp(AC, 0, instruction.Opcode))))
  case isa.OpcodeNop:
    cu.doInOneTick(func() {})
  default:
    // unary arithmetic operation
    cu.doInOneTick(// TODO: think about better way to pass left/right if unnecessary
      cu.SigLatchRegFunc(AC, cu.calculate(cu.toAluOp(AC, AC, instruction.Opcode))))
}
  return nil
}

func (cu *ControlUnit) decodeAndExecuteBranchInstruction(instruction isa.MachineCodeTerm) error {
  flags := cu.dataPath.GetFlags()
  opcode := instruction.Opcode

  condition := opcode == isa.OpcodeJc && flags.CARRY || opcode == isa.OpcodeJnc && !flags.CARRY || opcode == isa.OpcodeJn && flags.NEGATIVE || opcode == isa.OpcodeJnneg && !flags.NEGATIVE

  if condition {
    // AR -> IP
    cu.doInOneTick(cu.SigLatchRegFunc(IP, cu.calculate(cu.aluRegisterPassthrough(AR))))
  }
  return nil
}

func (cu *ControlUnit) aluIncrement(register Register) ExecutionParams {
  return *NewAluOp(AluOperationAdd).SetLeft(1).SetRight(cu.GetReg(register))
}

func (cu *ControlUnit) aluDecrement(register Register) ExecutionParams {
  return *NewAluOp(AluOperationSub).SetLeft(cu.GetReg(register)).SetRight(1)
}

func (cu *ControlUnit) aluRegisterPassthrough(register Register) ExecutionParams {
  // called for DR for address commands
  return *NewAluOp(AluOperationAdd).SetRight(cu.GetReg(register))
}

func (cu *ControlUnit) toAluOp(left Register, right Register, operation isa.Opcode) ExecutionParams {
  aluOp := opcodeToAluOperation[operation]
  return *NewAluOp(aluOp).SetLeft(cu.GetReg(left)).SetRight(cu.GetReg(right))
}

func (cu *ControlUnit) doInOneTick(singleTickOperation ...SingleTickOperation) {
  for _, op := range singleTickOperation {
    op()
  }
  cu.Tick()
}
