<?php
// test double for a class the library type-hints but composer does not install
class Swift { const VERSION = '6.2.0'; }

class Swift_Message
{
    public ?string $body = null;
    public function setBody($b, $mime = null) { $this->body = $b; return $this; }
    public function setDate($d) { return $this; }
    public function __toString(): string { return (string) $this->body; }
}
