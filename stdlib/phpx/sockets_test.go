//go:build phpx

package phpx_test

import (
	"net"
	"strconv"
	"testing"
)

// TestFsockopen pins a TCP round trip through fsockopen, fwrite, fread and fclose, and the false of a refusal.
func TestFsockopen(t *testing.T) {
	t.Parallel()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		buf := make([]byte, 4)
		_, _ = c.Read(buf)
		_, _ = c.Write([]byte("re:" + string(buf)))
		_ = c.Close()
	}()
	port := strconv.Itoa(ln.Addr().(*net.TCPAddr).Port)
	got := runCurl(t, `<?php
$s = fsockopen("127.0.0.1", `+port+`, $no, $str, 2);
fwrite($s, "ping"); echo fread($s, 16), "\n"; fclose($s);
var_dump(@fsockopen("127.0.0.1", 1, $no, $str, 1));`)
	if want := "re:ping\nbool(false)\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
