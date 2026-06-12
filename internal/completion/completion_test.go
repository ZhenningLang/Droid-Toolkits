package completion

import (
	"strings"
	"testing"
)

func TestGenerateIncludesSupportedCommandsForAllShells(t *testing.T) {
	shells := []string{"bash", "zsh", "fish"}

	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			script, err := Generate(shell)
			if err != nil {
				t.Fatalf("Generate(%q) error = %v", shell, err)
			}
			for _, command := range []string{"inspect", "fork", "droid", "claude", "opencode", "kilo", "all"} {
				if !strings.Contains(script, command) {
					t.Fatalf("Generate(%q) missing command %q in script: %s", shell, command, script)
				}
			}
			if strings.Contains(script, "compress") {
				t.Fatalf("Generate(%q) should not include removed command compress: %s", shell, script)
			}
		})
	}
}

func TestGenerateRejectsUnknownShell(t *testing.T) {
	_, err := Generate("powershell")
	if err == nil || !strings.Contains(err.Error(), "unsupported shell") {
		t.Fatalf("Generate() error = %v, want unsupported shell", err)
	}
}
