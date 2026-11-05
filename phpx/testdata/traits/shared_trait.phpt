name: a trait reached by two paths composes once
description: >
  A class uses T directly and also U, which uses T. The method T declares is one method, not a collision. Livewire components that use WithFileUploads and a Filament trait that uses it too hit this. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
trait T { public function hello() { return 'T::hello'; } protected function guard() { return 'guarded'; } }
trait U { use T; public function other() { return 'U::other'; } }
class C { use T, U; public function both() { return $this->hello() . '/' . $this->other() . '/' . $this->guard(); } }
echo (new C)->both(), "\n";
class D { use U, T; }
echo (new D)->hello(), "\n";
---
T::hello/U::other/guarded
T::hello
