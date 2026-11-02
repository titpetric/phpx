# phpx/testdata

One folder per reversal, and the fixtures that settle it. `phpx test --matrix ./phpx/testdata/...` runs the tree on all three runtimes.

| Folder         | Reversal                               | State |
|----------------|----------------------------------------|-------|
| `globals/`     | global values and the `global` keyword | done  |
| `inheritance/` | `extends`                              | done  |
| `traits/`      | traits                                 | done  |
| `abstract/`    | `abstract class`                       | done  |
| `stdclass/`    | `stdClass`                             | done  |
| `syntax/`      | OOP syntax a framework snippet met     | done  |
| `frameworks/`  | one snippet tree per framework         | open  |

The first fixture for a reversal creates its folder, and `README.phpx.md` carries the box that ticks when the folder passes. `syntax/` collects the class syntax the framework snippets met on the way to a reversal. `frameworks/` holds a `composer.json` and an entry program per framework; a blocked framework holds no fixture, and [report.md](report.md) names the wall it meets.

## Writing one

A fixture is a `.phpt` document: the header, then the source, then the expected output, each separated by a line of three dashes. The expected section is what `php` printed, pasted from a run of the source through the php binary. Output copied from phpx locks in whatever phpx does today, which is the opposite of what these fixtures are for.

The matrix has three columns. The PHP column is the law. The Runtime column is the interpreter, where the overlay lives. The Flat stack column is the bytecode engine, and a `SKIP` there means the engine declined the program and the whole of it ran on the interpreter, which is the atomic fallback `docs/flatstack.md` requires.

A reversal the engine declines is a reversal with a performance cost, and the fixture's description says which construct the engine turned down.

## What a reversal touches

The overlay for a reversal lands in three places, and `README.phpx.md` argues each:

- a fork-owned Go file in the package that needs it, named `phpx.go` or `phpx_*.go`, which a follow leaves alone
- `patches/`, for the upstream files the change has to edit, copied over the tree on every follow
- this tree, for the fixtures that settle the behaviour against php
