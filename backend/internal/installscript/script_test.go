package installscript

import (
	"strings"
	"testing"
)

func TestScriptNonEmptyAndMentionsYmbot(t *testing.T) {
	if strings.TrimSpace(Script) == "" {
		t.Fatal("install script empty")
	}
	for _, need := range []string{
		"ymbot",
		"go install",
		"github.com/yumeng-tech/ai-collab-board/backend/cmd/ymbot",
		"#!/usr/bin/env bash",
	} {
		if !strings.Contains(Script, need) {
			t.Fatalf("script missing %q", need)
		}
	}
}

func TestGoInstallCmd(t *testing.T) {
	if !strings.Contains(GoInstallCmd, "github.com/yumeng-tech/ai-collab-board/backend/cmd/ymbot@latest") {
		t.Fatal(GoInstallCmd)
	}
}
