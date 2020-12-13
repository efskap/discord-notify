# ![](assets/icon.ico) discord-notify

### Shows your Discord notifications, because Ripcord kinda doesn't.

![Go](https://github.com/efskap/discord-notify/workflows/Go/badge.svg)

Companion app for [Ripcord](https://cancel.fm/ripcord/) to supplant its notification system.

Targets Linux, but should work on Windows (only tested in Wine, albeit with some systray issues).

* Uses regular Discord notification settings.
* Displays notifications (w/ avatar) through your OS.
* But not when Ripcord is focused.
* Plays a sound (comes with Discord and Skype sounds)
* Shows an icon in the system tray

[Ripcord](https://cancel.fm/ripcord/) is an amazing alternative Discord client, largely because native programs tend to be snappier. 

However, its notifications support is rather lacklustre, displaying them only for DMs, without a sound, and in plain text. Heck, until recently it didn't even tell you the name of the sender. Unfortunately it's closed source shareware, so I can't go in and tweak it to my liking.

## Install

### Pre-built binary

Binaries for Windows and GNU/Linux (x86-64) are automatically built and uploaded to: https://github.com/efskap/discord-notify/releases

### From source

```sh
git clone https://github.com/efskap/discord-notify
cd discord-notify
go generate && go install
```

(simply doing `go install https://github.com/efskap/discord-notify` as normal won't work because I'm using `replace` in `go.mod` for now... sorry)

## Usage

```sh
$ discord-notify -h
Usage of discord-notify:
  -list-sounds
        List available built-in sounds.
  -sound string
        MP3 to play on notifications (default "disc")
  -systray
        Show an icon in the system tray (default true)
  -t string
        Discord token

$ discord-notify -sound skaip -t your.token.abc
$ discord-notify -sound none -t your.token.abc -systray=false # muted with no system tray
```

There's no installer so just create a startup script that executes `discord-notify` with the token and sound you want. 

You can pass in the token with `-t`, or for convenience put it in a file called `discord.token` in `$XDG_CONFIG_HOME` or one of `XDG_CONFIG_DIRS`. 

e.g. On Linux it can go in `~/.config/discord.token` or `/etc/xdg/discord.token`, and on Windows it should be `%UserProfile%\Local Settings\Application Data\discord.token`

Built-in notification sounds can be selected through the system tray as well, but the chosen one is not saved as there's no mutable config yet. Custom sounds have to be passed on the command line because I haven't integrated a filepicker yet.

