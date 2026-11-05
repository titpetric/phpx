<?php
require_once 'late_trait_support.php';
class Late { use LateGreeter; }
echo (new Late)->greet(), "\n";
