# vibe
Cross-platform command line [Tidal music streaming service](https://tidal.com/) client

## Motivation

- Audio quality: 
    - web-based Tidal clients might suffer from suboptimal audio mixing / resampling in the browsers (see, e.g, [this chromecast review](https://www.audiosciencereview.com/forum/index.php?threads/review-and-measurements-of-chromecast-audio-digital-output.4544/));
    - official Windows desktop app might produce [choppy sound](https://www.canuckaudiomart.com/forum/viewtopic.php?t=54986) or [adjust volume erratically](https://www.reddit.com/r/TIdaL/comments/d7flaz/loudness_normalization_cant_be_turned_off_desktop/);
    - headless operation.
- Lightweight on resources: 
    - clients for big media projects, e.g., Kodi, require lots of dependencies;
    - web and `electron` based clients require significant amounts of RAM and CPU time;

## Note to the users
`vibe` is still under development and is subject to dramatic changes without prior notice.

## Prerequisites

- ffmpeg 
- sox 

### Regarding MQA
`vibe` can decode and render MQA streams using [mqa's](https://code.videolan.org/mansr/mqa) `mqadec` and `mqarender` binaries. It is user's responsibility to make sure those are accessible either via `PATH` (if using `--mqa-mode=host` flag) or via `localhost/mqa` container (using either `--mqa-mode=podman` or `--mqa-mode=docker` flags).

## Installation
```bash
go get -u github.com/godsic/vibe
```

## Authentication

`vibe` supports two authentication methods: 
- (default) OAuth2 authentication via a web browser. It requires user to copy-paste `code` from the address bar back to the input field in the terminal. This is still work-in-progress and might eventually be fully automated with the likes of [selenium](https://github.com/tebeka/selenium) or [chromedp](https://github.com/chromedp/chromedp).
- `--legacy-login` would trigger single factor authentification in case OAuth2 is absolutely not possible. It would likely be disabled by Tidal in the future. 


On successful login, `vibe` saves `~/.vibe/config/session.json` file and will reuse it for authomatic login. If one needs to use `vibe` on a headless machine, then it is advised to perform OAuth2 authentication elsewhere and then copy `session.json` file to the headless client.

**MQA encoded streams can only be accessed via OAuth2 authentication**.

## Audio quality

### The main feature of `vibe` is absolute loudness normalization to avoid hearing damage.
 By taking into account properies of the audio equipment, e.g., output voltages, gains, impedences and sensitivies, `vibe` can play music at requested absolute SPL (set via `--loudness=` flag, `75 dbC` by default). 
 
`vibe` resamples audio internally to the sample rate that offers the highest SINAD of the given DAC.  

 `vibe` maintains database of audio hardware and prompts users to specify their audio chain on first run. It is very limited at the moment and **user contributions to the hardware database are very welcome**.

 ## Usage

Play favorite tracks
```bash
vibe
```
Shuffle favorite tracks
```bash
vibe --shuffle
```
Override default playback loudness of `75 dbC`
```bash
vibe --loudness 85
```
Bypass PulseAudio on Linux
```bash
pasuspender -- vibe
```
For less common flags see 
```bash
vibe --help
```

