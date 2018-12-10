<a name="v1.6.0"></a>
# [v1.6.0](https://github.com/rhysd/notes-cli/releases/tag/v1.6.0) - 10 Dec 2018

- **New:** `notes` with no argument now shows `notes ls -o` if any note exists
- **Improve:** `list --full` output buffering
- **Fix:** Broken pipe on using pager in some case

[Changes][v1.6.0]


<a name="v1.5.1"></a>
# [v1.5.1](https://github.com/rhysd/notes-cli/releases/tag/v1.5.1) - 02 Dec 2018

Allow to disable Git, editor or pager by setting empty to the corresponding environment variable.

For example, following will disable pager on `notes ls`

```
export NOTES_CLI_PAGER=
```

[Changes][v1.5.1]


<a name="v1.5.0"></a>
# [v1.5.0](https://github.com/rhysd/notes-cli/releases/tag/v1.5.0) - 25 Nov 2018

- **New:** Do paging `list` command's long output using pager command. Default pager command is `less` and it can be customizable via `$NOTES_CLI_PAGER`
- **Improve:** Change layout of `list --oneline`. Category only field was removed and now category is unified with relative path. First field is not changed and category can be retrieved from the relative path. So this should not be breaking change
- **Improve:** Truncate note body lines in `list --full` by number of lines, not bytes
- **Improve:** Add `...` to the end of truncated note body in `list --full`
- **Improve:** Improve help description and error message
- **New:** Add `--no-edit` to `new` command for an editor plugin

[Changes][v1.5.0]


<a name="v1.4.0"></a>
# [v1.4.0](https://github.com/rhysd/notes-cli/releases/tag/v1.4.0) - 18 Nov 2018

- **New:** Now category can be nested with `/` like `blog/myown` or `blog/dev.to`.
- **New:** Allow to create an external subcommands like `git`. `notes-foo` in `$PATH` is called on `notes foo` with passing arguments and path to `notes` executable.
- **New:** Allow to put `.template.md` at root of notes-cli home directory tree. It is always used as template for creating a new note if category-specific `.template.md` is not found.
- **Improve:** When a template is starting with `-->` (closing comment), `notes` considers it is hiding metadata and automatically insert corresponding `<!--` before metadata
- **New:** (API) `Category` and `Categories` types are added and `CollectCategories()` factory function is added
- **Fix:** Category or file name or tags which contain wide characters such as CJK are now correctly aligned. NFD file paths on macOS are now correctly normalized also.
- **Fix:** Sort order at `--sort created` was opposite. Now it is in descending order hence `head -1` can take the latest note
- **Fix:** Close file after inline input is written to created note at `notes new`
- **Improve:** (Doc) Tweak README sections structure and add TOC since it gets bigger.

[Changes][v1.4.0]


<a name="v1.3.0"></a>
# [v1.3.0](https://github.com/rhysd/notes-cli/releases/tag/v1.3.0) - 14 Nov 2018

- **New:** Refer `$EDITOR` environment variable when `$NOTES_CLI_EDITOR` is not set
- **Improve:** Allow `$NOTES_CLI_EDITOR` to have options such as `"vim -g"`. Previously only command and path could be specified like `"code"` or `"/path/to/emacs"`
- **Improve:** Add more documents
- **Fix:** Add/Fix some tests

[Changes][v1.3.0]


<a name="v1.2.0"></a>
# [v1.2.0](https://github.com/rhysd/notes-cli/releases/tag/v1.2.0) - 09 Nov 2018

- **New:** Enable to put a template file for notes per category
- **New:** Enable to hide metadata by surrounding it with `<!-- ... -->` comment
- **Fix:** Improve and fix descriptions of commands in help texts
- **Fix:** Improve migration script
- **Fix:** Improve API documents

[Changes][v1.2.0]


<a name="v1.1.2"></a>
# [v1.1.2](https://github.com/rhysd/notes-cli/releases/tag/v1.1.2) - 06 Nov 2018

- Fix getting the executable path on `selfupdate` command

To avoid above fixed bug in earlier version, please use full-path of executable when you update.

```
$ ~/.go/bin/notes selfupdate
```

[Changes][v1.1.2]


<a name="v1.1.1"></a>
# [v1.1.1](https://github.com/rhysd/notes-cli/releases/tag/v1.1.1) - 06 Nov 2018

1. Validate category name as directory on `notes new`
2. Ignore horizontal rules (`---`) after metadata
3. Add migration script from [memolist.vim](https://github.com/glidenote/memolist.vim). It's put in `scripts/` directory
4. Fix description of `--color-always`
5. Fix checking the latest version on `selfupdate`
6. Add Zsh/Bash completions
7. Add `man` manual

By 2., you can add `<hr/>` after metadata to separate your list from metadata:

```markdown
Shopping
=======
- **Category:** memo
- **Tags:**
- **Created:** 2018-11-6T20:36:00+09:00
----------

- Milk
- Egg
- Meat
- Tomato
```

[Changes][v1.1.1]


<a name="v1.1.0"></a>
# [v1.1.0](https://github.com/rhysd/notes-cli/releases/tag/v1.1.0) - 04 Nov 2018

- Add `--edit` to `list` subcommand
  - You can edit listed notes immediately without piping result to editor's argument
- Add `selfupdate` subcommand
  - It updates itself. You don't need to update binary manually

[Changes][v1.1.0]


<a name="v1.0.0"></a>
# [v1.0.0](https://github.com/rhysd/notes-cli/releases/tag/v1.0.0) - 03 Nov 2018

First release :tada:

Please see README and `notes help` to know how to use.

https://github.com/rhysd/notes-cli/blob/master/README.md

[Changes][v1.0.0]


[v1.6.0]: https://github.com/rhysd/notes-cli/compare/v1.5.1...v1.6.0
[v1.5.1]: https://github.com/rhysd/notes-cli/compare/v1.5.0...v1.5.1
[v1.5.0]: https://github.com/rhysd/notes-cli/compare/v1.4.0...v1.5.0
[v1.4.0]: https://github.com/rhysd/notes-cli/compare/v1.3.0...v1.4.0
[v1.3.0]: https://github.com/rhysd/notes-cli/compare/v1.2.0...v1.3.0
[v1.2.0]: https://github.com/rhysd/notes-cli/compare/v1.1.2...v1.2.0
[v1.1.2]: https://github.com/rhysd/notes-cli/compare/v1.1.1...v1.1.2
[v1.1.1]: https://github.com/rhysd/notes-cli/compare/v1.1.0...v1.1.1
[v1.1.0]: https://github.com/rhysd/notes-cli/compare/v1.0.0...v1.1.0
[v1.0.0]: https://github.com/rhysd/notes-cli/tree/v1.0.0

 <!-- Generated by changelog-from-release -->
