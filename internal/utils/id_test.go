package utils

import "testing"

func TestNewIDIsV7(t *testing.T) {
	got, err := NewID()
	if err != nil {
		t.Fatal(err)
	}
	if got.Version() != 7 {
		t.Fatalf("version = %d", got.Version())
	}
}
