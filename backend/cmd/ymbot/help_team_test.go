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
		"team pull → seat enter → handoff get",
		"slug = 席位",
		"bot-team-template/playbook/看板席位迁绑.md",
		"team pull --uuid",
		"team join --code",
		"空手上场",
		"board_team_join",
		"seat enter --role",
		"handoff get",
		"migrate enter --role",
		"= seat enter（别名）",
		"seat leave --role",
		"seat expand --team",
		"patrol [--quiet-if-unchanged]",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("help missing %q\n%s", want, out)
		}
	}
	// migrate 不得抢主航线位：seat enter 须先于 migrate enter 出现在「命令」区
	seatIdx := strings.Index(out, "seat enter --role")
	migIdx := strings.Index(out, "migrate enter --role")
	if seatIdx < 0 || migIdx < 0 || seatIdx > migIdx {
		t.Fatalf("seat enter should appear before migrate enter alias\n%s", out)
	}
}
