package cmd

import (
	"os"
	"testing"

	"github.com/dopejs/opencc/internal/config"
)

func TestCompleteConfigNames(t *testing.T) {
	setTestHome(t)
	writeTestProvider(t, "alpha", &config.ProviderConfig{BaseURL: "https://a.com", AuthToken: "tok"})
	writeTestProvider(t, "beta", &config.ProviderConfig{BaseURL: "https://b.com", AuthToken: "tok"})

	names, directive := completeConfigNames(nil, nil, "")
	if directive != 4 { // cobra.ShellCompDirectiveNoFileComp = 4
		t.Errorf("directive = %d", directive)
	}
	if len(names) != 2 {
		t.Errorf("expected 2 names, got %d: %v", len(names), names)
	}
}

func TestRunCompletion(t *testing.T) {
	tests := []struct {
		shell   string
		wantErr bool
	}{
		{"zsh", false},
		{"bash", false},
		{"fish", false},
		{"powershell", false},
		{"invalid", false}, // prints error but doesn't return error
	}

	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			// Redirect stdout to avoid noise
			old := os.Stdout
			_, w, _ := os.Pipe()
			os.Stdout = w

			err := runCompletion(completionCmd, []string{tt.shell})

			w.Close()
			os.Stdout = old

			if (err != nil) != tt.wantErr {
				t.Errorf("runCompletion(%q) error = %v, wantErr %v", tt.shell, err, tt.wantErr)
			}
		})
	}
}
