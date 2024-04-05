package bedrock_test

import (
	"testing"

	"sirherobrine23.org/minecraft-server/go-bds/bedrock"
)

func TestPlayerParse(t *testing.T) {
	t.Run("Connect", func(t *testing.T) {
		connect1 := "[2024-04-01 20:50:26:198 INFO] Player connected: Sirherobrine, xuid: 2535413418839840"
		connect2 := "[2024-04-01 21:46:11:691 INFO] Player connected: nod dd, xuid:"

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
			return
		} else if playerInfo.Action != "connected" {
			t.Errorf("Invalid player name: %q", playerInfo.Username)
			return
		} else if playerInfo.Xuid != "2535413418839840" {
			t.Errorf("Invalid player name: %q", playerInfo.Username)
			return
		}

		if playerInfo2.Username != "nod dd" {
			t.Errorf("Invalid player name: %q", playerInfo2.Username)
			return
		} else if playerInfo2.Action != "connected" {
			t.Errorf("Invalid player name: %q", playerInfo2.Username)
			return
		} else if playerInfo2.Xuid != "" {
			t.Errorf("Invalid player name: %q", playerInfo2.Username)
			return
		}
	})

	t.Run("Disconnect", func(t *testing.T) {
		connect1 := "[2022-08-30 20:56:55:231 INFO] Player disconnected: Sirherobrine, xuid: 2535413418839840"
		connect2 := "[2024-04-01 21:46:33:199 INFO] Player disconnected: nod dd, xuid: , pfid: c31902da495f4549"

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
			return
		} else if playerInfo.Action != "disconnected" {
			t.Errorf("Invalid player name: %q", playerInfo.Username)
			return
		} else if playerInfo.Xuid != "2535413418839840" {
			t.Errorf("Invalid player name: %q", playerInfo.Username)
			return
		}

		if playerInfo2.Username != "nod dd" {
			t.Errorf("Invalid player name: %q", playerInfo2.Username)
			return
		} else if playerInfo2.Action != "disconnected" {
			t.Errorf("Invalid player name: %q", playerInfo2.Username)
			return
		} else if playerInfo2.Xuid != "" {
			t.Errorf("Invalid player name: %q", playerInfo2.Username)
			return
		}
	})
}

func TestVersion(t *testing.T) {
	t.Run("Get", func(t *testing.T) {
		versions, err := bedrock.GetMojangVersions()
		if err != nil {
			t.Error(err)
			return
		}

		if len(versions) == 0 {
			t.Error("Invalid data return")
			return
		}
	})
}
