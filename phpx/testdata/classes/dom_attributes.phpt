name: a DOM element answers its attributes list, and XPath calls a php function
description: >
  `$node->attributes` is a map with a length and one DOMAttr per attribute in
  document order, and a selector may call a string function through
  `php:functionString`, which is how embed's Document picks its oembed link by
  a case-insensitive rel and type. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
$html = '<html><body><link rel="alternate" type="application/json+oembed" href="https://example.test/oembed?format=json&amp;url=x"></body></html>';
$doc = new DOMDocument('1.0', 'UTF-8');
@$doc->loadHTML($html);

$xpath = new DOMXPath($doc);
$xpath->registerNamespace('php', 'http://php.net/xpath');
$xpath->registerPhpFunctions();
$nodes = $xpath->query('.//link[php:functionString("strtolower", @rel)="alternate"][php:functionString("strtolower", @type)="application/json+oembed"]');
var_dump($nodes->length);

$link = $nodes->item(0);
var_dump($link->attributes->length);
for ($i = 0; $i < $link->attributes->length; $i++) {
    $attr = $link->attributes->item($i);
    echo $attr->name, '=', $attr->nodeValue, "\n";
}
echo $link->getAttribute('href'), "\n";

$e = $doc->createElement('p');
$e->setAttribute('id', 'x');
var_dump($e->attributes->length, $e->attributes->item(0)->name);
---
int(1)
int(3)
rel=alternate
type=application/json+oembed
href=https://example.test/oembed?format=json&url=x
https://example.test/oembed?format=json&url=x
int(1)
string(2) "id"
