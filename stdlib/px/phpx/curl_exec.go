//go:build phpx

package phpx

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

// curlInfoOrder is the order curl_getinfo() lists its keys in.
var curlInfoOrder = []string{
	"url", "content_type", "http_code", "header_size", "request_size", "filetime", "ssl_verify_result",
	"redirect_count", "total_time", "namelookup_time", "connect_time", "pretransfer_time", "size_upload",
	"size_download", "speed_download", "speed_upload", "download_content_length", "upload_content_length",
	"starttransfer_time", "redirect_time", "redirect_url", "primary_ip", "certinfo", "primary_port",
	"local_ip", "local_port", "http_version", "protocol", "ssl_verifyresult", "scheme", "appconnect_time",
	"connect_time_us", "namelookup_time_us", "pretransfer_time_us", "redirect_time_us",
	"starttransfer_time_us", "total_time_us",
}

// curlInfoNames maps a CURLINFO_ constant to the key curl_getinfo reports it as.
var curlInfoNames = map[string]string{
	"CURLINFO_EFFECTIVE_URL": "url", "CURLINFO_CONTENT_TYPE": "content_type", "CURLINFO_HTTP_CODE": "http_code",
	"CURLINFO_RESPONSE_CODE": "http_code", "CURLINFO_HEADER_SIZE": "header_size", "CURLINFO_REQUEST_SIZE": "request_size",
	"CURLINFO_FILETIME": "filetime", "CURLINFO_SSL_VERIFYRESULT": "ssl_verify_result", "CURLINFO_REDIRECT_COUNT": "redirect_count",
	"CURLINFO_TOTAL_TIME": "total_time", "CURLINFO_NAMELOOKUP_TIME": "namelookup_time", "CURLINFO_CONNECT_TIME": "connect_time",
	"CURLINFO_PRETRANSFER_TIME": "pretransfer_time", "CURLINFO_SIZE_UPLOAD": "size_upload", "CURLINFO_SIZE_DOWNLOAD": "size_download",
	"CURLINFO_SPEED_DOWNLOAD": "speed_download", "CURLINFO_SPEED_UPLOAD": "speed_upload",
	"CURLINFO_CONTENT_LENGTH_DOWNLOAD": "download_content_length", "CURLINFO_CONTENT_LENGTH_UPLOAD": "upload_content_length",
	"CURLINFO_STARTTRANSFER_TIME": "starttransfer_time", "CURLINFO_REDIRECT_TIME": "redirect_time", "CURLINFO_REDIRECT_URL": "redirect_url",
	"CURLINFO_PRIMARY_IP": "primary_ip", "CURLINFO_PRIMARY_PORT": "primary_port", "CURLINFO_LOCAL_IP": "local_ip",
	"CURLINFO_LOCAL_PORT": "local_port", "CURLINFO_HTTP_VERSION": "http_version", "CURLINFO_PROTOCOL": "protocol",
	"CURLINFO_SCHEME": "scheme", "CURLINFO_APPCONNECT_TIME": "appconnect_time", "CURLINFO_CERTINFO": "certinfo",
}

// infoOption answers curl_getinfo($ch, $option).
func (h *CurlHandle) infoOption(option int64) any {
	if option == curlOpt("CURLINFO_HEADER_OUT") {
		if v, ok := h.info["request_header"]; ok {
			return v
		}
		return false
	}
	for name, key := range curlInfoNames {
		if phpxCurlIntConstants[name] == option {
			if v, ok := h.info[key]; ok {
				return v
			}
			return false
		}
	}
	return false
}

// curlFailure is a transfer that ends with a libcurl error code.
type curlFailure struct {
	code int64
	msg  string
}

func (f *curlFailure) Error() string { return f.msg }

func curlErrorText(code int64) string {
	if text, ok := curlErrorTexts[code]; ok {
		return text
	}
	return "Unknown error"
}

var curlErrorTexts = map[int64]string{
	0: "No error", 1: "Unsupported protocol", 2: "Failed initialization", 3: "URL using bad/illegal format or missing URL",
	5: "Couldn't resolve proxy name", 6: "Couldn't resolve host name", 7: "Couldn't connect to server",
	8: "Weird server reply", 18: "Transferred a partial file", 23: "Failed writing received data to disk/application",
	26: "Failed to open/read local data from file/application", 27: "Out of memory", 28: "Timeout was reached",
	33: "Requested range was not delivered by the server", 34: "Error occurred in the HTTP2 framing layer", 35: "SSL connect error",
	42: "Operation was aborted by an application callback", 43: "A libcurl function was given a bad argument",
	47: "Number of redirects hit maximum amount", 52: "Server returned nothing (no headers, no data)",
	55: "Failed sending data to the peer", 56: "Failure when receiving data from the peer", 58: "Problem with the local SSL certificate",
	60: "SSL peer certificate or SSH remote key was not OK", 61: "Unrecognized or bad HTTP Content or Transfer-Encoding",
}

func curlMultiErrorText(code int64) any {
	texts := map[int64]string{
		-1: "Please call curl_multi_perform() soon", 0: "No error", 1: "Invalid multi handle", 2: "Invalid easy handle",
		3: "Out of memory", 4: "Internal error", 5: "Invalid socket argument", 6: "Unknown option", 7: "The easy handle is already added to a multi handle",
	}
	if t, ok := texts[code]; ok {
		return t
	}
	return "Unknown error"
}

func curlShareErrorText(code int64) any {
	texts := map[int64]string{0: "No error", 1: "Unknown share option", 2: "Share currently in use", 3: "Invalid share handle", 4: "Out of memory", 5: "Feature not enabled in this library"}
	if t, ok := texts[code]; ok {
		return t
	}
	return "Unknown error"
}

func curlVersion() any {
	out := model.NewArray()
	out.Set("version_number", int64(527873))
	out.Set("age", int64(11))
	out.Set("features", int64(1472153501))
	features := model.NewArray()
	for _, name := range []string{"AsynchDNS", "IDN", "IPv6", "Largefile", "libz", "NTLM", "SSL", "HTTP2", "UnixSockets", "HTTPS-proxy", "alt-svc", "HSTS"} {
		features.Set(name, true)
	}
	out.Set("feature_list", features)
	out.Set("ssl_version_number", int64(0))
	out.Set("version", "8.4.0")
	out.Set("host", "x86_64-pc-linux-gnu")
	out.Set("ssl_version", "OpenSSL/3.0.0")
	out.Set("libz_version", "1.3")
	protocols := model.NewArray()
	for _, p := range []string{"file", "http", "https"} {
		protocols.Append(p)
	}
	out.Set("protocols", protocols)
	return out
}

func phpxURLEscape(s string) string {
	return strings.ReplaceAll(url.QueryEscape(s), "+", "%20")
}

func phpxURLUnescape(s string) string {
	out, err := url.PathUnescape(s)
	if err != nil {
		return s
	}
	return out
}

// exec runs the transfer the options describe and reports its result as
// curl_exec does: the body when CURLOPT_RETURNTRANSFER is set, true, or false
// with curl_errno() set.
func (h *CurlHandle) exec(rt *runner.Runtime) (any, error) {
	h.errno, h.errmsg, h.body = 0, "", ""
	h.info = map[string]any{}
	t := &curlTransfer{rt: rt, h: h, start: time.Now()}
	body, err := t.run()
	var scriptErr error
	var failure *curlFailure
	if err != nil {
		if errors.As(err, &failure) {
			h.errno, h.errmsg = failure.code, failure.msg
		} else {
			scriptErr = err
		}
	}
	t.fillInfo()
	if scriptErr != nil {
		return nil, scriptErr
	}
	if failure != nil {
		return false, nil
	}
	if _, ok := h.opts[curlOpt("CURLOPT_RETURNTRANSFER")]; ok && phpval.Truthy(h.opts[curlOpt("CURLOPT_RETURNTRANSFER")]) {
		h.body = body
		return body, nil
	}
	return true, nil
}

// curlTransfer is one run of a handle.
type curlTransfer struct {
	rt    *runner.Runtime
	h     *CurlHandle
	start time.Time

	headBlock                                         string
	tlsFailed                                         bool
	finalURL                                          string
	status                                            int
	proto                                             string
	contentType                                       string
	headerSize, requestSize, sizeDown, sizeUp         int64
	redirects                                         int
	redirectURL                                       string
	remoteIP, localIP                                 string
	remotePort, localPort                             int
	lookup, connect, appconnect, pretransfer, started time.Duration
	downloadLength                                    int64
	requestHeader                                     string
	sslVerify                                         int64
	secure                                            bool
}

func (t *curlTransfer) opt(name string) (any, bool) {
	v, ok := t.h.opts[curlOpt(name)]
	return v, ok
}

func (t *curlTransfer) str(name string) string {
	if v, ok := t.opt(name); ok {
		return phpval.String(v)
	}
	return ""
}

func (t *curlTransfer) num(name string) int64 {
	if v, ok := t.opt(name); ok {
		return phpval.Int(v)
	}
	return 0
}

func (t *curlTransfer) flag(name string) bool {
	v, ok := t.opt(name)
	return ok && phpval.Truthy(v)
}

func (t *curlTransfer) callback(name string) func(...any) (any, error) {
	v, ok := t.opt(name)
	if !ok || v == nil {
		return nil
	}
	fn, _ := t.rt.Callable(v)
	return fn
}

// run performs the request, following redirects when asked to, and delivers the
// header lines and the body to the callbacks.
func (t *curlTransfer) run() (string, error) {
	target := t.str("CURLOPT_URL")
	if target == "" {
		return "", &curlFailure{3, "URL using bad/illegal format or missing URL"}
	}
	if !strings.Contains(target, "://") {
		target = "http://" + target
	}
	method := "GET"
	switch {
	case t.str("CURLOPT_CUSTOMREQUEST") != "":
		method = t.str("CURLOPT_CUSTOMREQUEST")
	case t.flag("CURLOPT_POST") || t.hasOpt("CURLOPT_POSTFIELDS"):
		method = "POST"
	case t.flag("CURLOPT_UPLOAD") || t.flag("CURLOPT_PUT"):
		method = "PUT"
	case t.flag("CURLOPT_NOBODY"):
		method = "HEAD"
	}
	if t.flag("CURLOPT_HTTPGET") && t.str("CURLOPT_CUSTOMREQUEST") == "" {
		method = "GET"
	}
	maxRedirs := t.num("CURLOPT_MAXREDIRS")
	if _, set := t.opt("CURLOPT_MAXREDIRS"); !set {
		maxRedirs = 30
	}
	follow := t.flag("CURLOPT_FOLLOWLOCATION")
	var out strings.Builder
	for {
		resp, bodyOut, err := t.once(method, target)
		if err != nil {
			return "", err
		}
		out.WriteString(bodyOut)
		redirect := follow && resp.StatusCode >= 300 && resp.StatusCode < 400 && resp.Header.Get("Location") != ""
		if !redirect {
			return out.String(), nil
		}
		if maxRedirs >= 0 && int64(t.redirects) >= maxRedirs {
			return "", &curlFailure{47, fmt.Sprintf("Maximum (%d) redirects followed", maxRedirs)}
		}
		next, err := resp.Request.URL.Parse(resp.Header.Get("Location"))
		if err != nil {
			return "", &curlFailure{3, "URL rejected: Malformed input to a URL function"}
		}
		t.redirects++
		target = next.String()
		if resp.StatusCode == 303 || ((resp.StatusCode == 301 || resp.StatusCode == 302) && method == "POST" && t.num("CURLOPT_POSTREDIR") == 0) {
			if method != "HEAD" {
				method = "GET"
			}
		}
		if method != "GET" && method != "HEAD" && t.flag("CURLOPT_UPLOAD") && t.callback("CURLOPT_READFUNCTION") != nil {
			return "", &curlFailure{65, "necessary data rewind wasn't possible"}
		}
	}
}

func (t *curlTransfer) hasOpt(name string) bool {
	_, ok := t.opt(name)
	return ok
}

// requestBody builds the body reader and the content length the options give it.
func (t *curlTransfer) requestBody(method string) (io.Reader, int64, error) {
	if v, ok := t.opt("CURLOPT_POSTFIELDS"); ok {
		var data string
		if model.IsCollection(v) {
			values := url.Values{}
			model.RangeValues(v, func(k, val any) bool {
				values.Add(phpval.String(k), phpval.String(val))
				return true
			})
			data = values.Encode()
		} else {
			data = phpval.String(v)
		}
		return strings.NewReader(data), int64(len(data)), nil
	}
	if t.flag("CURLOPT_UPLOAD") || t.flag("CURLOPT_PUT") {
		size := int64(-1)
		if _, ok := t.opt("CURLOPT_INFILESIZE"); ok {
			size = t.num("CURLOPT_INFILESIZE")
		}
		if read := t.callback("CURLOPT_READFUNCTION"); read != nil {
			return &curlReader{t: t, read: read}, size, nil
		}
		if f, ok := t.opt("CURLOPT_INFILE"); ok {
			if r, ok := f.(io.Reader); ok {
				return r, size, nil
			}
		}
		return strings.NewReader(""), 0, nil
	}
	return nil, 0, nil
}

// curlReader feeds the READFUNCTION callback's chunks to the request.
type curlReader struct {
	t    *curlTransfer
	read func(...any) (any, error)
	buf  []byte
	done bool
	err  error
}

func (r *curlReader) Read(p []byte) (int, error) {
	for len(r.buf) == 0 && !r.done {
		out, err := r.read(r.t.h, nil, int64(len(p)))
		if err != nil {
			r.err = err
			return 0, err
		}
		s, _ := out.(string)
		if s == "" {
			r.done = true
			break
		}
		r.buf = []byte(s)
	}
	if len(r.buf) == 0 {
		return 0, io.EOF
	}
	n := copy(p, r.buf)
	r.buf = r.buf[n:]
	r.t.sizeUp += int64(n)
	return n, nil
}

// once sends one request and delivers its response.
func (t *curlTransfer) once(method, target string) (*http.Response, string, error) {
	u, err := url.Parse(target)
	if err != nil || u.Host == "" {
		if u != nil && u.Scheme == "file" {
			return t.file(u)
		}
		return nil, "", &curlFailure{3, "URL rejected: Malformed input to a URL function"}
	}
	switch u.Scheme {
	case "http", "https":
	case "file":
		return t.file(u)
	default:
		return nil, "", &curlFailure{1, fmt.Sprintf("Protocol \"%s\" not supported", u.Scheme)}
	}
	if port := t.num("CURLOPT_PORT"); port > 0 {
		u.Host = net.JoinHostPort(u.Hostname(), strconv.FormatInt(port, 10))
	}
	body, size, err := t.requestBody(method)
	if err != nil {
		return nil, "", err
	}
	var bodyReader io.Reader = body
	ctx := context.Background()
	if ms := t.num("CURLOPT_TIMEOUT_MS"); ms > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(ms)*time.Millisecond)
		defer cancel()
	} else if s := t.num("CURLOPT_TIMEOUT"); s > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(s)*time.Second)
		defer cancel()
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return nil, "", &curlFailure{3, "URL rejected: " + err.Error()}
	}
	if size >= 0 && body != nil {
		req.ContentLength = size
	} else if body != nil {
		req.ContentLength = -1
	}
	t.setHeaders(req, body != nil)
	var trace httptrace.ClientTrace
	var dnsStart, connStart, tlsStart time.Time
	trace.DNSStart = func(httptrace.DNSStartInfo) { dnsStart = time.Now() }
	trace.DNSDone = func(httptrace.DNSDoneInfo) {
		if !dnsStart.IsZero() {
			t.lookup = time.Since(t.start)
		}
	}
	trace.ConnectStart = func(string, string) { connStart = time.Now() }
	trace.ConnectDone = func(network, addr string, err error) {
		if err == nil && !connStart.IsZero() {
			t.connect = time.Since(t.start)
		}
	}
	trace.TLSHandshakeStart = func() { tlsStart = time.Now() }
	trace.TLSHandshakeDone = func(_ tls.ConnectionState, err error) {
		t.tlsFailed = t.tlsFailed || err != nil
		if !tlsStart.IsZero() {
			t.appconnect = time.Since(t.start)
		}
	}
	trace.GotConn = func(info httptrace.GotConnInfo) {
		t.pretransfer = time.Since(t.start)
		if addr, ok := info.Conn.RemoteAddr().(*net.TCPAddr); ok {
			t.remoteIP, t.remotePort = addr.IP.String(), addr.Port
		}
		if addr, ok := info.Conn.LocalAddr().(*net.TCPAddr); ok {
			t.localIP, t.localPort = addr.IP.String(), addr.Port
		}
	}
	trace.GotFirstResponseByte = func() { t.started = time.Since(t.start) }
	req = req.WithContext(httptrace.WithClientTrace(ctx, &trace))

	client, err := t.client(u)
	if err != nil {
		return nil, "", err
	}
	t.requestHeader = t.renderRequestHead(req)
	t.requestSize += int64(len(t.requestHeader))
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", t.translate(err, u, ctx)
	}
	defer resp.Body.Close()
	t.finalURL = u.String()
	t.status = resp.StatusCode
	t.proto = resp.Proto
	t.contentType = resp.Header.Get("Content-Type")
	t.downloadLength = resp.ContentLength
	t.secure = resp.TLS != nil
	if loc := resp.Header.Get("Location"); loc != "" {
		if next, err := resp.Request.URL.Parse(loc); err == nil {
			t.redirectURL = next.String()
		}
	}
	if err := t.deliverHeaders(resp); err != nil {
		return nil, "", err
	}
	if method == "HEAD" || resp.StatusCode == 204 || resp.StatusCode == 304 {
		return resp, "", nil
	}
	follow := t.flag("CURLOPT_FOLLOWLOCATION") && resp.StatusCode >= 300 && resp.StatusCode < 400 && resp.Header.Get("Location") != ""
	if follow {
		_, _ = io.Copy(io.Discard, resp.Body)
		return resp, "", nil
	}
	out, err := t.deliverBody(resp)
	if err != nil {
		return nil, "", err
	}
	return resp, out, nil
}

// file serves a file:// URL as libcurl does: the bytes, with no header lines.
func (t *curlTransfer) file(u *url.URL) (*http.Response, string, error) {
	data, err := os.ReadFile(u.Path)
	if err != nil {
		return nil, "", &curlFailure{37, "Couldn't open file " + u.Path}
	}
	t.finalURL = u.String()
	t.sizeDown = int64(len(data))
	resp := &http.Response{StatusCode: 0, Header: http.Header{}, Request: &http.Request{URL: u}}
	if write := t.callback("CURLOPT_WRITEFUNCTION"); write != nil {
		if _, err := write(t.h, string(data)); err != nil {
			return nil, "", err
		}
		return resp, "", nil
	}
	return resp, string(data), nil
}

// setHeaders applies CURLOPT_HTTPHEADER and the headers curl adds on its own.
func (t *curlTransfer) setHeaders(req *http.Request, hasBody bool) {
	removed := map[string]bool{}
	if v, ok := t.opt("CURLOPT_HTTPHEADER"); ok {
		model.RangeValues(v, func(_, line any) bool {
			text := phpval.String(line)
			name, value, found := strings.Cut(text, ":")
			if !found {
				if strings.HasSuffix(text, ";") {
					req.Header[http.CanonicalHeaderKey(strings.TrimSuffix(text, ";"))] = []string{""}
				}
				return true
			}
			name = strings.TrimSpace(name)
			value = strings.TrimSpace(value)
			if value == "" {
				removed[http.CanonicalHeaderKey(name)] = true
				req.Header.Del(name)
				return true
			}
			req.Header.Add(name, value)
			return true
		})
	}
	if strings.EqualFold(req.Header.Get("Host"), "") == false {
		req.Host = req.Header.Get("Host")
		req.Header.Del("Host")
	}
	if req.Header.Get("User-Agent") == "" && !removed["User-Agent"] {
		ua := t.str("CURLOPT_USERAGENT")
		if ua == "" {
			ua = "curl/8.4.0"
		}
		req.Header.Set("User-Agent", ua)
	}
	if req.Header.Get("Accept") == "" && !removed["Accept"] {
		req.Header.Set("Accept", "*/*")
	}
	if v, ok := t.opt("CURLOPT_ENCODING"); ok && req.Header.Get("Accept-Encoding") == "" {
		enc := phpval.String(v)
		if enc == "" {
			enc = "deflate, gzip"
		}
		req.Header.Set("Accept-Encoding", enc)
	}
	if hasBody && req.Header.Get("Content-Type") == "" && !removed["Content-Type"] && t.hasOpt("CURLOPT_POSTFIELDS") {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if ref := t.str("CURLOPT_REFERER"); ref != "" && req.Header.Get("Referer") == "" {
		req.Header.Set("Referer", ref)
	}
	if cookie := t.str("CURLOPT_COOKIE"); cookie != "" && req.Header.Get("Cookie") == "" {
		req.Header.Set("Cookie", cookie)
	}
	if up := t.str("CURLOPT_USERPWD"); up != "" && req.Header.Get("Authorization") == "" && t.preemptiveBasic() {
		user, pass, _ := strings.Cut(up, ":")
		req.SetBasicAuth(user, pass)
	}
	for name := range removed {
		req.Header.Del(name)
	}
}

// preemptiveBasic is true when the credentials go out with the first request:
// libcurl sends Basic at once only when it is the sole method CURLOPT_HTTPAUTH
// allows (the default), and waits for a challenge otherwise.
func (t *curlTransfer) preemptiveBasic() bool {
	auth, set := t.opt("CURLOPT_HTTPAUTH")
	return !set || phpval.Int(auth) == 1
}

// renderRequestHead is the request line and headers as sent, for HEADER_OUT.
func (t *curlTransfer) renderRequestHead(req *http.Request) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s %s HTTP/1.1\r\nHost: %s\r\n", req.Method, req.URL.RequestURI(), req.URL.Host)
	keys := make([]string, 0, len(req.Header))
	for k := range req.Header {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		for _, v := range req.Header[k] {
			fmt.Fprintf(&b, "%s: %s\r\n", k, v)
		}
	}
	b.WriteString("\r\n")
	return b.String()
}

// client builds the http.Client the options describe.
func (t *curlTransfer) client(u *url.URL) (*http.Client, error) {
	dialer := &net.Dialer{}
	if ms := t.num("CURLOPT_CONNECTTIMEOUT_MS"); ms > 0 {
		dialer.Timeout = time.Duration(ms) * time.Millisecond
	} else if s := t.num("CURLOPT_CONNECTTIMEOUT"); s > 0 {
		dialer.Timeout = time.Duration(s) * time.Second
	}
	network := ""
	switch t.num("CURLOPT_IPRESOLVE") {
	case 1:
		network = "tcp4"
	case 2:
		network = "tcp6"
	}
	resolve := map[string]string{}
	if v, ok := t.opt("CURLOPT_RESOLVE"); ok {
		model.RangeValues(v, func(_, line any) bool {
			parts := strings.SplitN(phpval.String(line), ":", 3)
			if len(parts) == 3 {
				resolve[parts[0]+":"+parts[1]] = strings.Trim(parts[2], "[]")
			}
			return true
		})
	}
	unix := t.str("CURLOPT_UNIX_SOCKET_PATH")
	tr := &http.Transport{
		DisableCompression: true,
		DisableKeepAlives:  t.flag("CURLOPT_FORBID_REUSE") || t.flag("CURLOPT_FRESH_CONNECT"),
		DialContext: func(ctx context.Context, netw, addr string) (net.Conn, error) {
			if unix != "" {
				return dialer.DialContext(ctx, "unix", unix)
			}
			if to, ok := resolve[addr]; ok {
				_, port, _ := net.SplitHostPort(addr)
				addr = net.JoinHostPort(to, port)
			}
			if network != "" {
				netw = network
			}
			return dialer.DialContext(ctx, netw, addr)
		},
	}
	if proxy := t.str("CURLOPT_PROXY"); proxy != "" {
		if !strings.Contains(proxy, "://") {
			proxy = "http://" + proxy
		}
		pu, err := url.Parse(proxy)
		if err != nil {
			return nil, &curlFailure{5, "Couldn't resolve proxy name"}
		}
		if up := t.str("CURLOPT_PROXYUSERPWD"); up != "" {
			user, pass, _ := strings.Cut(up, ":")
			pu.User = url.UserPassword(user, pass)
		}
		tr.Proxy = http.ProxyURL(pu)
	}
	cfg := &tls.Config{}
	if v, ok := t.opt("CURLOPT_SSL_VERIFYPEER"); ok && !phpval.Truthy(v) {
		cfg.InsecureSkipVerify = true
	}
	if v, ok := t.opt("CURLOPT_SSL_VERIFYHOST"); ok && phpval.Int(v) == 0 {
		cfg.InsecureSkipVerify = true
	}
	if ca := t.str("CURLOPT_CAINFO"); ca != "" {
		pem, err := os.ReadFile(ca)
		if err != nil {
			return nil, &curlFailure{77, "error setting certificate file: " + ca}
		}
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(pem)
		cfg.RootCAs = pool
	}
	if cert, key := t.str("CURLOPT_SSLCERT"), t.str("CURLOPT_SSLKEY"); cert != "" {
		if key == "" {
			key = cert
		}
		pair, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			return nil, &curlFailure{58, "could not load PEM client certificate"}
		}
		cfg.Certificates = []tls.Certificate{pair}
	}
	switch t.num("CURLOPT_SSLVERSION") {
	case 4:
		cfg.MinVersion = tls.VersionTLS10
	case 5:
		cfg.MinVersion = tls.VersionTLS11
	case 6:
		cfg.MinVersion = tls.VersionTLS12
	case 7:
		cfg.MinVersion = tls.VersionTLS13
	}
	tr.TLSClientConfig = cfg
	if v := t.num("CURLOPT_HTTP_VERSION"); v == 1 || v == 2 {
		tr.TLSNextProto = map[string]func(string, *tls.Conn) http.RoundTripper{}
	}
	return &http.Client{
		Transport:     tr,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}, nil
}

// translate maps a net/http error onto the libcurl code a script sees.
func (t *curlTransfer) translate(err error, u *url.URL, ctx context.Context) error {
	var failure *curlFailure
	if errors.As(err, &failure) {
		return failure
	}
	var dns *net.DNSError
	var certErr x509.UnknownAuthorityError
	var hostErr x509.HostnameError
	var invalid x509.CertificateInvalidError
	elapsed := time.Since(t.start).Milliseconds()
	switch {
	case errors.Is(ctx.Err(), context.DeadlineExceeded) || os.IsTimeout(err):
		var opErr *net.OpError
		if errors.As(err, &opErr) && opErr.Op == "dial" {
			return &curlFailure{28, fmt.Sprintf("Connection timed out after %d milliseconds", elapsed)}
		}
		return &curlFailure{28, fmt.Sprintf("Operation timed out after %d milliseconds with 0 bytes received", elapsed)}
	case errors.As(err, &dns):
		return &curlFailure{6, "Could not resolve host: " + u.Hostname()}
	case errors.As(err, &certErr), errors.As(err, &hostErr), errors.As(err, &invalid):
		return &curlFailure{60, "SSL certificate problem: " + err.Error()}
	case t.tlsFailed && !errors.As(err, &certErr) && !errors.As(err, &hostErr) && !errors.As(err, &invalid):
		return &curlFailure{35, "TLS connect error: " + err.Error()}
	case strings.Contains(err.Error(), "tls:"):
		return &curlFailure{35, "SSL connect error: " + err.Error()}
	case strings.Contains(err.Error(), "connection refused"), strings.Contains(err.Error(), "no route to host"), strings.Contains(err.Error(), "network is unreachable"), strings.Contains(err.Error(), "no suitable address"):
		port := u.Port()
		if port == "" {
			port = map[string]string{"https": "443"}[u.Scheme]
			if port == "" {
				port = "80"
			}
		}
		return &curlFailure{7, fmt.Sprintf("Failed to connect to %s port %s after %d ms: Couldn't connect to server", u.Hostname(), port, elapsed)}
	case errors.Is(err, io.EOF), strings.Contains(err.Error(), "EOF"):
		return &curlFailure{52, "Empty reply from server"}
	case strings.Contains(err.Error(), "proxyconnect"):
		return &curlFailure{5, "Couldn't resolve proxy name"}
	}
	return &curlFailure{56, "Failure when receiving data from the peer: " + err.Error()}
}

// deliverHeaders hands the status line and header lines to CURLOPT_HEADERFUNCTION
// or, with CURLOPT_HEADER, to the body.
func (t *curlTransfer) deliverHeaders(resp *http.Response) error {
	var lines []string
	lines = append(lines, fmt.Sprintf("%s %s\r\n", resp.Proto, resp.Status))
	keys := make([]string, 0, len(resp.Header))
	for k := range resp.Header {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		for _, v := range resp.Header[k] {
			lines = append(lines, k+": "+v+"\r\n")
		}
	}
	for _, te := range resp.TransferEncoding {
		lines = append(lines, "Transfer-Encoding: "+te+"\r\n")
	}
	lines = append(lines, "\r\n")
	hf := t.callback("CURLOPT_HEADERFUNCTION")
	if hf == nil && t.flag("CURLOPT_HEADER") {
		t.headBlock = strings.Join(lines, "")
	}
	for _, line := range lines {
		t.headerSize += int64(len(line))
		if hf == nil {
			continue
		}
		out, err := hf(t.h, line)
		if err != nil {
			return err
		}
		if phpval.Int(out) != int64(len(line)) {
			return &curlFailure{23, fmt.Sprintf("Failure writing output to destination, passed %d returned %d", len(line), phpval.Int(out))}
		}
	}
	return nil
}

// deliverBody streams the response body to the write callback, the CURLOPT_FILE
// handle, or the return value.
func (t *curlTransfer) deliverBody(resp *http.Response) (string, error) {
	var src io.Reader = resp.Body
	switch strings.ToLower(resp.Header.Get("Content-Encoding")) {
	case "gzip", "x-gzip":
		if _, ok := t.opt("CURLOPT_ENCODING"); ok {
			zr, err := gzip.NewReader(src)
			if err != nil {
				return "", &curlFailure{61, "Error while processing content unencoding: invalid code lengths set"}
			}
			src = zr
		}
	case "deflate":
		if _, ok := t.opt("CURLOPT_ENCODING"); ok {
			raw, _ := io.ReadAll(src)
			if zr, err := zlib.NewReader(bytes.NewReader(raw)); err == nil {
				src = zr
			} else {
				src = flate.NewReader(bytes.NewReader(raw))
			}
		}
	}
	write := t.callback("CURLOPT_WRITEFUNCTION")
	var sink io.Writer
	if f, ok := t.opt("CURLOPT_FILE"); ok {
		sink, _ = f.(io.Writer)
	}
	returnIt := t.flag("CURLOPT_RETURNTRANSFER") && write == nil && sink == nil
	var ret strings.Builder
	buf := make([]byte, 16384)
	var progress func(...any) (any, error)
	if v, ok := t.opt("CURLOPT_NOPROGRESS"); ok && !phpval.Truthy(v) {
		if progress = t.callback("CURLOPT_XFERINFOFUNCTION"); progress == nil {
			progress = t.callback("CURLOPT_PROGRESSFUNCTION")
		}
	}
	report := func() error {
		if progress == nil {
			return nil
		}
		total := resp.ContentLength
		if total < 0 {
			total = 0
		}
		out, perr := progress(t.h, total, t.sizeDown, int64(0), t.sizeUp)
		if perr != nil {
			return perr
		}
		if phpval.Int(out) != 0 {
			return &curlFailure{42, "Operation was aborted by an application callback"}
		}
		return nil
	}
	if err := report(); err != nil {
		return "", err
	}
	for {
		n, err := src.Read(buf)
		if n > 0 || (t.headBlock != "" && t.sizeDown == 0) {
			chunk := t.headBlock + string(buf[:n])
			t.headBlock = ""
			t.sizeDown += int64(n)
			if perr := report(); perr != nil {
				return "", perr
			}
			switch {
			case write != nil:
				out, werr := write(t.h, chunk)
				if werr != nil {
					return "", werr
				}
				if phpval.Int(out) != int64(len(chunk)) {
					return "", &curlFailure{23, fmt.Sprintf("Failure writing output to destination, passed %d returned %d", len(chunk), phpval.Int(out))}
				}
			case sink != nil:
				_, _ = io.WriteString(sink, chunk)
			case returnIt:
				ret.WriteString(chunk)
			default:
				_, _ = io.WriteString(t.rt.Output(), chunk)
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			if errors.Is(err, io.ErrUnexpectedEOF) {
				return "", &curlFailure{18, "transfer closed with outstanding read data remaining"}
			}
			return "", &curlFailure{56, "Failure when receiving data from the peer"}
		}
	}
	if err := t.deliverTrailers(resp); err != nil {
		return "", err
	}
	return ret.String(), nil
}

// deliverTrailers hands the trailer fields of a chunked response, known once
// the body is read, to CURLOPT_HEADERFUNCTION the way libcurl does: each field
// line, then the blank line.
func (t *curlTransfer) deliverTrailers(resp *http.Response) error {
	hf := t.callback("CURLOPT_HEADERFUNCTION")
	if len(resp.Trailer) == 0 || hf == nil {
		return nil
	}
	keys := make([]string, 0, len(resp.Trailer))
	for k := range resp.Trailer {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var lines []string
	for _, k := range keys {
		for _, v := range resp.Trailer[k] {
			lines = append(lines, k+": "+v+"\r\n")
		}
	}
	lines = append(lines, "\r\n")
	for _, line := range lines {
		t.headerSize += int64(len(line))
		out, err := hf(t.h, line)
		if err != nil {
			return err
		}
		if phpval.Int(out) != int64(len(line)) {
			return &curlFailure{23, fmt.Sprintf("Failure writing output to destination, passed %d returned %d", len(line), phpval.Int(out))}
		}
	}
	return nil
}

// fillInfo records what curl_getinfo reports.
func (t *curlTransfer) fillInfo() {
	secs := func(d time.Duration) float64 { return float64(d) / float64(time.Second) }
	total := time.Since(t.start)
	i := t.h.info
	i["url"] = t.finalURL
	i["content_type"] = nilIfEmpty(t.contentType)
	i["http_code"] = int64(t.status)
	i["header_size"] = t.headerSize
	i["request_size"] = t.requestSize
	i["filetime"] = int64(-1)
	i["ssl_verify_result"] = t.sslVerify
	i["redirect_count"] = int64(t.redirects)
	i["total_time"] = secs(total)
	i["namelookup_time"] = secs(t.lookup)
	i["connect_time"] = secs(t.connect)
	i["pretransfer_time"] = secs(t.pretransfer)
	i["size_upload"] = float64(t.sizeUp)
	i["size_download"] = float64(t.sizeDown)
	i["speed_download"] = float64(0)
	i["speed_upload"] = float64(0)
	i["download_content_length"] = float64(t.downloadLength)
	i["upload_content_length"] = float64(-1)
	i["starttransfer_time"] = secs(t.started)
	i["redirect_time"] = float64(0)
	i["redirect_url"] = t.redirectURL
	i["primary_ip"] = t.remoteIP
	i["certinfo"] = model.NewArray()
	i["primary_port"] = int64(t.remotePort)
	i["local_ip"] = t.localIP
	i["local_port"] = int64(t.localPort)
	i["http_version"] = int64(2)
	if t.proto == "HTTP/1.1" {
		i["http_version"] = int64(2)
	} else if t.proto == "HTTP/1.0" {
		i["http_version"] = int64(1)
	} else if strings.HasPrefix(t.proto, "HTTP/2") {
		i["http_version"] = int64(3)
	}
	i["protocol"] = int64(1)
	i["ssl_verifyresult"] = t.sslVerify
	i["scheme"] = "HTTP"
	if t.secure {
		i["protocol"] = int64(2)
		i["scheme"] = "HTTPS"
	}
	i["appconnect_time"] = secs(t.appconnect)
	i["connect_time_us"] = int64(t.connect / time.Microsecond)
	i["namelookup_time_us"] = int64(t.lookup / time.Microsecond)
	i["pretransfer_time_us"] = int64(t.pretransfer / time.Microsecond)
	i["redirect_time_us"] = int64(0)
	i["starttransfer_time_us"] = int64(t.started / time.Microsecond)
	i["total_time_us"] = int64(total / time.Microsecond)
	if t.flag("CURLOPT_HEADER_OUT") || t.hasOpt(strconv.Itoa(int(curlOpt("CURLINFO_HEADER_OUT")))) {
		i["request_header"] = t.requestHeader
	}
}

func nilIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
