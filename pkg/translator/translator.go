package translator

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/Moleus/comp-arch-lab3/pkg/isa"
)

type Translator interface {
	Translate(input string) (isa.Program, error)
	GetLinesOfCode() int
}

type AsmTranslator struct {
	instructions []ParsedInstruction
	currentIndex int

	LinesOfCode int
}

func NewTranslator() Translator {
	instructions := make([]ParsedInstruction, 0)
	return &AsmTranslator{instructions: instructions, currentIndex: 0}
}

func (t *AsmTranslator) GetLinesOfCode() int {
	return t.LinesOfCode
}

type ParsedInstruction struct {
	Index        int
	Label        string
	Opcode       string
	ValueType    isa.ValueType
	Operand      int
	LabelOperand string
	MetaInfo     isa.TermMetaInfo
}

func NewConstant(label string, operand int, valueType isa.ValueType) ParsedInstruction {
	return ParsedInstruction{Label: label, Operand: operand, ValueType: valueType, Opcode: isa.OpcodeNop.String()}
}

type ParseError struct {
	message     string
	lineContent string
	line        int
}

func (e ParseError) Error() string {
	return fmt.Sprintf("Parse error at %d ('%s'): %s", e.line, e.lineContent, e.message)
}

func NewParseError(message string, lineContent string, line int) error {
	return ParseError{message, lineContent, line}
}

func (t *AsmTranslator) Translate(input string) (isa.Program, error) {
	if err := t.ParseInstructions(input); err != nil {
		return isa.Program{}, err
	}
	t.instructions = addIndices(t.instructions)
	machineCode, err := t.convertTermsToMachineCode()
	if err != nil {
		return isa.Program{}, err
	}
	return addStartAddress(machineCode)
}

func (t *AsmTranslator) ParseInstructions(input string) error {
	lines := strings.Split(input, "\n")
	for i, line := range lines {
		line := strings.Split(line, ";")[0]
		line = strings.TrimSpace(line)
		err := t.parseLine(line, i+1)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *AsmTranslator) parseLine(line string, lineNumber int) error {
	metaInfo := isa.TermMetaInfo{LineNum: lineNumber, OriginalContent: line}
	parts := strings.SplitN(line, " ", 3)

	if len(parts) == 0 || parts[0] == "" {
		return nil
	}

	if parts[0] == "word:" {
		return NewParseError("Don't use `word` as a label. It's reserved", line, 0)
	}

	t.LinesOfCode++

	if isConstantDeclaration(parts) {
		instructions, err := parseConstantDeclaration(parts)
		if err != nil {
			return NewParseError(fmt.Sprintf("failed to parse constant: %s", err.Error()), line, lineNumber)
		}
		for _, instruction := range instructions {
			instruction.MetaInfo = metaInfo
			t.addConstant(instruction)
		}
	} else {
		instruction := t.parseInstructionDeclaration(parts)
		instruction.MetaInfo = metaInfo
		t.addInstruction(instruction)
	}

	return nil
}

func parseConstantDeclaration(parts []string) ([]ParsedInstruction, error) {
	label := strings.Split(parts[0], ":")[0]
	argument := strings.TrimSpace(parts[2])

	switch {
	case strings.HasPrefix(argument, "'") && strings.HasSuffix(argument, "'"):
		return parseConstString(label, argument), nil
	case isNumber(argument):
		return wrapInSlice(parseConstNumber(label, argument))
	default:
		return wrapInSlice(parseAddressConstantDeclaration(label, argument))
	}
}

func (t *AsmTranslator) parseInstructionDeclaration(parts []string) ParsedInstruction {
	instruction := ParsedInstruction{}
	if hasLabel(parts) {
		label := strings.Split(parts[0], ":")[0]
		instruction.Label = label
		parts = parts[1:]
	}
	instruction.Opcode = parts[0]
	if len(parts) > 1 {
		instruction = addLabelOperand(instruction, parts[1])
	}
	return instruction
}

func addLabelOperand(instruction ParsedInstruction, label string) ParsedInstruction {
	if isIndirectAddressing(label) {
		instruction.ValueType = isa.ValueTypeAddressIndirect
		label = strings.Trim(label, "()")
	} else {
		instruction.ValueType = isa.ValueTypeAddressDirect
	}
	instruction.LabelOperand = label
	return instruction
}

func parseConstString(label string, value string) []ParsedInstruction {
	value = strings.Trim(value, "'")
	instructions := make([]ParsedInstruction, 0)
	for _, char := range value {
		instructions = append(instructions, NewConstant("", int(char), isa.ValueTypeChar))
	}
	instructions[0].Label = label
	instructions = append(instructions, NewConstant("", 0, isa.ValueTypeChar))
	return instructions
}

func parseConstNumber(label string, value string) (ParsedInstruction, error) {
	number, err := strconv.Atoi(value)
	if err != nil {
		return ParsedInstruction{}, fmt.Errorf("failed to parse number: %s", value)
	}
	return NewConstant(label, number, isa.ValueTypeNumber), nil
}

func parseAddressConstantDeclaration(label string, argument string) (ParsedInstruction, error) {
	return ParsedInstruction{Label: label, Opcode: isa.OpcodeNop.String(), ValueType: isa.ValueTypeAddressDirect, LabelOperand: argument}, nil
}

func (t *AsmTranslator) addConstant(instruction ParsedInstruction) {
	t.addInstruction(instruction)
}

func (t *AsmTranslator) addInstruction(instruction ParsedInstruction) {
	instruction.Index = t.currentIndex
	t.instructions = append(t.instructions, instruction)
	t.currentIndex++
}

func (t *AsmTranslator) labelToAddress(label string) (int, error) {
	for _, instruction := range t.instructions {
		if instruction.Label == label {
			return instruction.Index, nil
		}
	}
	return 0, fmt.Errorf("label '%s' not found", label)
}

func (t *AsmTranslator) convertTermsToMachineCode() (machineCode []isa.MachineCodeTerm, err error) {
	machineCode = make([]isa.MachineCodeTerm, len(t.instructions))
	for i, instruction := range t.instructions {
		var label *string
		if instruction.Label != "" {
			label = new(string)
			*label = instruction.Label
		}
		operand, err := t.inferOperand(instruction)
		if err != nil {
			return []isa.MachineCodeTerm{}, err
		}

		opcode, err := isa.GetOpcodeFromString(instruction.Opcode)
		if err != nil {
			return []isa.MachineCodeTerm{}, err
		}

		operandType := instruction.ValueType
		newMachineCodeTerm := isa.MachineCodeTerm{
			Index:       instruction.Index,
			Label:       label,
			Opcode:      opcode,
			Operand:     operand,
			OperandType: operandType,
			TermInfo:    instruction.MetaInfo,
		}
		machineCode[i] = newMachineCodeTerm
	}
	return machineCode, nil
}

func (t *AsmTranslator) inferOperand(instruction ParsedInstruction) (*int, error) {
	var operand = new(int)
	var err error

	switch instruction.ValueType {
	case isa.ValueTypeNone:
		return nil, nil
	case isa.ValueTypeChar, isa.ValueTypeNumber:
		*operand = instruction.Operand
		return operand, nil
	case isa.ValueTypeAddressDirect, isa.ValueTypeAddressIndirect:
		if instruction.LabelOperand == "" {
			panic(fmt.Sprintf("label operand is empty: %s", instruction.Opcode))
		}
		*operand, err = t.labelToAddress(instruction.LabelOperand)
		return operand, err
	default:
		panic(fmt.Sprintf("unknown operand type: %d", instruction.ValueType))
	}
}

func addStartAddress(machineCode []isa.MachineCodeTerm) (isa.Program, error) {
	startTerm := slices.IndexFunc(machineCode, func(term isa.MachineCodeTerm) bool {
		return term.Label != nil && *term.Label == "start"
	})
	if startTerm == -1 {
		return isa.Program{}, fmt.Errorf("start label not found")
	}
	startAddress := machineCode[startTerm].Index

	return isa.Program{
		StartAddress: startAddress,
		Instructions: machineCode,
	}, nil
}
