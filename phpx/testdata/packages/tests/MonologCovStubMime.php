<?php
// test double for a class the library type-hints but composer does not install
namespace Symfony\Component\Mime;

class Email
{
    public ?string $subject = null;
    public ?string $html = null;
    public ?string $text = null;
    public ?string $htmlCharset = null;
    public ?string $textCharset = null;
    public $date = null;
    public function getSubject(): ?string { return $this->subject; }
    public function subject(string $s): static { $this->subject = $s; return $this; }
    public function html($body, ?string $charset = 'utf-8'): static { $this->html = $body; return $this; }
    public function text($body, ?string $charset = 'utf-8'): static { $this->text = $body; return $this; }
    public function getHtmlCharset(): ?string { return $this->htmlCharset; }
    public function getTextCharset(): ?string { return $this->textCharset; }
    public function date(\DateTimeInterface $d): static { $this->date = $d; return $this; }
}
