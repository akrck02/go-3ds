# go-3ds

## What is where

| Piece | What it does |
|---|---|
| `controls` | The buttons, the circle pad and the touch screen as go-renderer's `input.State`: the cross is the arrow keys, L and R are minus and plus, A is Enter, B is Escape and the system's back, and the touch screen is the pointer, so dragging works like dragging with a mouse. Turning a reading into input is plain Go, tested on any computer; reading the hardware (`read.go`) is built only for the 3DS. |
| `app` | The program loop: the New 3DS's fast clock, the controls into the input state every frame, START or the HOME menu to stop. |
| `cmd/explorer` | The sample island: the circle pad or the cross turn the camera round it, L and R move it away and closer, dragging on the bottom screen turns it, and the 3D slider shows it in 3D. |
| `tools/run-in-azahar.sh` | Runs a `.3dsx` in the Azahar emulator without a screen (under Xvfb) and prints what the program wrote to the debug log; it also warns about reads and writes of unmapped memory, which stop a program on the console but not in the emulator. |

## What you need

- A TinyGo with the `nintendo3ds` target (a fork of TinyGo, `feature/3ds`), built against a
  system LLVM.
- devkitPro's 3DS packages (devkitARM, libctru, citro3d, 3dstools) in `/opt/devkitpro`.

```bash
tinygo build -target=nintendo3ds -o explorer.3dsx ./cmd/explorer
```

Copy the `.3dsx` to `/3ds/` on the SD card and open it from the Homebrew Launcher.

## What was checked

On a New 3DS XL: go-renderer's sample island at 60 frames a second (GPU about 5 ms a frame) with
`cmd/island3ds`, whose drawing `cmd/explorer` shares. The explorer itself, its controls and 3D are
checked in the emulator (the cross turns the camera); the 3D needs a look on the console.

## Not yet

- The second core and the New 3DS's C-stick.
- Sound, files on the SD card (TinyGo's `os` has no file system there yet), the keyboard applet.
