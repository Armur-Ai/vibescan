# Records demo/menu.gif: the interactive menu.
#
# Prerequisite, from the repo root:
#   (cd cli && go build -o ../bin/vibescan .)
# Then:
#   vhs demo/menu.tape

Output demo/menu.gif

Set Shell bash
Set FontSize 16
Set Width 1000
Set Height 560
Set Padding 24
Set Theme "Catppuccin Mocha"
Set TypingSpeed 40ms

Hide
Type "export PATH=$PWD/bin:$PATH PS1='$ ' && clear"
Enter
Show

Sleep 400ms
Type "vibescan"
Sleep 400ms
Enter
Sleep 1500ms
Down
Sleep 450ms
Down
Sleep 450ms
Down
Sleep 450ms
Down
Sleep 450ms
Up
Sleep 450ms
Up
Sleep 1200ms
Type "q"
Sleep 600ms
