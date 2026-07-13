package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestConfirmOCR(t *testing.T) {
	tests := []struct {
		answer string
		want   bool
	}{
		{answer: "y\n", want: true},
		{answer: "YES\n", want: true},
		{answer: "n\n", want: false},
		{answer: "\n", want: false},
	}

	for _, tt := range tests {
		cmd := &cobra.Command{}
		cmd.SetIn(strings.NewReader(tt.answer))
		var out bytes.Buffer
		cmd.SetOut(&out)

		got, err := confirmOCR(cmd, `C:\Screenshots`, "both", true)
		if err != nil {
			t.Fatal(err)
		}
		if got != tt.want {
			t.Fatalf("answer %q: got %t, want %t", tt.answer, got, tt.want)
		}
		if !strings.Contains(out.String(), "[y/N]") {
			t.Fatalf("prompt missing from output: %q", out.String())
		}
	}
}
