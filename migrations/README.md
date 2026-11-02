# migrations

Each file here is one step that turns a clean phpscript clone into phpx. The numeric prefix orders the steps and `migrations/run.sh` runs the set.

| Step                    | What it rewrites                                     |
|-------------------------|------------------------------------------------------|
| `001-module-path.sh`    | the `module` line in go.mod, and every import of it  |
| `002-readme.sh`         | `README.md`, from `README.phpx.md`                   |
| `003-overlay.sh`        | the upstream files `phpx/testdata/patches/` carries  |
| `004-atkins-state.sh`   | `.gitignore`, for atkins' run database under `.atkins/` |

## The contract

- A step rewrites the tree and commits nothing. The caller commits.
- A step run twice leaves the same tree. It reads what it wrote last time and skips.
- A step reads the tree and never the network.
- A step that cannot finish exits non-zero and names the path that stopped it.
- A step rewrites what the fork has to rewrite. The fork's own files arrive as additions, and an addition belongs in a commit, not here.

## Why the steps run again on every follow

The go toolchain renames a module path halfway. `go mod edit -module` writes the go.mod line and leaves every import of the old path alone, and the import pass is a text rewrite over the `*.go` files that name it. That rewrite reaches most of the tree, and git has no way to hold it as an edit: any upstream commit that touches an import collides with it, one resolution per file.

So the fork never asks git to hold it. `phpx/scripts/follow.sh` takes upstream's version of every upstream file verbatim and runs these steps over the result, which reproduces the fork's half exactly and leaves no file with two authors. A migration is the fork's only way to change an upstream file; a committed edit to one disappears on the next follow.

## Running the steps

```sh
bash migrations/run.sh          # every step, in order
bash migrations/001-module-path.sh   # one step
```

`atkins -f atkins.phpx.yml phpx:migrate` is the same set under the pipeline, and `phpx:follow` runs it over the tree it took from upstream. A clean phpscript clone reaches phpx by taking the fork's own files on top of it and running `run.sh`:

```sh
git clone https://github.com/titpetric/phpscript.git phpx && cd phpx
git remote add fork git@github.com:titpetric/phpx.git && git fetch fork main
git checkout fork/main -- README.phpx.md ACCEPTANCE.md CHANGELOG.md atkins.phpx.yml lefthook.yml migrations phpx
bash migrations/run.sh
```

`follow.sh` holds that list of paths as `fork_owns`, and a new fork-owned path goes on it in the commit that creates it.
