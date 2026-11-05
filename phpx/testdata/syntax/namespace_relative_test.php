<?php

namespace Lib\Util;

function twice(int $x): int
{
    return $x * 2;
}

function run(int $x): int
{
    return namespace\twice($x) + namespace\ONE;
}

const ONE = 1;

class K
{
    public static function make(): string
    {
        return namespace\K::class;
    }
}

echo run(20), "\n";
echo K::make(), "\n";
echo \Lib\Util\twice(4), "\n";
