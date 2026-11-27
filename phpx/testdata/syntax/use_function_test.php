<?php
namespace Demo;
function version() { return 'helper'; }
use function Demo\version;
final class Version { }
var_dump(class_exists('Demo\Version'));
var_dump(Version::class);
var_dump(version());
