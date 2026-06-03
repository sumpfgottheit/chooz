Output demo/simple.gif

Set FontSize 14
Set Width 1300
Set Height 550
Set Theme "Catppuccin Mocha"
Set Padding 20

Type "SELECTION=$(./chooz testdata/big.yaml)"
Enter
Sleep 1s

Down
Sleep 500ms
Down
Sleep 500ms
Down
Sleep 500ms
Up
Sleep 500ms

Type "b"
Sleep 500ms
Type "a"
Sleep 500ms
Type "c"
Sleep 800ms

Backspace
Sleep 400ms
Backspace
Sleep 800ms

Escape
Sleep 800ms

Type "log"
Sleep 1s
Enter
Sleep 800ms

Type "echo RETURNCODE=$?"
Enter
Type "echo Selection was: $SELECTION"
Enter
Sleep 4s
