package replayparser

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCommandParsesReplaySummary(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture is Unix-specific")
	}

	dir := t.TempDir()
	parserPath := filepath.Join(dir, "rl-replay")
	script := `#!/bin/sh
test -s "$1"
cat <<'JSON'
{"date":"2026-08-25","version":"868.32","map":"Stadium_P","game_type":"TAGame.Replay_Soccar_TA","duration_seconds":310.5,"winner":"Blue","teams":[{"index":0,"name":"Blue","score":3},{"index":1,"name":"Orange","score":1}],"players":[{"identity":{"name":"Alice","platform":"OnlinePlatform_Steam","online_id":"123","epic_id":null},"stats":{"goals":2,"assists":1,"shots":4,"saves":3,"score":567,"shots_on_target":null,"goals_against":null,"shots_against":null,"shooting_percentage":null},"team":0}],"goals":[]}
JSON
`
	if err := os.WriteFile(parserPath, []byte(script), 0o700); err != nil {
		t.Fatalf("write parser fixture: %v", err)
	}

	result, err := NewCommand(parserPath).Parse(context.Background(), []byte("replay-bytes"))
	if err != nil {
		t.Fatalf("parse replay: %v", err)
	}
	if result.Summary.Teams[0].Score != 3 {
		t.Fatalf("expected home score 3, got %d", result.Summary.Teams[0].Score)
	}
	if got := *result.Summary.Players[0].Identity.OnlineID; got != "123" {
		t.Fatalf("expected online id 123, got %q", got)
	}
	if !strings.Contains(string(result.RawJSON), `"winner":"Blue"`) {
		t.Fatalf("expected raw parser output, got %s", result.RawJSON)
	}
}

func TestCommandReportsParserFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture is Unix-specific")
	}

	parserPath := filepath.Join(t.TempDir(), "rl-replay")
	if err := os.WriteFile(parserPath, []byte("#!/bin/sh\necho malformed replay >&2\nexit 2\n"), 0o700); err != nil {
		t.Fatalf("write parser fixture: %v", err)
	}

	_, err := NewCommand(parserPath).Parse(context.Background(), []byte("bad-replay"))
	if err == nil || !strings.Contains(err.Error(), "malformed replay") {
		t.Fatalf("expected parser stderr, got %v", err)
	}
}

func TestCommandParsesConfiguredRealReplay(t *testing.T) {
	parserPath := os.Getenv("TEST_REPLAY_PARSER_PATH")
	fixturePath := os.Getenv("TEST_REPLAY_FIXTURE_PATH")
	if parserPath == "" || fixturePath == "" {
		t.Skip("TEST_REPLAY_PARSER_PATH and TEST_REPLAY_FIXTURE_PATH are not set")
	}

	replay, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read real replay fixture: %v", err)
	}
	result, err := NewCommand(parserPath).Parse(context.Background(), replay)
	if err != nil {
		t.Fatalf("parse real replay fixture: %v", err)
	}
	if len(result.Summary.Teams) < 2 || len(result.Summary.Players) == 0 {
		t.Fatalf("expected real teams and players, got %#v", result.Summary)
	}
}
