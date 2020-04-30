# discord-notify

### Shows your Discord notifications, because Ripcord kinda doesn't.

Companion app for [Ripcord](https://cancel.fm/ripcord/) to supplant its notification system.

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
        Sound to play on notifications (default "none.mp3")
  -t string
        Discord token
```

I've included the sound I use in this repo (needs to be downloaded separately from `go install`), and although it is not native to Discord, it appeals to my boomerish sensibilities. 

