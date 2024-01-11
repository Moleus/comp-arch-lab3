package translator

import (
	"gotest.tools/v3/assert"
	"strings"
	"testing"

	"github.com/Moleus/comp-arch-lab3/pkg/isa"
)

func TestParseConstant(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []ParsedInstruction
	}{
		{
			name:  "parse numeric constant",
			input: "label: word: 123",
			expected: []ParsedInstruction{{
				Index:     0,
				Label:     "label",
				Opcode:    isa.OpcodeNop.String(),
				Operand:   123,
				ValueType: isa.ValueTypeNumber,
			}},
		},
		{
			name:  "parse string constant",
			input: "label: word: 'he'",
			expected: []ParsedInstruction{{
				Index:     0,
				Label:     "label",
				Opcode:    isa.OpcodeNop.String(),
				Operand:   'h',
				ValueType: isa.ValueTypeChar,
			}, {
				Index:     0,
				Label:     "",
				Opcode:    isa.OpcodeNop.String(),
				Operand:   'e',
				ValueType: isa.ValueTypeChar,
			}, {
				Index:     0,
				Label:     "",
				Opcode:    isa.OpcodeNop.String(),
				Operand:   0,
				ValueType: isa.ValueTypeChar,
			}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parts := strings.Fields(test.input)
			instructions, err := parseConstantDeclaration(parts)
			if err != nil {
				t.Fatal(err)
			}
			for i, instruction := range instructions {
				assert.Equal(t, instruction, test.expected[i])
			}
		})
	}
}
