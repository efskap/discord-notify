# discord-notify
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fefskap%2Fdiscord-notify.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fefskap%2Fdiscord-notify?ref=badge_shield)


### Shows your Discord notifications, because Ripcord kinda doesn't.

Companion app for [Ripcord](https://cancel.fm/ripcord/) to supplant its notification system.

Targets Linux, but should work on Windows (only tested in Wine).

* Uses regular Discord notification settings.
* Displays notifications (w/ avatar) through your OS.
* But not when Ripcord is focused.
* Can play a sound.

[Ripcord](https://cancel.fm/ripcord/) is an amazing alternative Discord client, largely because native programs tend to be snappier. 

However, its notifications support is rather lacklustre, displaying them only for DMs, without a sound, and in plain text. Heck, until recently it didn't even tell you the name of the sender. Unfortunately it's closed source shareware, so I can't go in and tweak it to my liking.



---

```sh
$ go install github.com/efskap/discord-notify
$ discord-notify -h
Usage of discord-notify:
  -sound string
        MP3 to play on notifications (default "none")
  -t string
        Discord token
```

You can pass in the token with `-t`, or for convenience put it in a file called `discord.token` in `$XDG_CONFIG_HOME` or one of `XDG_CONFIG_DIRS`. 

e.g. On Linux it can go in `~/.config/discord.token` or `/etc/xdg/discord.token`, and on Windows it should be `%UserProfile%\Local Settings\Application Data\discord.token`

I've included the sound I use in this repo (needs to be downloaded separately from `go install`), and although it is not native to Discord, it appeals to my boomerish sensibilities. 


## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fefskap%2Fdiscord-notify.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fefskap%2Fdiscord-notify?ref=badge_large)