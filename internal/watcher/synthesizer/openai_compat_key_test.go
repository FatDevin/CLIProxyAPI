package synthesizer

import (
	"strings"
	"testing"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
)

func TestConfigSynthesizer_OpenAICompatAPIKeyResolution(t *testing.T) {
	const envName = "CLIPROXY_TEST_OPENAI_COMPAT_API_KEY"

	tests := []struct {
		name      string
		apiKey    string
		apiKeyEnv string
		envValue  string
		wantKey   string
		wantErr   string
	}{
		{
			name:    "uses plaintext API key when API key env is empty",
			apiKey:  "plaintext-placeholder",
			wantKey: "plaintext-placeholder",
		},
		{
			name:      "uses environment API key over plaintext",
			apiKey:    "plaintext-placeholder",
			apiKeyEnv: envName,
			envValue:  " env-placeholder ",
			wantKey:   "env-placeholder",
		},
		{
			name:      "errors when API key environment is empty",
			apiKey:    "plaintext-placeholder",
			apiKeyEnv: envName,
			envValue:  "   ",
			wantErr:   "api-key-entries[0]",
		},
		{
			name:    "errors when both API key sources are empty",
			apiKey:  "",
			wantErr: "missing api-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(envName, tt.envValue)
			ctx := &SynthesisContext{
				Config: &config.Config{
					OpenAICompatibility: []config.OpenAICompatibility{{
						Name:    "test-provider",
						BaseURL: "https://provider.example.com/v1",
						APIKeyEntries: []config.OpenAICompatibilityAPIKey{{
							APIKey:    tt.apiKey,
							APIKeyEnv: tt.apiKeyEnv,
						}},
					}},
				},
				Now:         time.Now(),
				IDGenerator: NewStableIDGenerator(),
			}

			auths, err := NewConfigSynthesizer().Synthesize(ctx)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("expected an error")
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %q, want it to contain %q", err, tt.wantErr)
				}
				if len(auths) != 0 {
					t.Fatalf("expected no auths on error, got %d", len(auths))
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(auths) != 1 {
				t.Fatalf("expected 1 auth, got %d", len(auths))
			}
			if got := auths[0].Attributes["api_key"]; got != tt.wantKey {
				t.Fatalf("api_key = %q, want %q", got, tt.wantKey)
			}
		})
	}
}
