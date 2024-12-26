package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/uuid"
	"sirherobrine23.com.br/go-bds/go-bds/playitgg/playitapi"
)

var configFile = "playitgg.json"

func root() error {
	var api playitapi.Api
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		claim, err := api.Claim()
		if err != nil {
			return err
		}

		fmt.Printf("Open %s in your browser\n", claim.String())
		if err = claim.Do(playitapi.AgentTypeDefault); err != nil {
			return err
		}
		config, err := json.MarshalIndent(api, "", "  ")
		if err != nil {
			return err
		} else if err = os.WriteFile(configFile, config, 0644); err != nil {
			return err
		}
	} else {
		file, err := os.Open(configFile)
		if err != nil {
			return err
		}
		defer file.Close()
		if err = json.NewDecoder(file).Decode(&api); err != nil {
			return err
		}
	}

	agentData, err := api.RundataAgents()
	if err != nil {
		return err
	}

	tuns, err := api.Tunnels(uuid.Nil, agentData.ID)
	if err != nil {
		return err
	}

	route, err := api.Routing(agentData.ID)
	if err != nil {
		return err
	}

	d, _ := json.MarshalIndent(map[string]any{
		"agent": agentData,
		"tuns":  tuns,
		"route": route,
	}, "", "  ")
	fmt.Println(string(d))

	return nil
}

func main() {
	if err := root(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
