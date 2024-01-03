/*
Package isa: Instruction Set Architecture (Система команд)
Фон-Неймановская архитектура

Задачи:
- читает машинный код из файла
- записывает машинный код в файл

# По сути занимается сериализацией и десериализаций программы в JSON

Используется в machine.go, controlunit.go и datapath.go
*/
package isa

import (
	"encoding/json"
	"io"
)

/* accumulator based ISA */

// Instruction represents all supported instructions for our architecture
const (
	WORD_WIDTH     = 16
	WORD_MAX_VALUE = 1<<(WORD_WIDTH-1) - 1
	WORD_MIN_VALUE = -1 << (WORD_WIDTH - 1)
)

type Opcode int

const (
  OpcodeAnd Opcode = iota
  OpcodeOr
  OpcodeAdd
  OpcodeSub
  OpcodeCmp
  OpcodeJmp
  OpcodeHlt
  OpcodeIret
  OpcodeIn
  OpcodeOut

  OpcodeLoad
  OpcodeStore

  OpcodePush
  OpcodePop

  OpcodeEi
  OpcodeDi
  OpcodeCla
  OpcodeNop
)

type OpcodeType int

const (
  OpcodeTypeAddress OpcodeType = iota
  OpcodeTypeAddressless
  OpcodeTypeBranch
  OpcodeTypeIO
)

type OpcodeInfo struct {
  instructionType OpcodeType
  stringRepresentation string
}

var (
  opcodeToInfo = map[Opcode]OpcodeInfo{
    OpcodeAnd: {
      instructionType: OpcodeTypeAddress,
      stringRepresentation: "AND",
    },
    OpcodeOr: {
      instructionType: OpcodeTypeAddress,
      stringRepresentation: "OR",
    },
    OpcodeAdd: {
      instructionType: OpcodeTypeAddress,
      stringRepresentation: "ADD",
    },
    OpcodeSub: {
      instructionType: OpcodeTypeAddress,
      stringRepresentation: "SUB",
    },
    OpcodeCmp: {
      instructionType: OpcodeTypeAddress,
      stringRepresentation: "CMP",
    },
    OpcodeJmp: {
      instructionType: OpcodeTypeBranch,
      stringRepresentation: "JMP",
    },
    OpcodeHlt: {
      instructionType: OpcodeTypeAddressless,
      stringRepresentation: "HLT",
    },
    OpcodeIret: {
      instructionType: OpcodeTypeAddressless,
      stringRepresentation: "IRET",
    },
    OpcodeIn: {
      instructionType: OpcodeTypeIO,
      stringRepresentation: "IN",
    },
    OpcodeOut: {
      instructionType: OpcodeTypeIO,
      stringRepresentation: "OUT",
    },
    OpcodeLoad: {
      instructionType: OpcodeTypeAddress,
      stringRepresentation: "LD",
    },
    OpcodeStore: {
      instructionType: OpcodeTypeAddress,
      stringRepresentation: "ST",
    },
    OpcodePush: {
      instructionType: OpcodeTypeAddressless,
      stringRepresentation: "PUSH",
    },
    OpcodePop: {
      instructionType: OpcodeTypeAddressless,
      stringRepresentation: "POP",
    },
    OpcodeEi: {
      instructionType: OpcodeTypeAddressless,
      stringRepresentation: "EI",
    },
    OpcodeDi: {
      instructionType: OpcodeTypeAddressless,
      stringRepresentation: "DI",
    },
    OpcodeCla: {
      instructionType: OpcodeTypeAddressless,
      stringRepresentation: "CLA",
    },
    OpcodeNop: {
      instructionType: OpcodeTypeAddressless,
      stringRepresentation: "NOP",
    },
  }
)

func (o Opcode) Type() OpcodeType {
  return opcodeToInfo[o].instructionType
}

func (o Opcode) String() string {
  return opcodeToInfo[o].stringRepresentation
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

type TermMetaInfo struct {
	LineNum         int    `json:"line_num"`
	OriginalContent string `json:"original_content"`
}

type MachineCodeTerm struct {
	Index    int          `json:"index"`
	Label    string       `json:"label,omitempty"`
	Opcode   Opcode       `json:"opcode"`
	Operand  int          `json:"operand,omitempty"`
	TermInfo TermMetaInfo `json:"term_info"`
}

// TODO: think about dependencies and move MachineCodeTerm in ISA
func ReadCode(input io.Reader) ([]MachineCodeTerm, error) {
	var machineCode []MachineCodeTerm
	decoder := json.NewDecoder(input)
	err := decoder.Decode(&machineCode)
	if err != nil {
		return []MachineCodeTerm{}, err
	}
	return machineCode, nil
}

func WriteCode(target io.Writer, machineCode []MachineCodeTerm) error {
	encodedCode, err := json.MarshalIndent(machineCode, "", "  ")
	if err != nil {
		return err
	}
	_, err = target.Write(encodedCode)
	if err != nil {
		return err
	}
	return nil
}
