package isa

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Instruction represents all supported instructions for our architecture
const (
	WordWidth    = 32
	WordMaxValue = 1<<(WordWidth-1) - 1
	WordMinValue = -1 << (WordWidth - 1)
	AddrWidth    = 11
	AddrMaxValue = 1<<(AddrWidth) - 1
)

type Opcode int

const (
	OpcodeNop Opcode = iota
	OpcodeAdd
	OpcodeSub
	OpcodeCmp
	OpcodeMod
	OpcodeHlt
	OpcodeIret
	OpcodeInc
	OpcodeDec

	OpcodeIn
	OpcodeOut

	OpcodeLoad
	OpcodeStore

	OpcodePush
	OpcodePop

	OpcodeEi
	OpcodeDi
	OpcodeCla

	OpcodeJmp
	OppcodeJz
	OpcodeJnz
	OpcodeJc
	OpcodeJnc
	OpcodeJn
	OpcodeJnneg
)

type OpcodeType int

const (
	OpcodeTypeAddress OpcodeType = iota
	OpcodeTypeAddressless
	OpcodeTypeBranch
	OpcodeTypeIO
)

type OpcodeInfo struct {
	instructionType      OpcodeType
	stringRepresentation string
}

var (
	opcodeToInfo = map[Opcode]OpcodeInfo{
		OpcodeAdd: {
			instructionType:      OpcodeTypeAddress,
			stringRepresentation: "ADD",
		},
		OpcodeSub: {
			instructionType:      OpcodeTypeAddress,
			stringRepresentation: "SUB",
		},
		OpcodeCmp: {
			instructionType:      OpcodeTypeAddress,
			stringRepresentation: "CMP",
		},
		OpcodeMod: {
			instructionType:      OpcodeTypeAddress,
			stringRepresentation: "MOD",
		},
		OpcodeHlt: {
			instructionType:      OpcodeTypeAddressless,
			stringRepresentation: "HLT",
		},
		OpcodeIret: {
			instructionType:      OpcodeTypeAddressless,
			stringRepresentation: "IRET",
		},
		OpcodeInc: {
			instructionType:      OpcodeTypeAddressless,
			stringRepresentation: "INC",
		},
		OpcodeDec: {
			instructionType:      OpcodeTypeAddressless,
			stringRepresentation: "DEC",
		},
		OpcodeIn: {
			instructionType:      OpcodeTypeIO,
			stringRepresentation: "IN",
		},
		OpcodeOut: {
			instructionType:      OpcodeTypeIO,
			stringRepresentation: "OUT",
		},
		OpcodeLoad: {
			instructionType:      OpcodeTypeAddress,
			stringRepresentation: "LD",
		},
		OpcodeStore: {
			instructionType:      OpcodeTypeAddress,
			stringRepresentation: "ST",
		},
		OpcodePush: {
			instructionType:      OpcodeTypeAddressless,
			stringRepresentation: "PUSH",
		},
		OpcodePop: {
			instructionType:      OpcodeTypeAddressless,
			stringRepresentation: "POP",
		},
		OpcodeEi: {
			instructionType:      OpcodeTypeAddressless,
			stringRepresentation: "EI",
		},
		OpcodeDi: {
			instructionType:      OpcodeTypeAddressless,
			stringRepresentation: "DI",
		},
		OpcodeCla: {
			instructionType:      OpcodeTypeAddressless,
			stringRepresentation: "CLA",
		},
		OpcodeNop: {
			instructionType:      OpcodeTypeAddressless,
			stringRepresentation: "NOP",
		},
		OpcodeJmp: {
			instructionType:      OpcodeTypeBranch,
			stringRepresentation: "JMP",
		},
		OppcodeJz: {
			instructionType:      OpcodeTypeBranch,
			stringRepresentation: "JZ",
		},
		OpcodeJnz: {
			instructionType:      OpcodeTypeBranch,
			stringRepresentation: "JNZ",
		},
		OpcodeJc: {
			instructionType:      OpcodeTypeBranch,
			stringRepresentation: "JC",
		},
		OpcodeJnc: {
			instructionType:      OpcodeTypeBranch,
			stringRepresentation: "JNC",
		},
		OpcodeJn: {
			instructionType:      OpcodeTypeBranch,
			stringRepresentation: "JN",
		},
		OpcodeJnneg: {
			instructionType:      OpcodeTypeBranch,
			stringRepresentation: "JNN",
		},
	}
)

//goland:noinspection GoMixedReceiverTypes
func (o Opcode) Type() OpcodeType {
	return opcodeToInfo[o].instructionType
}

//goland:noinspection GoMixedReceiverTypes
func (o Opcode) String() string {
	return opcodeToInfo[o].stringRepresentation
}

//goland:noinspection GoMixedReceiverTypes
func (o Opcode) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.String())
}

func (o *Opcode) UnmarshalJSON(data []byte) error {
	var opcode string
	err := json.Unmarshal(data, &opcode)
	if err != nil {
		return err
	}
	for opcodeObj, opcodeInfo := range opcodeToInfo {
		if opcodeInfo.stringRepresentation == opcode {
			*o = opcodeObj
		}
	}

	return nil
}

func GetOpcodeFromString(opcode string) (Opcode, error) {
	for opcodeObj, opcodeInfo := range opcodeToInfo {
		if strings.EqualFold(opcodeInfo.stringRepresentation, opcode) {
			return opcodeObj, nil
		}
	}
	return OpcodeNop, fmt.Errorf("unknown opcode: '%s'", opcode)
}

type Program struct {
	StartAddress int
	Instructions []MachineCodeTerm
}

type MachineWord struct {
	Opcode    Opcode
	Value     int
	ValueType ValueType
}

func NewConstantNumber(value int) MachineWord {
	return MachineWord{
		Opcode:    OpcodeNop,
		Value:     value,
		ValueType: ValueTypeNumber,
	}
}

func NewMemoryWordFromIO(ioData IoData) MachineWord {
	return MachineWord{
		Opcode:    OpcodeNop,
		Value:     int(ioData.Char[0]),
		ValueType: ValueTypeChar,
	}
}

func NewMemoryWord(instruction MachineCodeTerm) MachineWord {
	if instruction.Opcode.Type() == OpcodeTypeAddress && instruction.Operand == nil {
		panic(fmt.Sprintf("address instruction without operand: %s", instruction.Opcode))
	}

	operand := 0
	if instruction.Operand != nil {
		operand = *instruction.Operand
	}
	return MachineWord{
		Opcode:    instruction.Opcode,
		Value:     operand,
		ValueType: instruction.OperandType,
	}
}

type ValueType int

const (
	ValueTypeNone ValueType = iota
	ValueTypeNumber
	ValueTypeChar
	ValueTypeAddressDirect
	ValueTypeAddressIndirect
)

type MachineCodeTerm struct {
	Index       int          `json:"index"`
	Label       *string      `json:"label,omitempty"`
	Opcode      Opcode       `json:"opcode"`
	Operand     *int         `json:"operand,omitempty"`
	OperandType ValueType    `json:"operand_type,omitempty"`
	TermInfo    TermMetaInfo `json:"term_info"`
}

type TermMetaInfo struct {
	LineNum         int    `json:"line_num"`
	OriginalContent string `json:"original_content"`
}

type IoData struct {
	ArrivesAt int
	Char      string
}

func ReadCode(input io.Reader) (Program, error) {
	var program Program
	decoder := json.NewDecoder(input)
	err := decoder.Decode(&program)
	if err != nil {
		return Program{}, err
	}
	return program, nil
}

func SerializeCode(program Program) ([]byte, error) {
	return json.MarshalIndent(program, "", "  ")
}

func ReadIoData(input io.Reader) ([]IoData, error) {
	var ioData = make([]IoData, 0)
	decoder := json.NewDecoder(input)
	err := decoder.Decode(&ioData)
	if err != nil && err != io.EOF {
		return []IoData{}, err
	}
	return ioData, nil
}
