package auth

import "testing"

func TestRandomTokenHashRoundTrip(t *testing.T) {
	raw, err := RandomToken()
	if err != nil {
		t.Fatal(err)
	}
	if raw == "" {
		t.Fatal("empty token")
	}
	h := HashToken(raw)
	if !EqualHash(h, HashToken(raw)) {
		t.Fatal("hash mismatch")
	}
	if EqualHash(h, HashToken(raw+"x")) {
		t.Fatal("expected different hash")
	}
}

func TestNormalizeEmail(t *testing.T) {
	got, err := NormalizeEmail("  Foo.Bar@Example.COM ")
	if err != nil {
		t.Fatal(err)
	}
	if got != "foo.bar@example.com" {
		t.Fatalf("got %q", got)
	}
	if _, err := NormalizeEmail("not-an-email"); err == nil {
		t.Fatal("expected error")
	}
}

func TestValidatePassword(t *testing.T) {
	if err := ValidatePassword("short"); err == nil {
		t.Fatal("expected error")
	}
	if err := ValidatePassword("long enough password"); err != nil {
		t.Fatal(err)
	}
}
