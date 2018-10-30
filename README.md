A small CLI note taking tool with your favorite editor
======================================================

**This repository is an experimental until version reaches v1.0.0. Features may be changed in the future and tests are not sufficient.**

This is a small CLI tool for note taking in terminal with your favorite editor.
You can create/list notes via this tool.
This tool also optionally can save your notes thanks to Git to avoid losing your notes.

This tool is intended to be used nicely with other commands such as `grep` (or [ag][], [rg][]),
`rm`, filtering tools such as [fzf][] or [peco][] and editors which can be started from command line.

## Installation

For now, please install by building from source directly as follows:

```
$ go get github.com/rhysd/notes-cli/cmd/notes
```

## Basic Usage

Every note must have one category. And it can have zero or more tags.

For example,

```
$ notes new blog how-to-handle-files golang,file
```

creates note file at `<HOME>/notes-cli/blog/how-to-handle-files.md` where `<HOME>` is
[XDG Data directory][xdg-dirs] by default and can be specified by `$NOTES_CLI_HOME` environment
variable.

Directory structure is something like:

```
<HOME>
├── category1
│   ├── note1.md
│   ├── note2.md
│   └── note3.md
├── category2
│   ├── note4.md
│   └── note5.md
└── category3
    └── note6.md
```

If you set your favorite editor by `$NOTES_CLI_EDITOR` environment variable, it opens the newly
created note file with your favorite editor. You can seamlessly edit the file.

Note file is something like:

```markdown
How to handle files in Go
=========================

- Category: blog
- Tags: golang, file
- Created: 2018-10-28T07:19:27+09:00
<!-- Do not touch list above -->

Please read documentation.
GoDoc explains everything.
```

Title and tags can be modified as you like, but when you change category, you also need to move
the file to appropriate directory manually. So changing category after created is not recommended.

You can show the list of note paths with:

```
$ notes list # or `notes ls`
```

Now there is only one note so it shows one path

```
/Users/me/.local/share/notes-cli/blog/how-to-handle-files.md
```

Note that `/Users/<NAME>/.local/share` is a default XDG data directory on macOS or Linux.

You can also show the full information of note with:

```
$ notes list --full
```

It shows:

```
/Users/me/.local/share/notes-cli/blog/how-to-handle-files.md
Category: blog
Tags: golang, file
Created: 2018-10-28T07:19:27+09:00

How to handle files in Go

```

Finally you can save your notes as revision of Git repository.

```
$ notes save
```

It adds all changes in notes and automatically caretes commit.
When `origin` is set as tracking remote, it pushes the changes to the remote.

This is just a basic usage. Please see `--help` for more details.

```
usage: notes [<flags>] <command> [<args> ...]

Simple note taking tool for command line with your favorite editor

Flags:
  -h, --help      Show context-sensitive help (also try --help-long and --help-man).
      --no-color  Disable color output
      --version   Show application version.

Commands:
  help [<command>...]
    Show help.

  new [<flags>] <category> <filename> [<tags>]
    Create a new note

  list [<flags>]
    List note paths with filtering by categories and/or tags with regular expressions

  categories
    List all categories

  tags [<category>]
    List all tags

  save [<flags>]
    Save memo data with Git

  config [<name>]
    Output config value to stdout

```

### How to integrate with Vim

Please write following code in your `.vimrc`.

```vim
function! s:notes_selection_done(selected) abort
    silent! autocmd! plugin-notes-cli
    let home = substitute(system('notes config home'), '\n$', '', '')
    let sep = has('win32') ? '\' : '/'
    let path = home . sep . a:selected
    execute 'split' '+setf\ markdown' path
    echom 'Note opened: ' . a:selected
endfunction
function! s:notes_open(args) abort
    execute 'terminal ++close bash -c "notes list -r | peco"'
    augroup plugin-notes-cli
        autocmd!
        autocmd BufWinLeave <buffer> call <SID>notes_selection_done(getline(1))
    augroup END
endfunction

command! -nargs=* Notes call <SID>notes_open(<q-args>)
```

`:Notes [args]` command will be available. `args` is passed to `notes list` command.
So you can easily filter paths by categories (`-c`) or tags (`-t`). It shows the list of paths with
`peco` command.

After you choose one note from the list with `peco`, it automatically opens the note in new buffer.

## FAQ

### I want to specify `/path/to/dir` as home

Please set it to environment variable.

```sh
export NOTES_CLI_HOME=/path/to/dir
```

### I want to grep notes

Please combine grep tools with `notes list` on your command line. For example,

```sh
$ grep -E some word $(notes list)
$ ag some word $(notes list)
```

If you want to filter with categories or tags, please use `-c` and/or `-t` of `list` command.

### I want to filter notes interactively and open it with my editor

Please pipe the list of paths from `notes list`. Following is an example with `peco` and Vim.

```sh
$ notes list | peco | xargs -o vim --not-a-term
```

### I want to remove some notes

Please use `rm` and `notes list`. Following is an example that all notes of specific category `foo`
are removed.

```sh
$ rm $(notes list -c foo)
```

Thanks to Git repository, this does not remove your notes completely until you run `notes save`
next time.

### I want to nest categories

Categories cannot be nested. Instead, you can define your own nested naming rule for categories.
For example `blog-personal-public` can indicate blog entries which is personal and publicly posted.
Other categories would be named like `blog-personal-private`, `blog-company-public`, ...
It's up to you.

## License

[MIT License](LICENSE.txt)

[ag]: https://github.com/ggreer/the_silver_searcher
[rg]: https://github.com/BurntSushi/ripgrep
[fzf]: https://github.com/junegunn/fzf
[peco]: https://github.com/peco/peco
[xdg-dirs]: https://wiki.archlinux.org/index.php/XDG_Base_Directory
