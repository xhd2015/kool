// Package rules provides functionality for managing Cursor rule files.
//
// The package supports:
// - Adding rule files to the user's ~/.kool/rules/files directory
// - Listing available rule files
// - Using rule files by copying them to the current project's .cursor/rules directory
// - Removing rule files from the ~/.kool/rules/files directory
// - Renaming rule files in the ~/.kool/rules/files directory
// - Printing the content of rule files
// - Entering an interactive shell for working with rules
//
// When running 'kool rule' without arguments, an interactive bash sub-shell
// is launched with a custom prompt showing the rules directory. In this shell,
// users can execute all standard bash commands plus special rule management
// commands like 'use <file>' to apply a rule.
package rules
