package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"

	playitAPI "sirherobrine23.com.br/go-bds/go-bds/playitgg/api"
)

func printJSON(v any) {
	js, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(js))
}

func main() {
	setup := false
	config, err := os.OpenFile("./config.json", os.O_RDONLY, 0644)
	if err != nil {
		if config != nil {
			config.Close()
		}
		setup = true
	}
	defer config.Close()

	var clientAPI playitAPI.Api
	if setup {
		if err = clientAPI.CreateClaimCode(); err != nil {
			panic(err)
		}
		fmt.Printf("Open %q\n", clientAPI.ClaimUrl())
		if err = clientAPI.SetupSecret(playitAPI.AgentDefault); err != nil {
			panic(err)
		}
		js, _ := json.MarshalIndent(clientAPI, "", "  ")
		config, _ = os.Create("./config.json")
		config.Write(js)
	} else if err = json.NewDecoder(config).Decode(&clientAPI); err != nil {
		panic(err)
	}

	id, err := clientAPI.CreateTunnel(playitAPI.Tunnel{
		Name: "test",
		TunnelType: playitAPI.TunnelTypeMinecraftBedrock,
		PortType: 0,
		PortCount: 1,
		Enabled: true,
		Alloc: &playitAPI.TunnelCreateUseAllocation{
			Type: "region",
			Data: playitAPI.TunnelCreateUseAllocationDetails{
				UseRegion: &playitAPI.UseRegion{
					Region: playitAPI.RegionGlobal,
				},
			},
		},
		Origin: playitAPI.TunnelOriginCreate{
			Type: "default",
			Agent: playitAPI.AgentMerged{
				AssignedDefaultCreate: &playitAPI.AssignedDefaultCreate{
					Ip: net.IPv4(192, 16, 0, 3),
				},
			},
		},
	})
	if err != nil {
		panic(err)
	}
	printJSON(id)
}
