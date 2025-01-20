package mclog_test

import (
	"encoding/json"
	"strings"
	"testing"

	. "sirherobrine23.com.br/go-bds/go-bds/mclog"
	_ "sirherobrine23.com.br/go-bds/go-bds/bedrock"
	_ "sirherobrine23.com.br/go-bds/go-bds/java"
)

func TestLog(t *testing.T) {
	var Insights []*Insights

	for _, logString := range Loger {
		insight, err := ParseLog(strings.NewReader(logString))
		if err != nil {
			t.Errorf("%s:\n%s", err, logString)
			return
		}
		Insights = append(Insights, insight)
	}

	d, _ := json.MarshalIndent(Insights, "", "  ")
	println(string(d))
}

var (
	Loger = []string{
		`NO LOG FILE! - setting up server logging...
[2022-11-16 19:34:11:606 INFO] Starting Server
[2022-11-16 19:34:11:606 INFO] Version 1.19.41.01
[2022-11-16 19:34:11:606 INFO] Session ID d914c8d5-2ad1-46fd-b213-e76ed4e6caca
[2022-11-16 19:34:11:606 INFO] Level Name: Sirherobrine23
[2022-11-16 19:34:11:609 INFO] Game mode: 0 Survival
[2022-11-16 19:34:11:609 INFO] Difficulty: 2 NORMAL
[2022-11-16 19:34:11:729 INFO] opening worlds/Sirherobrine23/db
[2022-11-16 19:34:12:336 INFO] IPv4 supported, port: 19132
[2022-11-16 19:34:12:337 INFO] IPv6 supported, port: 19133
[2022-11-16 19:34:13:002 INFO] Server started.
[2022-11-16 19:34:13:008 INFO] Please note that LAN discovery will not function for this server.
[2022-11-16 19:34:13:008 INFO] Server IP must be specified in Servers tab in game.
[2022-11-16 19:34:13:052 INFO] ================ TELEMETRY MESSAGE ===================
[2022-11-16 19:34:13:052 INFO] Server Telemetry is currently not enabled.
[2022-11-16 19:34:13:052 INFO] Enabling this telemetry helps us improve the game.
[2022-11-16 19:34:13:052 INFO]
[2022-11-16 19:34:13:052 INFO] To enable this feature, add the line 'emit-server-telemetry=true'
[2022-11-16 19:34:13:052 INFO] to the server.properties file in the handheld/src-server directory
[2022-11-16 19:34:13:052 INFO] ======================================================
[2022-11-16 19:55:44:503 INFO] Player connected: Sirherobrine, xuid: 2535413418839840
[2022-11-16 19:55:49:400 INFO] Player Spawned: Sirherobrine xuid: 2535413418839840
[2022-11-16 19:56:59:378 INFO] Player disconnected: Sirherobrine, xuid: 2535413418839840
[2022-11-16 20:01:48:937 INFO] Running AutoCompaction...
[2022-11-16 20:07:48:959 INFO] Running AutoCompaction...
[2022-11-19 19:23:22:600 INFO] Running AutoCompaction...
[2022-11-19 19:29:22:601 INFO] Running AutoCompaction...
[2022-11-19 19:35:22:602 INFO] Running AutoCompaction...
[2022-11-19 19:41:23:746 INFO] Running AutoCompaction...
[2022-11-19 19:47:23:746 INFO] Running AutoCompaction...
[2022-11-19 19:53:23:747 INFO] Running AutoCompaction...
[2022-11-19 19:59:23:748 INFO] Running AutoCompaction...
[2022-11-19 20:05:24:904 INFO] Running AutoCompaction...
[2022-11-19 20:07:15:776 INFO] Player connected: Sirherobrine, xuid: 2535413418839840
[2022-11-19 20:07:30:911 INFO] Player Spawned: Sirherobrine xuid: 2535413418839840`,
		`NO LOG FILE! - setting up server logging...
NO LOG FILE! - [2025-01-19 23:35:13 INFO] Starting Server
NO LOG FILE! - [2025-01-19 23:35:13 INFO] Version 1.6.1.0
NO LOG FILE! - [2025-01-19 23:35:13 INFO] Level Name: Bedrock level
NO LOG FILE! - [2025-01-19 23:35:13 ERROR] Error opening whitelist file: whitelist.json
NO LOG FILE! - [2025-01-19 23:35:13 ERROR] Error opening ops file: ops.json
NO LOG FILE! - [2025-01-19 23:35:13 INFO] Game mode: 0 Survival
NO LOG FILE! - [2025-01-19 23:35:13 INFO] Difficulty: 1 EASY
NO LOG FILE! - [2025-01-19 23:35:13 INFO] IPv4 supported, port: 19132
NO LOG FILE! - [2025-01-19 23:35:13 INFO] IPv6 supported, port: 19133
NO LOG FILE! - [2025-01-19 23:35:14 INFO] Listening on IPv6 port: 19133
NO LOG FILE! - [2025-01-19 23:35:14 INFO] Listening on IPv4 port: 19132
NO LOG FILE! - [2025-01-19 23:35:18 INFO] Player connected: 2535413418839840
NO LOG FILE! - [2025-01-19 23:38:54 INFO] Player disconnected: 2535413418839840
NO LOG FILE! - [2025-01-19 23:40:13 INFO] Player connected:
NO LOG FILE! - [2025-01-19 23:40:54 INFO] Player disconnected:
`,
		`Unpacking 1.21.4/server-1.21.4.jar (versions:1.21.4) to versions/1.21.4/server-1.21.4.jar
Unpacking com/fasterxml/jackson/core/jackson-annotations/2.13.4/jackson-annotations-2.13.4.jar (libraries:com.fasterxml.jackson.core:jackson-annotations:2.13.4) to libraries/com/fasterxml/jackson/core/jackson-annotations/2.13.4/jackson-annotations-2.13.4.jar
Unpacking com/fasterxml/jackson/core/jackson-core/2.13.4/jackson-core-2.13.4.jar (libraries:com.fasterxml.jackson.core:jackson-core:2.13.4) to libraries/com/fasterxml/jackson/core/jackson-core/2.13.4/jackson-core-2.13.4.jar
Unpacking com/fasterxml/jackson/core/jackson-databind/**.**.**.**/jackson-databind-2.13.4.2.jar (libraries:com.fasterxml.jackson.core:jackson-databind:**.**.**.**) to libraries/com/fasterxml/jackson/core/jackson-databind/**.**.**.**/jackson-databind-2.13.4.2.jar
Unpacking com/github/oshi/oshi-core/6.6.5/oshi-core-6.6.5.jar (libraries:com.github.oshi:oshi-core:6.6.5) to libraries/com/github/oshi/oshi-core/6.6.5/oshi-core-6.6.5.jar
Unpacking com/github/stephenc/jcip/jcip-annotations/1.0-1/jcip-annotations-1.0-1.jar (libraries:com.github.stephenc.jcip:jcip-annotations:1.0-1) to libraries/com/github/stephenc/jcip/jcip-annotations/1.0-1/jcip-annotations-1.0-1.jar
Unpacking com/google/code/gson/gson/2.11.0/gson-2.11.0.jar (libraries:com.google.code.gson:gson:2.11.0) to libraries/com/google/code/gson/gson/2.11.0/gson-2.11.0.jar
Unpacking com/google/guava/failureaccess/1.0.2/failureaccess-1.0.2.jar (libraries:com.google.guava:failureaccess:1.0.2) to libraries/com/google/guava/failureaccess/1.0.2/failureaccess-1.0.2.jar
Unpacking com/google/guava/guava/33.3.1-jre/guava-33.3.1-jre.jar (libraries:com.google.guava:guava:33.3.1-jre) to libraries/com/google/guava/guava/33.3.1-jre/guava-33.3.1-jre.jar
Unpacking com/microsoft/azure/msal4j/1.17.2/msal4j-1.17.2.jar (libraries:com.microsoft.azure:msal4j:1.17.2) to libraries/com/microsoft/azure/msal4j/1.17.2/msal4j-1.17.2.jar
Unpacking com/mojang/authlib/6.0.57/authlib-6.0.57.jar (libraries:com.mojang:authlib:6.0.57) to libraries/com/mojang/authlib/6.0.57/authlib-6.0.57.jar
Unpacking com/mojang/brigadier/1.3.10/brigadier-1.3.10.jar (libraries:com.mojang:brigadier:1.3.10) to libraries/com/mojang/brigadier/1.3.10/brigadier-1.3.10.jar
Unpacking com/mojang/datafixerupper/8.0.16/datafixerupper-8.0.16.jar (libraries:com.mojang:datafixerupper:8.0.16) to libraries/com/mojang/datafixerupper/8.0.16/datafixerupper-8.0.16.jar
Unpacking com/mojang/jtracy/1.0.29/jtracy-1.0.29.jar (libraries:com.mojang:jtracy:1.0.29) to libraries/com/mojang/jtracy/1.0.29/jtracy-1.0.29.jar
Unpacking com/mojang/logging/1.5.10/logging-1.5.10.jar (libraries:com.mojang:logging:1.5.10) to libraries/com/mojang/logging/1.5.10/logging-1.5.10.jar
Unpacking com/nimbusds/content-type/2.3/content-type-2.3.jar (libraries:com.nimbusds:content-type:2.3) to libraries/com/nimbusds/content-type/2.3/content-type-2.3.jar
Unpacking com/nimbusds/lang-tag/1.7/lang-tag-1.7.jar (libraries:com.nimbusds:lang-tag:1.7) to libraries/com/nimbusds/lang-tag/1.7/lang-tag-1.7.jar
Unpacking com/nimbusds/nimbus-jose-jwt/9.40/nimbus-jose-jwt-9.40.jar (libraries:com.nimbusds:nimbus-jose-jwt:9.40) to libraries/com/nimbusds/nimbus-jose-jwt/9.40/nimbus-jose-jwt-9.40.jar
Unpacking com/nimbusds/oauth2-oidc-sdk/11.18/oauth2-oidc-sdk-11.18.jar (libraries:com.nimbusds:oauth2-oidc-sdk:11.18) to libraries/com/nimbusds/oauth2-oidc-sdk/11.18/oauth2-oidc-sdk-11.18.jar
Unpacking commons-io/commons-io/2.17.0/commons-io-2.17.0.jar (libraries:commons-io:commons-io:2.17.0) to libraries/commons-io/commons-io/2.17.0/commons-io-2.17.0.jar
Unpacking io/netty/netty-buffer/4.1.115.Final/netty-buffer-4.1.115.Final.jar (libraries:io.netty:netty-buffer:4.1.115.Final) to libraries/io/netty/netty-buffer/4.1.115.Final/netty-buffer-4.1.115.Final.jar
Unpacking io/netty/netty-codec/4.1.115.Final/netty-codec-4.1.115.Final.jar (libraries:io.netty:netty-codec:4.1.115.Final) to libraries/io/netty/netty-codec/4.1.115.Final/netty-codec-4.1.115.Final.jar
Unpacking io/netty/netty-common/4.1.115.Final/netty-common-4.1.115.Final.jar (libraries:io.netty:netty-common:4.1.115.Final) to libraries/io/netty/netty-common/4.1.115.Final/netty-common-4.1.115.Final.jar
Unpacking io/netty/netty-handler/4.1.115.Final/netty-handler-4.1.115.Final.jar (libraries:io.netty:netty-handler:4.1.115.Final) to libraries/io/netty/netty-handler/4.1.115.Final/netty-handler-4.1.115.Final.jar
Unpacking io/netty/netty-resolver/4.1.115.Final/netty-resolver-4.1.115.Final.jar (libraries:io.netty:netty-resolver:4.1.115.Final) to libraries/io/netty/netty-resolver/4.1.115.Final/netty-resolver-4.1.115.Final.jar
Unpacking io/netty/netty-transport/4.1.115.Final/netty-transport-4.1.115.Final.jar (libraries:io.netty:netty-transport:4.1.115.Final) to libraries/io/netty/netty-transport/4.1.115.Final/netty-transport-4.1.115.Final.jar
Unpacking io/netty/netty-transport-classes-epoll/4.1.115.Final/netty-transport-classes-epoll-4.1.115.Final.jar (libraries:io.netty:netty-transport-classes-epoll:4.1.115.Final) to libraries/io/netty/netty-transport-classes-epoll/4.1.115.Final/netty-transport-classes-epoll-4.1.115.Final.jar
Unpacking io/netty/netty-transport-native-epoll/4.1.115.Final/netty-transport-native-epoll-4.1.115.Final-linux-x86_64.jar (libraries:io.netty:netty-transport-native-epoll:4.1.115.Final:linux-x86_64) to libraries/io/netty/netty-transport-native-epoll/4.1.115.Final/netty-transport-native-epoll-4.1.115.Final-linux-x86_64.jar
Unpacking io/netty/netty-transport-native-epoll/4.1.115.Final/netty-transport-native-epoll-4.1.115.Final-linux-aarch_64.jar (libraries:io.netty:netty-transport-native-epoll:4.1.115.Final:linux-aarch_64) to libraries/io/netty/netty-transport-native-epoll/4.1.115.Final/netty-transport-native-epoll-4.1.115.Final-linux-aarch_64.jar
Unpacking io/netty/netty-transport-native-unix-common/4.1.115.Final/netty-transport-native-unix-common-4.1.115.Final.jar (libraries:io.netty:netty-transport-native-unix-common:4.1.115.Final) to libraries/io/netty/netty-transport-native-unix-common/4.1.115.Final/netty-transport-native-unix-common-4.1.115.Final.jar
Unpacking it/unimi/dsi/fastutil/8.5.15/fastutil-8.5.15.jar (libraries:it.unimi.dsi:fastutil:8.5.15) to libraries/it/unimi/dsi/fastutil/8.5.15/fastutil-8.5.15.jar
Unpacking net/java/dev/jna/jna/5.15.0/jna-5.15.0.jar (libraries:net.java.dev.jna:jna:5.15.0) to libraries/net/java/dev/jna/jna/5.15.0/jna-5.15.0.jar
Unpacking net/java/dev/jna/jna-platform/5.15.0/jna-platform-5.15.0.jar (libraries:net.java.dev.jna:jna-platform:5.15.0) to libraries/net/java/dev/jna/jna-platform/5.15.0/jna-platform-5.15.0.jar
Unpacking net/minidev/accessors-smart/2.5.1/accessors-smart-2.5.1.jar (libraries:net.minidev:accessors-smart:2.5.1) to libraries/net/minidev/accessors-smart/2.5.1/accessors-smart-2.5.1.jar
Unpacking net/minidev/json-smart/2.5.1/json-smart-2.5.1.jar (libraries:net.minidev:json-smart:2.5.1) to libraries/net/minidev/json-smart/2.5.1/json-smart-2.5.1.jar
Unpacking net/sf/jopt-simple/jopt-simple/5.0.4/jopt-simple-5.0.4.jar (libraries:net.sf.jopt-simple:jopt-simple:5.0.4) to libraries/net/sf/jopt-simple/jopt-simple/5.0.4/jopt-simple-5.0.4.jar
Unpacking org/apache/commons/commons-lang3/3.17.0/commons-lang3-3.17.0.jar (libraries:org.apache.commons:commons-lang3:3.17.0) to libraries/org/apache/commons/commons-lang3/3.17.0/commons-lang3-3.17.0.jar
Unpacking org/apache/logging/log4j/log4j-api/2.24.1/log4j-api-2.24.1.jar (libraries:org.apache.logging.log4j:log4j-api:2.24.1) to libraries/org/apache/logging/log4j/log4j-api/2.24.1/log4j-api-2.24.1.jar
Unpacking org/apache/logging/log4j/log4j-core/2.24.1/log4j-core-2.24.1.jar (libraries:org.apache.logging.log4j:log4j-core:2.24.1) to libraries/org/apache/logging/log4j/log4j-core/2.24.1/log4j-core-2.24.1.jar
Unpacking org/apache/logging/log4j/log4j-slf4j2-impl/2.24.1/log4j-slf4j2-impl-2.24.1.jar (libraries:org.apache.logging.log4j:log4j-slf4j2-impl:2.24.1) to libraries/org/apache/logging/log4j/log4j-slf4j2-impl/2.24.1/log4j-slf4j2-impl-2.24.1.jar
Unpacking org/joml/joml/1.10.8/joml-1.10.8.jar (libraries:org.joml:joml:1.10.8) to libraries/org/joml/joml/1.10.8/joml-1.10.8.jar
Unpacking org/lz4/lz4-java/1.8.0/lz4-java-1.8.0.jar (libraries:org.lz4:lz4-java:1.8.0) to libraries/org/lz4/lz4-java/1.8.0/lz4-java-1.8.0.jar
Unpacking org/ow2/asm/asm/9.6/asm-9.6.jar (libraries:org.ow2.asm:asm:9.6) to libraries/org/ow2/asm/asm/9.6/asm-9.6.jar
Unpacking org/slf4j/slf4j-api/2.0.16/slf4j-api-2.0.16.jar (libraries:org.slf4j:slf4j-api:2.0.16) to libraries/org/slf4j/slf4j-api/2.0.16/slf4j-api-2.0.16.jar
Starting net.minecraft.server.Main
[18:02:07] [ServerMain/INFO]: Environment: Environment[sessionHost=https://sessionserver.mojang.com, servicesHost=https://api.minecraftservices.com, name=PROD]
[18:02:09] [ServerMain/INFO]: No existing world data, creating new world
[18:02:10] [ServerMain/INFO]: Loaded 1370 recipes
[18:02:10] [ServerMain/INFO]: Loaded 1481 advancements
[18:02:10] [Server thread/INFO]: Starting minecraft server version 1.21.4
[18:02:10] [Server thread/INFO]: Loading properties
[18:02:10] [Server thread/INFO]: Default game type: SURVIVAL
[18:02:10] [Server thread/INFO]: Generating keypair
[18:02:10] [Server thread/INFO]: Starting Minecraft server on *:25565
[18:02:11] [Server thread/INFO]: Using epoll channel type
[18:02:11] [Server thread/INFO]: Preparing level "world"
[18:02:21] [Server thread/INFO]: Preparing start region for dimension minecraft:overworld
[18:02:21] [Worker-Main-3/INFO]: Preparing spawn area: 2%
[18:02:21] [Worker-Main-1/INFO]: Preparing spawn area: 2%
[18:02:22] [Worker-Main-3/INFO]: Preparing spawn area: 2%
[18:02:22] [Worker-Main-3/INFO]: Preparing spawn area: 2%
[18:02:23] [Worker-Main-3/INFO]: Preparing spawn area: 12%
[18:02:23] [Worker-Main-2/INFO]: Preparing spawn area: 18%
[18:02:24] [Worker-Main-1/INFO]: Preparing spawn area: 18%
[18:02:24] [Worker-Main-1/INFO]: Preparing spawn area: 18%
[18:02:25] [Worker-Main-2/INFO]: Preparing spawn area: 28%
[18:02:26] [Worker-Main-2/INFO]: Preparing spawn area: 51%
[18:02:26] [Worker-Main-1/INFO]: Preparing spawn area: 51%
[18:02:26] [Worker-Main-3/INFO]: Preparing spawn area: 51%
[18:02:27] [Worker-Main-3/INFO]: Preparing spawn area: 51%
[18:02:27] [Server thread/INFO]: Time elapsed: 6164 ms
[18:02:27] [Server thread/INFO]: Done (16.414s)! For help, type "help"
[18:02:27] [Server thread/INFO]: Starting remote control listener
[18:02:27] [Server thread/INFO]: Thread RCON Listener started
[18:02:27] [Server thread/INFO]: RCON running on 0.0.0.0:25575
[18:03:27] [Server thread/INFO]: Server empty for 60 seconds, pausing
[18:04:23] [Server thread/INFO]: Stopping the server
[18:04:23] [Server thread/INFO]: Stopping server
[18:04:23] [Server thread/INFO]: Saving players
[18:04:23] [Server thread/INFO]: Saving worlds
[18:04:24] [Server thread/INFO]: Saving chunks for level 'ServerLevel[world]'/minecraft:overworld
[18:04:28] [Server thread/INFO]: Saving chunks for level 'ServerLevel[world]'/minecraft:the_end
[18:04:28] [Server thread/INFO]: Saving chunks for level 'ServerLevel[world]'/minecraft:the_nether
[18:04:28] [Server thread/INFO]: ThreadedAnvilChunkStorage (world): All chunks are saved
[18:04:28] [Server thread/INFO]: ThreadedAnvilChunkStorage (DIM1): All chunks are saved
[18:04:28] [Server thread/INFO]: ThreadedAnvilChunkStorage (DIM-1): All chunks are saved
[18:04:28] [Server thread/INFO]: ThreadedAnvilChunkStorage: All dimensions are saved
[18:04:28] [Server thread/INFO]: Thread RCON Listener stopped`,
		`[Log4jPatcher] [INFO] Transforming org/apache/logging/log4j/core/lookup/JndiLookup
[Log4jPatcher] [INFO] Transforming org/apache/logging/log4j/core/pattern/MessagePatternConverter
[Log4jPatcher] [WARN]  Unable to find noLookups:Z field in org/apache/logging/log4j/core/pattern/MessagePatternConverter
Jan 19, 2025 6:36:38 PM io.netty.util.internal.PlatformDependent <clinit>
INFO: Your platform does not provide complete low-level API for accessing direct buffers reliably. Unless explicitly requested, heap buffer will always be preferred to avoid potential system unstability.
[18:36:38] [Server thread/INFO]: Starting minecraft server version 1.7.10
[18:36:38] [Server thread/INFO]: Loading properties
[18:36:38] [Server thread/INFO]: Default game type: SURVIVAL
[18:36:38] [Server thread/INFO]: Generating keypair
[18:36:38] [Server thread/INFO]: Starting Minecraft server on *:25565
[18:36:38] [Server thread/WARN]: Failed to load user banlist:
java.io.FileNotFoundException: banned-players.json (No such file or directory)
	at java.base/java.io.FileInputStream.open0(Native Method) ~[?:?]
	at java.base/java.io.FileInputStream.open(Unknown Source) ~[?:?]
	at java.base/java.io.FileInputStream.<init>(Unknown Source) ~[?:?]
	at com.google.common.io.Files.newReader(Files.java:86) ~[minecraft_server.1.7.10.jar:?]
	at om.g(SourceFile:124) ~[minecraft_server.1.7.10.jar:?]
	at ls.y(SourceFile:99) [minecraft_server.1.7.10.jar:?]
	at ls.<init>(SourceFile:25) [minecraft_server.1.7.10.jar:?]
	at lt.e(SourceFile:166) [minecraft_server.1.7.10.jar:?]
	at net.minecraft.server.MinecraftServer.run(SourceFile:339) [minecraft_server.1.7.10.jar:?]
	at lj.run(SourceFile:628) [minecraft_server.1.7.10.jar:?]
[18:36:38] [Server thread/WARN]: Failed to load ip banlist:
java.io.FileNotFoundException: banned-ips.json (No such file or directory)
	at java.base/java.io.FileInputStream.open0(Native Method) ~[?:?]
	at java.base/java.io.FileInputStream.open(Unknown Source) ~[?:?]
	at java.base/java.io.FileInputStream.<init>(Unknown Source) ~[?:?]
	at com.google.common.io.Files.newReader(Files.java:86) ~[minecraft_server.1.7.10.jar:?]
	at om.g(SourceFile:124) ~[minecraft_server.1.7.10.jar:?]
	at ls.x(SourceFile:91) [minecraft_server.1.7.10.jar:?]
	at ls.<init>(SourceFile:27) [minecraft_server.1.7.10.jar:?]
	at lt.e(SourceFile:166) [minecraft_server.1.7.10.jar:?]
	at net.minecraft.server.MinecraftServer.run(SourceFile:339) [minecraft_server.1.7.10.jar:?]
	at lj.run(SourceFile:628) [minecraft_server.1.7.10.jar:?]
[18:36:38] [Server thread/WARN]: Failed to load operators list:
java.io.FileNotFoundException: ops.json (No such file or directory)
	at java.base/java.io.FileInputStream.open0(Native Method) ~[?:?]
	at java.base/java.io.FileInputStream.open(Unknown Source) ~[?:?]
	at java.base/java.io.FileInputStream.<init>(Unknown Source) ~[?:?]
	at com.google.common.io.Files.newReader(Files.java:86) ~[minecraft_server.1.7.10.jar:?]
	at om.g(SourceFile:124) ~[minecraft_server.1.7.10.jar:?]
	at ls.z(SourceFile:107) [minecraft_server.1.7.10.jar:?]
	at ls.<init>(SourceFile:29) [minecraft_server.1.7.10.jar:?]
	at lt.e(SourceFile:166) [minecraft_server.1.7.10.jar:?]
	at net.minecraft.server.MinecraftServer.run(SourceFile:339) [minecraft_server.1.7.10.jar:?]
	at lj.run(SourceFile:628) [minecraft_server.1.7.10.jar:?]
[18:36:38] [Server thread/WARN]: Failed to load white-list:
java.io.FileNotFoundException: whitelist.json (No such file or directory)
	at java.base/java.io.FileInputStream.open0(Native Method) ~[?:?]
	at java.base/java.io.FileInputStream.open(Unknown Source) ~[?:?]
	at java.base/java.io.FileInputStream.<init>(Unknown Source) ~[?:?]
	at com.google.common.io.Files.newReader(Files.java:86) ~[minecraft_server.1.7.10.jar:?]
	at om.g(SourceFile:124) ~[minecraft_server.1.7.10.jar:?]
	at ls.B(SourceFile:123) [minecraft_server.1.7.10.jar:?]
	at ls.<init>(SourceFile:30) [minecraft_server.1.7.10.jar:?]
	at lt.e(SourceFile:166) [minecraft_server.1.7.10.jar:?]
	at net.minecraft.server.MinecraftServer.run(SourceFile:339) [minecraft_server.1.7.10.jar:?]
	at lj.run(SourceFile:628) [minecraft_server.1.7.10.jar:?]
[18:36:38] [Server thread/INFO]: Preparing level "world"
[18:36:38] [Server thread/INFO]: Preparing start region for level 0
[18:36:39] [Server thread/INFO]: Preparing spawn area: 12%
[18:36:40] [Server thread/INFO]: Preparing spawn area: 28%
[18:36:41] [Server thread/INFO]: Preparing spawn area: 48%
[18:36:42] [Server thread/INFO]: Preparing spawn area: 69%
[18:36:43] [Server thread/INFO]: Preparing spawn area: 91%
[18:36:44] [Server thread/INFO]: Done (5.518s)! For help, type "help" or "?"
[18:36:44] [Server thread/INFO]: Starting remote control listener
[18:36:44] [RCON Listener #1/INFO]: RCON running on 0.0.0.0:25575`,
		`Unpacking 1.21.4/server-1.21.4.jar (versions:1.21.4) to versions/1.21.4/server-1.21.4.jar
Unpacking com/fasterxml/jackson/core/jackson-annotations/2.13.4/jackson-annotations-2.13.4.jar (libraries:com.fasterxml.jackson.core:jackson-annotations:2.13.4) to libraries/com/fasterxml/jackson/core/jackson-annotations/2.13.4/jackson-annotations-2.13.4.jar
Unpacking com/fasterxml/jackson/core/jackson-core/2.13.4/jackson-core-2.13.4.jar (libraries:com.fasterxml.jackson.core:jackson-core:2.13.4) to libraries/com/fasterxml/jackson/core/jackson-core/2.13.4/jackson-core-2.13.4.jar
Unpacking com/fasterxml/jackson/core/jackson-databind/2.13.4.2/jackson-databind-2.13.4.2.jar (libraries:com.fasterxml.jackson.core:jackson-databind:2.13.4.2) to libraries/com/fasterxml/jackson/core/jackson-databind/2.13.4.2/jackson-databind-2.13.4.2.jar
Unpacking com/github/oshi/oshi-core/6.6.5/oshi-core-6.6.5.jar (libraries:com.github.oshi:oshi-core:6.6.5) to libraries/com/github/oshi/oshi-core/6.6.5/oshi-core-6.6.5.jar
Unpacking com/github/stephenc/jcip/jcip-annotations/1.0-1/jcip-annotations-1.0-1.jar (libraries:com.github.stephenc.jcip:jcip-annotations:1.0-1) to libraries/com/github/stephenc/jcip/jcip-annotations/1.0-1/jcip-annotations-1.0-1.jar
Unpacking com/google/code/gson/gson/2.11.0/gson-2.11.0.jar (libraries:com.google.code.gson:gson:2.11.0) to libraries/com/google/code/gson/gson/2.11.0/gson-2.11.0.jar
Unpacking com/google/guava/failureaccess/1.0.2/failureaccess-1.0.2.jar (libraries:com.google.guava:failureaccess:1.0.2) to libraries/com/google/guava/failureaccess/1.0.2/failureaccess-1.0.2.jar
Unpacking com/google/guava/guava/33.3.1-jre/guava-33.3.1-jre.jar (libraries:com.google.guava:guava:33.3.1-jre) to libraries/com/google/guava/guava/33.3.1-jre/guava-33.3.1-jre.jar
Unpacking com/microsoft/azure/msal4j/1.17.2/msal4j-1.17.2.jar (libraries:com.microsoft.azure:msal4j:1.17.2) to libraries/com/microsoft/azure/msal4j/1.17.2/msal4j-1.17.2.jar
Unpacking com/mojang/authlib/6.0.57/authlib-6.0.57.jar (libraries:com.mojang:authlib:6.0.57) to libraries/com/mojang/authlib/6.0.57/authlib-6.0.57.jar
Unpacking com/mojang/brigadier/1.3.10/brigadier-1.3.10.jar (libraries:com.mojang:brigadier:1.3.10) to libraries/com/mojang/brigadier/1.3.10/brigadier-1.3.10.jar
Unpacking com/mojang/datafixerupper/8.0.16/datafixerupper-8.0.16.jar (libraries:com.mojang:datafixerupper:8.0.16) to libraries/com/mojang/datafixerupper/8.0.16/datafixerupper-8.0.16.jar
Unpacking com/mojang/jtracy/1.0.29/jtracy-1.0.29.jar (libraries:com.mojang:jtracy:1.0.29) to libraries/com/mojang/jtracy/1.0.29/jtracy-1.0.29.jar
Unpacking com/mojang/logging/1.5.10/logging-1.5.10.jar (libraries:com.mojang:logging:1.5.10) to libraries/com/mojang/logging/1.5.10/logging-1.5.10.jar
Unpacking com/nimbusds/content-type/2.3/content-type-2.3.jar (libraries:com.nimbusds:content-type:2.3) to libraries/com/nimbusds/content-type/2.3/content-type-2.3.jar
Unpacking com/nimbusds/lang-tag/1.7/lang-tag-1.7.jar (libraries:com.nimbusds:lang-tag:1.7) to libraries/com/nimbusds/lang-tag/1.7/lang-tag-1.7.jar
Unpacking com/nimbusds/nimbus-jose-jwt/9.40/nimbus-jose-jwt-9.40.jar (libraries:com.nimbusds:nimbus-jose-jwt:9.40) to libraries/com/nimbusds/nimbus-jose-jwt/9.40/nimbus-jose-jwt-9.40.jar
Unpacking com/nimbusds/oauth2-oidc-sdk/11.18/oauth2-oidc-sdk-11.18.jar (libraries:com.nimbusds:oauth2-oidc-sdk:11.18) to libraries/com/nimbusds/oauth2-oidc-sdk/11.18/oauth2-oidc-sdk-11.18.jar
Unpacking commons-io/commons-io/2.17.0/commons-io-2.17.0.jar (libraries:commons-io:commons-io:2.17.0) to libraries/commons-io/commons-io/2.17.0/commons-io-2.17.0.jar
Unpacking io/netty/netty-buffer/4.1.115.Final/netty-buffer-4.1.115.Final.jar (libraries:io.netty:netty-buffer:4.1.115.Final) to libraries/io/netty/netty-buffer/4.1.115.Final/netty-buffer-4.1.115.Final.jar
Unpacking io/netty/netty-codec/4.1.115.Final/netty-codec-4.1.115.Final.jar (libraries:io.netty:netty-codec:4.1.115.Final) to libraries/io/netty/netty-codec/4.1.115.Final/netty-codec-4.1.115.Final.jar
Unpacking io/netty/netty-common/4.1.115.Final/netty-common-4.1.115.Final.jar (libraries:io.netty:netty-common:4.1.115.Final) to libraries/io/netty/netty-common/4.1.115.Final/netty-common-4.1.115.Final.jar
Unpacking io/netty/netty-handler/4.1.115.Final/netty-handler-4.1.115.Final.jar (libraries:io.netty:netty-handler:4.1.115.Final) to libraries/io/netty/netty-handler/4.1.115.Final/netty-handler-4.1.115.Final.jar
Unpacking io/netty/netty-resolver/4.1.115.Final/netty-resolver-4.1.115.Final.jar (libraries:io.netty:netty-resolver:4.1.115.Final) to libraries/io/netty/netty-resolver/4.1.115.Final/netty-resolver-4.1.115.Final.jar
Unpacking io/netty/netty-transport/4.1.115.Final/netty-transport-4.1.115.Final.jar (libraries:io.netty:netty-transport:4.1.115.Final) to libraries/io/netty/netty-transport/4.1.115.Final/netty-transport-4.1.115.Final.jar
Unpacking io/netty/netty-transport-classes-epoll/4.1.115.Final/netty-transport-classes-epoll-4.1.115.Final.jar (libraries:io.netty:netty-transport-classes-epoll:4.1.115.Final) to libraries/io/netty/netty-transport-classes-epoll/4.1.115.Final/netty-transport-classes-epoll-4.1.115.Final.jar
Unpacking io/netty/netty-transport-native-epoll/4.1.115.Final/netty-transport-native-epoll-4.1.115.Final-linux-x86_64.jar (libraries:io.netty:netty-transport-native-epoll:4.1.115.Final:linux-x86_64) to libraries/io/netty/netty-transport-native-epoll/4.1.115.Final/netty-transport-native-epoll-4.1.115.Final-linux-x86_64.jar
Unpacking io/netty/netty-transport-native-epoll/4.1.115.Final/netty-transport-native-epoll-4.1.115.Final-linux-aarch_64.jar (libraries:io.netty:netty-transport-native-epoll:4.1.115.Final:linux-aarch_64) to libraries/io/netty/netty-transport-native-epoll/4.1.115.Final/netty-transport-native-epoll-4.1.115.Final-linux-aarch_64.jar
Unpacking io/netty/netty-transport-native-unix-common/4.1.115.Final/netty-transport-native-unix-common-4.1.115.Final.jar (libraries:io.netty:netty-transport-native-unix-common:4.1.115.Final) to libraries/io/netty/netty-transport-native-unix-common/4.1.115.Final/netty-transport-native-unix-common-4.1.115.Final.jar
Unpacking it/unimi/dsi/fastutil/8.5.15/fastutil-8.5.15.jar (libraries:it.unimi.dsi:fastutil:8.5.15) to libraries/it/unimi/dsi/fastutil/8.5.15/fastutil-8.5.15.jar
Unpacking net/java/dev/jna/jna/5.15.0/jna-5.15.0.jar (libraries:net.java.dev.jna:jna:5.15.0) to libraries/net/java/dev/jna/jna/5.15.0/jna-5.15.0.jar
Unpacking net/java/dev/jna/jna-platform/5.15.0/jna-platform-5.15.0.jar (libraries:net.java.dev.jna:jna-platform:5.15.0) to libraries/net/java/dev/jna/jna-platform/5.15.0/jna-platform-5.15.0.jar
Unpacking net/minidev/accessors-smart/2.5.1/accessors-smart-2.5.1.jar (libraries:net.minidev:accessors-smart:2.5.1) to libraries/net/minidev/accessors-smart/2.5.1/accessors-smart-2.5.1.jar
Unpacking net/minidev/json-smart/2.5.1/json-smart-2.5.1.jar (libraries:net.minidev:json-smart:2.5.1) to libraries/net/minidev/json-smart/2.5.1/json-smart-2.5.1.jar
Unpacking net/sf/jopt-simple/jopt-simple/5.0.4/jopt-simple-5.0.4.jar (libraries:net.sf.jopt-simple:jopt-simple:5.0.4) to libraries/net/sf/jopt-simple/jopt-simple/5.0.4/jopt-simple-5.0.4.jar
Unpacking org/apache/commons/commons-lang3/3.17.0/commons-lang3-3.17.0.jar (libraries:org.apache.commons:commons-lang3:3.17.0) to libraries/org/apache/commons/commons-lang3/3.17.0/commons-lang3-3.17.0.jar
Unpacking org/apache/logging/log4j/log4j-api/2.24.1/log4j-api-2.24.1.jar (libraries:org.apache.logging.log4j:log4j-api:2.24.1) to libraries/org/apache/logging/log4j/log4j-api/2.24.1/log4j-api-2.24.1.jar
Unpacking org/apache/logging/log4j/log4j-core/2.24.1/log4j-core-2.24.1.jar (libraries:org.apache.logging.log4j:log4j-core:2.24.1) to libraries/org/apache/logging/log4j/log4j-core/2.24.1/log4j-core-2.24.1.jar
Unpacking org/apache/logging/log4j/log4j-slf4j2-impl/2.24.1/log4j-slf4j2-impl-2.24.1.jar (libraries:org.apache.logging.log4j:log4j-slf4j2-impl:2.24.1) to libraries/org/apache/logging/log4j/log4j-slf4j2-impl/2.24.1/log4j-slf4j2-impl-2.24.1.jar
Unpacking org/joml/joml/1.10.8/joml-1.10.8.jar (libraries:org.joml:joml:1.10.8) to libraries/org/joml/joml/1.10.8/joml-1.10.8.jar
Unpacking org/lz4/lz4-java/1.8.0/lz4-java-1.8.0.jar (libraries:org.lz4:lz4-java:1.8.0) to libraries/org/lz4/lz4-java/1.8.0/lz4-java-1.8.0.jar
Unpacking org/ow2/asm/asm/9.6/asm-9.6.jar (libraries:org.ow2.asm:asm:9.6) to libraries/org/ow2/asm/asm/9.6/asm-9.6.jar
Unpacking org/slf4j/slf4j-api/2.0.16/slf4j-api-2.0.16.jar (libraries:org.slf4j:slf4j-api:2.0.16) to libraries/org/slf4j/slf4j-api/2.0.16/slf4j-api-2.0.16.jar
Starting net.minecraft.server.Main
[21:40:55] [ServerMain/INFO]: Environment: Environment[sessionHost=https://sessionserver.mojang.com, servicesHost=https://api.minecraftservices.com, name=PROD]
[21:40:59] [ServerMain/INFO]: No existing world data, creating new world
[21:41:01] [ServerMain/INFO]: Loaded 1370 recipes
[21:41:01] [ServerMain/INFO]: Loaded 1481 advancements
[21:41:01] [Server thread/INFO]: Starting minecraft server version 1.21.4
[21:41:01] [Server thread/INFO]: Loading properties
[21:41:01] [Server thread/INFO]: Default game type: SURVIVAL
[21:41:01] [Server thread/INFO]: Generating keypair
[21:41:01] [Server thread/INFO]: Starting Minecraft server on *:25565
[21:41:02] [Server thread/INFO]: Using epoll channel type
[21:41:02] [Server thread/INFO]: Preparing level "world"
[21:41:22] [Server thread/INFO]: Preparing start region for dimension minecraft:overworld
[21:41:23] [Worker-Main-2/INFO]: Preparing spawn area: 2%
[21:41:23] [Worker-Main-2/INFO]: Preparing spawn area: 2%
[21:41:24] [Worker-Main-2/INFO]: Preparing spawn area: 2%
[21:41:24] [Worker-Main-3/INFO]: Preparing spawn area: 2%
[21:41:24] [Worker-Main-1/INFO]: Preparing spawn area: 2%
[21:41:25] [Worker-Main-1/INFO]: Preparing spawn area: 2%
[21:41:25] [Worker-Main-1/INFO]: Preparing spawn area: 2%
[21:41:26] [Worker-Main-1/INFO]: Preparing spawn area: 2%
[21:41:26] [Worker-Main-3/INFO]: Preparing spawn area: 2%
[21:41:27] [Worker-Main-1/INFO]: Preparing spawn area: 2%
[21:41:27] [Worker-Main-1/INFO]: Preparing spawn area: 2%
[21:41:28] [Worker-Main-1/INFO]: Preparing spawn area: 2%
[21:41:28] [Worker-Main-1/INFO]: Preparing spawn area: 2%
[21:41:29] [Worker-Main-3/INFO]: Preparing spawn area: 8%
[21:41:30] [Worker-Main-2/INFO]: Preparing spawn area: 18%
[21:41:30] [Worker-Main-2/INFO]: Preparing spawn area: 18%
[21:41:30] [Worker-Main-1/INFO]: Preparing spawn area: 18%
[21:41:31] [Worker-Main-1/INFO]: Preparing spawn area: 18%
[21:41:31] [Worker-Main-3/INFO]: Preparing spawn area: 18%
[21:41:32] [Worker-Main-1/INFO]: Preparing spawn area: 18%
[21:41:32] [Worker-Main-1/INFO]: Preparing spawn area: 18%
[21:41:33] [Worker-Main-3/INFO]: Preparing spawn area: 18%
[21:41:33] [Worker-Main-3/INFO]: Preparing spawn area: 18%
[21:41:34] [Worker-Main-3/INFO]: Preparing spawn area: 18%
[21:41:34] [Worker-Main-3/INFO]: Preparing spawn area: 20%
[21:41:35] [Worker-Main-3/INFO]: Preparing spawn area: 51%
[21:41:35] [Worker-Main-1/INFO]: Preparing spawn area: 51%
[21:41:36] [Worker-Main-3/INFO]: Preparing spawn area: 51%
[21:41:36] [Worker-Main-2/INFO]: Preparing spawn area: 51%
[21:41:37] [Worker-Main-3/INFO]: Preparing spawn area: 51%
[21:41:37] [Worker-Main-1/INFO]: Preparing spawn area: 51%
[21:41:38] [Worker-Main-3/INFO]: Preparing spawn area: 51%
[21:41:38] [Worker-Main-3/INFO]: Preparing spawn area: 51%
[21:41:39] [Worker-Main-3/INFO]: Preparing spawn area: 51%
[21:41:39] [Worker-Main-2/INFO]: Preparing spawn area: 51%
[21:41:40] [Server thread/INFO]: Time elapsed: 17596 ms
[21:41:40] [Server thread/INFO]: Done (38.008s)! For help, type "help"
[21:41:40] [Server thread/INFO]: Starting remote control listener
[21:41:40] [Server thread/INFO]: Thread RCON Listener started
[21:41:40] [Server thread/INFO]: RCON running on 0.0.0.0:25575
[21:42:17] [User Authenticator #1/INFO]: UUID of player Sirherobrine23 is 0dc9df8f-9f5a-45d8-8848-9262a4357ae0
[21:42:20] [Server thread/INFO]: Sirherobrine23[/[0:0:0:0:0:0:0:1]:54662] logged in with entity id 41 at (37.5, 76.0, -57.5)
[21:42:20] [Server thread/INFO]: Sirherobrine23 joined the game
[21:43:42] [Server thread/WARN]: Sirherobrine23 moved too quickly! 9.663497831602484,3.176759275064242,0.8707508869438136
[21:43:42] [Server thread/WARN]: Can't keep up! Is the server overloaded? Running 4510ms or 90 ticks behind
[21:45:18] [Server thread/INFO]: Sirherobrine23 lost connection: Disconnected
[21:45:18] [Server thread/INFO]: Sirherobrine23 left the game`,
	}
)
