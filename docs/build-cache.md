# Build cache

A run parses a PHP file once per process. The build cache carries that parsed
program to the next process as a file, under the user's cache directory. A
process that finds an entry for an unchanged file skips the lexer, the parser
and their allocations. The runtime, the formatter and flatstack read the same
`model.Program` either way.

## Naming

An entry is `$XDG_CACHE_HOME/phpx/<build>_<key>.phpb`, which is
`$HOME/.cache/phpx` when `XDG_CACHE_HOME` is unset. `PHPScript_CACHE` names
another directory.

- `<build>` is the source revision a Go build stamped into the binary, twelve
  hex characters. A dirty tree or a build without version information takes the
  executable's size and modification time instead, which changes on every
  rebuild.
- `<key>` is the SHA-256 of the source name the runtime resolved and the parse
  options. A file parsed as `Runtime\Go` keeps another entry than the same file
  parsed as phpscript.

The entry header repeats the build, the name, the options and the SHA-256 of
the source bytes. A read compares all four. One field that differs answers a
miss.

## Failure is a miss

The cache decides nothing. A missing entry, a stale entry, a corrupt entry and
a payload the decoder refuses all fall back to the parse, and the parse reports
its errors as it always did. A parse whose program cannot be encoded writes no
entry. No run fails because of the cache.

A write lands through a temporary file and a rename in the same directory, and
two processes may use one cache directory at once.

## Turning it on

The library leaves it off: `runner.Options.BuildCache` names the directory and
is empty by default.

`phpscript run` and `phpscript test` turn it on, under the default directory.
`--build-cache-dir` moves it and `--no-build-cache` turns it off. A configuration
file sets `runner.build_cache`; `PHPScript_CACHE` sets the directory for every
command.

## Cost

A cold read pays one parse, one encode and one write. An encode of the 34 KB
prelude costs 0.2 ms and 140 KB of allocations; a parse of the same file costs
0.83 ms and 364 KB. The second process pays the file read, the content hash and
the decode: 0.32 ms and 384 KB through `LoadFile`, against 0.97 ms and 414 KB
to parse.

The entry is smaller than the source: 22.9 KB for the 34 KB prelude, 34.5 KB
for a 73 KB Symfony YAML parser. A suite of 327 fixtures writes 347 entries and
1.5 MB.

## Limits

- The entry holds the AST, and flatstack compiles that AST on every process.
  A bytecode section is the next step.
- Nothing prunes the directory yet. `buildcache.Store.Prune` drops old entries
  by age and size, and no command calls it on a schedule.
- A source name the runtime spells the same from two roots shares one key. The
  content hash and the options in the header keep that correct, at the cost of
  one entry replacing the other.
