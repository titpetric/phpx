name: an inaccessible method goes to __call and __callStatic
description: >
  A protected or private method called from outside its scope is sent to __call when the class declares one, which is how Laravel's Router answers Route::prefix(). Without __call php raises the visibility error. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
class Router {
    protected function prefix($p) { return "protected:$p"; }
    private function name($n) { return "private:$n"; }
    public function __call($m, $a) { return "__call:$m(" . implode(',', $a) . ")"; }
    protected static function sfn($x) { return "s:$x"; }
    public static function __callStatic($m, $a) { return "__callStatic:$m(" . implode(',', $a) . ")"; }
    public function inside() { return $this->prefix('a') . ' ' . static::sfn('b'); }
}
$r = new Router;
echo $r->prefix('x'), "\n";
echo $r->name('y'), "\n";
echo Router::sfn('z'), "\n";
echo $r->inside(), "\n";
class Plain { protected function p() {} }
try { (new Plain)->p(); } catch (Error $e) { echo $e->getMessage(), "\n"; }
---
__call:prefix(x)
__call:name(y)
__callStatic:sfn(z)
protected:a s:b
Call to protected method Plain::p() from global scope
