<?php

class A {
	public $p = 5;

	function a() {
		return array('k' => 0, 'n' => null);
	}

	function o() {
		return new A;
	}

	function v() {
		return null;
	}
}
$x = new A;

var_dump(($__phpx_c1 = $x->a()['k']) !== null ? $__phpx_c1 : 'd');
var_dump(($__phpx_c2 = $x->a()['n']) !== null ? $__phpx_c2 : 'd');
var_dump(($__phpx_c3 = $x->a()['zz']) !== null ? $__phpx_c3 : 'd');
var_dump(($__phpx_c4 = $x->v()) !== null ? $__phpx_c4 : 'nul');
var_dump(($__phpx_c5 = $x->o()->p) !== null ? $__phpx_c5 : 'd');
var_dump(($__phpx_c6 = $x->o()->q) !== null ? $__phpx_c6 : 'd');
var_dump(($__phpx_c7 = strlen('ab')) !== null ? $__phpx_c7 : 9);
var_dump(($__phpx_c8 = $x->v()['a']['b']) !== null ? $__phpx_c8 : 'deep');
