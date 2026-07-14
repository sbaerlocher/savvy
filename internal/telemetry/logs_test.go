package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

// newTestLogger wires the production handler chain (sanitizing + enrichment)
// over a buffer so the sanitization behaviour can be asserted on real emitted
// output — mirroring how InitLogger composes handlers.
func newTestLogger(buf *bytes.Buffer) *slog.Logger {
	base := slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	return slog.New(&enrichedLogHandler{
		handler:        NewSanitizingHandler(base),
		serviceName:    "test",
		serviceVersion: "0.0.0",
	})
}

func TestHandlerStripsCRLFFromStringAttrs(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	logger.Info("user action", "email", "victim\n{\"level\":\"ERROR\",\"msg\":\"forged\"}")

	out := buf.String()
	// The emitted record must be a single JSON line: no embedded newline from
	// the attribute value, so exactly one line of output.
	if lines := strings.Count(strings.TrimRight(out, "\n"), "\n"); lines != 0 {
		t.Fatalf("expected single log line, got %d embedded newlines:\n%s", lines, out)
	}
	if strings.Contains(out, "forged\"}\n") {
		t.Errorf("forged newline survived sanitization: %s", out)
	}

	var rec map[string]any
	if err := json.Unmarshal([]byte(out), &rec); err != nil {
		t.Fatalf("output is not a single valid JSON object: %v\n%s", err, out)
	}
	if got := rec["email"]; got != "victim{\"level\":\"ERROR\",\"msg\":\"forged\"}" {
		t.Errorf("unexpected sanitized email value: %v", got)
	}
}

func TestHandlerStripsCRLFFromMessageAndGroups(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	logger.LogAttrs(context.Background(), slog.LevelInfo,
		"msg\nwith newline",
		slog.Group("meta", slog.String("name", "a\r\nb")),
	)

	out := strings.TrimRight(buf.String(), "\n")
	if strings.Contains(out, "\n") {
		t.Fatalf("record contains embedded newline after sanitization:\n%s", out)
	}

	var rec map[string]any
	if err := json.Unmarshal([]byte(out), &rec); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out)
	}
	if rec["msg"] != "msgwith newline" {
		t.Errorf("message not sanitized: %v", rec["msg"])
	}
	meta, ok := rec["meta"].(map[string]any)
	if !ok {
		t.Fatalf("meta group missing or wrong type: %v", rec["meta"])
	}
	if meta["name"] != "ab" {
		t.Errorf("nested group value not sanitized: %v", meta["name"])
	}
}

func TestHandlerStripsCRLFFromWithAttrs(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	// Values bound via With() must be sanitized too, since they are copied into
	// every subsequent record.
	logger.With("tenant", "acme\nINJECTED").Info("scoped")

	out := strings.TrimRight(buf.String(), "\n")
	if strings.Contains(out, "\n") {
		t.Fatalf("bound attribute forged a newline:\n%s", out)
	}
	var rec map[string]any
	if err := json.Unmarshal([]byte(out), &rec); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out)
	}
	if rec["tenant"] != "acmeINJECTED" {
		t.Errorf("bound attribute not sanitized: %v", rec["tenant"])
	}
}
