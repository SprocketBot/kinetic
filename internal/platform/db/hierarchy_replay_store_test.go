package db

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/kineticbot/kinetic-v3/internal/domain/hierarchy"
	"github.com/kineticbot/kinetic-v3/internal/platform/replayparser"
)

type replayParserStub struct {
	result replayparser.Result
	err    error
}

func (p replayParserStub) Parse(context.Context, []byte) (replayparser.Result, error) {
	return p.result, p.err
}

func parsedReplayFixture() replayparser.Result {
	team := int32(0)
	platform := "OnlinePlatform_Steam"
	accountID := "unlinked-steam-id"
	raw := json.RawMessage(`{"version":"868.32","game_type":"TAGame.Replay_Soccar_TA","duration_seconds":300,"winner":"Blue","teams":[{"index":0,"name":"Blue","score":3},{"index":1,"name":"Orange","score":1}],"players":[{"identity":{"name":"Unlinked Player","platform":"OnlinePlatform_Steam","online_id":"unlinked-steam-id","epic_id":null},"stats":{"goals":2,"assists":1,"shots":4,"saves":3,"score":567},"team":0}],"goals":[]}`)
	return replayparser.Result{
		RawJSON: raw,
		Summary: replayparser.MatchSummary{
			Version:         "868.32",
			GameType:        "TAGame.Replay_Soccar_TA",
			DurationSeconds: float64Ptr(300),
			Teams: []replayparser.Team{
				{Index: 0, Name: "Blue", Score: 3},
				{Index: 1, Name: "Orange", Score: 1},
			},
			Players: []replayparser.Player{{
				Identity: replayparser.PlayerIdentity{Name: "Unlinked Player", Platform: &platform, OnlineID: &accountID},
				Stats:    replayparser.PlayerStats{Goals: 2, Assists: 1, Shots: 4, Saves: 3, Score: 567},
				Team:     &team,
			}},
		},
	}
}

func float64Ptr(value float64) *float64 { return &value }

func TestDecodeReplayInputBase64(t *testing.T) {
	decoded, err := decodeReplayInput(hierarchy.IngestReplayEvidenceInput{ReplayBodyBase64: "cmVwbGF5LWJ5dGVz"})
	if err != nil {
		t.Fatalf("decode replay: %v", err)
	}
	if string(decoded) != "replay-bytes" {
		t.Fatalf("expected replay bytes, got %q", decoded)
	}
}

func TestReplayIdentityNormalizesParserPlatforms(t *testing.T) {
	steamPlatform := "OnlinePlatform_Steam"
	steamID := "76561198000000000"
	provider, accountID := replayIdentity(replayparser.PlayerIdentity{
		Name:     "Steam Player",
		Platform: &steamPlatform,
		OnlineID: &steamID,
	})
	if provider != "steam" || accountID != steamID {
		t.Fatalf("expected steam identity, got %q:%q", provider, accountID)
	}

	epicID := "epic-account-id"
	provider, accountID = replayIdentity(replayparser.PlayerIdentity{Name: "Epic Player", EpicID: &epicID})
	if provider != "epic" || accountID != epicID {
		t.Fatalf("expected epic identity, got %q:%q", provider, accountID)
	}

	dingoPlatform := "OnlinePlatform_Dingo"
	xboxID := "2535459589302009"
	provider, accountID = replayIdentity(replayparser.PlayerIdentity{
		Name:     "Xbox Player",
		Platform: &dingoPlatform,
		OnlineID: &xboxID,
	})
	if provider != "xbox" || accountID != xboxID {
		t.Fatalf("expected dingo to normalize to xbox, got %q:%q", provider, accountID)
	}

	epicPlatform := "OnlinePlatform_Epic"
	zeroID := "0"
	provider, accountID = replayIdentity(replayparser.PlayerIdentity{
		Name:     "Unresolved Epic Player",
		Platform: &epicPlatform,
		OnlineID: &zeroID,
	})
	if provider != "epic" || accountID != "" {
		t.Fatalf("expected zero Epic id to remain unresolved, got %q:%q", provider, accountID)
	}
}

func TestReplayTeamScoresUsesParserIndexes(t *testing.T) {
	scores := replayTeamScores([]replayparser.Team{
		{Index: 0, Name: "Blue", Score: 4},
		{Index: 1, Name: "Orange", Score: 2},
	})
	if scores[0] != 4 || scores[1] != 2 {
		t.Fatalf("expected 4-2, got %d-%d", scores[0], scores[1])
	}
}

func TestResolveReplayTeamIDsSupportsEitherColorAssignment(t *testing.T) {
	teamIDs, ok := resolveReplayTeamIDs(map[int32]map[int64]int{
		0: {20: 2},
		1: {10: 2},
	}, 10, 20)
	if !ok || *teamIDs[0] != 20 || *teamIDs[1] != 10 {
		t.Fatalf("expected blue=away and orange=home, got %#v ok=%t", teamIDs, ok)
	}

	_, ok = resolveReplayTeamIDs(map[int32]map[int64]int{
		0: {10: 1, 20: 1},
		1: {20: 2},
	}, 10, 20)
	if ok {
		t.Fatal("expected mixed parser team membership to require review")
	}
}
