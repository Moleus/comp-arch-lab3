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
	"github.com/Moleus/comp-arch-lab3/pkg/isa"
	"io"
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
	program  isa.Program
	dataPath *DataPath

	// Счетчик команд
	instructionCounter int
	// Счетчик тактов
	tickCounter int

	stateOutput io.Writer
}

func NewControlUnit(program isa.Program, dataPath *DataPath, stateOutput io.Writer) *ControlUnit {
	return &ControlUnit{program: program, dataPath: dataPath, stateOutput: stateOutput}
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

func (cu *ControlUnit) calculate(aluParams ExecutionParams) int {
	return cu.dataPath.Alu.Execute(aluParams)
}

func (cu *ControlUnit) RunInstructionCycle() error {
	cu.presetInstructionCounter(cu.program.StartAddress)
	for cu.instructionCounter < len(cu.program.Instructions) {
		err := cu.DecodeAndExecuteInstruction()
		if err != nil {
			return err
		}
		for cu.dataPath.IsInterruptRequired() {
			cu.processInterrupt()
		}
		err = cu.dumpInstructionEnd()
		if err != nil {
			return err
		}
		cu.instructionCounter++
	}
	fmt.Println("Program finished")
	return nil
}

func (cu *ControlUnit) DecodeAndExecuteInstruction() error {
	instruction := cu.program.Instructions[cu.GetReg(IP)]
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
	cu.doInOneTick("0 -> PS[3]",
		cu.SigLatchRegFunc(PS, cu.calculate(*NewAluOp(AluOperationAnd).SetLeft(cu.GetReg(PS)).SetRight(^(1 << 3)))))

	cu.pushOnStack(IP)
	cu.pushOnStack(PS)

	if err := cu.RunInstructionCycle(); err != nil {
		panic(err)
	}

	cu.popFromStack(PS)
	cu.popFromStack(IP)

	// TODO: check PS bits and offset
	// enable interrupts
	cu.doInOneTick("1 -> PS[3]", cu.SigLatchRegFunc(PS, cu.calculate(*NewAluOp(AluOperationOr).SetLeft(cu.GetReg(PS)).SetRight(1 << 3))))
}

func (cu *ControlUnit) pushOnStack(register Register) {
	cu.doInOneTick("SP -> AR; SP - 1 -> SP",
		cu.SigLatchRegFunc(SP, cu.calculate(cu.aluDecrement(SP))),
		cu.SigLatchRegFunc(AR, cu.calculate(cu.aluRegisterPassthrough(SP))),
	)
	cu.doInOneTick(fmt.Sprintf("%s -> DR", register), cu.SigLatchRegFunc(DR, cu.calculate(cu.aluRegisterPassthrough(register))))
	cu.doInOneTick("DR -> mem[AR]", cu.SigWriteMemoryFunc())
}

func (cu *ControlUnit) popFromStack(target Register) {
	cu.doInOneTick("SP + 1 -> SP; SP -> AR",
		cu.SigLatchRegFunc(SP, cu.calculate(cu.aluIncrement(SP))),
		cu.SigLatchRegFunc(AR, cu.calculate(cu.aluRegisterPassthrough(SP))),
	)
	cu.doInOneTick("mem[AR] -> DR",
		cu.SigReadMemoryFunc())
	cu.doInOneTick(fmt.Sprintf("DR -> %s", target),
		cu.SigLatchRegFunc(target, cu.calculate(cu.aluRegisterPassthrough(DR))),
	)
}

func (cu *ControlUnit) InstructionFetch() {
	// цикл выборки команды
	cu.doInOneTick("IP -> AR", cu.SigLatchRegFunc(AR, cu.calculate(cu.aluRegisterPassthrough(IP))))
	cu.doInOneTick("IP + 1 -> IP; mem[AR] -> DR", cu.SigLatchRegFunc(IP, cu.calculate(cu.aluIncrement(IP))), cu.SigLatchRegFunc(DR, cu.dataPath.ReadMemory(cu.GetReg(AR))))
	cu.doInOneTick("DR -> CR", cu.SigLatchRegFunc(CR, cu.calculate(cu.aluRegisterPassthrough(DR))))
}

func (cu *ControlUnit) AddressFetch() {
	// цикл выборки адреса
	// TODO: у нас только абсолютная адресация, поэтому цикл выборки адреса не используется. Мб его удалить?
	cu.doInOneTick("CR -> AR", cu.SigLatchRegFunc(AR, cu.calculate(cu.aluRegisterPassthrough(CR))))
	cu.doInOneTick("IP + 1 -> IP", cu.SigLatchRegFunc(IP, cu.calculate(cu.aluIncrement(IP))))
}

func (cu *ControlUnit) OperandFetch() {
	// цикл выборки операнда
	cu.doInOneTick("DR -> AR", cu.SigLatchRegFunc(AR, cu.calculate(cu.aluRegisterPassthrough(DR))))
	cu.doInOneTick("mem[AR] -> DR", cu.SigLatchRegFunc(DR, cu.dataPath.ReadMemory(cu.GetReg(AR))))
	// значение лежит в DR
}

// Пробуем реализовать без косвенной адресации. Только прямая абсолютная.
func (cu *ControlUnit) decodeAndExecuteAddressInstruction(instruction isa.MachineCodeTerm) error {
	//cu.AddressFetch()
	cu.OperandFetch()

	opcode := instruction.Opcode
	switch {
	case opcode == isa.OpcodeLoad:
		cu.doInOneTick("DR -> AC", cu.SigLatchRegFunc(AC, cu.calculate(cu.aluRegisterPassthrough(DR))))
	case opcode == isa.OpcodeStore:
		cu.doInOneTick("AC -> DR", cu.SigLatchRegFunc(DR, cu.calculate(cu.aluRegisterPassthrough(AC))))
		cu.doInOneTick("DR -> mem[AR]", cu.SigWriteMemoryFunc())
	case opcode.Type() == isa.OpcodeTypeIO:
	default:
		// арифметическая
		cu.doInOneTick("AC +- DR -> AC", cu.SigLatchRegFunc(AC, cu.calculate(cu.toAluOp(AC, DR, instruction.Opcode))))
	}
	return nil
}

func (cu *ControlUnit) decodeAndExecuteAddresslessInstruction(instruction isa.MachineCodeTerm) error {
	switch instruction.Opcode {
	case isa.OpcodeHlt:
		cu.doInOneTick("HLT", func() {})
		return NewControlUnitError("Halt")
	case isa.OpcodeIret:
		cu.doInOneTick("IRET", func() {})
		return NewControlUnitError("Interrupt return")
	case isa.OpcodePush:
		//  AC -> DR, SP -> AR, SP - 1 -> SP, DR -> mem[AR]
		cu.doInOneTick("AC -> DR",
			cu.SigLatchRegFunc(DR, cu.calculate(cu.aluRegisterPassthrough(AC))),
		)
		cu.doInOneTick("SP -> AR",
			cu.SigLatchRegFunc(AR, cu.calculate(cu.aluRegisterPassthrough(SP))),
		)
		cu.doInOneTick("SP - 1 -> SP; DR -> mem[AR]",
			cu.SigLatchRegFunc(SP, cu.calculate(cu.aluDecrement(SP))),
			cu.SigWriteMemoryFunc(),
		)
	case isa.OpcodePop:
		// SP + 1 -> SP, SP -> AR, mem[SP] -> DR, DR -> AC
		cu.doInOneTick("SP + 1 -> SP",
			cu.SigLatchRegFunc(SP, cu.calculate(cu.aluIncrement(SP))),
		)
		cu.doInOneTick("SP -> AR; mem[AR] -> DR",
			cu.SigLatchRegFunc(AR, cu.calculate(cu.aluRegisterPassthrough(SP))),
			cu.SigReadMemoryFunc(),
		)

	case isa.OpcodeEi:
		cu.doInOneTick("1 -> PS[4]", cu.SigLatchRegFunc(PS, cu.calculate(*NewAluOp(AluOperationOr).SetLeft(cu.GetReg(PS)).SetRight(1 << 4))))
	case isa.OpcodeDi:
		cu.doInOneTick("0 -> PS[4]", cu.SigLatchRegFunc(PS, cu.calculate(*NewAluOp(AluOperationAnd).SetLeft(cu.GetReg(PS)).SetRight(^(1 << 4)))))
	case isa.OpcodeCla:
		// 0 -> AC
		cu.doInOneTick("0 -> AC", cu.SigLatchRegFunc(AC, cu.calculate(cu.toAluOp(AC, 0, instruction.Opcode))))
	case isa.OpcodeNop:
		cu.doInOneTick("NOP", func() {})
	default:
		// unary arithmetic operation
		cu.doInOneTick("AC +- -> AC", // TODO: think about better way to pass left/right if unnecessary
			cu.SigLatchRegFunc(AC, cu.calculate(cu.toAluOp(AC, AC, instruction.Opcode))))
	}
	return nil
}

func (cu *ControlUnit) decodeAndExecuteBranchInstruction(instruction isa.MachineCodeTerm) error {
	flags := cu.dataPath.GetFlags()
	opcode := instruction.Opcode

	condition := opcode == isa.OpcodeJc && flags.CARRY || opcode == isa.OpcodeJnc && !flags.CARRY || opcode == isa.OpcodeJn && flags.NEGATIVE || opcode == isa.OpcodeJnneg && !flags.NEGATIVE

	if condition {
		cu.doInOneTick("AR -> IP", cu.SigLatchRegFunc(IP, cu.calculate(cu.aluRegisterPassthrough(AR))))
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

func (cu *ControlUnit) presetInstructionCounter(value int) {
	cu.dataPath.SigLatchRegister(IP, value)
	cu.instructionCounter = value
}

func (cu *ControlUnit) doInOneTick(description string, singleTickOperation ...SingleTickOperation) {
	for _, op := range singleTickOperation {
		op()
	}
	if err := cu.dumpState(description); err != nil {
		fmt.Println(err)
	}
	cu.tick()
}

func (cu *ControlUnit) tick() {
	cu.tickCounter++
}

func (cu *ControlUnit) dumpState(currentOperationDescription string) error {
	// tick number.
	// PS: NZC
	memByAR := cu.dataPath.ReadMemory(cu.GetReg(AR))
	statusFlags := cu.dataPath.GetFlags()
	formattedFlags := formatFlags(statusFlags)

	instructionRepr := cu.formatCurrentInstruction()
	outputRow := fmt.Sprintf("%-29s | AC: %d | IP: %d | SP: %d | AR: %d | DR: %d | PS: %s | mem[AR]: %d | CR: %s", currentOperationDescription, cu.GetReg(AC), cu.GetReg(IP), cu.GetReg(SP), cu.GetReg(AR), cu.GetReg(DR), formattedFlags, memByAR, instructionRepr)
	_, err := cu.stateOutput.Write([]byte(outputRow + "\n"))
	if err != nil {
		return fmt.Errorf("failed to write state: %w", err)
	}
	return nil
}

func (cu *ControlUnit) dumpInstructionEnd() error {
	if _, err := cu.stateOutput.Write([]byte("\n")); err != nil {
		return err
	}
	return nil
}

func formatFlags(flags BitFlags) string {
	var result string
	if flags.ZERO {
		result += "Z"
	} else {
		result += "!Z"
	}
	if flags.NEGATIVE {
		result += " N"
	} else {
		result += " !N"
	}
	if flags.CARRY {
		result += " C"
	} else {
		result += " !C"
	}
	return result
}

func (cu *ControlUnit) formatCurrentInstruction() string {
	currentInstruction := cu.program.Instructions[cu.instructionCounter]
	var instructionRepr string = "-"
	if currentInstruction.Constant != nil {
		instructionRepr = fmt.Sprintf("%s %s", currentInstruction.Opcode, *currentInstruction.Constant)
	} else if currentInstruction.Operand != nil {
		instructionRepr = fmt.Sprintf("%s %d", currentInstruction.Opcode, *currentInstruction.Operand)
	} else {
		instructionRepr = fmt.Sprintf("%s", currentInstruction.Opcode.String())
	}
	return instructionRepr
}
