name: json_decode answers stdClass, the class a JSON object names
description: >
  PHP's default shape for a JSON object is an stdClass instance, and framework
  code reads a decoded body with `->name`. The expected section is what `php`
  printed. `$associative = true` still asks for an array, and false, null and an
  omitted argument all answer the object shape; a nested object takes the same
  shape at any depth, the document's order survives a decode and re-encode, and
  a decoded object is an ordinary property bag: it answers instanceof stdClass,
  takes a dynamic property, unsets one, foreachs and casts to an array in the
  order the document named them. The two limits the decoder keeps are
  upstream's: an invalid document raises where php answers null, and the
  `$depth` and `$flags` arguments are accepted and ignored.
---
<?php

$body = '{"name":"phpscript","count":2,"tags":["a","b"],"child":{"deep":true},"empty":{}}';
$o = json_decode($body);

echo get_class($o), "\n";
echo $o->name, ":", $o->count, "\n";
echo $o instanceof stdClass ? "stdClass" : "other", "\n";
echo get_class($o->child), " ", $o->child->deep ? "deep" : "shallow", "\n";
echo get_class($o->empty), " ", count(get_object_vars($o->empty)), "\n";
echo implode(",", array_keys(get_object_vars($o))), "\n";

foreach ($o as $key => $value) {
	echo "  ", $key, "\n";
}

// A decoded object is a property bag like any other.
$o->added = "by the script";
unset($o->count);
echo json_encode($o), "\n";

// $associative selects the array shape instead.
$arr = json_decode($body, true);
echo gettype($arr), " ", gettype($arr["child"]), " ", $arr["tags"][1], "\n";

// false and null ask for objects, and an omitted argument does too.
echo get_class(json_decode($body, false)), " ", get_class(json_decode($body, null)), "\n";

// The document's order is preserved, so a round trip reads as it arrived.
echo json_encode(json_decode('{"z":1,"a":2,"m":3}')), "\n";

// A list decodes as a list, and an object inside it is an object.
$rows = json_decode('[{"c":3},{"c":4}]');
echo gettype($rows), " ", get_class($rows[0]), " ", $rows[0]->c, "\n";
echo json_encode($rows), "\n";

// The properties cast back out in the order they were set.
$back = (array) json_decode('{"b":1,"a":2}');
echo implode(",", array_keys($back)), "\n";
---
stdClass
phpscript:2
stdClass
stdClass deep
stdClass 0
name,count,tags,child,empty
  name
  count
  tags
  child
  empty
{"name":"phpscript","tags":["a","b"],"child":{"deep":true},"empty":{},"added":"by the script"}
array array b
stdClass stdClass
{"z":1,"a":2,"m":3}
array stdClass 3
[{"c":3},{"c":4}]
b,a
