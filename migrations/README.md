# migrations

Each file here is one step that turns a clean phpscript clone into phpx. The numeric prefix orders the steps and `migrations/run.sh` runs the set.

| Step                 | What it rewrites                                    |
|----------------------|-----------------------------------------------------|
| `001-module-path.sh` | the `module` line in go.mod, and every import of it |
| `002-readme.sh`      | `README.md`, from `README.phpx.md`                  |

## The contract

- A step rewrites the tree and commits nothing. The caller commits.
- A step run twice leaves the same tree. It reads what it wrote last time and skips.
- A step reads the tree and never the network.
- A step that cannot finish exits non-zero and names the path that stopped it.
- A step rewrites what the fork has to rewrite. The fork's own files arrive as additions, and an addition belongs in a commit, not here.

## Why the steps run again on every pull

The go toolchain renames a module path halfway. `go mod edit -module` writes the go.mod line and leaves every import of the old path alone, and the import pass is a text rewrite over the `*.go` files that name it. That rewrite reaches most of the tree. A commit holding it conflicts with any upstream commit that touches an import, and the rebase turns into one resolution per file.

`phpx/scripts/pull.sh` drops that commit before it rebases and writes it again after. The fork then carries two kinds of commit:

- the fork's own files, under paths upstream never writes, which the rebase replays
- the output of these steps, which the next pull drops and writes again

## Running the steps

```sh
bash migrations/run.sh          # every step, in order
bash migrations/001-module-path.sh   # one step
```

`atkins -f atkins.phpx.yml phpx:migrate` is the same set under the pipeline, and `phpx:pull` runs it after a rebase. A clean phpscript clone reaches phpx by taking the fork's additive commit on top of it and running `run.sh`:

```sh
git clone https://github.com/titpetric/phpscript.git phpx && cd phpx
git remote add fork git@github.com:titpetric/phpx.git && git fetch fork main
git cherry-pick $(git rev-list --reverse origin/main..fork/main --grep '^chore(phpx): generated' --invert-grep)
bash migrations/run.sh
```
