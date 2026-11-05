<?php
// test double for a class the library type-hints but composer does not install
namespace Gelf;

interface PublisherInterface { public function publish(Message $message); }

class Message
{
    public array $data = [];
    public array $additional = [];
    public function setTimestamp($t) { $this->data["timestamp"] = $t; return $this; }
    public function setShortMessage($m) { $this->data["short"] = $m; return $this; }
    public function setHost($h) { $this->data["host"] = $h; return $this; }
    public function setLevel($l) { $this->data["level"] = $l; return $this; }
    public function setAdditional($k, $v) { $this->additional[$k] = $v; return $this; }
    public function hasAdditional($k) { return isset($this->additional[$k]); }
    public function toArray() { return $this->data + ["additional" => $this->additional]; }
}
