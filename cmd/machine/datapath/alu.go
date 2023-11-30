package datapath

import "github.com/Moleus/comp-arch-lab3/pkg/isa"

type BinaryOperationExec func(left int, right int) int

type AluOperation int

const (
	AluOperationAdd AluOperation = iota
	AluOperationSub
	AluOperationMul
	AluOperationDiv
	AluOperationMod
	AluOperationRight
	AluOperationLeft
)

type Alu struct {
	bitFlags       int
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

func takeRight(left int, right int) int {
	return right
}

func takeLeft(left int, right int) int {
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
		},
	}
}

func wrapOverflow(value int) int {
	if value > isa.WORD_MAX_VALUE || value < isa.WORD_MIN_VALUE {
		return (value+(isa.WORD_MAX_VALUE+1))%(2*(isa.WORD_MAX_VALUE+1)) - isa.WORD_MAX_VALUE - 1
	}
	return value
}

type FlagBit int

const (
	ZERO FlagBit = iota
	NEGATIVE
)

func (a *Alu) getBit(bit FlagBit) bool {
	return (a.bitFlags >> bit) & 1
}

func (a *Alu) setBit(bit FlagBit, value bool) {
	if value {
		a.bitFlags |= 1 << bit
	} else {
		a.bitFlags &= ^(1 << bit)
	}
}

func (a *Alu) setFlags(value int) {
	a.setBit(ZERO, value == 0)
	a.setBit(NEGATIVE, value < 0)
}

func (a *Alu) Execute(operation AluOperation, left int, right int) int {
	if a.operation2func[operation] == nil {
		panic("unknown operation")
	}
	output := a.operation2func[operation](left, right)
	wrapOverflow(output)
	a.setFlags(output)
	return output
}
