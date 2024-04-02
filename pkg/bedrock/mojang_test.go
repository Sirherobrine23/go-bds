package bedrock_test

import (
	"testing"

	"sirherobrine23.org/minecraft-server/go-bds/pkg/bedrock"
)

func TestPlayerParse(t *testing.T) {
	t.Run("Connect", func(t *testing.T) {
		connect1 := "[2024-04-01 20:50:26:198 INFO] Player connected: Sirherobrine, xuid: 2535413418839840";
		connect2 := "[2024-04-01 21:46:11:691 INFO] Player connected: nod dd, xuid:";

		playerInfo, err := bedrock.ParseBedrockPlayerAction(connect1)
		if err != nil {
			t.Errorf("Player 1: %s", err)
			return
		}

		playerInfo2, err := bedrock.ParseBedrockPlayerAction(connect2)
		if err != nil {
			t.Errorf("Player 2: %s", err)
			return
		}

		if playerInfo.Username != "Sirherobrine" {
			t.Errorf("Invalid player name: %q", playerInfo.Username)
			return;
		} else if playerInfo.Action != "connected" {
			t.Errorf("Invalid player name: %q", playerInfo.Username)
			return;
		} else if playerInfo.Xuid != "2535413418839840" {
			t.Errorf("Invalid player name: %q", playerInfo.Username)
			return;
		}

		if playerInfo2.Username != "nod dd" {
			t.Errorf("Invalid player name: %q", playerInfo2.Username)
			return;
		} else if playerInfo2.Action != "connected" {
			t.Errorf("Invalid player name: %q", playerInfo2.Username)
			return;
		} else if playerInfo2.Xuid != "" {
			t.Errorf("Invalid player name: %q", playerInfo2.Username)
			return;
		}
	})
}