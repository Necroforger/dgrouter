# soundboard
Example music bot using dca-rs to stream and transcode music
- Responds to mentions
- Plays music from youtube

See the sounds directory for an example of how to include some media files.
If the file extension ends in .json, the bot will create a command that will use ytdl play the supplied youtube links. The json should be in the form of
```json
[
	["name", "youtube-url"],
	["name2", "youtube-url"]
]
```

## Dependencies
This bot depends on [ffmpeg](https://ffmpeg.org) to transcode the audio to opus.
It should be installed to your path.

[Here is a guide for installing on windows](https://github.com/adaptlearning/adapt_authoring/wiki/Installing-FFmpeg)

## Flags

| Flag | description |
|------|-------------|
| t    | bot token   |
| p    | bot prefix  |
