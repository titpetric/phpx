<?php
NAMESPACE KeywordCase;

CONST TOP = "const";

FUNCTION greet($name) {
    GLOBAL $prefix;
    RETURN $prefix . $name;
}

TRAIT Says {
    PUBLIC FUNCTION who() {
        RETURN "trait";
    }
}

INTERFACE Contract {
    CONST LIMIT = 3;
    PUBLIC FUNCTION id();
}

ABSTRACT CLASS Base IMPLEMENTS Contract {
    USE Says;
    PROTECTED $id = 1;
    PUBLIC FUNCTION id() {
        RETURN $this->id;
    }
}

FINAL CLASS Child EXTENDS Base {
    PUBLIC STATIC $made = 0;
    PUBLIC FUNCTION __construct() {
        STATIC::$made++;
    }
    PUBLIC FUNCTION kind() {
        RETURN STATIC::CLASS;
    }
}

ENUM Suit: string {
    CASE Hearts = 'H';
    CASE Spades = 'S';
}

$prefix = "hi ";
$c = NEW Child();
ECHO greet("x"), "\n";
ECHO $c->who(), "\n";
ECHO $c->id(), "\n";
ECHO Child::LIMIT, "\n";
ECHO TOP, "\n";
ECHO Suit::Hearts->value, "\n";
ECHO $c::CLASS, "\n";
ECHO $c->kind(), "\n";
IF ($c INSTANCEOF Base) { ECHO "is base\n"; } ELSE { ECHO "no\n"; }
FOREACH ([1, 2] AS $v) { ECHO $v, " "; }
ECHO "\n";
SWITCH ($c->id()) { CASE 1: ECHO "one\n"; BREAK; DEFAULT: ECHO "other\n"; }
TRY {
    THROW NEW \RuntimeException("boom");
} CATCH (\RuntimeException $e) {
    ECHO $e->getMessage(), "\n";
}
