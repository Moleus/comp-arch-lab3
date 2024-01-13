package machine

import (
	"fmt"
	"io"

	"github.com/Moleus/comp-arch-lab3/pkg/isa"
)

type Register int

const (
	StatusRegisterCarryBit           = 1 << 0
	StatusRegisterZeroBit            = 1 << 2
	StatusRegisterNegativeBit        = 1 << 3
	StatusRegisterEnableInterruptBit = 1 << 5
)

const (
	AC Register = iota
	IP
	CR
	PS
	SP
	DR
	AR
)

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

type BitFlags struct {
	Zero             bool
	Negative         bool
	Carry            bool
	EnableInterrupts bool
}

type AccumulatorSel int

const (
	AccumulatorSelInput AccumulatorSel = iota
	AccumulatorSelAlu
)

type DataPath struct {
	inputBuffer  []isa.IoData
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
		Zero:             dp.registers[PS].Value&StatusRegisterZeroBit > 0,
		Negative:         dp.registers[PS].Value&StatusRegisterNegativeBit > 0,
		Carry:            dp.registers[PS].Value&StatusRegisterCarryBit > 0,
		EnableInterrupts: dp.registers[PS].Value&StatusRegisterEnableInterruptBit > 0,
	}
}

func (dp *DataPath) IsInterruptEnabled() bool {
	return dp.registers[PS].Value&StatusRegisterEnableInterruptBit > 0
}

func (dp *DataPath) SigLatchRegister(register Register, value isa.MachineWord) {
	dp.registers[register] = value
}

func (dp *DataPath) SigLatchAC(aluData isa.MachineWord, sel AccumulatorSel) {
	if sel == AccumulatorSelInput {
		dp.registers[AC] = isa.NewMemoryWordFromIO(dp.inputBuffer[0])
		dp.inputBuffer = dp.inputBuffer[1:]
	} else {
		dp.registers[AC] = aluData
	}
}

func (dp *DataPath) isInputReady() bool {
	return len(dp.inputBuffer) > 0 && dp.inputBuffer[0].ArrivesAt <= dp.clock.GetCurrentTick()
}

func (dp *DataPath) SigWritePortOut() {
	ac := dp.registers[AC]
	if ac.ValueType == isa.ValueTypeChar || ac.Value == 10 {
		if _, err := dp.outputBuffer.Write([]byte{byte(ac.Value)}); err != nil {
			panic(err)
		}
	} else {
		if _, err := dp.outputBuffer.Write([]byte(fmt.Sprintf("%d", ac.Value))); err != nil {
			panic(err)
		}
	}
}

func (dp *DataPath) GetRegister(register Register) isa.MachineWord {
	return dp.registers[register]
}

func (dp *DataPath) ReadMemory(address int) isa.MachineWord {
	return dp.memory[address]
}

func (dp *DataPath) WriteMemory() {
	dp.memory[dp.GetRegister(AR).Value] = dp.GetRegister(DR)
}

func (dp *DataPath) SigExecuteAluOp(aluParams ExecutionParams) isa.MachineWord {
	result, bitFlags := dp.Alu.Execute(aluParams)
	oldPs := dp.registers[PS]
	oldPs.Value = updatePsWithBitFlags(oldPs.Value, bitFlags)
	dp.registers[PS] = oldPs
	return result
}

func updatePsWithBitFlags(oldPs int, bitFlags BitFlags) int {
	if bitFlags.Carry {
		oldPs |= StatusRegisterCarryBit
	} else {
		oldPs &= ^StatusRegisterCarryBit
	}
	if bitFlags.Zero {
		oldPs |= StatusRegisterZeroBit
	} else {
		oldPs &= ^StatusRegisterZeroBit
	}
	if bitFlags.Negative {
		oldPs |= StatusRegisterNegativeBit
	} else {
		oldPs &= ^StatusRegisterNegativeBit
	}
	return oldPs
}
