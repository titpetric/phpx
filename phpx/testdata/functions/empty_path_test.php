<?php

// An empty filename names nothing: every path question answers false, where a
// cleaned path would name the current directory.
var_dump(
    file_exists(''),
    is_file(''),
    is_dir(''),
    is_readable(''),
    is_writable(''),
    is_executable(''),
);
