name: call_user_func of a class-scoped name runs on the calling instance
description: >
  php binds the calling scope to `call_user_func([self::class, 'method'])`, so a
  protected method reads the receiver's properties, and a non-static method
  named from outside the class is not callable at all. CarbonPeriod keeps its
  filters as class-name arrays and reads `$this->options` in one of them, so a
  call that built a fresh empty instance ended urlhub's chart render on
  `Unknown getter 'options'`. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
class Scoped {
    protected ?int $options = 7;
    protected function read() { return $this->options; }
    public function run() { return call_user_func([self::class, 'read']); }
    public static function stat() { return 'static'; }
}
echo (new Scoped)->run(), "\n";
var_dump(is_callable([Scoped::class, 'read']));
var_dump(is_callable([Scoped::class, 'stat']));
echo call_user_func([Scoped::class, 'stat']), "\n";
---
7
bool(false)
bool(true)
static
