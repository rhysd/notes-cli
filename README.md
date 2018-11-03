A small CLI note taking tool with your favorite editor
======================================================
[![Linux/Mac Build Status][travisci-badge]][travisci]
[![Windows Build Status][appveyor-badge]][appveyor]
[![Coverage Report][codecov-badge]][codecov]
[![Documentation][doc-badge]][doc]

This is a small CLI tool for note taking in terminal with your favorite editor.
You can create/list notes via this tool.
This tool also optionally can save your notes thanks to Git to avoid losing your notes.

This tool is intended to be used nicely with other commands such as `grep` (or [ag][], [rg][]),
`rm`, filtering tools such as [fzf][] or [peco][] and editors which can be started from command line.



## Installation

Download an archive for your OS from [release page](https://github.com/rhysd/notes-cli/releases).
It contains an executable. Please unzip the archive and put the executable in a directory in `$PATH`.

Or you can install by building from source directly as follows. Go toolchain is necessary.

```
$ go get -u github.com/rhysd/notes-cli/cmd/notes
```



## Usage

### Basic Usage

`notes` provides some subcommands to manage your markdown notes.

- Create a new note with `notes new <category> <filename> [<tags>]`.
- Open existing note by `notes ls` and your favorite editor. e.g., `vim $(notes ls)` opens all notes in Vim.
- Check existing notes on terminal with `notes ls -o` (`-o` means showing one line information for each note).

By `notes list` (omitting `-o`), it shows paths separated by newline. By choosing one line from the
output and pass it to your editor's command as argument, you can easily open the note in your editor.

Every note must have one category. And it can have zero or more tags.


### Create a new note

For example,

```
$ notes new blog how-to-handle-files golang,file
```

creates a note file at `<HOME>/notes-cli/blog/how-to-handle-files.md` where `<HOME>` is
[XDG Data directory][xdg-dirs] (on macOS, `~/.local/share/notes-cli`) by default and can be specified
by `$NOTES_CLI_HOME` environment variable. The home directory is automatically created.

Category is `blog`. Every note must belong to one category.

Tags are `golang` and `file`. Tags can be omitted.

Directories structure under home is something like:

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

If you set your favorite editor to `$NOTES_CLI_EDITOR` environment variable, it opens the newly
created note file with it. You can seamlessly edit the file.

```markdown
how-to-handle-files
===================
- Category: blog
- Tags: golang, file
- Created: 2018-10-28T07:19:27+09:00

```

Please do not remove `- Category: ...`, `- Tags: ...` and `- Created: ...` lines and title.
They are used by `notes` command (modifying them is OK).
Default title is file name. You can edit the title and body of note as follows:

```markdown
How to handle files in Go
=========================
- Category: blog
- Tags: golang, file
- Created: 2018-10-28T07:19:27+09:00

Please read documentation.
GoDoc explains everything.
```

Note that every note is under the category directory of the note. When you change a category of note,
you also need to adjust directory structure manually (move the note file to new category directory).

For more details, please check `notes new --help`.


### Open notes you created flexibly

Let's say to open some notes you created.

You can show the list of note paths with:

```
$ notes list # or `notes ls`
```

For example, now there is only one note so it shows one path

```
/Users/me/.local/share/notes-cli/blog/how-to-handle-files.md
```

Note that `/Users/<NAME>/.local/share` is a default XDG data directory on macOS or Linux and you can
change it by setting `$NOTES_CLI_HOME` environment variable.

When there are multiple notes, note is output per line. So you can easily retrieve some notes from
them by filtering the list with `grep`, `head`, `peco`, `fzf`, ...

```
$ notes ls | grep -l file | xargs -o vim
```

Or following also works.

```
vim $(notes ls | xargs grep file)
```

And searching notes is also easy by using `grep`, `rg`, `ag`, ...

```
$ notes ls | xargs ag documentation
```

When you want to search and open it with Vim, it's also easy.

```
$ notes ls | xargs ag -l documentation | xargs -o vim
```

`notes ls` accepts `--sort` option and changes the order of list. By default, the order is created
date time of note. By ordering with modified time of note, you can instantly open last-modified note
as follows since first line is a path to the note most recently modified.

```
$ note ls --sort modified | head -1 | xargs -o vim
```

For more details, please check `notes list --help`.


### Check notes you created as list

`notes list` also can show brief of notes to terminal.

You can also show the full information of notes on terminal with `--full` (or `-f`) option.

```
$ notes list --full
```

For example,

```
/Users/me/.local/share/notes-cli/blog/how-to-handle-files.md
Category: blog
Tags: golang, file
Created: 2018-10-28T07:19:27+09:00

How to handle files in Go
=========================

Please read documentation.
GoDoc explains everything.

```

It shows

- Full path to the note file
- Metadata `Category`, `Tags` and `Created`
- Title of note
- Body of note (upto 200 bytes)

with colors.

When there are many notes, it outputs many lines. In the acse, a pager tool like `less` is useful
to see the output per page. `-A` global option is short of `--always-color`.

```
$ notes -A ls --full | less -R
```

When you want to see the all notes quickly, `--oneline` (or `-o`) may be more useful than `--full`.
`notes ls --oneline` shows one brief of note per line.

For example,

```
blog/how-to-handle-files.md blog golang,file How to handle files in Go
```

- First field indicates a relative path of note file from home directory.
- Second field indicates a category of the note.
- Third field indicates comma-separated tags of the note. When note has no tag, it leaves as blank.
- Rest is a title of the note

This is useful for checking many notes at a glance.

For more details, please see `notes list --help`.


### Save notes to Git repository

Finally you can save your notes as revision of Git repository.

```
$ notes save
```

It saves all your notes under your `notes-cli` diretory as Git repository.
It adds all changes in notes and automatically creates commit.

By default, it only adds and commits your notes to the repository. But if you set `origin` remote to
the repository, it automatically pushes the notes to the remote.

For more details, please see `notes save --help`.


### Use from Go program

This command can be used from Go program as a library. Please read [API documentation][doc] to know
the interfaces.



## FAQ

### I want to specify `/path/to/dir` as home

Please set it to environment variable.

```sh
export NOTES_CLI_HOME=/path/to/dir
```


### How can I grep notes?

Please combine grep tools with `notes list` on your command line. For example,

```sh
$ grep -E some word $(notes list)
$ ag some word $(notes list)
```

If you want to filter with categories or tags, please use `-c` and/or `-t` of `list` command.


### How can I filter notes interactively and open it with my editor?

Please pipe the list of paths from `notes list`. Following is an example with `peco` and Vim.

```sh
$ notes list | peco | xargs -o vim --not-a-term
```


### Can I open the latest note without selecting it from list?

Output of `notes list` is sorted by created date time by default. By using `head` command, you can
choose the latest note in the list.

```sh
$ vim "$(notes list | head -1)"
```

If you want to access to the last modified note, sorting by `modified` and taking first item by `head`
should work.

```sh
$ vim "$(notes list --sort modified | head -1)"
```

By giving `--sort` (or `-s`) option to `notes list`, you can change how to sort. Please see
`notes list --help` for more details.


### How can I remove some notes?

Please use `rm` and `notes list`. Following is an example that all notes of specific category `foo`
are removed.

```sh
$ rm $(notes list -c foo)
```

Thanks to Git repository, this does not remove your notes completely until you run `notes save`
next time.


### Can I nest categories?

Categories cannot be nested. Instead, you can define your own nested naming rule for categories.
For example `blog-personal-public` can indicate blog entries which is personal and publicly posted.
Other categories would be named like `blog-personal-private`, `blog-company-public`, ...
It's up to you.


### How can I integrate with Vim?

Please write following code in your `.vimrc`.

```vim
function! s:notes_selection_done(selected) abort
    silent! autocmd! plugin-notes-cli
    let home = substitute(system('notes config home'), '\n$', '', '')
    let sep = has('win32') ? '\' : '/'
    let path = home . sep . split(a:selected, ' ')[0]
    execute 'split' '+setf\ markdown' path
    echom 'Note opened: ' . a:selected
endfunction
function! s:notes_open(args) abort
    execute 'terminal ++close bash -c "notes list --oneline | peco"'
    augroup plugin-notes-cli
        autocmd!
        autocmd BufWinLeave <buffer> call <SID>notes_selection_done(getline(1))
    augroup END
endfunction
command! -nargs=* NotesOpen call <SID>notes_open(<q-args>)

function! s:notes_new(...) abort
    if has_key(a:, 1)
        let cat = a:1
    else
        let cat = input('category?: ')
    endif
    if has_key(a:, 2)
        let name = a:2
    else
        let name = input('filename?: ')
    endif
    let tags = get(a:, 3, '')
    let cmd = printf('notes new --no-inline-input %s %s %s', cat, name, tags)
    let out = system(cmd)
    if v:shell_error
        echohl ErrorMsg | echomsg string(cmd) . ' failed: ' . out | echohl None
        return
    endif
    let path = split(out)[-1]
    execute 'edit!' path
    normal! Go
endfunction
command! -nargs=* NotesNew call <SID>notes_new(<f-args>)

function s:notes_last_mod(args) abort
    let out = system('notes list --sort modified ' . a:args)
    if v:shell_error
        echohl ErrorMsg | echomsg string(cmd) . ' failed: ' . out | echohl None
        return
    endif
    let last = split(out)[0]
    execute 'edit!' last
endfunction
command! -nargs=* NotesLastMod call <SID>notes_last_mod(<q-args>)
```

- `:NotesOpen [args]`: It shows notes as list with incremental filtering thanks to `peco` command.
  The chosen note is opened in a new buffer. `args` is passed to `notes list` command. So you can
  easily filter paths by categories (`-c`) or tags (`-t`).
- `:NotesNew [args]`: It creates a new note and opens it with a new buffer. `args` is the same as
  `notes new` but category and file name can be empty. In the case, Vim ask you to input them after
  starting the command.
- `:NotesLastMod [args]`: It opens the last modified note in new buffer. When `args` is given, it
  is passed to underlying `notes list` command execution so you can filter result by categories
  and/or tags with `-c` or `-t`.



## License

[MIT License](LICENSE.txt)




[ag]: https://github.com/ggreer/the_silver_searcher
[rg]: https://github.com/BurntSushi/ripgrep
[fzf]: https://github.com/junegunn/fzf
[peco]: https://github.com/peco/peco
[xdg-dirs]: https://wiki.archlinux.org/index.php/XDG_Base_Directory
[appveyor-badge]: https://ci.appveyor.com/api/projects/status/5pbcku1buw8gnqu9/branch/master?svg=true
[appveyor]: https://ci.appveyor.com/project/rhysd/notes-cli
[travisci-badge]: https://travis-ci.org/rhysd/notes-cli.svg?branch=master
[travisci]: https://travis-ci.org/rhysd/notes-cli
[codecov-badge]: https://codecov.io/gh/rhysd/notes-cli/branch/master/graph/badge.svg
[codecov]: https://codecov.io/gh/rhysd/notes-cli
[doc-badge]: https://godoc.org/github.com/rhysd/notes-cli?status.svg
[doc]: http://godoc.org/github.com/rhysd/notes-cli
