<?php
namespace T;
interface Skipped extends \Throwable {}
class AssertionFailed extends \Exception {}
final class ExpectationFailed extends AssertionFailed {}
final class SkippedWith extends AssertionFailed implements Skipped {}
final class Wrapped extends \RuntimeException implements Skipped {}
function run(callable $f) {
    try { $f(); } catch (\Throwable $t) {
        echo get_class($t), ':', $t instanceof AssertionFailed ? 'AF' : '-', $t instanceof Skipped ? 'S' : '-', $t instanceof \Exception ? 'E' : '-', $t instanceof \Throwable ? 'T' : '-', $t instanceof \RuntimeException ? 'R' : '-', "\n";
    }
}
run(fn() => throw new ExpectationFailed('x'));
run(fn() => throw new SkippedWith('x'));
run(fn() => throw new Wrapped('x'));
run(fn() => throw new \InvalidArgumentException('x'));
try { throw new ExpectationFailed('m'); } catch (AssertionFailed $e) { echo "caught AF\n"; }
try { throw new SkippedWith('m'); } catch (Skipped $e) { echo "caught Skipped\n"; }
try { throw new ExpectationFailed('m'); } catch (Skipped | AssertionFailed $e) { echo "caught multi\n"; }
