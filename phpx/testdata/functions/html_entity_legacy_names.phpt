name: html_entity_decode leaves a terminated name go only matched in part
description: >
  A name go does not know whole, but whose legacy prefix it does, is not an
  entity and php leaves it as written: "&parag;" stays, while "&fjlig;" decodes
  to "fj" and "&semi;" to a semicolon. Expected output is php 8.4's.
---
<?php
echo html_entity_decode('&parag;', ENT_QUOTES | ENT_HTML5, 'UTF-8'), "\n";
echo html_entity_decode('&fjlig;', ENT_QUOTES | ENT_HTML5, 'UTF-8'), "\n";
echo html_entity_decode('&semi;', ENT_QUOTES | ENT_HTML5, 'UTF-8'), "|\n";
echo html_entity_decode('&notit;', ENT_QUOTES | ENT_HTML5, 'UTF-8'), "\n";
var_dump(html_entity_decode('&copy;', ENT_QUOTES | ENT_HTML5, 'UTF-8'));
var_dump(html_entity_decode('&MadeUpEntity;', ENT_QUOTES | ENT_HTML5, 'UTF-8'));
---
&parag;
fj
;|
&notit;
string(2) "©"
string(14) "&MadeUpEntity;"
