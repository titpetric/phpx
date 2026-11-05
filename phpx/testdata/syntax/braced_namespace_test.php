<?php
namespace Lib\Ex {
    class RuntimeException extends \Exception {}
    function helper() { return __FUNCTION__; }
    const LEVEL = 3;
}
namespace Lib {
    use Lib\Ex\RuntimeException;
    class Svc {
        public function run() {
            try { throw new \RuntimeException('global'); }
            catch (RuntimeException $e) { return 'wrong'; }
            catch (\RuntimeException $e) { return 'right:' . __CLASS__ . ':' . __NAMESPACE__; }
        }
    }
}
namespace {
    echo (new Lib\Svc)->run(), "\n";
    echo Lib\Ex\helper(), " ", Lib\Ex\LEVEL, " ", get_class(new Lib\Ex\RuntimeException), "\n";
    foreach ([[1, 2], [3, 4]] as $row) { array_shift($row); echo count($row); }
    echo "\n";
}
