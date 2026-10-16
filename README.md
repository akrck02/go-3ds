# go-3ds

> **Note:** this project was written with an AI assistant and is pending thorough review. See the [AI disclosure](#ai-disclosure).

Go programs on the **Nintendo 3DS**, as homebrew for the Homebrew Launcher. The drawing is
[go-renderer](https://github.com/akrck02/go-renderer)'s (`ctr`, through citro3d); this repository
holds the rest of the platform: the controls as go-renderer's `input.State`, the program loop, and
an explorer of the sample island.

```bash
tinygo build -target=nintendo3ds -o explorer.3dsx ./cmd/explorer
```

It needs a TinyGo with the `nintendo3ds` target and devkitPro's 3DS libraries. Everything is in
[docs/](docs/README.md).

## AI disclosure

I write most of my code by hand, so I want to be clear about this project: **it was written by an
AI assistant (Claude, by Anthropic) under my direction.** Every commit made this way carries a
`Co-Authored-By: Claude` line.

**Status: pending thorough review.** The code compiles and its tests pass, but I have not yet
reviewed it line by line. Treat it as a working draft, not as code I vouch for.
