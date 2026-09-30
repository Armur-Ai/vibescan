# Records demo/scan.gif: a real scan of the example app.
#
# Prerequisites, from the repo root:
#   docker compose up -d --build
#   (cd cli && go build -o ../bin/vibescan .)
# Then:
#   vhs demo/scan.tape

Output demo/scan.gif

Set Shell bash
Set FontSize 13
Set Width 1500
Set Height 800
Set Padding 20
Set Theme "Catppuccin Mocha"
Set TypingSpeed 35ms

Hide
Type "export PATH=$PWD/bin:$PATH PS1='$ ' && clear"
Enter
Show

Sleep 500ms
Type "vibescan scan examples/vibe-coded-app/server.js"
Sleep 600ms
Enter
Sleep 16s
