package machine

import (
	"github.com/Moleus/comp-arch-lab3/pkg/isa"
)

type BinaryOperationExec func(left int, right int) int

type AluOperation int

const (
	AluOperationNone AluOperation = iota
	AluOperationAdd
	AluOperationSub
	AluOperationMul
	AluOperationDiv
	AluOperationMod
	AluOperationRight
	AluOperationLeft
	AluOperationOr
	AluOperationAnd
)

var (
	opcodeToAluOperation = map[isa.Opcode]AluOperation{
		isa.OpcodeAdd: AluOperationAdd,
		isa.OpcodeSub: AluOperationSub,
		isa.OpcodeCla: AluOperationRight,
		isa.OpcodeCmp: AluOperationSub,
		isa.OpcodeMod: AluOperationMod,
	}
)

type Alu struct {
	bitFlags       BitFlags
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

func or(left int, right int) int {
	return left | right
}

func and(left int, right int) int {
	return left & right
}

func takeRight(_ int, right int) int {
	return right
}

func takeLeft(left int, _ int) int {
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
			AluOperationOr:    or,
			AluOperationAnd:   and,
		},
	}
}

func wrapOverflow(value int) int {
	if value > isa.WordMaxValue || value < isa.WordMinValue {
		return (value+(isa.WordMaxValue+1))%(2*(isa.WordMaxValue+1)) - isa.WordMaxValue - 1
	}
	return value
}

type FlagBit int

const (
	ZERO FlagBit = iota
	NEGATIVE
	CARRY
)

func (a *Alu) setFlags(value int) {
	a.bitFlags.CARRY = value > isa.WordMaxValue || value < isa.WordMinValue
	a.bitFlags.ZERO = value == 0
	a.bitFlags.NEGATIVE = value < 0
}

type ExecutionParams struct {
	operation       AluOperation
	left            isa.MachineWord
	right           isa.MachineWord
	updateRegisters bool
}

func NewAluOp(operation AluOperation) *ExecutionParams {
	return &ExecutionParams{
		operation:       operation,
		left:            isa.NewConstantNumber(0),
		right:           isa.NewConstantNumber(0),
		updateRegisters: false,
	}
}

func (p *ExecutionParams) SetLeft(left isa.MachineWord) *ExecutionParams {
	p.left = left
	return p
}

func (p *ExecutionParams) SetLeftValue(left int) *ExecutionParams {
	p.left = isa.NewConstantNumber(left)
	return p
}

func (p *ExecutionParams) SetRight(right isa.MachineWord) *ExecutionParams {
	p.right = right
	return p
}

func (p *ExecutionParams) SetRightValue(right int) *ExecutionParams {
	p.right = isa.NewConstantNumber(right)
	return p
}

func (p *ExecutionParams) UpdateRegisters(updateRegisters bool) *ExecutionParams {
	p.updateRegisters = updateRegisters
	return p
}

func (a *Alu) Execute(executionParams ExecutionParams) (isa.MachineWord, BitFlags) {
	if a.operation2func[executionParams.operation] == nil {
		panic("unknown operation")
	}
	output := a.operation2func[executionParams.operation](executionParams.left.Value, executionParams.right.Value)
	result := executionParams.left
	result.Value = output
	wrapOverflow(output)
	if executionParams.updateRegisters {
		a.setFlags(output)
	}
	return result, a.bitFlags
}
