package auth

import "testing"

func TestHashAndVerify(t *testing.T) {
	h, err := HashPassword("Admin123!")
	if err != nil {
		t.Fatal(err)
	}
	ok, err := VerifyPassword(h, "Admin123!")
	if err != nil || !ok {
		t.Fatalf("verify failed: ok=%v err=%v", ok, err)
	}
	ok, err = VerifyPassword(h, "wrong")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected mismatch")
	}
}
