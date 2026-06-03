Output demo/simple.gif

Set FontSize 14
Set Width 1100
Set Height 500
Set Theme "Catppuccin Mocha"
Set Padding 20

Type "SELECTION=$(./chooz testdata/example.yaml)"
Enter
Sleep 1s
Down
Sleep 600ms
Down
Sleep 600ms
Up
Sleep 600ms
Enter
Sleep 800ms
Type "echo RETURNCODE=$?"
Enter
Type "echo Selection was $SELECTION"
Enter
Sleep 5000ms
