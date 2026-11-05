name: an interface that extends others holds their constants
description: >
  `CI::SATURDAY` and `C::SUNDAY` resolve through `interface CI extends A, B`.
  carbon's CarbonInterface takes its constants from five extended interfaces.
  Expected output is php 8.4's.
---
<?php
interface UnitValue { public const SATURDAY = 6; const SUNDAY = 0; }
interface Fmt { const ISO = 'c'; }
interface CI extends UnitValue, Fmt { const OWN = 1; }
class C implements CI { function d() { return static::SATURDAY . self::ISO . self::OWN; } }
echo CI::SATURDAY, CI::ISO, CI::OWN, (new C)->d(), C::SUNDAY, "\n";
---
6c16c10
