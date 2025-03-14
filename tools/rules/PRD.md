plan only: under tools/rules, I'd like to implement a functionality that supports cursor rule management.it should be called from @main.go via CLI.It should support the following commands:
- `kool rule add SOME_FILE.mdc` -> copy the file to ~/.kool/rules/files/SOME_FILE.mdc
- `kool rule list` -> show a list of added files
- `kool rule use SOME_FILE.mdc` -> check if there is existing file with same name under .cursor/rules.if not, copy that from ~/.kool/rules/files/SOME_FILE.mdc to current dir's .cursor/rules/SOME_FILE.mdc, otherwise, report error


add two subcommands:
- `kool rule dir` --> show the files directory
- `kool rule rm SOME_FILE.mdc` --> remove the specified file

improve the kool add so that when there is existsing file, append a date-time suffix to the copied file

TODO:
- when executing `kool rule`, enter a bash sub-shell(like python's venv), user can execute all commands, plus a special `use`.
