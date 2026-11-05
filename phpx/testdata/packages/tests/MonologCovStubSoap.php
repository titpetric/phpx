<?php
// test double for a class the library special-cases but the php build does not provide (ext-soap)
class SoapFault extends Exception
{
    public $faultcode = null;
    public $faultactor = null;
    public $detail = null;

    public function __construct($code = null, $string = '', $actor = null, $detail = null)
    {
        parent::__construct((string) $string);
        $this->faultcode = $code;
        $this->faultactor = $actor;
        $this->detail = $detail;
    }
}
