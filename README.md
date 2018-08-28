# Discord Wipe - Wipe your messages off Discord with a selfbot

### Why?
Discord does not provide an easy way to delete all of your own messages from the service. There isn't away to mass delete messages prior to 2 weeks old, and that's just a bulk solution with a bot - you can't target your own messages with bulk deletion.

### The Program
This is a self bot that you run on your terminal. There are a few settings you can tweak on each run, which will be detailed later.

### The Advantages
You can run this in any server you are in, no bot account needed. You just need to find your authentication token (which is still very easy)

## The Dangers (Read This)
This program loads every message in all channels you specify, deleting your messages one by one. This can cause strain if you do it too quickly. Discord Wipe automatically waits between 1100 and 1900 milliseconds (1.1 and 1.9 seconds) between message deletions. You can adjust these all the way down to a minimum of 700 milliseconds.

**There's no guarantee you won't get flagged or banned for doing this.** Use this program at your own risk and discretion. Discord doesn't like selfbots, and I'm sure they won't like it if 100 people go through the history of millions of messages.

### Running/Configuration
There are various command-line flags you can use to configure each run. Items in **bold** are required (only serverID)

| Flag | Purpose | Example |
|------------------|-----------------------------------------------------------------------|-------------------------------|
| **`-server`** | The ID of the server to run on | -server="12345678" |
| `-wipechannels`* | Comma-separated IDs of channels to wipe | -wipechannels="123,456,789" |
| `-exemptchannels`** | Comma-separated IDs of channels to exempt | -exemptchannels="123,456,789" |
| `-waitmin` | Minimum wait time between deletion in milliseconds. Min 700 | -waitmin=1000 |
| `-waitmax` | Maximum wait time between deletion in milliseconds | -waitmax=2000 |
| `-retrieve`*** | Number of messages to retrieve per loop. Default 50. Min 25. Max 100. | -retrieve=75 |

\* If you set `wipechannels`, all channels not in this list are exempt, i.e. don't set `exemptchannels`

\** If you set `exemptchannels`, all channels not in this list are to be wiped, i.e. don't set `wipechannels`

\*** `retrieve` defaults to 50. Set this higher if you want the program to run faster, caps at 100. Don't touch this if you don't know what it means, probably.

## How to compile (not needed if you just download the release)
1. [Install Go](https://golang.org/doc/install)
2. Open the root directory and run `go get github.com/bwmarrin/discordgo`
3. Run `go build`

## How to run
1. Grab your Discord authentication token - use the handy tutorial at [this link](https://github.com/appu1232/Discord-Selfbot/wiki/Installation-&-Setup#grab-your-token-from-discord)
2. Put that token in a new file called `token.txt`
3. Run the server from the command line using the flags above - for example:

Windows: `wipe-windows.exe -server="1234" -exemptchannels="123,456" -waitmin=1000`

*nix: `./wipe -server="1234" -exemptchannels="123,456" -waitmin=1000`


## FAQ

[Go here](https://github.com/custom-salad/discord-wipe/wiki/Frequently-Asked-Questions)

## TODO
Make this work for personal DMs

## Current Checksums

Compiled on `go version go1.10 linux/amd64`

Note: Cross compilation seems to provide different binaries even on the same version of Go. Treat these as insurance that the Github releases are "correct". When in doubt, compile from source.

### Windows `wipe-windows.exe`
```
md5: 7f1d14ddbf1473538ce6433a0e87102e
sha256: 4a054b00e9d2dccc584e74380840a1f9e5345765f13d02ea5a864b2ee388a1a5
```

### macOS `wipe-darwin`
```
md5: e98c56792d366daf7352268151d03216
sha256: 72fc33b09864db85b1ea876d09ba97b0c09ca04df980a9245c132dafb34033aa
```

### Linux `wipe-linux`
```
md5: 4a87f8a4ba1d328240f4fe67b569f430
sha256: e8ad8243817b6e1b7de671414ef178c9e8a63dfbccb196243fab187c8ff53644
```
