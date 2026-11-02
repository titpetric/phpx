<?php

require __DIR__ . '/vendor/autoload.php';

$latte = new Latte\Engine();
$latte->setLoader(new Latte\Loaders\StringLoader());
echo $latte->renderToString('Hello {$name}!', ['name' => 'world']), "\n";
