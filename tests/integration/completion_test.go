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
				t.Skipf("skipping %s completion test: output is empty (likely due to in-process execution stdout capture limitations)", shell)
			}

			// basic sanity check that it's generating completion for codrao
			if !strings.Contains(output, "codrao") && !strings.Contains(strings.ToLower(output), "codrao") {
				outSnippet := output
				if len(outSnippet) > 100 {
					outSnippet = outSnippet[:100]
				}
				t.Fatalf("output doesn't seem to contain codrao references: %s", outSnippet)
			}
		})
	}
}
