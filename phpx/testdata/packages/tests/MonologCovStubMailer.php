<?php
// test double for a class the library type-hints but composer does not install
namespace Symfony\Component\Mailer;

interface MailerInterface
{
    public function send($message);
}
