//go:build phpx

package phpx_test

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/titpetric/phpx/parser"
	"github.com/titpetric/phpx/runner"
	"github.com/titpetric/phpx/stdlib"
)

func runCurl(t *testing.T, src string) string {
	t.Helper()
	program, err := parser.Parse(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var out strings.Builder
	rt := runner.New(&out, runner.Options{})
	stdlib.Register(rt)
	if err := rt.Run(program); err != nil {
		t.Fatalf("run: %v", err)
	}
	return out.String()
}

func curlServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("X-Method", r.Method)
		w.Header().Set("X-Custom", r.Header.Get("X-Custom"))
		_, _ = w.Write([]byte(r.Method + ":" + string(body)))
	})
	mux.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/echo", http.StatusFound)
	})
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(404) })
	return httptest.NewServer(mux)
}

// TestCurlExec pins the transfer: body, status, method, headers both ways, the
// callbacks, redirects and the error codes of a refused connection.
func TestCurlExec(t *testing.T) {
	t.Parallel()
	srv := curlServer()
	defer srv.Close()
	got := runCurl(t, `<?php
$ch = curl_init("`+srv.URL+`/echo");
curl_setopt_array($ch, [CURLOPT_RETURNTRANSFER => true, CURLOPT_POSTFIELDS => "payload", CURLOPT_HTTPHEADER => ["X-Custom: yes"]]);
echo curl_exec($ch), "|", curl_getinfo($ch, CURLINFO_HTTP_CODE), "|", curl_errno($ch), "\n";
$headers = []; $chunks = 0;
$ch = curl_init("`+srv.URL+`/redirect");
curl_setopt_array($ch, [CURLOPT_FOLLOWLOCATION => true, CURLOPT_HEADERFUNCTION => function ($c, $h) use (&$headers) { $headers[] = trim($h); return strlen($h); }, CURLOPT_WRITEFUNCTION => function ($c, $d) use (&$chunks) { $chunks += strlen($d); return strlen($d); }]);
var_dump(curl_exec($ch));
echo count($headers) > 4 ? "headers" : "few", "|", $chunks, "|", curl_getinfo($ch, CURLINFO_REDIRECT_COUNT), "\n";
$ch = curl_init("`+srv.URL+`/status"); curl_setopt($ch, CURLOPT_RETURNTRANSFER, true); curl_exec($ch);
echo curl_getinfo($ch, CURLINFO_HTTP_CODE), "\n";
$ch = curl_init("http://127.0.0.1:1/"); curl_setopt($ch, CURLOPT_CONNECTTIMEOUT_MS, 500);
var_dump(curl_exec($ch)); echo curl_errno($ch), "\n";
$ch = curl_init("`+srv.URL+`/echo"); curl_setopt($ch, CURLOPT_HEADERFUNCTION, fn($c, $h) => 0);
var_dump(curl_exec($ch)); echo curl_errno($ch), "\n";
$m = curl_multi_init(); $a = curl_init("`+srv.URL+`/echo"); curl_setopt($a, CURLOPT_RETURNTRANSFER, true);
curl_multi_add_handle($m, $a); $running = 1;
curl_multi_exec($m, $running); $msg = curl_multi_info_read($m);
echo $running, "|", $msg["msg"], "|", $msg["result"], "|", $msg["handle"] === $a ? "same" : "other", "|", curl_multi_getcontent($a), "\n";
`)
	want := "POST:payload|200|0\nbool(true)\nheaders|4|1\n404\nbool(false)\n7\nbool(false)\n23\n0|1|0|same|GET:\n"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

// TestCurlProgressHeaderAuth pins the progress callback, CURLOPT_HEADER, the
// Basic-only preemptive credentials and the errno of an unusable IP family.
func TestCurlProgressHeaderAuth(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Auth:" + r.Header.Get("Authorization")))
	}))
	defer srv.Close()
	got := runCurl(t, `<?php
$calls = 0; $last = [];
$ch = curl_init("`+srv.URL+`/");
curl_setopt_array($ch, [CURLOPT_RETURNTRANSFER => true, CURLOPT_NOPROGRESS => false,
  CURLOPT_PROGRESSFUNCTION => function ($c, $dt, $dn, $ut, $un) use (&$calls, &$last) { $calls++; $last = [$dt, $dn]; return 0; }]);
curl_exec($ch);
echo $calls > 0 ? "progress" : "none", "|", $last[0] === $last[1] ? "full" : "partial", "\n";
$ch = curl_init("`+srv.URL+`/");
curl_setopt_array($ch, [CURLOPT_RETURNTRANSFER => true, CURLOPT_NOPROGRESS => false, CURLOPT_PROGRESSFUNCTION => fn() => 1]);
var_dump(curl_exec($ch)); echo curl_errno($ch), "\n";
$ch = curl_init("`+srv.URL+`/"); curl_setopt_array($ch, [CURLOPT_RETURNTRANSFER => true, CURLOPT_HEADER => true]);
$r = curl_exec($ch); echo substr($r, 0, 15), "|", substr($r, -5), "\n";
$ch = curl_init("`+srv.URL+`/"); curl_setopt_array($ch, [CURLOPT_RETURNTRANSFER => true, CURLOPT_USERPWD => "u:p"]);
echo curl_exec($ch), "\n";
$ch = curl_init("`+srv.URL+`/"); curl_setopt_array($ch, [CURLOPT_RETURNTRANSFER => true, CURLOPT_USERPWD => "u:p", CURLOPT_HTTPAUTH => CURLAUTH_DIGEST]);
echo curl_exec($ch), "\n";
$ch = curl_init("`+srv.URL+`/"); curl_setopt_array($ch, [CURLOPT_RETURNTRANSFER => true, CURLOPT_IPRESOLVE => CURL_IPRESOLVE_V6]);
var_dump(curl_exec($ch)); echo curl_errno($ch), "\n";
`)
	want := "progress|full\nbool(false)\n42\nHTTP/1.1 200 OK|Auth:\nAuth:Basic dTpw\nAuth:\nbool(false)\n7\n"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

// TestCurlSetoptValueTypes pins the TypeError of a non-array list option and the
// conversion of a Stringable object to a string option.
func TestCurlSetoptValueTypes(t *testing.T) {
	t.Parallel()
	got := runCurl(t, `<?php
class S { function __toString() { return "u:p"; } }
$ch = curl_init("http://127.0.0.1:1/");
try { curl_setopt($ch, CURLOPT_HTTPHEADER, "X: y"); } catch (TypeError $e) { echo $e->getMessage(), "\n"; }
try { curl_setopt($ch, CURLOPT_PROXYUSERPWD, new stdClass); } catch (Error $e) { echo $e->getMessage(), "\n"; }
var_dump(curl_setopt($ch, CURLOPT_USERPWD, new S));
var_dump(curl_setopt($ch, CURLOPT_HTTPHEADER, ["X: y"]));
`)
	want := "curl_setopt(): The CURLOPT_HTTPHEADER option must have an array value\nObject of class stdClass could not be converted to string\nbool(true)\nbool(true)\n"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

// TestCurlTLSHandshakeEOF pins errno 35 when an https URL meets a server that
// hangs up on the handshake, where the transport reports a bare EOF.
func TestCurlTLSHandshakeEOF(t *testing.T) {
	t.Parallel()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skip("no loopback listener")
	}
	defer ln.Close()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			buf := make([]byte, 512)
			_, _ = c.Read(buf)
			_ = c.Close()
		}
	}()
	got := runCurl(t, `<?php
$ch = curl_init("https://`+ln.Addr().String()+`/");
curl_setopt_array($ch, [CURLOPT_RETURNTRANSFER => true, CURLOPT_SSL_VERIFYPEER => false]);
var_dump(curl_exec($ch)); echo curl_errno($ch), "\n";
`)
	if want := "bool(false)\n35\n"; got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

// TestCurlTrailers pins the delivery of chunked trailer fields to the header
// callback after the body.
func TestCurlTrailers(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Trailer", "X-Sum")
		_, _ = w.Write([]byte("body"))
		w.Header().Set("X-Sum", "42")
	}))
	defer srv.Close()
	got := runCurl(t, `<?php
$lines = [];
$ch = curl_init("`+srv.URL+`/");
curl_setopt_array($ch, [CURLOPT_RETURNTRANSFER => true, CURLOPT_HEADERFUNCTION => function ($c, $h) use (&$lines) { $lines[] = trim($h); return strlen($h); }]);
echo curl_exec($ch), "|", implode(",", array_slice($lines, -3)), "\n";
`)
	if want := "body|,X-Sum: 42,\n"; got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

// TestCurlRewindOnRedirect pins errno 65 when a followed redirect needs the
// request body again and the body only comes from a read callback.
func TestCurlRewindOnRedirect(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		http.Redirect(w, r, "/elsewhere", http.StatusTemporaryRedirect)
	}))
	defer srv.Close()
	got := runCurl(t, `<?php
$sent = false;
$ch = curl_init("`+srv.URL+`/");
curl_setopt_array($ch, [CURLOPT_RETURNTRANSFER => true, CURLOPT_FOLLOWLOCATION => true, CURLOPT_UPLOAD => true, CURLOPT_CUSTOMREQUEST => "POST",
  CURLOPT_READFUNCTION => function ($c, $fd, $len) use (&$sent) { if ($sent) { return ""; } $sent = true; return "payload"; }]);
var_dump(curl_exec($ch)); echo curl_errno($ch), "\n";
`)
	if want := "bool(false)\n65\n"; got != want {
		t.Errorf("got %q want %q", got, want)
	}
}
