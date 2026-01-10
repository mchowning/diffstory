# demo/navigation.tape
Output navigation.gif

Set Shell "bash"
Set FontSize 24
Set Width 1200
Set Height 700
Set Theme "Catppuccin Mocha"
Set Padding 20
Set TypingSpeed 75ms
Set CursorBlink false

# Build binary (hidden) - ensures binary is available
Hide
Type "go build -o ./diffstory ./cmd/diffstory && clear"
Enter
Sleep 1s
Show

# Launch with pre-loaded review
Type "./diffstory --review demo/review.json"
Sleep 500ms
Enter

# Wait for UI to render
Sleep 2s

# Navigate between sections
Type "j"
Sleep 700ms
Type "j"
Sleep 700ms
Type "k"
Sleep 700ms

# Focus Files panel (key "2") then collapse/expand with Enter
Type "2"
Sleep 500ms
Type "j"
Sleep 500ms
Enter
Sleep 700ms
Enter
Sleep 700ms

# Back to sections panel
Type "1"
Sleep 500ms

# Scroll diff content
Type "J"
Sleep 500ms
Type "J"
Sleep 500ms

Type "q"
Sleep 500ms
