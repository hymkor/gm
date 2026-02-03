Release notes
=============

v0.0.4
------
Feb 3, 2026

- Update go-readline-ny from v1.12.3 to v1.14.1 (#4)
- Update go-multiline-ny from v0.22.2 to v0.22.4
- Update go-ttyadapter from v0.1.0 to v0.3.0
  - Use `tty10pe.Tty` instead of `tty10.Tty`

v0.0.3
------
Nov 23, 2025

- (#3) Updated go-multiline-ny to v0.22.2 and aligned with API changes:
    - Replaced `readline.Editor.Out` with `multiline.Editor.Out()`
    - Replaced `readline.Editor.Writer` with `multiline.Editor.Writer()`
    - Replaced access to `.Tty` with `multiline.Editor.SetTty()`
    - Replaced `readline.Editor.Highlight` with `multiline.Editor.Highlight`

v0.0.2
------
Nov 28, 2024

- Fix: Dirty flag was not set on inserting an empty line after an empty line[^1]
- Use [go-windows1x-virtualterminal] instead of [go-colorable]
- Use [go-readline-ny/tty10] instead of [go-readline-ny/tty8]
- Update [go-multiline-ny v0.18.1]
- Update [go-readline-skk v0.4.2]

[^1]: https://github.com/hymkor/go-multiline-ny/commit/67be40253991c85482f2d4ff698ff4fb862d6404

[go-windows1x-virtualterminal]: https://github.com/hymkor/go-windows1x-virtualterminal
[go-readline-ny/tty10]: https://github.com/nyaosorg/go-readline-ny/tree/master/tty10
[go-readline-ny/tty8]: https://github.com/nyaosorg/go-readline-ny/tree/master/tty8.go
[go-colorable]: https://github.com/mattn/go-colorable
[go-multiline-ny v0.18.1]: https://github.com/hymkor/go-multiline-ny/releases/tag/v0.18.1
[go-readline-skk v0.4.2]: https://github.com/nyaosorg/go-readline-skk/releases/tag/v0.4.2

v0.0.1
------
Aug 17, 2023

- The first release
