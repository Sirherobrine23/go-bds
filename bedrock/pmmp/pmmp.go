// Pocketmine-MP PMMP is Minecraft bedrock server writed in PHP code with many plugins create by community,
// PMMP fork original pocketimine after many time under maintence code
//
// Pocketmine-MP code: https://github.com/pmmp/PocketMine-MP
//
// Original Pocketmine-MP Code: https://github.com/PocketMine/PocketMine-MP
package pmmp

import "errors"

var (
	ErrNoVersion error = errors.New("version not found")
	ErrPlatform  error = errors.New("current platform no supported")
)
