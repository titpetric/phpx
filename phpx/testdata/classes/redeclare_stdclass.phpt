name: stdClass cannot be declared by a script
description: >
  php owns the name stdClass and fatals `Cannot redeclare class stdClass`.
error: 'Cannot redeclare class stdClass'
runner:
  php: false
---
<?php

class stdClass {}

echo "unreachable\n";
---
Internal Server Error
