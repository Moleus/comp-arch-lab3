package translator

import "strings"

func isConstantDeclaration(parts []string) bool {
	return hasLabel(parts) && parts[1] == "word:"
}

func hasLabel(parts []string) bool {
	return strings.HasSuffix(parts[0], ":")
}

func addIndices(instructions []ParsedInstruction) []ParsedInstruction {
	for i, instruction := range instructions {
		instruction.Index = i
	}
	return instructions
}

func isNumber(value string) bool {
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func wrapInSlice(instruction ParsedInstruction, err error) ([]ParsedInstruction, error) {
	instructions := make([]ParsedInstruction, 0)
	instructions = append(instructions, instruction)
	return instructions, err
}

func isIndirectAddressing(label string) bool {
	return strings.HasPrefix(label, "(") && strings.HasSuffix(label, ")")
}
