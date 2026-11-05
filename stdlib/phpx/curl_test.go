//go:build phpx

package phpx_test

import (
	"io"
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
