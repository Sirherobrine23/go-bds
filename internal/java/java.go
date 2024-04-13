package java

import (
	"errors"

	"sirherobrine23.org/Minecraft-Server/go-bds/internal/java/adoptium"
	"sirherobrine23.org/Minecraft-Server/go-bds/internal/java/globals"
	"sirherobrine23.org/Minecraft-Server/go-bds/internal/java/liberica"
	"sirherobrine23.org/Minecraft-Server/go-bds/internal/java/microsoft"
	"sirherobrine23.org/Minecraft-Server/go-bds/internal/java/zulu"
)

var (
	ErrInvalidDistro error = errors.New("distro not found")
)

func Release(distro string) ([]globals.Version, error) {
	if distro == "adoptium" {
		return adoptium.Releases()
	} else if distro == "liberica" {
		return liberica.Releases()
	} else if distro == "zulu" {
		return zulu.Releases()
	} else if distro == "microsoft" {
		return microsoft.Releases()
	}
	return nil, ErrInvalidDistro
}

func AllReleases() (map[string][]globals.Version, error) {
	maped := map[string][]globals.Version{}
	channelError := make(chan error)
	dones := 0

	go (func() {
		dones++
		vers, err := adoptium.Releases()
		maped["adoptium"] = vers
		channelError <- err
	})()

	go (func() {
		dones++
		vers, err := liberica.Releases()
		maped["liberica"] = vers
		channelError <- err
	})()

	go (func() {
		dones++
		vers, err := microsoft.Releases()
		maped["microsoft"] = vers
		channelError <- err
	})()

	go (func() {
		dones++
		vers, err := zulu.Releases()
		maped["zulu"] = vers
		channelError <- err
	})()

	for {
		err := <-channelError
		dones--
		if err != nil {
			return nil, err
		} else if dones == 0 {
			break
		}
	}

	return maped, nil
}
