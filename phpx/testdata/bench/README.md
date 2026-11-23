# bench

A comparison of one workload written two ways: `flat.php` in phpscript's style (two unrelated classes), `oop.php` in phpx's (an interface, an abstract base, a trait, `extends`, `parent::`, `static::`). Both build 200,000 shapes and print the same sum, which php prints too. `./run.sh [binary]` takes the best of three.

`./upstream.sh [phpx] [phpscript]` adds the runtime the fork came from. It builds
`upstream/main` from this checkout when no phpscript binary is given, and answers one question the php column
cannot: what the fork costs over phpscript on the flat program, and what it buys that phpscript refuses.

| Run (2026-11-23, `-tags phpx`, interpreter, best of 3) | php | phpscript | phpx |
|---|---:|---:|---:|
| `flat.php` | 67 ms | 794 ms | 1166 ms |
| `oop.php` | 69 ms | refused | 1405 ms |

phpscript refuses `oop.php` at line 4, `interface Shape { function area(): float; }`, with `expected "(", got
3("name")`. The OOP program is the one feature this table cannot time on upstream: there is no run to time. The
fork runs both and prints php's sum on both.

The fork is **47% slower than phpscript on the flat program**, which uses none of the fork's features. Two of the
changes since the last table are measured here: copy on write adds 3% (1,130 ms at `1c0d6cb8` to 1,163 ms) and the
destructor adds 2% (1,141 ms to 1,163 ms), the rest is the overlay set the fork carried before them. The
inheritance, trait and `static::` machinery costs phpx about 21% over its own flat program, where php pays 3%. The
engine declines `extends`, traits and a destructor, so `oop.php` has no bytecode path; that is the next lever, and
it would move the 1.4 s and not the 0.07 s.
