# Go Bds Maneger

A module that brings together many functions to manage your Minecraft server efficiently, including several tools.

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

1. Minecraft Java run in any platform and architecture on java avaible.
1. The official Minecraft bedrock server ***CANNOT*** run on all platforms and CPU architectures, this tool tries as many ways to run this server as efficiently as possible.

## Tools

- [Aternos Mclog](https://mclo.gs/) ([Git repo](https://github.com/aternosorg/mclogs))
   - Server handler
   - Client
- [playit proxy client](https://playit.gg)
   - Add generic client to UDP and TCP packets
- binfmt: Attempt check file exec info and more
- Mount overlay: overlayfs or similar in platforms supported
   - Linux
   - MacOS (soon)
   - Windows (soon)