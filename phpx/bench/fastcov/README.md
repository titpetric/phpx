# fastcov

The commonmark suite, run on php and on phpx, compared one test at a time.

[turbocov](../turbocov/) is the php driver that collects line coverage inside
the engine. phpx cannot load an extension, so it records statement coverage in
its own interpreter; what both stacks share is the workload. This directory
runs that workload and answers two questions:

- **1-1 parity**: which test does php pass that phpx does not? `compare.py`
  reads a JUnit report from each stack and joins them by test ID.
- **the cost of one invocation**: what does a single test cost phpx, and what
  does that invocation allocate? `one_by_one.py` runs exactly one test per
  process on both stacks and records the wall time, PHPUnit's own test time,
  the outcome and, for phpx, the Go heap from its own `--memprofile`.

The two clocks answer different questions. The functional test tree takes phpx
about 30 seconds to load before the first test runs, while php loads it in a
fraction of a second; the process wall carries that start-up and PHPUnit's
`Time:` line does not. Allocations read the same way: the `--memprofile` heap
covers the whole process, so the start-up's share dominates a single test.
Record what a conversion allocates with a script of its own
(`--memprofile` around a loop) when the number has to describe the
interpreter and not PHPUnit's loader.

## Running

```sh
export COMMONMARK=/data/covbench/commonmark   # a checkout with vendor/
./run.sh unit                                  # php vs phpx, unit suite
./run.sh both                                  # both suites
```

`run.sh` builds phpx from this tree, writes a JUnit report per stack, and
writes `report.md`. Its exit status is 1 when a test passes on php and not on
phpx, so a pipeline can gate on parity.

One test per invocation, for the tests phpx did not pass:

```sh
python3 -I one_by_one.py --checkout "$COMMONMARK" \
    --from-junit /tmp/phpx.xml --phpx /tmp/phpx --shim /tmp/phpx-shim.sh \
    --out one-by-one.md --jsonl one-by-one.jsonl
```

Without `--from-junit`, `--tests FILE` takes one test ID a line in the
`Class::method with data set "..."` form the JUnit report writes. `--limit`
caps the run; a full suite is over three thousand processes and takes hours.

## What phpx needed for the suite

The suite is 4,447 tests: 3,228 unit and 1,219 functional. phpx passes all of
them; the functional suite reports 6 skipped, which it skips on php too.

The functional suite went from 1 error and 11 failures to none with these
fixes:

- **`strncmp` stops at its limit.** php answers 0 whenever the first `$length`
  bytes agree, whatever follows, and only a string shorter than `$length` lets
  the length break the tie. The Footnote renderer compares a destination
  against its prefix this way, so `#Fn:` was read as a mismatch and the
  lowercased branch ran.
- **A reference in an array literal is a slot, and a spread carries it.**
  `[$a, $b, &$m]` stores a cell, and `f(...$array)` hands a php function its
  by-reference parameter and a binding its setter. nette's `Strings::match`
  fills `$m` through `preg_match(...$args)`, and on its output the schema
  validation of the Mention and HeadingPermalink extensions turned a
  configuration error into a `TypeError`.
- **`html_entity_decode` reads go's partial match as no match.** Go matches a
  legacy name without its semicolon, so `&parag;` decoded as `&para` plus `g;`.
  A terminated name that does not decode whole stays as written, while
  `&fjlig;` and `&semi;` decode whole.
- **A pattern closes at its first unescaped delimiter, and its modifiers are
  checked.** `//x//i` is an empty body with the modifiers `x//i` and does not
  compile; an unknown modifier, a missing delimiter, an alphanumeric delimiter
  and an empty pattern answer false. The Mention extension's assertion relies
  on that.
- **XPath calls a php function, and an element answers its attributes.**
  `php:functionString("strtolower", @rel)` is how embed's Document selects its
  oembed link, and `$node->attributes` is a map with a length and a `DOMAttr`
  per attribute in document order. Both were missing, so the adapter fell
  through to a network request.

The unit suite went from 3,129/3,228 passing to 3,228/3,228 with these fixes:

- **A reference in an array element is a slot.** `$a[] = &$x` stores a cell,
  a by-reference `foreach` binds the loop variable to the element's cell and
  leaves it bound after the loop, and a value assigned to a referenced element
  writes through it. `array_pop` through a stored reference removes the value
  everywhere. This was 94 of the 101 unit failures.
- **Objects serialize.** `serialize()` mangles protected and private property
  keys the way php does, and `unserialize()` builds the instance without a
  constructor and honours `allowed_classes`. Two `#[RunInSeparateProcess]`
  tests needed it.
- **`clone` gives a copy its own arrays.** A reference element from a setter
  that has returned is copied as its value, so a write through the copy leaves
  the original alone.
- **Reflection declares properties.** `ReflectionProperty::getDeclaringClass()`
  answers the ancestor that declared an inherited property, and
  `getProperties()` lists a class's own properties before the inherited ones.
- **JUnit XML.** `DOMDocument::appendChild()` lets PHPUnit's JUnit logger build
  its tree, which is what makes the 1-1 comparison possible.

## Start-up and allocations

Once a program declares one by-reference parameter, every argument of every
call goes through `__phpx_refarg`. The helper was installed with `adapt()`, so
each argument paid a reflection call; on the functional tree's start-up that
was a third of the CPU. Installing it in the uniform call shape took the
no-test start-up from 44.3 s to 28.7 s on the bench host.

A conversion loop is the honest profile for allocation. Two hundred conversions
of a page of markdown, with the table extension, allocate about 511 MB: 18%
`phpxCollectDropped`, 8% `NewScope`, 8% `Scope.Set`, 8% the reference argument,
7% `model.NewObject` in total. The reference wrapper exists because a
by-reference program pays it for every variable argument; a program without a
by-reference parameter compiles the plain value.

## Measure it on a quiet host

The suite is start-up bound. On a shared host another run moves a single
test's wall time by more than the interpreter difference does; pin the CPU or
compare medians, not one sample. The heap profile is per invocation and does
not depend on the neighbour's load.
