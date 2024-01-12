package machine

import (
	"github.com/Moleus/comp-arch-lab3/pkg/isa"
)

type AluOperation int

const (
	AluOperationNone AluOperation = iota
	AluOperationAdd
	AluOperationSub
	AluOperationMul
	AluOperationMod
	AluOperationRight
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

type BinaryOperationExec func(left int, right int) int

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

func NewAlu() *Alu {
	return &Alu{
		operation2func: map[AluOperation]BinaryOperationExec{
			AluOperationAdd:   add,
			AluOperationSub:   sub,
			AluOperationMul:   mul,
			AluOperationMod:   mod,
			AluOperationRight: takeRight,
			AluOperationOr:    or,
			AluOperationAnd:   and,
		},
	}
}

type FlagBit int

func (a *Alu) setFlags(value int) {
	a.bitFlags.Carry = value > isa.WordMaxValue || value < isa.WordMinValue
	a.bitFlags.Zero = value == 0
	a.bitFlags.Negative = value < 0
}

type ExecutionParams struct {
	operation   AluOperation
	left        isa.MachineWord
	right       isa.MachineWord
	updateFlags bool
}

func NewAluOp(operation AluOperation) *ExecutionParams {
	return &ExecutionParams{
		operation:   operation,
		left:        isa.NewConstantNumber(0),
		right:       isa.NewConstantNumber(0),
		updateFlags: false,
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

func (p *ExecutionParams) UpdateFlags(updateFlags bool) *ExecutionParams {
	p.updateFlags = updateFlags
	return p
}

func (a *Alu) Execute(executionParams ExecutionParams) (isa.MachineWord, BitFlags) {
	if a.operation2func[executionParams.operation] == nil {
		panic("unknown operation")
	}
	output := a.operation2func[executionParams.operation](executionParams.left.Value, executionParams.right.Value)
	result := executionParams.left
	result.Value = output
	if executionParams.updateFlags {
		a.setFlags(output)
	}
	return result, a.bitFlags
}
