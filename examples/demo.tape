Output examples/demo.gif

Set Shell "bash"
Set FontSize 20
Set Width 1600
Set Height 600
Set Padding 80
Set Theme "Everblush"

# Create file and stage it
Type "touch file"
Sleep 1
Enter
Type "git add file"
Sleep 1
Enter

# Enter program
Type "cometary"
Sleep 500ms
Enter
Sleep 3

# Choose correct type of commit
Down@500ms 1
Enter
Sleep 3

# Enter commit scope
Type "file"
Sleep 2
Enter
Sleep 2

# Enter commit message
Type "add new file"
Sleep 2
Enter
Sleep 2

# Confirm no ammend
Enter
Sleep 4
