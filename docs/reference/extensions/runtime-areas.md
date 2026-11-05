# Runtime areas

The runtime's bindings sit in namespaces called areas. An area is a group of functions, classes and constants that one Go installer registers, and its name is a namespace path. Three namespaces hold them:

| Namespace    | Holds                                                                    | Installed                      |
|--------------|--------------------------------------------------------------------------|--------------------------------|
| `Runtime\PS` | phpscript's own surface: php's standard library and the extensions       | by default, first              |
| `Runtime\PX` | the phpx overlay: the php compatibility work installed over `Runtime\PS` | by default, after `Runtime\PS` |
| `Runtime\Go` | a subset of go's standard library, written in php                        | by `use`, per file             |

`phpscript list --areas` prints the areas of a build with what each defines.

## Folders map to areas

A folder of `stdlib` is an area, and a file of a large folder is a sub-area:

| Folder               | Area                                                        |
|----------------------|-------------------------------------------------------------|
| `stdlib/core`        | `Runtime\PS\Core\<File>`, such as `Runtime\PS\Core\Strings` |
| `stdlib/compat`      | `Runtime\PS\Compat`                                         |
| `stdlib/crypto`      | `Runtime\PS\Crypto`, with `Hash`, `Random`, `Identifiers`   |
| `stdlib/files`       | `Runtime\PS\Files`                                          |
| `stdlib/http`        | `Runtime\PS\HTTP`, with `SAPI`                              |
| `stdlib/pexec`       | `Runtime\PS\Exec`                                           |
| `stdlib/phpx`        | `Runtime\PX\<File>`, such as `Runtime\PX\Datetime`          |
| `stdlib/gostd/<pkg>` | `Runtime\Go\<Pkg>`, such as `Runtime\Go\Strings`            |

The remaining folders (`auth`, `database`, `gd`, `info`, `internals`, `mail`, `regexp`, `session`, `span`, `time`) map to the area of the same name under `Runtime\PS`. A file named `phpx_<name>.go` inside a `Runtime\PS` folder registers `Runtime\PX\<Folder>\<Name>`: it overlays the folder's area and installs with the rest of `Runtime\PX`.

The [implemented API inventory](implemented-apis.md) groups every function and class by area.

## The overlay installs last

Every `Runtime\PS` area installs first, in registration order. Every `Runtime\PX` area installs after them, in registration order. A `Runtime\PX` binding that registers a name `Runtime\PS` already defines replaces it, and the area listing counts it under Replaces. The order follows the namespace and does not depend on which Go package initialised first.

## Calling a function through its area

An area is a namespace, and a function answers under it as well as under its bare name:

```php
echo \Runtime\PS\Core\Strings\strlen("four");   // 4
var_dump(function_exists('Runtime\PS\Core\Strings\strlen')); // bool(true)
```

The area has to be the one that defines the function, or the one that replaces it. `Runtime\PS\Core\Arrays\strlen` names nothing. Both runtimes resolve the name the same way. php has no `Runtime` namespace, and source that must also run on php calls the bare name.

## Configuring the default runtime

The `runner.areas` key of the configuration names the namespaces the default runtime installs. It is empty by default, which installs every area:

```yaml
runner:
  areas:
    - Runtime\PS
```

That runtime carries `Runtime\PS` and no overlay. A deeper name narrows further: `Runtime\PS\Core\Strings` installs the string functions and no other part of the standard library, which suits a runtime for a template engine that never calls a file function. A host sets `Options.Areas` on `runner.Options` for the same effect.

A name outside the installed areas stays undefined, and `function_exists` answers false. The `--stdlib` flag is the other axis: it selects profiles, such as `secure` and `exec`, and the two compose.

## Go's standard library, opt-in

`Runtime\Go` holds go's builtins (`make`, `len`, `append`; see [the Go runtime](go-runtime.md)) and the areas below it map go's import paths:

| Area                 | Go package | Names                                                                 |
|----------------------|------------|-----------------------------------------------------------------------|
| `Runtime\Go\Strings` | `strings`  | `Fields`, `Split`, `Contains`, `HasPrefix`, `Index`, `Repeat`, `Trim` |
| `Runtime\Go\Path`    | `path`     | `Base`, `Dir`, `Ext`, `Clean`, `Join`, `Match`                        |

A script imports one by its name, and the alias calls the function:

```php
use Runtime\Go\Strings;

echo implode("|", Strings\Fields("  a b   c "));   // a|b|c
```

The default runtime carries none of these names. The import installs the area on the runtime that runs the file, and `function_exists('Runtime\Go\Strings\Fields')` is false before it.

The mapping continues down go's import paths: `net/http` is `Runtime\Go\Net\HTTP`, `path/filepath` is `Runtime\Go\Path\Filepath`. A package that reaches outside the sandbox, such as a network client, registers under its own profile area, as `stdlib/pexec` does. A host serving untrusted scripts leaves that profile out.

## Writing an area

A binding package registers each installer under a name in its `init`:

```go
func init() {
	runner.RegisterArea(`Runtime\PS\Core\Strings`, Register)
}
```

`RegisterArea` takes the name and the installer; `RegisterProfileArea` adds the profile mask for an area that leaves the sandbox. Several files may register under one name. An installer that registers no name goes under `Runtime\PS\Unnamed`, which `runner.RegisterBinding` still does for a host.

An opt-in area registers with `runner.RegisterNamespace(name, installer)`, and names its functions with the full path (`Runtime\Go\Strings\Fields`), which the `use` alias shortens.

## Listing the areas

```bash
phpscript list --areas                 # Every area, in install order
phpscript list --areas 'Runtime\PX'    # The overlay only
phpscript list --areas --json          # The rows as data
phpscript list --bindings              # Every name, with the area that defines it
```

The Replaces column counts the names an area registers over. `stdlib.Areas` returns the same data to a Go program, and `stdlib.Owners` maps each name to its first definer.
