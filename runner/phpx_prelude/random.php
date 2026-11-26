<?php
namespace Random;

interface Engine {
    public function generate(): string;
}

class Secure implements Engine {
    public function generate(): string { return random_bytes(8); }
}

class Mt19937 implements Engine {
    public function __construct($seed = null) {}
    public function generate(): string { return random_bytes(8); }
}

class Xoshiro256StarStar implements Engine {
    public function __construct($seed = null) {}
    public function generate(): string { return random_bytes(8); }
}

class PcgOneseq128XslRr64 implements Engine {
    public function __construct($seed = null) {}
    public function generate(): string { return random_bytes(8); }
}

class Randomizer {
    public Engine $engine;
    public function __construct($engine = null) {
        $this->engine = $engine instanceof Engine ? $engine : new Secure;
    }
    public function getBytes(int $length): string {
        return random_bytes($length);
    }
    public function getBytesFromString(string $string, int $length): string {
        $size = strlen($string);
        if ($size === 0 || $length <= 0) {
            return '';
        }
        $out = '';
        for ($i = 0; $i < $length; $i++) {
            $out .= $string[random_int(0, $size - 1)];
        }
        return $out;
    }
    public function getInt(int $min, int $max): int {
        return random_int($min, $max);
    }
    public function nextInt(): int {
        return random_int(0, PHP_INT_MAX);
    }
    public function getFloat(float $min = 0.0, float $max = 1.0, $boundary = null): float {
        return $min + (random_int(0, PHP_INT_MAX) / (float) PHP_INT_MAX) * ($max - $min);
    }
    public function shuffleArray(array $array): array {
        $keys = array_keys($array);
        shuffle($keys);
        $out = [];
        foreach ($keys as $key) {
            $out[$key] = $array[$key];
        }
        return $out;
    }
    public function shuffleBytes(string $bytes): string {
        $parts = str_split($bytes);
        shuffle($parts);
        return implode('', $parts);
    }
    public function pickArrayKeys(array $array, int $num): array {
        $keys = array_keys($array);
        shuffle($keys);
        return array_slice($keys, 0, $num);
    }
}
