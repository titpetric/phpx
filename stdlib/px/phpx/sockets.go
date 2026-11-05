//go:build phpx

package phpx

import (
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterArea(`Runtime\PX\Sockets`, registerPhpxSockets)
}

// registerPhpxSockets installs fsockopen over net.Dial: a handle fread, fwrite
// and fclose accept.
func registerPhpxSockets(rt *runner.Runtime) {
	rt.RegisterFunc("fsockopen", func(host string, rest ...any) any {
		port := int64(-1)
		if len(rest) > 0 && rest[0] != nil {
			port = phpval.Int(rest[0])
		}
		var setErrno, setErrstr func(any)
		if len(rest) > 1 {
			setErrno, _ = rest[1].(func(any))
		}
		if len(rest) > 2 {
			setErrstr, _ = rest[2].(func(any))
		}
		timeout := 60.0
		if len(rest) > 3 && rest[3] != nil {
			timeout = phpval.Float(rest[3])
		}
		network := "tcp"
		switch {
		case strings.HasPrefix(host, "udp://"):
			network, host = "udp", strings.TrimPrefix(host, "udp://")
		case strings.HasPrefix(host, "tcp://"):
			host = strings.TrimPrefix(host, "tcp://")
		case strings.HasPrefix(host, "unix://"):
			network, host = "unix", strings.TrimPrefix(host, "unix://")
		}
		addr := host
		if network != "unix" {
			addr = net.JoinHostPort(host, strconv.FormatInt(port, 10))
		}
		conn, err := net.DialTimeout(network, addr, time.Duration(timeout*float64(time.Second)))
		if err != nil {
			if setErrno != nil {
				setErrno(int64(111))
			}
			if setErrstr != nil {
				setErrstr(err.Error())
			}
			return false
		}
		return conn
	})
}
