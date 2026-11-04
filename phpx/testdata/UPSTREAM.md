# Upstream seams

phpx carries 40 overlaid files (`phpx/testdata/patches/`). An overlay is a whole file copied over upstream's after
every follow, so an upstream edit to it goes unseen and the fork's edit is only as fresh as the last capture.
The way out is the one the bindings already took: upstream owns a registry, the embedder fills it. Each pull request
below replaces one overlay with a registration the fork makes from a file it owns.

## Open

| PR | Seam | Overlay it retires |
|----|------|--------------------|
| [#204](https://github.com/titpetric/phpscript/pull/204) | `files.RegisterWrapper(scheme, open)`: `fopen('php://<scheme>')` asks a registry | `stdlib/files/stream.go` (`stdlib/files/phpx_wrappers.go` becomes a registration of `stdout` and `stderr`) |
| [#205](https://github.com/titpetric/phpscript/pull/205) | `Runtime.RegisterClassSource(file, source, classes...)`: autoload declares PHP source on first use | `runner/runner_load.go` (`runner/phpx_prelude.go` becomes two registrations, Reflection and SPL) |

Neither changes upstream behaviour: nothing is registered there.

## Next, in the order the overlays cost

1. **Expression compile hook** (`runner/expr/direct.go`, `direct_operators.go`). The fork's entry is one line at the top
   of `directCompiler.compile`, `dc.phpxSpread(e)`, which now fronts spread arguments, spread array literals, named
   arguments, assignment expressions on non-variable targets and the lazy constant arms. An exported
   `expr.RegisterCompileHook(func(*Compiler, model.Expr) (Closure, bool, error))` retires the file if `closure` and
   `Env` become exported types, and a second hook for the right operand of `&&`/`||`/`?:` retires
   `direct_operators.go`.
2. **Environment helpers** (`runner/runtime_eval.go`). `buildEnv` installs `__set`, `__get` and the rest from `rt`; the
   fork adds `__phpx_assign` and `__phpx_offset` the same way, and `func_get_args` and the string-offset write call
   fork code. `Runtime.RegisterEnvHelper(name, func(*scopeRef) any)` is the whole seam.
3. **Parser extension points** (`parser/expr.go`, `parser.go`, `lexer.go`, 87 + 80 changed lines). The fork's syntax
   splits into three kinds, and each is one hook:
   - token-stream rewrites (`enum` becomes a class): `parser.RegisterTokenRewrite(func([]Token) []Token)` with `Token`
     exported;
   - a prefix keyword in expression position (`match`, `print`, `clone`, `throw`): `parser.RegisterPrefix(keyword, func(*Parser) (model.Expr, error))`;
   - a postfix or infix form (`?->`, `<=>`, `::` on a value, `xor`): `RegisterPostfix` and a binary-operator table.
   Exporting the cursor (`Cur`, `Next`, `IsOp`, `ParseExpr`) is most of the work.
4. **Assignment and scope** (`runner/runner_assign.go`, `runner_frame.go`, `runner_control.go`, `dispatch.go`,
   `instanceof.go`, `helpers.go`). Each carries one to three seam lines (`phpxStringOffset`, `phpxArrayAccessSet`,
   `phpxNamedArgs`, `phpxOwnedArray`, `phpxForeachObject`, `phpxHostThrowableIs`). A `runner.Extension` interface with
   those methods and a no-op default would turn all of them into one registration.
5. **Test files that pin a reversed decision** (`stdlib/core/json_test.go`, `tests/bindings_test.go`, `closure_test.go`,
   `platform_test.go`). These are not seams. They are the decisions in `docs/design.md` the fork reverses, and the
   fork's answer is to drop or reword them, which an overlay does and an interface cannot. If upstream ever wants the
   JSON constants, native slices as writable arrays or `instanceof Throwable`, the tests change there.

## Not proposed

Everything in the "Won't build" table stays upstream's decision: `eval`, `goto`, generators, `curl_*`, PDO,
`${var}` interpolation. The fork has not touched them.
