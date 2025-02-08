# Go-bds

A complete module to manage your Minecraft server easily!

this module will always try to support as many servers as possible, and any contribution would always be welcome

as one of the goals of this project is to simplify as much as possible for possible integration with other projects.

## Servers

For now we have a few supported servers, but enough to quickly deploy a server.

`Basically we are implementing basic features like, version list, installation and server startup software.`

### Bedrock

- [Mojang](minecraft.net/en-us/download/server/bedrock) server
- [Pocketmine-PMMP](https://github.com/pmmp/PocketMine-MP) server (`implemented more features, but it is already possible to start`)

### Java

- [Mojang](https://www.minecraft.net/en-us/download/server) server
- [Spigot](https://www.spigotmc.org/) server (`the server is compiled on the fly, so have git installed on your machine`)
- [Purpur](https://purpurmc.org/) server
- [Paper](https://papermc.io/software/paper) server
- [Folia](https://papermc.io/software/folia) server
- [Velocity](https://papermc.io/software/velocity) server

## Tools

- Mergefs/Overlayfs (***only some systems are supported***)
   - Golang fs.FS and os extends
- Mclog
   - Parse logs (experimental)
   - Client
