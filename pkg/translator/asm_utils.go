package translator

import "strings"

// <label>: word: <value>
func isConstantDeclaration(parts []string) bool {
  return hasLabel(parts) && parts[1] == "word:"
}

func hasLabel(parts []string) bool {
  return strings.HasSuffix(parts[0], ":")
}

func addIndicies(instructions []ParsedInstruction) []ParsedInstruction {
  for i, instruction := range instructions {
    instruction.Index = i
  }
  return instructions
}

