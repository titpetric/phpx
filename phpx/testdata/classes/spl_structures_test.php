<?php
$s = new SplObjectStorage;
$a = new stdClass; $b = new stdClass; $b->n = 2;
$s->attach($a, 'first');
$s->attach($b);
$s[$b] = 'second';
echo count($s), $s->contains($a) ? 'Y' : 'N', $s[$a], $s[$b], "\n";
$s->detach($a);
echo count($s), $s->contains($a) ? 'Y' : 'N', "\n";
foreach ($s as $i => $o) { echo $i, get_class($o), $o->n, ' '; }
echo "\n";

$st = new SplStack;
$st->push(1); $st->push(2); $st->push(3);
echo $st->pop(), $st->top(), count($st), "\n";
foreach ($st as $v) { echo $v; }
echo "\n";
$q = new SplQueue;
$q->enqueue('a'); $q->enqueue('b'); $q->enqueue('c');
echo $q->dequeue(), count($q), $q->isEmpty() ? 'E' : 'F', "\n";
foreach ($q as $k => $v) { echo $k, $v, ' '; }
echo "\n";
try { (new SplStack)->pop(); } catch (RuntimeException $e) { echo $e->getMessage(), "\n"; }

$ao = new ArrayObject(['x' => 1, 'y' => 2]);
$ao['z'] = 3;
echo count($ao), isset($ao['x']) ? 'Y' : 'N', $ao['z'], "\n";
foreach ($ao as $k => $v) { echo $k, $v; }
echo "\n";
unset($ao['x']);
echo implode(',', array_keys($ao->getArrayCopy())), "\n";

$it = new ArrayIterator([10, 20, 30]);
foreach ($it as $k => $v) { echo $k, ':', $v, ' '; }
echo count($it), "\n";

$fa = new SplFixedArray(3);
$fa[0] = 'p'; $fa[2] = 'r';
echo $fa->getSize(), implode(',', array_map(fn($x) => $x ?? '-', $fa->toArray())), "\n";
try { $fa[3] = 'x'; } catch (RuntimeException $e) { echo $e->getMessage(), "\n"; }

class Subj implements SplSubject {
    private $obs;
    public $state = 0;
    public function __construct() { $this->obs = new SplObjectStorage; }
    public function attach(SplObserver $o): void { $this->obs->attach($o); }
    public function detach(SplObserver $o): void { $this->obs->detach($o); }
    public function notify(): void { foreach ($this->obs as $o) { $o->update($this); } }
}
class Obs implements SplObserver {
    public function update(SplSubject $s): void { echo 'saw ', $s->state, "\n"; }
}
$subj = new Subj; $subj->attach(new Obs); $subj->state = 7; $subj->notify();
