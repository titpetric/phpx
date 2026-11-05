<?php

use GuzzleHttp\Handler\CurlFactory;
use GuzzleHttp\Psr7\Request;
use PHPUnit\Framework\TestCase;

/**
 * Helpers for the GuzzleCurl factory tests: build an EasyHandle offline.
 */
abstract class GuzzleCurlFactoryBase extends TestCase
{
    /** @var list<string> */
    protected array $deprecations = [];
    /** @var list<string> */
    protected array $tmp = [];

    protected function tearDown(): void
    {
        foreach ($this->tmp as $p) {
            if (is_link($p) || is_file($p)) {
                unlink($p);
            } elseif (is_dir($p)) {
                rmdir($p);
            }
        }
        $this->tmp = [];
    }

    protected function tmpFile(string $suffix = '.pem', string $data = 'x'): string
    {
        $p = __DIR__.'/gcurl2-'.getmypid().'-'.count($this->tmp).$suffix;
        file_put_contents($p, $data);
        $this->tmp[] = $p;

        return $p;
    }

    /** @return array{0: ?\GuzzleHttp\Handler\EasyHandle, 1: ?\Throwable} */
    protected function make(array $options = [], array $req = [], ?CurlFactory $f = null): array
    {
        $this->deprecations = [];
        set_error_handler(function ($no, $msg) {
            $this->deprecations[] = $msg;

            return true;
        }, E_USER_DEPRECATED);
        try {
            $request = new Request(
                $req['method'] ?? 'GET',
                $req['uri'] ?? 'http://example.test/path',
                $req['headers'] ?? [],
                $req['body'] ?? null,
                $req['version'] ?? '1.1'
            );

            return [($f ?? new CurlFactory(3))->create($request, $options), null];
        } catch (\Throwable $e) {
            return [null, $e];
        } finally {
            restore_error_handler();
        }
    }

    protected function ok(array $options = [], array $req = [], ?CurlFactory $f = null): \GuzzleHttp\Handler\EasyHandle
    {
        [$easy, $e] = $this->make($options, $req, $f);
        $this->assertNull($e, $e ? get_class($e).': '.$e->getMessage() : '');

        return $easy;
    }

    protected function fails(string $class, string $needle, array $options = [], array $req = [], ?CurlFactory $f = null): \Throwable
    {
        [$easy, $e] = $this->make($options, $req, $f);
        $this->assertNotNull($e, 'expected '.$class.' with '.$needle);
        $this->assertInstanceOf($class, $e, $e->getMessage());
        $this->assertStringContainsString($needle, $e->getMessage());

        return $e;
    }
}
