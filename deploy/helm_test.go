package deploy_test

import (
	"bytes"
	"io"
	"os/exec"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDatabaseRendering(t *testing.T) {
	helm, err := exec.LookPath("helm")
	if err != nil {
		t.Skip("helm is required for chart rendering tests")
	}
	for _, tc := range []struct{ mode, urlSecret, credential, secret, key string }{
		{"URL", "url-secret", "DATABASE_URL", "url-secret", "connection-uri"},
		{"parameters", "", "DATABASE_PASSWORD", "password-secret", "password"},
	} {
		t.Run(tc.mode, func(t *testing.T) {
			out, err := exec.CommandContext(t.Context(), helm, "template", "savvy", "helm", "-f", "helm/values-test.yaml",
				"--set", "database.existingSecret="+tc.urlSecret+",database.databaseUrlKey=connection-uri",
				"--set", "database.host=postgres,database.port=6543,database.name=savvy,database.user=user",
				"--set", "database.passwordSecret.existingSecret=password-secret,database.passwordSecret.passwordKey=password").CombinedOutput()
			if err != nil {
				t.Fatalf("helm template failed: %s", out)
			}
			got := map[string]string{}
			credentialFound := false
			decoder := yaml.NewDecoder(bytes.NewReader(out))
			for {
				var doc map[string]any
				if err := decoder.Decode(&doc); err == io.EOF {
					break
				} else if err != nil {
					t.Fatal(err)
				}
				for _, field := range []string{"data", "stringData"} {
					data, _ := doc[field].(map[string]any)
					if data["DATABASE_PASSWORD"] != nil || data["DATABASE_URL"] != nil {
						t.Fatal("database credentials must not be stored in generated ConfigMaps or Secrets")
					}
				}
				if doc["kind"] != "Deployment" {
					continue
				}
				template := doc["spec"].(map[string]any)["template"].(map[string]any)
				containers := template["spec"].(map[string]any)["containers"].([]any)
				for _, item := range containers[0].(map[string]any)["env"].([]any) {
					env := item.(map[string]any)
					name := env["name"].(string)
					if name == tc.credential {
						ref := env["valueFrom"].(map[string]any)["secretKeyRef"].(map[string]any)
						if _, literal := env["value"]; literal || ref["name"] != tc.secret || ref["key"] != tc.key {
							t.Fatal("database credential must use the configured secretKeyRef, never a literal value")
						}
						credentialFound = true
					} else if strings.HasPrefix(name, "DATABASE_") {
						got[name] = env["value"].(string)
					}
				}
			}
			want := map[string]string{}
			if tc.mode == "parameters" {
				want = map[string]string{"DATABASE_HOST": "postgres", "DATABASE_PORT": "6543", "DATABASE_NAME": "savvy", "DATABASE_USER": "user", "DATABASE_SSLMODE": "verify-full"}
			}
			if !credentialFound || len(got) != len(want) {
				t.Fatal("incorrect database environment variables")
			}
			for name, value := range want {
				if got[name] != value {
					t.Fatalf("incorrect value for %s", name)
				}
			}
		})
	}
}
