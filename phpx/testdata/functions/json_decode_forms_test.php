<?php
echo json_encode(json_decode('{"a":{"b":[1,2,{"c":null}]},"d":1.5,"e":"xé😀","f":12345678901234567890}')), "\n";
var_dump(json_decode('{"a":{"b":[1]}}', true), json_decode('[1,2', true), json_last_error_msg(), json_decode(''), json_last_error(), json_decode('nul'), json_decode(' true '), json_decode('"\x01"'), json_last_error(), json_decode("\"\xff\""), json_last_error());
var_dump(json_decode('[[1]]', true, 1), json_last_error_msg(), json_decode('12345678901234567890', false, 512, JSON_BIGINT_AS_STRING), json_decode('{"0":1,"b":2}', true));
echo json_encode(json_decode('{"0":1}', false, 512, JSON_OBJECT_AS_ARRAY)), "\n";
try { json_decode('{', false, 512, JSON_THROW_ON_ERROR); } catch (JsonException $e) { echo $e->getMessage(), $e->getCode(), "\n"; }
var_dump(json_decode('1e2'), json_decode('-0'), json_decode('0.1'), json_decode('01'), json_decode('[1,]'), json_decode("{'a':1}"));
