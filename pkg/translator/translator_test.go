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
				Index:      0,
				Label:      "label",
				Opcode:     "nop",
				Operand:    "123",
				IsConstant: true,
			}},
		},
		{
			name:  "parse string constant",
			input: "label: word: 2, 'he'",
			expected: []ParsedInstruction{{
				Index:      0,
				Label:      "label",
				Opcode:     "nop",
				Operand:    "2",
				IsConstant: true,
			}, {
				Index:      0,
				Label:      "",
				Opcode:     "nop",
				Operand:    "h",
				IsConstant: true,
			}, {
				Index:      0,
				Label:      "",
				Opcode:     "nop",
				Operand:    "e",
				IsConstant: true,
			}},
		},
		{
			name:  "parse other string constant",
			input: "label: word: 3, 'abc'",
			expected: []ParsedInstruction{{
				Index:      0,
				Label:      "label",
				Opcode:     "nop",
				Operand:    "3",
				IsConstant: true,
			}, {
				Index:      0,
				Label:      "",
				Opcode:     "nop",
				Operand:    "a",
				IsConstant: true,
			}, {
				Index:      0,
				Label:      "",
				Opcode:     "nop",
				Operand:    "b",
				IsConstant: true,
			}, {
				Index:      0,
				Label:      "",
				Opcode:     "nop",
				Operand:    "c",
				IsConstant: true,
			},
			},
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
