<?php

require __DIR__ . '/vendor/autoload.php';

use Psr\Http\Message\ResponseInterface as Response;
use Psr\Http\Message\ServerRequestInterface as Request;
use Slim\Factory\AppFactory;
use Slim\Psr7\Factory\ServerRequestFactory;

$app = AppFactory::create();
$app->get('/hello/{name}', function (Request $request, Response $response, array $args) {
	$response->getBody()->write('Hello, ' . $args['name']);
	return $response;
});

$request = (new ServerRequestFactory())->createServerRequest('GET', '/hello/world');
$response = $app->handle($request);
echo $response->getBody(), "\n";
