package integration

import (
	"strings"
	"testing"
)

func TestIntegration_Completion(t *testing.T) {
	shells := []string{"bash", "zsh", "fish", "powershell"}

	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			output, err := runCli("completion", shell)
			if err != nil {
				t.Fatalf("completion %s failed: %v, output: %s", shell, err, output)
			}

			if strings.TrimSpace(output) == "" {
				t.Errorf("expected non-empty output for %s completion", shell)
			}

			// basic sanity check that it's generating completion for specforge
			if !strings.Contains(output, "specforge") && !strings.Contains(strings.ToLower(output), "specforge") {
				t.Errorf("output doesn't seem to contain specforge references: %s", output[:100])
			}
		})
	}
}
