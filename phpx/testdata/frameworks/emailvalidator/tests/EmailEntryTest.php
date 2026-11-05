<?php
use PHPUnit\Framework\TestCase;
use Egulias\EmailValidator\EmailValidator;
use Egulias\EmailValidator\Validation\RFCValidation;

final class EmailEntryTest extends TestCase
{
    public function testValid(): void
    {
        $v = new EmailValidator();
        $this->assertTrue($v->isValid('user@example.com', new RFCValidation()));
        $this->assertTrue($v->isValid('first.last+tag@sub.example.org', new RFCValidation()));
    }
    public function testInvalid(): void
    {
        $v = new EmailValidator();
        $this->assertFalse($v->isValid('not-an-email', new RFCValidation()));
        $this->assertFalse($v->isValid('a@', new RFCValidation()));
        $this->assertFalse($v->isValid('@example.com', new RFCValidation()));
        $this->assertFalse($v->isValid('a b@example.com', new RFCValidation()));
    }
}
