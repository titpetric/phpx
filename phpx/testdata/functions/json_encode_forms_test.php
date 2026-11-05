<?php
$v = ["a" => "x/y é 😀 <b> & ' \" \n \x01", "n" => [1, 2.5, 1.0, 1e25, 0.00001, -0.0, 100000000000000000000.0, true, null], "o" => new stdClass, "e" => [], "l" => [3 => 1], 12 => ['k' => ['z' => []]]];
foreach ([0, JSON_UNESCAPED_SLASHES, JSON_UNESCAPED_UNICODE, JSON_PRETTY_PRINT, JSON_HEX_TAG | JSON_HEX_AMP | JSON_HEX_APOS | JSON_HEX_QUOT, JSON_FORCE_OBJECT, JSON_PRESERVE_ZERO_FRACTION, JSON_NUMERIC_CHECK] as $f) { echo json_encode($v, $f), "\n"; }
echo json_encode("123", JSON_NUMERIC_CHECK), json_encode(["1.5e2"], JSON_NUMERIC_CHECK), "\n";
var_dump(json_encode("\xB1\x31"), json_last_error(), json_last_error_msg());
var_dump(json_encode("\xB1\x31", JSON_INVALID_UTF8_SUBSTITUTE), json_encode("a\xB1b", JSON_INVALID_UTF8_IGNORE), json_encode("\xB1", JSON_PARTIAL_OUTPUT_ON_ERROR));
var_dump(json_encode(NAN), json_last_error_msg(), json_encode(NAN, JSON_PARTIAL_OUTPUT_ON_ERROR));
try { json_encode(NAN, JSON_THROW_ON_ERROR); } catch (JsonException $e) { echo get_class($e), $e->getMessage(), $e->getCode(), "\n"; }
class P implements JsonSerializable { public $a = 1; protected $b = 2; private $c = 3; function jsonSerialize(): mixed { return ['x' => $this->a]; } }
class Q { public $a = 1; protected $b = 2; public $n = null; }
enum S: string { case A = 'a'; }
echo json_encode([new P, new Q, S::A]), "\n";
