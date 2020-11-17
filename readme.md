# discord-notify

### Shows your Discord notifications, because Ripcord kinda doesn't.

![Go](https://github.com/efskap/discord-notify/workflows/Go/badge.svg)

Companion app for [Ripcord](https://cancel.fm/ripcord/) to supplant its notification system.

Targets Linux, but should work on Windows (only tested in Wine).

* Uses regular Discord notification settings.
* Displays notifications (w/ avatar) through your OS.
* But not when Ripcord is focused.
* Can play a sound.

[Ripcord](https://cancel.fm/ripcord/) is an amazing alternative Discord client, largely because native programs tend to be snappier. 

However, its notifications support is rather lacklustre, displaying them only for DMs, without a sound, and in plain text. Heck, until recently it didn't even tell you the name of the sender. Unfortunately it's closed source shareware, so I can't go in and tweak it to my liking.

## Install

### Pre-built binary

Binaries for Windows and GNU/Linux (x86-64) are automatically built and uploaded to: https://github.com/efskap/discord-notify/releases

### From source

```sh
git clone https://github.com/efskap/discord-notify
cd discord-notify
go install
```

(simply doing `go install https://github.com/efskap/discord-notify` as normal won't work because I'm using `replace` in `go.mod` for now... sorry)

## Usage

```sh
$ discord-notify -h
Usage of discord-notify:
  -sound string
        MP3 to play on notifications (default "none")
  -t string
        Discord token
$ discord-notify -t your.tok.en -sound /path/to/sound.mp3
```

There's no installer so just create a startup script that executes `discord-notify` with the token and sound you want. 

You can pass in the token with `-t`, or for convenience put it in a file called `discord.token` in `$XDG_CONFIG_HOME` or one of `XDG_CONFIG_DIRS`. 

e.g. On Linux it can go in `~/.config/discord.token` or `/etc/xdg/discord.token`, and on Windows it should be `%UserProfile%\Local Settings\Application Data\discord.token`

I've included the sound I use in the repo (and generated releases), and although it is not native to Discord, it appeals to my boomer mindset. 

**You can download the normal Discord notification sound from here**: https://discord.com/assets/dd920c06a01e5bb8b09678581e29d56f.mp3
