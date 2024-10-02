# Go Bds Maneger

Maneger Minecraft server easy and more eficient

This is just a base package, without cli or even http api, for this you must create a go project and import this module, if not use one of our ready-made projects here on the server

## Server suports

- Bedrock
  1. [Mojang](https://minecraft.net/en-us/download/server/bedrock)
  1. [PocketMine-PMMP](https://github.com/pmmp/PocketMine-MP) (Partial)

- Java
  1. [Mojang](https://www.minecraft.net/en-us/download/server)
  1. [Spigot](https://www.spigotmc.org/) (Experimental)
  1. [Purpur server](https://purpurmc.org/)
  1. [Paper project](https://papermc.io/)
      - `paper`
      - `folia`
      - `velocity`

## System packages

We implement system calls to set up filesystems of the type or similar to Linux's OverlayFS on possible platforms and in a way that is mostly compatible with servers, and another tools to maneger server easyly.

1. Binfmt
1. OverlayFS - Filesystem implementation
   1. MergeFS - Golang Overlayfs implementations