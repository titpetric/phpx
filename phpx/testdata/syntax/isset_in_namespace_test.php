<?php
namespace N;
class T {
    private $b = '';
    private function empty(): string { $c = substr($this->b, -2); if (!isset($c[0])) { return 'empty'; } return 'non'; }
    public function f() { $c = ''; $d = 'ab'; $a = ['k' => null, 'z' => 0]; return [$this->empty(), isset($c[0]), isset($d[1]), isset($d[2]), empty($c[0]), empty($a['k']), empty($a['z']), isset($a['k']), empty($a['none'])]; }
}
echo json_encode((new T)->f()), "\n";
