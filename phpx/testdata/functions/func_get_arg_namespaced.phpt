name: func_get_arg reads the call's arguments from a namespaced file
description: >
  The bytecode engine has no func_get_arg, so the fixture opts it out.
runner:
  flatstack: false
---
<?php
namespace Pkg;
function va() { return [func_get_arg(0), func_get_arg(1), func_num_args()]; }
echo json_encode(va('x', 'y', 'z')), "\n";
$f = function () { return func_get_arg(1); };
echo $f('p', 'q'), "\n";
---\n["x","y",3]
q\n