# Rebuild with: make demo
Output docs/demo.gif
Output docs/demo.mp4
Require bash

Set Shell bash
Set Width 1440
Set Height 960
Set FontSize 18
Set FontFamily "DejaVu Sans Mono"
Set Theme "Dracula"
Set Padding 16
Set WindowBar Colorful
Set Framerate 15
Set TypingSpeed 65ms
Set CursorBlink false

Hide
Type "bash docs/demo.sh"
Enter
Wait+Screen /Last light on the radio/
Sleep 1s
Show
Sleep 2s
Screenshot docs/screenshot.png

# Replace the current style with the 80s preset.
Type "j"
Type "d"
Type "s"
Sleep 1s
Type "/80s"
Sleep 1s
Enter
Sleep 2s

# Choose a hook-first arrangement.
Type "r"
Sleep 1s
Type "/hook first"
Sleep 1s
Enter
Sleep 2s

# Rewrite the chorus; the preview updates as the lyrics change.
Type "j"
Type "c"
Type "We are the last light on the radio"
Enter
Type "A little louder when the city goes"
Enter
Type "If all we have is one more song"
Enter
Type "Turn it up, we're coming home"
Escape
Sleep 2s
Type "w"
Sleep 2s
Type "l"
Sleep 2s
Type "l"
Sleep 2s

Hide
Type "q"
Sleep 1s
