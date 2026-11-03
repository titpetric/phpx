name: late static binding names the called class
description: >
  static:: resolves to the class a call was made on, not the class that wrote
  the method body: new static, static::class, static::CONST, get_called_class(),
  forwarding through self:: and static::, a closure created in a static method,
  and a trait method. new self still names the declaring class. The expected
  section is what `php` printed.
---
<?php
class Model {
    public static function make() { return new static(); }
    public static function who() { return static::class; }
    public static function viaSelf() { return self::who(); }
    public static function viaStatic() { return static::who(); }
    public function name() { return static::class . '/' . self::class . '/' . get_called_class(); }
    public static function fresh() { return new self(); }
    public static function later() { return function () { return static::class; }; }
    public static function limit() { return static::LIMIT; }
    const LIMIT = 1;
}
class User extends Model { const LIMIT = 5; }
class Admin extends User {}
echo get_class(User::make()), "\n";
echo get_class(Admin::make()), "\n";
echo Admin::who(), "|", Admin::viaSelf(), "|", Admin::viaStatic(), "|", Model::who(), "\n";
echo (new Admin)->name(), "\n";
echo get_class(Admin::fresh()), "\n";
echo (Admin::later())(), "\n";
echo User::limit(), Admin::limit(), Model::limit(), "\n";
trait Builds { public static function build(): static { return new static(); } }
class Widget { use Builds; }
class Button extends Widget {}
echo get_class(Button::build()), "\n";
---
User
Admin
Admin|Admin|Admin|Model
Admin/Model/Admin
Model
Admin
551
Button
