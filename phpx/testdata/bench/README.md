# bench

A comparison of one workload written two ways: `flat.php` in phpscript's style (two unrelated classes), `oop.php` in phpx's (an interface, an abstract base, a trait, `extends`, `parent::`, `static::`). Both build 200,000 shapes and print the same sum, which php prints too. `./run.sh [binary]` takes the best of three.

| Run (2026-11-15, `-tags phpx`, interpreter) | php | phpx |
|---|---:|---:|
| flat.php | 68 ms | 1171 ms |
| oop.php | 72 ms | 1425 ms |

The inheritance, trait and `static::` machinery costs phpx about 22% over the flat program. php pays about 6%. The gap to php is the interpreter's, not the fork's: the flat program is 17 times slower on phpx before any fork feature runs. The engine declines `extends` and traits, so `oop.php` has no bytecode path; that is the next lever.
