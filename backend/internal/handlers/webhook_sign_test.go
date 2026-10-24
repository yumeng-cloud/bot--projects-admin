package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestSignPayload(t *testing.T) {
	body := []byte(`{"event":"issue.created"}`)
	sig := SignPayload("secret", body)
	mac := hmac.New(sha256.New, []byte("secret"))
	mac.Write(body)
	want := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if sig != want {
		t.Fatalf("got %s want %s", sig, want)
	}
}
