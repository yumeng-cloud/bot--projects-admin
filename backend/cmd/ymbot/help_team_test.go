package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestHelpListsTeamSeatCommands(t *testing.T) {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	printHelp()
	_ = w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	out := buf.String()
	for _, want := range []string{
		"team pull --uuid",
		"migrate enter --role",
		"seat enter --role",
		"seat leave --role",
		"seat expand --team",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("help missing %q\n%s", want, out)
		}
	}
}
