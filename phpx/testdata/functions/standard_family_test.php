<?php
function t($label, $f) { try { echo $label, ': ', json_encode($f()), "\n"; } catch (Throwable $e) { echo $label, ': ', get_class($e), "\n"; } }
t('gmdate', fn() => gmdate('D, d M Y H:i:s \G\M\T c r O U e T P N jS z t L o y a A g G h B', 86400 * 365 + 3725));
t('date letters', fn() => [gmdate('l jS \o\f F Y', 0), gmdate('W', 1704067200), gmdate('S', 86400 * 1), gmdate('S', 86400 * 2), gmdate('S', 86400 * 3), gmdate('S', 86400 * 10)]);
t('filter', fn() => [filter_var('1.2.3.4', FILTER_VALIDATE_IP), filter_var('x', FILTER_VALIDATE_IP), filter_var('::1', FILTER_VALIDATE_IP, FILTER_FLAG_IPV6), filter_var('1.2.3.4', FILTER_VALIDATE_IP, FILTER_FLAG_IPV6), filter_var('10.0.0.1', FILTER_VALIDATE_IP, FILTER_FLAG_NO_PRIV_RANGE), filter_var('12', FILTER_VALIDATE_INT), filter_var('012', FILTER_VALIDATE_INT), filter_var('5', FILTER_VALIDATE_INT, ['options' => ['min_range' => 6]]), filter_var('a@b.co', FILTER_VALIDATE_EMAIL), filter_var('nope', FILTER_VALIDATE_EMAIL), filter_var('yes', FILTER_VALIDATE_BOOLEAN), filter_var('maybe', FILTER_VALIDATE_BOOLEAN, FILTER_NULL_ON_FAILURE), filter_var('1.5', FILTER_VALIDATE_FLOAT), filter_var('http://x.y/z', FILTER_VALIDATE_URL), filter_var('x.y', FILTER_VALIDATE_URL)]);
t('consts', fn() => [IDNA_DEFAULT, STREAM_CRYPTO_METHOD_TLSv1_2_CLIENT, STREAM_NOTIFY_CONNECT, INTL_IDNA_VARIANT_UTS46]);
t('inet', fn() => [inet_ntop("\x7f\0\0\1"), bin2hex(inet_pton('::1')), ip2long('1.2.3.4'), long2ip(16909060)]);
t('gz', fn() => [gzdecode(gzencode('abc')), gzinflate(gzdeflate('abc')), gzuncompress(gzcompress('abc'))]);
t('ini', fn() => [ini_get('allow_url_fopen'), ini_set('precision', '10'), ini_get('precision'), ini_get('nope.nope')]);
t('data', fn() => [file_get_contents('data://text/plain,hello'), file_get_contents('data://text/plain;base64,aGk='), fread(fopen('data://text/plain,rs', 'r'), 10)]);
t('preg_quote', fn() => [preg_quote('a/b.c', '~'), preg_quote('a/b'), preg_quote('a#b', '#')]);
t('isset offset', function () { $s = 'src/**'; return [isset($s[6]), isset($s[5]), isset($s[-1]), isset($s[-7]), isset($s['x']), isset($s['1']), $s[9] ?? 'dflt']; });
t('strnat', fn() => [strnatcmp('e2', 'e10'), strnatcasecmp('A1', 'a2'), strnatcmp('a', 'a'), strnatcmp('img12', 'img10')]);
t('datetime tz', fn() => [(new DateTimeImmutable('2000-01-01 UTC'))->getTimestamp(), (new DateTimeImmutable('2000-01-01 12:00:00 GMT'))->getTimestamp()]);
t('get_class_methods', fn() => get_class_methods('Cm'));
class Cm { public function b() {} public function a() {} protected function c() {} private function d() {} }
t('warn fopen', function () { $got = null; set_error_handler(function ($n, $s) use (&$got) { $got = [$n, $s]; return true; }); $r = fopen('/nonexistent/x', 'r'); restore_error_handler(); return [$r, $got]; });
t('meta', function () { $m = stream_get_meta_data(fopen('php://temp', 'r+')); return [$m['uri'], $m['mode'], $m['seekable'], $m['stream_type'], $m['wrapper_type']]; });
t('hash ctx', function () { $c = hash_init('sha256'); hash_update($c, 'ab'); hash_update($c, 'c'); $d = hash_copy($c); return [hash_final($c) === hash('sha256', 'abc'), hash_final($d) === hash('sha256', 'abc'), hash_final(hash_init('md5')) === md5('')]; });
class Base2 { protected function d() { return 'd'; } }
class A2 extends Base2 { protected function d() { return 'A::d'; } }
class U2 extends Base2 { function __construct(private Base2 $c) {} function r() { return $this->c->d(); } }
t('sibling protected', fn() => (new U2(new A2))->r());
t('mb', fn() => [mb_strimwidth('Hello World', 0, 8, '...'), mb_strimwidth('Hi', 0, 8, '...')]);
