# hitmaker

A small Go / Bubble Tea TUI for writing songs for Suno. Write reusable lyric sections, arrange them by name, and copy the finished lyrics and style into Suno.

## Install

Requires Go 1.25 or newer and a terminal. Hitmaker is built with Bubble Tea v2, Bubbles v2, and Lip Gloss v2. Three columns appear at 110+ terminal columns; narrower terminals show the active editor at full width. Minimum size: 48 × 18.

Install directly with Go:

```sh
go install github.com/fabean/hitmaker@latest
hitmaker
```

Go puts the executable in `GOBIN`, or `$(go env GOPATH)/bin` by default (usually `~/go/bin`). Add that directory to your shell's `PATH` if `hitmaker` is not found. Run the same install command to update.

### Build from source

Clone with HTTPS, or use `git@github.com:fabean/hitmaker.git` if you prefer SSH:

```sh
git clone https://github.com/fabean/hitmaker.git
cd hitmaker
make build
./hitmaker
./hitmaker --demo
```

On Linux or macOS, install into `~/.local/bin`:

```sh
make install
export PATH="$HOME/.local/bin:$PATH"
hitmaker
```

Add the PATH setting to your shell configuration to keep it across sessions. Set `BINDIR` to choose another destination, for example `make install BINDIR="$HOME/bin"`. To update a source checkout, run `git pull --ff-only` and `make install` again.

Make is optional. You can run `go run .` from the checkout or build directly:

```sh
go build -trimpath -o hitmaker .
```

On Windows, use `go build -trimpath -o hitmaker.exe .` and run `.\hitmaker.exe`, or use the direct `go install` command above.

### Clipboard and browser helpers

On Linux, install `wl-clipboard` for Wayland, or `xclip` / `xsel` for X11. Opening Suno uses `xdg-open` (usually provided by `xdg-utils`). macOS uses `pbcopy` and `open`; Windows uses `clip.exe` and the default browser. Without a clipboard helper, Hitmaker tries the terminal's OSC 52 clipboard support. If copying fails, export with **x** and copy from the text files.

## Write a song

1. On first launch, enter a working title to start editing the chorus. Existing songs open in **NORMAL** mode.
2. Press **i** to edit a field and **Esc** to return to normal mode. **j/k** selects a field in the left pane; **h/l** switches between lyric fields, arrangement, and output. **Tab** also cycles through fields.
3. Press **t** to add a named section, or **r** to choose an arrangement. Use **l** to reach the arrangement and **i** to edit it. Changing a section updates every reference to it.
4. Press **s** for style presets. In pickers, **j/k** selects a row, **/** starts searching, and **Enter** chooses the result. Styles append to the style field, which grows with its text up to the terminal height.
5. Press **Y** to copy all Suno lyrics, then **b** to open Suno. In its custom/advanced editor, paste the lyrics. Back in hitmaker, focus the style field and press **y** to copy its text.

For example:

```text
[sad cinematic intro]
Verse 1
Chorus x2
[swelling strings]
Bridge
Chorus
[Outro]
```

Names are case-insensitive. `Chorus x4` repeats the chorus four times, including its `[Chorus]` heading. Repeats are limited to 1–16. Blank arrangement lines are ignored; output blocks have blank lines between them. Bracketed directions and other standalone text pass through unchanged. A missing standard section (such as `Verse 9`), an unknown repeated section, or an empty referenced section is flagged. Y copies the visible output even when sections are unfinished, and reports what was skipped. File export waits until those issues are fixed. To use a word such as “Chorus” as a direction rather than a reference, write `[Chorus]`.

Style presets supply genre, mood, instrumentation, and vocal descriptors. They do not generate lyrics. Hitmaker works locally without an account, API key, or network connection; opening Suno launches your browser. Song generation happens in Suno. See [Suno's Custom mode guide](https://help.suno.com/en/articles/2415873).

## Style library

**s** in normal mode opens 85 presets, including **80s** for bright, upbeat Wham!-inspired dance-pop, plus electronic, guitar, groove, acoustic, cinematic, experimental, and artist-inspired options. Search names, descriptions, or the full prompts. Multiple words narrow the results; try `80s`, `banjo`, `instrumental`, `warm bass`, `choir`, or `oddball`. PgUp/PgDn browses a page at a time.

Some starting points:

- **Liquid chiptune / SNES jungle:** game-console melodies over rolling or chopped breaks.
- **Cathedral garage:** two-step drums, sub bass, organ chords, and distant choir.
- **Space-western trip-hop:** dusty drums, baritone guitar, pedal steel, and sci-fi atmosphere.
- **Pirate disco:** concertina, fiddle, disco bass, and a crew-sung hook.
- **Elevator boss fight:** instrumental lounge jazz growing into a prog-rock theme.
- **Cosmic bluegrass:** banjo, mandolin, close vocal harmonies, and spacious synth drones.

The electronic direction takes inspiration from [Beginbot's public profile](https://suno.com/@beginbot), the liquid D&B / house / glitch tags on [Teej Made a Language](https://suno.com/song/c71ab8c0-c6ed-48e3-852d-6026b1c77afb), and the liquid D&B / footwork / glitch / warm-pad combination on [Up like RAM prices – instrumental](https://suno.com/song/e59e5f00-b1a4-4fdc-bf91-0354d9088f35). The preset prompts are newly written for hitmaker, including the original experimental combinations.

The artist presets are searchable by band name. Their copied prompts describe the instrumentation, vocal delivery, dynamics, and production:

| Preset | Direction |
| --- | --- |
| mewithoutYou | Raw spoken-word post-hardcore: jagged electric guitars, melodic bass ostinatos, restless drum-kit rhythms, and speech escalating into impassioned shouts. Focused on abrasive electric-band dynamics. |
| Chevelle | Low-tuned riffs, tense verses, and heavy melodic choruses. Guitar direction informed by [Pete Loeffler's rig rundown](https://www.premierguitar.com/gear/rig-rundown-chevelle). |
| Thrice | Atmospheric guitars, gritty soulful vocals, and dynamic post-hardcore, leaning toward their expansive alternative-rock side. See their [guitar and arrangement interview](https://www.musicradar.com/news/thrice-when-we-were-17-we-wanted-to-fit-as-many-notes-as-possible-in-were-minimalists-nowadays). |
| Vulfpeck | Sparse bass-led funk, dry drums, clean rhythmic guitar, warm keys, and soulful vocals. Inspired by their [rhythm-section approach](https://jambands.com/features/2014/01/12/vulfpeck-keep-it-beastly/). |

Presets marked **instrumental** include a no-vocals direction; choose one when you want an instrumental track. All prompts remain editable in the style field. These are creative starting points, not tested guarantees of a particular generated sound.

## Arrangements

**r** opens 23 arrangements: five general structures, twelve genre-focused recipes, and six artist-inspired templates. **/** searches by genre, band, or performance direction. Applying a template keeps the previous arrangement in the scratchpad and leaves your lyric sections and style intact.

| Presets | Pacing |
| --- | --- |
| Thrice · quiet to heavy / shifting sections | Restrained openings, atmospheric transitions, changing rhythmic emphasis, and heavy final releases or evolving codas. |
| Chevelle · tension and release / riff-driven slow burn | Muted verses and huge hooks, or a persistent riff with a delayed chorus and escalating outro. |
| mewithoutYou · spoken-word crescendo / angular post-hardcore | An escalating spoken narrative with a late shouted refrain, or jagged electric riffs, abrupt stops, and an explosive distorted coda. |
| Pop, hip-hop, country, folk | Different balances of story, hook, bridge, and repeated refrain. |
| Funk, gospel, metal, dance | Vamps, choir responses, breakdowns, builds, and instrumental drops. |
| Ambient, post-rock, jazz, cinematic | Instrumental forms built around motifs, development, solo sections, or long crescendos. |

The artist templates are original recipes informed by the musical directions described above, rather than transcriptions of specific songs. They use the existing Chorus, Verse 1, Verse 2, and Bridge sections. Instrumental templates use only performance directions. All templates remain editable, and Suno's interpretation can vary.

## Custom styles and arrangements

Create `styles.json` and/or `arrangements.json` in `~/.config/hitmaker/`. If `XDG_CONFIG_HOME` is an absolute path, Hitmaker uses `$XDG_CONFIG_HOME/hitmaker/` instead. `--config-dir /path/to/config` overrides this location independently of `--data-dir`.

From a source checkout, copy the examples without replacing existing files:

```sh
config_dir="${XDG_CONFIG_HOME:-$HOME/.config}/hitmaker"
mkdir -p "$config_dir"
cp -n examples/config/styles.json examples/config/arrangements.json "$config_dir/"
```

Both files contain a JSON array of objects with required `name` and `value` strings and an optional `detail` string. `detail` appears in the picker; `value` is the actual style prompt or arrangement. Names, details, and values are all searchable.

For example, `styles.json`:

```json
[
  {
    "name": "My space funk",
    "detail": "Custom · cosmic keys · tight groove",
    "value": "space funk, nimble electric bass, dry drums, warm clavinet, shimmering synth pads, playful soulful vocals"
  }
]
```

And `arrangements.json`:

```json
[
  {
    "name": "My big finish",
    "detail": "Quiet opening, double final chorus",
    "value": "[Quiet intro]\nVerse 1\nChorus\nVerse 2\nBridge\n[Full band]\nChorus x2\n[Outro]"
  }
]
```

Use `\n` for line breaks inside JSON strings. Arrangement values use the same section references, repeat syntax, and bracketed directions as the editor. Templates referencing extra sections require you to add those sections with **t**.

Files reload automatically whenever you open **s** or **r**. Close and reopen the picker after editing a file; no restart is needed. New names are appended after the built-ins. A matching name overrides an existing preset, ignoring case and surrounding whitespace; the last valid duplicate wins. Removing an override restores the built-in next time you open the picker. Applying a preset changes the current song; editing its config file does not change songs that already used it.

Missing files are fine. Invalid entries are skipped with a warning in the picker while valid entries still load. Malformed JSON falls back to built-ins for that picker. Hitmaker never rewrites these files. See the complete [example styles](examples/config/styles.json) and [example arrangements](examples/config/arrangements.json).

## Keys

Hitmaker uses Vim-style **NORMAL** and **INSERT** modes. These are whole-field controls; it does not implement Vim's full text-editing language. Normal commands never fire while you type lyrics in insert mode.

| Normal-mode key | Action |
| --- | --- |
| h / l | Previous / next pane |
| j / k | Next / previous field; scroll arrangement or output |
| Tab / Shift+Tab | Next / previous field across all panes |
| i / a / Enter | Edit the focused field |
| Esc | Return to normal mode |
| d | Clear the focused field, keeping its section name |
| y | Copy the focused title, style, section, arrangement, scratchpad, or generated output |
| u | Restore the last clear of this field while it remains empty |
| c | Clear the field and start editing |
| Y | Copy all rendered Suno lyrics from any pane |
| n / e | New song / saved-song picker |
| t | Add a lyric section |
| s / r | Style presets / arrangements |
| m | Open scratchpad |
| w | Save now |
| b | Open Suno |
| x | Export lyrics, style, and title text files |
| q | Save and quit |
| ? | Help |

In pickers, use **j/k** or **PgUp/PgDn** to browse, **/** to search, and **Enter** to choose. While searching, letters are literal text. **Esc** stops searching; another Esc closes the picker. **q** closes a picker while not searching. The scratchpad uses the same i/Esc, d/y/u/c controls; q returns to the workspace in normal mode.

In insert mode, use the usual text navigation, selection, and paste keys. Ctrl+A selects all in a multiline editor. Previous Control/function-key shortcuts remain available as aliases, but none are required for the main workflow. Ctrl+C remains an emergency save-and-quit shortcut.

Clipboard support uses `wl-copy`, `xclip`, or `xsel` on Linux; `pbcopy` on macOS; and `clip.exe` on Windows. If unavailable, hitmaker requests the terminal clipboard via OSC 52. Terminal support varies; **x** always provides a file export.

Clearing keeps lyric section names and arrangement references intact, and autosaves the empty text. **u** restores the most recently cleared text for that field during the current song session. It does not overwrite new text; restore backups reset when you open another song or quit. The output is generated, so clear its source sections or arrangement instead.

## Your files

Edits save after 800 ms of inactivity, on w, and before switching songs or quitting. Hitmaker refuses to switch or quit if a required save fails. On startup it opens the most recently saved song.

Data lives in `$XDG_DATA_HOME/hitmaker`, or `~/.local/share/hitmaker` by default:

```text
hitmaker/
  songs/<id>.json
  exports/<timestamp>_<song-title>_<id>.md
  exports/<id>/lyrics.txt
  exports/<id>/style.txt
  exports/<id>/title.txt
```

Every save, including autosave and **w**, also writes a new Markdown snapshot under `exports/`. The filename contains the local date/time (including fractions of a second and UTC offset), a safe version of the song title, and its ID. Earlier snapshots are kept. Each Markdown file contains the title, save time, style, and rendered lyrics; unfinished drafts include a list of skipped sections. Scratchpad notes are excluded. Existing songs get their first snapshot the next time they are saved.

JSON files include the title, style, lyric sections, arrangement, and scratchpad. Files use stable IDs, so renaming a title updates the same file. Saves use a temporary file and atomic rename. Exported text excludes the scratchpad. Corrupt song files are reported and left untouched.

```sh
./hitmaker --data-dir ./my-songs
./hitmaker --config-dir ./examples/config
./hitmaker --list
./hitmaker --export <song-id>
./hitmaker --version
```

`--demo` opens a fresh example. Editing or explicitly saving it adds it to your library. Use `--demo --data-dir /tmp/hitmaker-demo` for an isolated playground. Run one hitmaker instance per song library to avoid overwriting a draft in another instance.

## Development

```sh
make test
make build
make install # ~/.local/bin/hitmaker, or override BINDIR
make dist    # cross-compile binaries into dist/
```

`make test` runs the suite with Go's race detector and then `go vet`. The race detector needs a working C compiler. Tests use temporary song libraries and stub clipboard commands. Coverage includes arrangement expansion, save/load, Markdown snapshots, text exports, Vim controls, clipboard fallback, custom preset loading and overrides, automatic config reloads, failed saves, and terminal dimensions.

`make dist` builds standalone executables for Linux (amd64, arm64), macOS (amd64, arm64), and Windows (amd64), with CGO disabled. Files are named `dist/hitmaker-<os>-<arch>` (plus `.exe` on Windows). These are compilation targets; desktop clipboard and terminal behavior should be checked on the target OS. Build outputs are ignored by Git. `make clean` removes the local `hitmaker` executable.

The built-in style library lives in `styles.go`, arrangements in `structures.go`, rendering in `song.go`, and persistence in `storage.go` / `markdown.go`. Custom preset loading is in `config.go`.
