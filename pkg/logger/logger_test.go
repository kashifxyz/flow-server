package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestJSONInProduction(t *testing.T) {
	var buf bytes.Buffer
	log := NewWithWriter("production", "info", &buf)
	log.Info().Msg("server started")
	out := buf.String()
	if !strings.Contains(out, `"message":"server started"`) && !strings.Contains(out, `"msg":"server started"`) {
		t.Fatalf("unexpected log: %s", out)
	}
}
