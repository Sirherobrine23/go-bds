package mojang_test

import (
	"fmt"
	"testing"

	"sirherobrine23.org/Minecraft-Server/go-bds/bedrock/mojang"
)

func TestPlayerParse(t *testing.T) {
	t.Run("Connect", func(t *testing.T) {
		Sirherobrine23Connect := "[2024-04-01 20:50:26:198 INFO] Player connected: Sirherobrine, xuid: 2535413418839840"
		err := mojang.ParseBedrockPlayerAction(Sirherobrine23Connect, func(username string, playerInfo mojang.PlayerConnections) {
			if username != "Sirherobrine" {
				t.Error(fmt.Errorf("cannot parse username"))
			} else if playerInfo.Action != mojang.PlayerActionConnect {
				t.Error(fmt.Errorf("cannot parse action"))
			} else if playerInfo.XUID != "2535413418839840" {
				t.Error(fmt.Errorf("cannot parse XUID"))
			} else if playerInfo.TimeConnection.Year() != 2024 || playerInfo.TimeConnection.Month() != 4 || playerInfo.TimeConnection.Day() != 1 || playerInfo.TimeConnection.Hour() != 20 || playerInfo.TimeConnection.Minute() != 50 {
				t.Error(fmt.Errorf("cannot parse Date"))
			}
		})
		if err != nil {
			t.Errorf("Player 1: %s", err)
			return
		}

		RandomSpace := "[2024-04-01 21:46:11:691 INFO] Player connected: nod dd, xuid:"
		err = mojang.ParseBedrockPlayerAction(RandomSpace, func(username string, playerInfo mojang.PlayerConnections) {
			if username != "nod dd" {
				t.Error(fmt.Errorf("cannot parse username"))
			} else if playerInfo.Action != mojang.PlayerActionConnect {
				t.Error(fmt.Errorf("cannot parse action"))
			} else if playerInfo.XUID != "" {
				t.Error(fmt.Errorf("cannot parse XUID"))
			}
		})
		if err != nil {
			t.Errorf("Player 2: %s", err)
			return
		}
	})

	t.Run("Disconnect", func(t *testing.T) {
		SirherobrineDisconnect := "[2022-08-30 20:56:55:231 INFO] Player disconnected: Sirherobrine, xuid: 2535413418839840"
		err := mojang.ParseBedrockPlayerAction(SirherobrineDisconnect, func(username string, playerInfo mojang.PlayerConnections) {
			if username != "Sirherobrine" {
				t.Error(fmt.Errorf("cannot parse username"))
			} else if playerInfo.Action != mojang.PlayerActionDisconnect {
				t.Error(fmt.Errorf("cannot parse action"))
			} else if playerInfo.XUID != "2535413418839840" {
				t.Error(fmt.Errorf("cannot parse XUID"))
			}
		})
		if err != nil {
			t.Errorf("Player 2: %s", err)
			return
		}

		RandomSpace := "[2024-04-01 21:46:33:199 INFO] Player disconnected: nod dd, xuid: , pfid: c31902da495f4549"
		err = mojang.ParseBedrockPlayerAction(RandomSpace, func(username string, playerInfo mojang.PlayerConnections) {
			if username != "nod dd" {
				t.Error(fmt.Errorf("cannot parse username"))
			} else if playerInfo.Action != mojang.PlayerActionDisconnect {
				t.Error(fmt.Errorf("cannot parse action"))
			} else if playerInfo.XUID != "" {
				t.Error(fmt.Errorf("cannot parse XUID"))
			}
		})
		if err != nil {
			t.Errorf("Player 2: %s", err)
			return
		}
	})
}
