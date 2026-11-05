<?php
// Test 4: boot the framework the way artisan does and report how far it got.
// Run from the app root; each stage prints one line, the last line is where it stopped.
//   php phpx_boot.php   |   phpx run phpx_boot.php
putenv('APP_KEY=base64:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=');
putenv('APP_ENV=testing');
putenv('DB_CONNECTION=sqlite');
putenv('DB_DATABASE=:memory:');
putenv('CACHE_STORE=array');
putenv('SESSION_DRIVER=array');
putenv('QUEUE_CONNECTION=sync');
$_ENV['APP_KEY'] = 'base64:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=';
$_ENV['APP_ENV'] = 'testing';
$_ENV['DB_CONNECTION'] = 'sqlite';
$_ENV['DB_DATABASE'] = ':memory:';
$_ENV['CACHE_STORE'] = 'array';
$_ENV['SESSION_DRIVER'] = 'array';
echo "1 autoload\n";
require 'vendor/autoload.php';
echo "2 application\n";
$app = require 'bootstrap/app.php';
echo "3 kernel\n";
$kernel = $app->make(Illuminate\Contracts\Console\Kernel::class);
echo "4 bootstrap (env, config, providers, boot)\n";
$kernel->bootstrap();
echo "5 container: ", get_class($app->make('config')), "\n";
echo "6 version: ", $app->version(), "\n";
echo "7 router: ", count($app->make('router')->getRoutes()), " routes\n";
echo "8 request\n";
$response = $app->make(Illuminate\Contracts\Http\Kernel::class)->handle(Illuminate\Http\Request::create('/', 'GET'));
echo "9 response: ", $response->getStatusCode(), "\n";
