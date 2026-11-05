<?php
// test double for a class the library type-hints but composer does not install
namespace PhpConsole;

class Dumper
{
    public $levelLimit;
    public $itemsCountLimit;
    public $itemSizeLimit;
    public $dumpSizeLimit;
    public $detectCallbacks;
    public function dump($var) { return ['dumped' => $var]; }
}

class DebugDispatcher
{
    public $detectTraceAndSource = false;
    public array $debugs = [];
    public function dispatchDebug($message, $tags, $ignore) { $this->debugs[] = [$message, $tags, $ignore]; }
}

class ErrorsDispatcher
{
    public array $errors = [];
    public array $exceptions = [];
    public function dispatchError($code, $message, $file, $line, $ignore) { $this->errors[] = [$code, $message, $file, $line]; }
    public function dispatchException($e) { $this->exceptions[] = $e; }
}

class Connector
{
    public static ?Connector $instance = null;
    public static $storage = null;
    public bool $active = true;
    public array $calls = [];
    public Dumper $dumper;
    public DebugDispatcher $debug;
    public ErrorsDispatcher $errors;
    public function __construct() { $this->dumper = new Dumper(); $this->debug = new DebugDispatcher(); $this->errors = new ErrorsDispatcher(); }
    public static function getInstance() { return self::$instance ??= new Connector(); }
    public static function setPostponeStorage($s) { self::$storage = $s; }
    public function isActiveClient() { return $this->active; }
    public function setSourcesBasePath($p) { $this->calls[] = ['base', $p]; }
    public function setServerEncoding($e) { $this->calls[] = ['enc', $e]; }
    public function setPassword($p) { $this->calls[] = ['pw', $p]; }
    public function enableSslOnlyMode() { $this->calls[] = ['ssl']; }
    public function setAllowedIpMasks($m) { $this->calls[] = ['ips', $m]; }
    public function setHeadersLimit($l) { $this->calls[] = ['limit', $l]; }
    public function startEvalRequestsListener() { $this->calls[] = ['eval']; }
    public function getDebugDispatcher() { return $this->debug; }
    public function getErrorsDispatcher() { return $this->errors; }
    public function getDumper() { return $this->dumper; }
}

class Helper
{
    public static bool $registered = false;
    public static function isRegistered() { return self::$registered; }
    public static function register() { self::$registered = true; }
}

class Handler
{
    public static ?Handler $instance = null;
    public array $calls = [];
    public static function getInstance() { return self::$instance ??= new Handler(); }
    public function setHandleErrors($b) { $this->calls[] = ['errors', $b]; }
    public function setHandleExceptions($b) { $this->calls[] = ['exceptions', $b]; }
    public function start() { $this->calls[] = ['start']; }
}
