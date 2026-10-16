# go-3ds

## What is where

| Piece | What it does |
|---|---|
| `controls` | The buttons, the circle pad and the touch screen as go-renderer's `input.State`: the cross is the arrow keys, L and R are minus and plus, A is Enter, B is Escape and the system's back, and the touch screen is the pointer, so dragging works like dragging with a mouse. Turning a reading into input is plain Go, tested on any computer; reading the hardware (`read.go`) is built only for the 3DS. |
| `bottom` | A go-ui interface on the bottom screen: go-ui lays it out, go-renderer's `ctr.FlatRenderer` paints it, the touch screen is its pointer. `bottom.Theme` is go-ui's dark theme made for 320 × 240 pixels a finger touches, in the sea's colours. |
| `app` | The program loop: the New 3DS's fast clock, the controls into the input state every frame, START or the HOME menu to stop. |
| `cmd/explorer` | The sample island on the top screen, in 3D with the slider, and a go-ui panel on the bottom one: places to fly to, the time of day and the fog, and how fast frames go. The circle pad or the cross turn the camera, L and R move it away and closer. |
| `tools/run-in-azahar.sh` | Runs a `.3dsx` in the Azahar emulator without a screen (under Xvfb) and prints what the program wrote to the debug log; it also warns about reads and writes of unmapped memory, which stop a program on the console but not in the emulator. |

## The explorer

<p align="center">
  <img alt="The island from afar; below, the View tab: a map of the island coloured by height, with dots on the village and the peak, beside buttons to fly there" src="images/explorer-map.png" width="300">
  <img alt="The same view at nine in the evening, dark, with the lamp lit; below, the Light tab with the time of day at 21:00, a moon icon and the fog on" src="images/explorer-evening.png" width="300">
  <img alt="The Figures tab: chips with 56 frames a second, 0.4 ms of GPU, 31 732 triangles and 0 % 3D" src="images/explorer-figures.png" width="300">
</p>

The map is a picture the explorer draws when it starts, from the island's heights
(`world.GroundHeight`): coloured by height like a chart, shaded as if lit from the north-west,
with a dot on the village and one on the peak. Flying to a place eases the camera there in under
a second. The time of day changes the light
when the finger leaves the screen: a new light works the meshes' colours out again on the
processor, which would make every step of a drag stutter.

Both screens are drawn in one frame: the island with `ctr.SceneRenderer.PaintStereo`, the panel
with `bottom.Interface.Paint`, between `ctr.BeginFrame` and `ctr.EndFrame`.

## What you need

- A TinyGo with the `nintendo3ds` target (a fork of TinyGo, `feature/3ds`), built against a
  system LLVM.
- devkitPro's 3DS packages (devkitARM, libctru, citro3d, 3dstools) in `/opt/devkitpro`.

```bash
tinygo build -target=nintendo3ds -o explorer.3dsx ./cmd/explorer
```

Copy the `.3dsx` to `/3ds/` on the SD card and open it from the Homebrew Launcher.

## What was checked

On a New 3DS XL: the explorer at 60 frames a second, steady while the camera moves fast (GPU about
6 ms a frame, processor under 1 ms), in 3D with the slider. Its go-ui panel is checked in the
Azahar emulator: tabs, buttons, the slider and the switch answer clicks on the bottom screen.

## Not yet

- The second core and the New 3DS's C-stick.
- Sound, files on the SD card (TinyGo's `os` has no file system there yet), the keyboard applet.
