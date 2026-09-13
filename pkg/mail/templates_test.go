package mail

import "testing"

func TestRenderVerifyEmail(t *testing.T) {
	body := render(Message{
		Template: TemplateVerifyEmail,
		Payload:  map[string]any{"link": "http://localhost/verify-email?token=abc", "expires": "24 hours"},
	})
	if body == "" || subject(Message{Template: TemplateVerifyEmail}) == "" {
		t.Fatal("empty template")
	}
}
