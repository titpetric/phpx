<?php
// Router for `php -S` used by the GuzzleCurl* tests. Offline, deterministic.
$uri = $_SERVER['REQUEST_URI'];
$path = parse_url($uri, PHP_URL_PATH);
$q = $_GET;
$method = $_SERVER['REQUEST_METHOD'];

function gc_echo(array $extra = []): void
{
    header('Content-Type: application/json');
    echo json_encode($extra + [
        'method' => $_SERVER['REQUEST_METHOD'],
        'path' => parse_url($_SERVER['REQUEST_URI'], PHP_URL_PATH),
        'query' => $_SERVER['QUERY_STRING'] ?? '',
        'body' => file_get_contents('php://input'),
        'headers' => getallheaders(),
        'cookies' => $_COOKIE,
        'post' => $_POST,
        'files' => array_map(fn ($f) => $f['name'], $_FILES),
    ]);
}

switch (true) {
    case $path === '/echo':
        gc_echo();
        break;
    case $path === '/status':
        http_response_code((int) ($q['code'] ?? 200));
        header('X-Status: yes');
        echo 'status '.($q['code'] ?? 200);
        break;
    case $path === '/redirect':
        // /redirect?code=301&to=/echo&n=3 -> chain of n hops
        $n = (int) ($q['n'] ?? 1);
        $code = (int) ($q['code'] ?? 302);
        if ($n > 1) {
            $to = '/redirect?code='.$code.'&n='.($n - 1).'&to='.urlencode($q['to'] ?? '/echo');
        } else {
            $to = $q['to'] ?? '/echo';
        }
        header('Location: '.$to, true, $code);
        echo 'moved';
        break;
    case $path === '/loop':
        header('Location: /loop', true, 302);
        break;
    case $path === '/gzip':
        header('Content-Encoding: gzip');
        header('Content-Type: text/plain');
        echo gzencode('zipped body');
        break;
    case $path === '/deflate':
        header('Content-Encoding: deflate');
        echo gzcompress('deflated body');
        break;
    case $path === '/slow':
        usleep((int) ($q['us'] ?? 1500000));
        echo 'slow';
        break;
    case $path === '/large':
        header('Content-Type: application/octet-stream');
        $n = (int) ($q['n'] ?? 300000);
        echo str_repeat('0123456789abcdef', intdiv($n, 16));
        break;
    case $path === '/chunked':
        header('Content-Type: text/plain');
        header('Transfer-Encoding: chunked');
        for ($i = 0; $i < 5; $i++) {
            $c = "chunk$i\n";
            echo dechex(strlen($c))."\r\n$c\r\n";
            flush();
        }
        echo "0\r\n\r\n";
        break;
    case $path === '/trailers':
        header('Transfer-Encoding: chunked');
        header('Trailer: X-Trailer-One, X-Trailer-Two');
        header('Content-Type: text/plain');
        echo "5\r\nhello\r\n0\r\nX-Trailer-One: one\r\nX-Trailer-Two: two\r\n\r\n";
        break;
    case $path === '/ctl':
        header("X-Ctl: a\x01b");
        echo 'ctl';
        break;
    case $path === '/close':
        // announce more than we send, then drop the connection
        header('Content-Length: 1000');
        echo 'partial';
        flush();
        exit;
    case $path === '/cookie':
        setcookie('a', '1');
        setcookie('b', '2', ['path' => '/']);
        echo 'cookies set';
        break;
    case $path === '/multi':
        header('X-Dup: one', false);
        header('X-Dup: two', false);
        header('X-Dup: three', false);
        echo 'multi';
        break;
    case $path === '/head':
        header('Content-Length: 5');
        header('X-Head: 1');
        if ($method !== 'HEAD') {
            echo 'hello';
        }
        break;
    case $path === '/auth':
        $u = $_SERVER['PHP_AUTH_USER'] ?? null;
        if ($u === null) {
            header('WWW-Authenticate: Basic realm="gc"');
            http_response_code(401);
            echo 'denied';
        } else {
            echo $u.':'.$_SERVER['PHP_AUTH_PW'];
        }
        break;
    case $path === '/empty':
        http_response_code(204);
        break;
    case $path === '/text':
        header('Content-Type: text/plain');
        echo (string) ($q['t'] ?? 'text');
        break;
    default:
        http_response_code(404);
        echo 'not found';
}
