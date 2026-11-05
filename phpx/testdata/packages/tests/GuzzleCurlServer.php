<?php

/**
 * Starts `php -S` once per test class on a free port and stops it afterwards.
 */
final class GuzzleCurlServer
{
    /** @var resource|null */
    private $proc;
    public int $port = 0;

    public static function start(): self
    {
        $s = new self();
        for ($try = 0; $try < 20; $try++) {
            $port = self::freePort();
            $s->proc = proc_open(
                ['env', 'PHP_CLI_SERVER_WORKERS=16', 'php', '-S', "127.0.0.1:$port", self::router()],
                [0 => ['file', '/dev/null', 'r'], 1 => ['file', '/dev/null', 'w'], 2 => ['file', '/dev/null', 'w']],
                $pipes
            );
            for ($i = 0; $i < 100; $i++) {
                $c = @fsockopen('127.0.0.1', $port, $en, $es, 1);
                if ($c) {
                    fclose($c);
                    $s->port = $port;

                    return $s;
                }
                usleep(50000);
            }
            $s->stop();
        }
        throw new \RuntimeException('cannot start php -S');
    }

    private static function router(): string
    {
        // __DIR__ is root-relative under phpx; the tests run with the packages dir as cwd.
        foreach (['shims/guzzle-curl-router.php', __DIR__.'/../shims/guzzle-curl-router.php'] as $p) {
            if (is_file($p)) {
                return $p;
            }
        }
        throw new \RuntimeException('router not found');
    }

    private static function freePort(): int
    {
        for ($i = 0; $i < 200; $i++) {
            $port = random_int(20000, 60000);
            $c = @fsockopen('127.0.0.1', $port, $en, $es, 1);
            if (!$c) {
                return $port;
            }
            fclose($c);
        }
        throw new \RuntimeException('no free port');
    }

    public function url(string $path = ''): string
    {
        return 'http://127.0.0.1:'.$this->port.$path;
    }

    public function stop(): void
    {
        if ($this->proc !== null) {
            proc_terminate($this->proc);
            proc_close($this->proc);
            $this->proc = null;
        }
    }
}

/**
 * Per-class server lifecycle for the GuzzleCurl* tests.
 */
abstract class GuzzleCurlBase extends \PHPUnit\Framework\TestCase
{
    protected static ?GuzzleCurlServer $srv = null;

    public static function setUpBeforeClass(): void
    {
        self::$srv = GuzzleCurlServer::start();
    }

    public static function tearDownAfterClass(): void
    {
        if (self::$srv !== null) {
            self::$srv->stop();
            self::$srv = null;
        }
    }

    protected static function url(string $path = ''): string
    {
        return self::$srv->url($path);
    }
}
