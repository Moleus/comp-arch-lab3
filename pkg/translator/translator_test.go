package translator

import (
	"bytes"
	"gotest.tools/v3/assert"
	"strings"
	"testing"

	"github.com/Moleus/comp-arch-lab3/pkg/isa"
	"gotest.tools/v3/golden"
)

func TestTranslator(t *testing.T) {
	tr := NewTranslator()
	f := golden.Open(t, "program.input.asm")
	defer f.Close()

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(f)
	if err != nil {
		t.Fatal(err)
	}
	contents := buf.String()
	if err != nil {
		t.Fatal(err)
	}
	machineCode, err := tr.Translate(contents)
	if err != nil {
		t.Fatal(err)
	}
	serialized, err := isa.SerializeCode(machineCode)

	golden.Assert(t, string(serialized), "program.golden.json")
}

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
