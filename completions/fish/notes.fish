# Global flags
complete -c notes -s h -l help -d "Show context-sensitive help."
complete -c notes -s A -l color-always -d "Enable color output always"
complete -c notes -l no-color -d "Disable color output"
complete -c notes -l version -d "Show application version."

# Subcommands
complete -c notes -n '__fish_use_subcommand' -xa 'help' -d "Show help."
complete -c notes -n '__fish_use_subcommand' -xa 'new' -d "Create a new note with given category and file name"
complete -c notes -n '__fish_use_subcommand' -xa 'list' -d "List notes with filtering by categories and/or tags with regular expressions. By default, it shows full path of notes (alias: ls)"
complete -c notes -n '__fish_use_subcommand' -xa 'ls' -d "List notes with filtering by categories and/or tags with regular expressions. By default, it shows full path of notes (alias: ls)"
complete -c notes -n '__fish_use_subcommand' -xa 'categories' -d "List all categories to stdout (alias: cats)"
complete -c notes -n '__fish_use_subcommand' -xa 'cats' -d "List all categories to stdout (alias: cats)"
complete -c notes -n '__fish_use_subcommand' -xa 'tags' -d "List all tags"
complete -c notes -n '__fish_use_subcommand' -xa 'save' -d "Save notes using Git. It adds all notes and creates a commit to Git repository at home directory"
complete -c notes -n '__fish_use_subcommand' -xa 'config' -d "Output config values to stdout. By default output all values with KEY=VALUE style"
complete -c notes -n '__fish_use_subcommand' -xa 'selfupdate' -d "Update myself to the latest version. It downloads the latest version executable and replaces current executable with it"

# Flags for subcommands
complete -c notes -n '__fish_seen_subcommand_from new' -l no-inline-input -d "Does not request inline input even if no editor is set"

complete -c notes -n '__fish_seen_subcommand_from ls list' -l no-inline-input -d "Does not request inline input even if no editor is set"
complete -c notes -n '__fish_seen_subcommand_from ls list' -s f -l full -d "Show full information of note instead of path"
complete -c notes -n '__fish_seen_subcommand_from ls list' -l category -d "Filter category name by regular expression"
complete -c notes -n '__fish_seen_subcommand_from ls list; and [ \'--category\' = (string split " " (commandline))[-2] ]' -xa (notes categories)
complete -c notes -n '__fish_seen_subcommand_from ls list' -l tag -d "Filter tag name by regular expression"
complete -c notes -n '__fish_seen_subcommand_from ls list; and [ \'--tag\' = (string split " " (commandline))[-2] ]' -xa (notes tags)
complete -c notes -n '__fish_seen_subcommand_from ls list' -s r -l relative -d 'Show relative paths from $NOTES_CLI_HOME directory'
complete -c notes -n '__fish_seen_subcommand_from ls list' -s o -l oneline -d "Show oneline information of note instead of path"
complete -c notes -n '__fish_seen_subcommand_from ls list' -l sort -d "Sort results by 'modified', 'created', 'filename' or 'category'. 'created' is default"
complete -c notes -n '__fish_seen_subcommand_from ls list' -s e -l edit -d 'Open listed notes with an editor. $NOTES_CLI_EDITOR must be set'

complete -c notes -n '__fish_seen_subcommand_from save' -l message -d "Commit message on save"

complete -c notes -n '__fish_seen_subcommand_from selfupdate' -l dry -d 'Dry run update. Only check the newer version is available'

# Candidates for subcommands
complete -c notes -n '__fish_seen_subcommand_from config' -xa 'home' -d "Home directory of notes-cli"
complete -c notes -n '__fish_seen_subcommand_from config' -xa 'editor' -d "Editor command path to open note"
complete -c notes -n '__fish_seen_subcommand_from config' -xa 'git' -d "Git command path to save notes"

complete -c notes -n '__fish_seen_subcommand_from help' -xa 'help' -d "Show help."
complete -c notes -n '__fish_seen_subcommand_from help' -xa 'new' -d "Create a new note with given category and file name"
complete -c notes -n '__fish_seen_subcommand_from help' -xa 'list' -d "List notes with filtering by categories and/or tags with regular expressions. By default, it shows full path of notes (alias: ls)"
complete -c notes -n '__fish_seen_subcommand_from help' -xa 'ls' -d "List notes with filtering by categories and/or tags with regular expressions. By default, it shows full path of notes (alias: ls)"
complete -c notes -n '__fish_seen_subcommand_from help' -xa 'categories' -d "List all categories to stdout (alias: cats)"
complete -c notes -n '__fish_seen_subcommand_from help' -xa 'cats' -d "List all categories to stdout (alias: cats)"
complete -c notes -n '__fish_seen_subcommand_from help' -xa 'tags' -d "List all tags"
complete -c notes -n '__fish_seen_subcommand_from help' -xa 'save' -d "Save notes using Git. It adds all notes and creates a commit to Git repository at home directory"
complete -c notes -n '__fish_seen_subcommand_from help' -xa 'config' -d "Output config values to stdout. By default output all values with KEY=VALUE style"
complete -c notes -n '__fish_seen_subcommand_from help' -xa 'selfupdate' -d "Update myself to the latest version. It downloads the latest version executable and replaces current executable with it"
