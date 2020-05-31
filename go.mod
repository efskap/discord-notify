module github.com/efskap/discord-notify

go 1.14

require (
	github.com/BurntSushi/xgb v0.0.0-20200324125942-20f126ea2843
	github.com/bwmarrin/discordgo v0.20.3
	github.com/faiface/beep v1.0.2
	github.com/gen2brain/beeep v0.0.0-20200420150314-13046a26d502
	github.com/kyoh86/xdg v1.2.0
	golang.org/x/sys v0.0.0-20200413165638-669c56c373c4
)

replace github.com/bwmarrin/discordgo => github.com/efskap/discordgo v0.20.3
