package machine

import (
	"errors"
	"fmt"
	"github.com/Moleus/comp-arch-lab3/pkg/isa"
	"io"
	"strings"
)

const (
	InterruptVectorFirst = 0x0
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

type ControlUnit struct {
	program  isa.Program
	dataPath *DataPath

	ExecutedInstructions int
	clock                *Clock

	stateOutput io.Writer
}

const MaxInstructions = 1_000_000

func NewControlUnit(program isa.Program, dataPath *DataPath, stateOutput io.Writer, clock *Clock) *ControlUnit {
	mapMemory(dataPath, program.Instructions)
	return &ControlUnit{program: program, dataPath: dataPath, stateOutput: stateOutput, clock: clock}
}

func mapMemory(dataPath *DataPath, instructions []isa.MachineCodeTerm) {
	for _, instruction := range instructions {
		instructionWord := isa.NewMemoryWord(instruction)
		dataPath.memory[instruction.Index] = instructionWord
	}
}

type SingleTickOperation func()

func (cu *ControlUnit) SigLatchRegFunc(register Register, value isa.MachineWord) func() {
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
	return cu.SigLatchRegFunc(DR, cu.dataPath.ReadMemory(cu.GetReg(AR).Value))
}

func (cu *ControlUnit) GetReg(register Register) isa.MachineWord {
	return cu.dataPath.GetRegister(register)
}

func (cu *ControlUnit) RunInstructionCycle() error {
	for cu.ExecutedInstructions < MaxInstructions {
		err := cu.DecodeAndExecuteInstruction()
		if err != nil {
			return err
		}
		cu.interruption()
		err = cu.dumpInstructionEnd()
		if err != nil {
			return err
		}
		cu.ExecutedInstructions++
	}
	return errors.New("instructions limit exceeded")
}

func (cu *ControlUnit) DecodeAndExecuteInstruction() error {
	cu.InstructionFetch()
	instruction := cu.GetReg(CR)
	instructionType := instruction.Opcode.Type()

	switch instructionType {
	case isa.OpcodeTypeAddress:
		return cu.decodeAndExecuteAddressInstruction(instruction)
	case isa.OpcodeTypeAddressless:
		return cu.decodeAndExecuteAddresslessInstruction(instruction)
	case isa.OpcodeTypeBranch:
		return cu.decodeAndExecuteBranchInstruction(instruction)
	case isa.OpcodeTypeIO:
		return cu.executeIOInstruction(instruction)
	default:
		panic(fmt.Sprintf("unknown instruction type: %d", instructionType))
	}
}

func (cu *ControlUnit) pushOnStack(register Register) {
	cu.doInOneTick("SP - 1 -> SP",
		cu.SigLatchRegFunc(SP, cu.dataPath.SigExecuteAluOp(*cu.aluDecrement(SP))),
	)
	cu.doInOneTick("SP -> AR",
		cu.SigLatchRegFunc(AR, cu.dataPath.SigExecuteAluOp(*cu.aluRegisterPassThrough(SP))),
	)
	cu.doInOneTick(fmt.Sprintf("%s -> DR", register), cu.SigLatchRegFunc(DR, cu.dataPath.SigExecuteAluOp(*cu.aluRegisterPassThrough(register))))
	cu.doInOneTick("DR -> mem[AR]", cu.SigWriteMemoryFunc())
}

func (cu *ControlUnit) popFromStack(target Register) {
	cu.doInOneTick("SP -> AR",
		cu.SigLatchRegFunc(AR, cu.dataPath.SigExecuteAluOp(*cu.aluRegisterPassThrough(SP))),
	)
	cu.doInOneTick("mem[AR] -> DR; SP + 1 -> SP",
		cu.SigLatchRegFunc(SP, cu.dataPath.SigExecuteAluOp(*cu.aluIncrement(SP))),
		cu.SigReadMemoryFunc())
	cu.doInOneTick(fmt.Sprintf("DR -> %s", target),
		cu.SigLatchRegFunc(target, cu.dataPath.SigExecuteAluOp(*cu.aluRegisterPassThrough(DR))),
	)
}

func (cu *ControlUnit) InstructionFetch() {
	cu.doInOneTick("IP -> AR", cu.SigLatchRegFunc(AR, cu.dataPath.SigExecuteAluOp(*cu.aluRegisterPassThrough(IP))))
	cu.doInOneTick("IP + 1 -> IP; mem[AR] -> DR", cu.SigLatchRegFunc(IP, cu.dataPath.SigExecuteAluOp(*cu.aluIncrement(IP))), cu.SigLatchRegFunc(DR, cu.dataPath.ReadMemory(cu.GetReg(AR).Value)))
	cu.doInOneTick("DR -> CR", cu.SigLatchRegFunc(CR, cu.dataPath.SigExecuteAluOp(*cu.aluRegisterPassThrough(DR))))
}

func (cu *ControlUnit) AddressFetch() {
	cu.doInOneTick("DR -> AR", cu.SigLatchRegFunc(AR, cu.dataPath.SigExecuteAluOp(*cu.aluRegisterPassThrough(DR))))
	cu.doInOneTick("mem[AR] -> DR", cu.SigReadMemoryFunc())
}

func (cu *ControlUnit) OperandFetch() {
	cu.doInOneTick("DR -> AR", cu.SigLatchRegFunc(AR, cu.dataPath.SigExecuteAluOp(*cu.aluRegisterPassThrough(DR))))
	cu.doInOneTick("mem[AR] -> DR", cu.SigLatchRegFunc(DR, cu.dataPath.ReadMemory(cu.GetReg(AR).Value)))
}

func (cu *ControlUnit) decodeAndExecuteAddressInstruction(instruction isa.MachineWord) error {
	if instruction.ValueType == isa.ValueTypeAddressIndirect {
		cu.AddressFetch()
	}
	cu.OperandFetch()

	opcode := instruction.Opcode
	switch {
	case opcode == isa.OpcodeLoad:
		cu.doInOneTick("DR -> AC", cu.SigLatchRegFunc(AC, cu.dataPath.SigExecuteAluOp(*cu.aluRegisterPassThrough(DR).UpdateFlags(true))))
	case opcode == isa.OpcodeStore:
		cu.doInOneTick("AC -> DR", cu.SigLatchRegFunc(DR, cu.dataPath.SigExecuteAluOp(*cu.aluRegisterPassThrough(AC))))
		cu.doInOneTick("DR -> mem[AR]", cu.SigWriteMemoryFunc())
	case opcode == isa.OpcodeCmp:
		cu.doInOneTick("AC - DR -> NZC", func() { cu.dataPath.SigExecuteAluOp(*cu.toAluOp(AC, DR, instruction.Opcode).UpdateFlags(true)) })
	default:
		cu.doInOneTick("AC +- DR -> AC", cu.SigLatchRegFunc(AC, cu.dataPath.SigExecuteAluOp(*cu.toAluOp(AC, DR, instruction.Opcode).UpdateFlags(true))))
	}
	return nil
}

func (cu *ControlUnit) decodeAndExecuteAddresslessInstruction(instruction isa.MachineWord) error {
	switch instruction.Opcode {
	case isa.OpcodeHlt:
		return NewControlUnitError("Halt")
	case isa.OpcodeIret:
		return NewControlUnitError("Interrupt return")
	case isa.OpcodePush:
		cu.pushOnStack(AC)
	case isa.OpcodePop:
		cu.popFromStack(AC)
	case isa.OpcodeEi:
		cu.doInOneTick("1 -> PS[EI]", cu.SigLatchRegFunc(PS, cu.dataPath.SigExecuteAluOp(*NewAluOp(AluOperationOr).SetLeft(cu.GetReg(PS)).SetRightValue(StatusRegisterEnableInterruptBit))))
	case isa.OpcodeDi:
		cu.doInOneTick("0 -> PS[EI]", cu.SigLatchRegFunc(PS, cu.dataPath.SigExecuteAluOp(*NewAluOp(AluOperationAnd).SetLeft(cu.GetReg(PS)).SetRightValue(^(StatusRegisterEnableInterruptBit)))))
	case isa.OpcodeCla:
		cu.doInOneTick("0 -> AC", cu.SigLatchRegFunc(AC, cu.dataPath.SigExecuteAluOp(*cu.toAluOp(AC, 0, instruction.Opcode).UpdateFlags(true))))
	case isa.OpcodeNop:
		cu.doInOneTick("NOP", func() {})
	case isa.OpcodeInc:
		cu.doInOneTick("AC + 1 -> AC", cu.SigLatchRegFunc(AC, cu.dataPath.SigExecuteAluOp(*cu.aluIncrement(AC).UpdateFlags(true))))
	case isa.OpcodeDec:
		cu.doInOneTick("AC - 1 -> AC", cu.SigLatchRegFunc(AC, cu.dataPath.SigExecuteAluOp(*cu.aluDecrement(AC).UpdateFlags(true))))
	default:
		panic(fmt.Sprintf("unknown addressless instruction: %s", instruction.Opcode))
	}
	return nil
}

func (cu *ControlUnit) decodeAndExecuteBranchInstruction(instruction isa.MachineWord) error {
	flags := cu.dataPath.GetFlags()
	opcode := instruction.Opcode

	condition := opcode == isa.OpcodeJc && flags.Carry || opcode == isa.OpcodeJnc && !flags.Carry || opcode == isa.OpcodeJn && flags.Negative || opcode == isa.OpcodeJnneg && !flags.Negative || opcode == isa.OppcodeJz && flags.Zero || opcode == isa.OpcodeJnz && !flags.Zero

	if condition || opcode == isa.OpcodeJmp {
		cu.doInOneTick("DR -> IP", cu.SigLatchRegFunc(IP, cu.dataPath.SigExecuteAluOp(*cu.aluRegisterPassThrough(DR))))
	}
	return nil
}

func (cu *ControlUnit) interruption() {
	for cu.dataPath.isInputReady() && cu.dataPath.IsInterruptEnabled() {
		cu.processInterrupt()
	}
}

func (cu *ControlUnit) processInterrupt() {
	cu.doInOneTick("0 -> PS[EI]",
		cu.SigLatchRegFunc(PS, cu.dataPath.SigExecuteAluOp(*NewAluOp(AluOperationAnd).SetLeft(cu.GetReg(PS)).SetRightValue(^(StatusRegisterEnableInterruptBit)))))

	cu.pushOnStack(IP)
	cu.pushOnStack(PS)

	cu.doInOneTick("intVec -> AR", cu.SigLatchRegFunc(AR, cu.dataPath.SigExecuteAluOp(*NewAluOp(AluOperationAdd).SetLeftValue(InterruptVectorFirst))))
	cu.doInOneTick("mem[AR] -> DR", cu.SigReadMemoryFunc())

	cu.doInOneTick("DR -> IP", cu.SigLatchRegFunc(IP, cu.dataPath.SigExecuteAluOp(*cu.aluRegisterPassThrough(DR))))

	if err := cu.RunInstructionCycle(); err != nil {
		var controlUnitError *ControlUnitError
		if !errors.As(err, &controlUnitError) {
			panic(err)
		}
	}

	cu.popFromStack(PS)
	cu.popFromStack(IP)

	cu.doInOneTick("1 -> PS[EI]", cu.SigLatchRegFunc(PS, cu.dataPath.SigExecuteAluOp(*NewAluOp(AluOperationOr).SetLeft(cu.GetReg(PS)).SetRightValue(StatusRegisterEnableInterruptBit))))
}

func (cu *ControlUnit) executeIOInstruction(instruction isa.MachineWord) error {
	switch instruction.Opcode {
	case isa.OpcodeIn:
		cu.doInOneTick("IN -> AC", cu.dataPath.SigLatchACInput)
	case isa.OpcodeOut:
		cu.doInOneTick("AC -> OUT", cu.dataPath.SigWritePortOut)
	default:
		return fmt.Errorf("unknown IO instruction: %s", instruction.Opcode)
	}
	return nil
}

func (cu *ControlUnit) aluIncrement(register Register) *ExecutionParams {
	return NewAluOp(AluOperationAdd).SetLeft(cu.GetReg(register)).SetRightValue(1)
}

func (cu *ControlUnit) aluDecrement(register Register) *ExecutionParams {
	return NewAluOp(AluOperationSub).SetLeft(cu.GetReg(register)).SetRightValue(1)
}

func (cu *ControlUnit) aluRegisterPassThrough(register Register) *ExecutionParams {
	return NewAluOp(AluOperationAdd).SetLeft(cu.GetReg(register))
}

func (cu *ControlUnit) toAluOp(left Register, right Register, operation isa.Opcode) *ExecutionParams {
	aluOp := opcodeToAluOperation[operation]
	if aluOp == AluOperationNone {
		panic(fmt.Sprintf("unknown opcode: %s", operation))
	}
	return NewAluOp(aluOp).SetLeft(cu.GetReg(left)).SetRight(cu.GetReg(right))
}

func (cu *ControlUnit) PresetInstructionCounter(value int) {
	cu.dataPath.SigLatchRegister(IP, isa.NewConstantNumber(value))
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
	cu.clock.currentTick++
}

func (cu *ControlUnit) dumpState(currentOperationDescription string) error {
	tick := cu.clock.GetCurrentTick()
	memByAR := cu.formatMemByAR(cu.GetReg(AR))
	statusFlags := cu.dataPath.GetFlags()
	formattedFlags := formatFlags(statusFlags)

	formattedRegisters := cu.formatRegistersState()
	outputRow := fmt.Sprintf("t%-4d | %-29s | %s | %s | mem[AR]: %s", tick, currentOperationDescription, formattedRegisters, formattedFlags, memByAR)
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
	if flags.Zero {
		result += "Z"
	} else {
		result += "!Z"
	}
	if flags.Negative {
		result += " N"
	} else {
		result += " !N"
	}
	if flags.Carry {
		result += " C"
	} else {
		result += " !C"
	}
	if flags.EnableInterrupts {
		result += " EI"
	} else {
		result += " DI"
	}
	return result
}

func (cu *ControlUnit) formatRegistersState() string {
	var strRegisters = make([]string, 0)
	registers := []Register{AC, IP, CR, PS, SP, DR, AR}
	for _, register := range registers {
		value := cu.GetReg(register)
		valueToPrint := fmt.Sprintf("%2d", value.Value)
		if register == CR {
			valueToPrint = printInstruction(value)
		}
		strRegisters = append(strRegisters, fmt.Sprintf("%s: %s", register, valueToPrint))
	}
	return strings.Join(strRegisters, ", ")
}

func (cu *ControlUnit) formatMemByAR(arRegister isa.MachineWord) string {
	memContent := cu.dataPath.ReadMemory(arRegister.Value)

	argument := fmt.Sprintf("%d", memContent.Value)
	if isa.ValueTypeChar == memContent.ValueType {
		if memContent.Value == 0 {
			argument = "0"
		} else {
			argument = fmt.Sprintf("'%c'", memContent.Value)
		}
	}

	if memContent.Opcode == isa.OpcodeNop {
		return argument
	}

	if memContent.ValueType == isa.ValueTypeNone {
		return memContent.Opcode.String()
	}

	return fmt.Sprintf("%s %s", memContent.Opcode, argument)
}

func printInstruction(word isa.MachineWord) string {
	if word.ValueType != isa.ValueTypeNone {
		return fmt.Sprintf("%3s %d", word.Opcode, word.Value)
	}
	return fmt.Sprintf("%5s", word.Opcode)
}
