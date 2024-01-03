package translator_test

import (
	"bytes"
	"github.com/Moleus/comp-arch-lab3/cmd/translator"
	"gotest.tools/v3/golden"
	"os"
	"testing"
)

func updateGoldenProgram(t *testing.T, content string) {
	_ = os.Mkdir("testdata", 0755)
	f, err := os.Create("testdata/program.golden.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	_, err = f.WriteString(content)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTranslator(t *testing.T) {
	tr := translator.NewTranslator()
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
	result, err := tr.Translate(contents)
	if err != nil {
		t.Fatal(err)
	}
	if golden.FlagUpdate() {
		updateGoldenProgram(t, result)
	}
	golden.Assert(t, result, "program.golden.json")
}
