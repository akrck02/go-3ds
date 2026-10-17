# Welcome to go-3ds 👋

**go-3ds runs Go programs on the Nintendo 3DS**: a 3D island on the top screen, buttons and sliders you touch on the bottom one, and real 3D with the slider.

<p align="center"><img alt="The 3DS's two screens: an island in the evening above, a panel with the time of day below" src="images/explorer-evening.png" width="320"></p>

It is the console's side of the family: the program's loop, the buttons and the touch screen, and a go-ui interface on the bottom screen. go-renderer draws on the console through its own graphics chip. Programs are built with a special version of TinyGo and opened from the Homebrew Launcher.

## Is it for me?

- ✅ You have a **3DS** (or an emulator) and want to put **your own program** on it.
- ✅ You want to see how one program runs on **very small hardware**.
- ❌ You want to play commercial games: this is homebrew, for your own programs.

## Where to start

Pick the line that sounds like you.

| If you… | Read | Time |
|---|---|---|
| are **new** and want a program on the console | 🏴‍☠️ [The quest](2-the-quest.md): an island in your pocket, in 5 levels | 45 min |
| want to know **what each package does** and **what you need** | [What is where](3-what-is-where.md) | 10 min |
| plan a **larger program** for the console | [Before writing a larger program](3-what-is-where.md#before-writing-a-larger-program) | 5 min |

## All the pages, in reading order

The menu on the left (or the files in `docs/`) follows this order.

**Start**

- [Overview](1-overview.md): this page
- [The quest](2-the-quest.md): five levels, from words on the screen to real 3D
- [What is where](3-what-is-where.md): the packages, the explorer, what you need, what was checked, and limits to know before a larger program

## 🚧 Not done yet

See [what is where](3-what-is-where.md#not-yet).

## Words you will meet

| Word | What it means |
|---|---|
| Go | A programming language. To try these projects you install it once, from go.dev, and type `go run …` in a terminal. |
| Terminal | The window where you type commands: *Terminal* on macOS and Linux, *PowerShell* on Windows. |
| Repository | A project's folder of files, kept on GitHub. `git clone <address>` copies it to your computer. |
| Homebrew | Programs for a game console that it was not sold with, opened from the Homebrew Launcher. |
| Quest | The tutorial of these docs: short levels, each ending with something that works and an achievement. |

## The family

go-3ds is one of a family of small projects that work together:

| Project | What it does |
|---|---|
| [go-renderer](https://github.com/akrck02/go-renderer) | draws 3D scenes and flat interfaces |
| [go-ui](https://github.com/akrck02/go-ui) | buttons, panels and whole applications |
| [go-ui-android](https://github.com/akrck02/go-ui-android) | Android apps that look like Android's own |
| [go-sound](https://github.com/akrck02/go-sound) | sound and music |
| [go-scene-builder](https://github.com/akrck02/go-scene-builder) | 3D scenes from a flat map |
| [go-simulator](https://github.com/akrck02/go-simulator) | simulations that change a scene |
| [go-simulator-terrain](https://github.com/akrck02/go-simulator-terrain) | ground shaped by rivers |
| [go-simulator-climate](https://github.com/akrck02/go-simulator-climate) | weather and biomes |
| [go-simulator-flora](https://github.com/akrck02/go-simulator-flora) | forests that grow |
| [go-simulator-urban](https://github.com/akrck02/go-simulator-urban) | towns that grow |
| [go-simulator-motion](https://github.com/akrck02/go-simulator-motion) | a scene that moves |
| **go-3ds** (this one) | Go on the Nintendo 3DS |

> ⚠️ All of them are **educational experiments**, written with an AI assistant and pending review by their author: good for learning and playing, not for production.
