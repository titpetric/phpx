//go:build phpx

package pexec

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterProfileBinding(runner.ProfileExec, registerPhpxProc)
}

// Process is the value proc_open returns, a resource in php.
type Process struct {
	cmd      *exec.Cmd
	command  string
	finished bool
	exitCode int
	parent   []*os.File // the child's ends, closed once it is started
	signaled bool
	termsig  int
}

// registerPhpxProc installs proc_open and its companions. They sit with the
// other process bindings: a host that runs untrusted scripts mounts without
// them.
func registerPhpxProc(rt *runner.Runtime) {
	r := root{rt: rt}
	rt.RegisterFunc("proc_open", func(command any, spec any, pipes func(any), rest ...any) (any, error) {
		var cmd *exec.Cmd
		var text string
		if model.IsCollection(command) {
			var argv []string
			model.RangeValues(command, func(_, v any) bool { argv = append(argv, phpval.String(v)); return true })
			if len(argv) == 0 {
				return false, nil
			}
			cmd = exec.Command(argv[0], argv[1:]...)
			text = strings.Join(argv, " ")
		} else {
			text = phpval.String(command)
			cmd = r.command(rt.Context(), text)
		}
		if len(rest) > 0 && rest[0] != nil {
			cmd.Dir = phpval.String(rest[0])
		}
		if len(rest) > 1 && rest[1] != nil {
			var env []string
			model.RangeValues(rest[1], func(k, v any) bool { env = append(env, phpval.String(k)+"="+phpval.String(v)); return true })
			cmd.Env = env
		}
		proc := &Process{cmd: cmd, command: text}
		out := model.NewArray()
		extra := map[int]*os.File{}
		var failed bool
		model.RangeValues(spec, func(k, v any) bool {
			fd := int(phpval.Int(k))
			child, parent, err := proc.descriptor(v)
			if err != nil {
				failed = true
				return false
			}
			switch fd {
			case 0:
				cmd.Stdin = child
			case 1:
				cmd.Stdout = child
			case 2:
				cmd.Stderr = child
			default:
				extra[fd] = child
			}
			if child != nil {
				proc.parent = append(proc.parent, child)
			}
			if parent != nil {
				out.Set(int64(fd), parent)
			}
			return true
		})
		if failed {
			return false, nil
		}
		for fd := 3; len(extra) > 0; fd++ {
			f, ok := extra[fd]
			if !ok {
				f, _ = os.Open(os.DevNull)
			} else {
				delete(extra, fd)
			}
			cmd.ExtraFiles = append(cmd.ExtraFiles, f)
		}
		if err := cmd.Start(); err != nil {
			return false, nil
		}
		for _, f := range proc.parent {
			_ = f.Close()
		}
		if pipes != nil {
			pipes(out)
		}
		return proc, nil
	})
	rt.RegisterFunc("proc_close", func(p *Process) int64 {
		p.wait()
		return int64(p.exitCode)
	})
	rt.RegisterFunc("proc_terminate", func(p *Process, rest ...any) bool {
		sig := syscall.SIGTERM
		if len(rest) > 0 {
			sig = syscall.Signal(phpval.Int(rest[0]))
		}
		if p.cmd.Process == nil {
			return false
		}
		return p.cmd.Process.Signal(sig) == nil
	})
	rt.RegisterFunc("proc_get_status", func(p *Process) *model.Array {
		running := !p.finished
		if running && p.cmd.Process != nil {
			if p.cmd.Process.Signal(syscall.Signal(0)) != nil {
				running = false
			}
		}
		out := model.NewArray()
		out.Set("command", p.command)
		out.Set("pid", int64(p.cmd.Process.Pid))
		out.Set("running", running)
		out.Set("signaled", p.signaled)
		out.Set("stopped", false)
		code := int64(-1)
		if p.finished {
			code = int64(p.exitCode)
		}
		out.Set("exitcode", code)
		out.Set("termsig", int64(p.termsig))
		out.Set("stopsig", int64(0))
		return out
	})
}

// descriptor builds the child's end and the script's end of one descriptor spec.
func (p *Process) descriptor(spec any) (child, parent *os.File, err error) {
	if f, ok := spec.(*os.File); ok {
		return f, nil, nil
	}
	var parts []string
	model.RangeValues(spec, func(_, v any) bool { parts = append(parts, phpval.String(v)); return true })
	if len(parts) == 0 {
		return nil, nil, os.ErrInvalid
	}
	switch parts[0] {
	case "pipe":
		mode := "r"
		if len(parts) > 1 {
			mode = parts[1]
		}
		rd, wr, err := os.Pipe()
		if err != nil {
			return nil, nil, err
		}
		if strings.HasPrefix(mode, "r") {
			return rd, wr, nil
		}
		return wr, rd, nil
	case "file":
		if len(parts) < 3 {
			return nil, nil, os.ErrInvalid
		}
		flag := os.O_RDONLY
		switch strings.TrimSuffix(parts[2], "b") {
		case "w", "w+":
			flag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
		case "a", "a+":
			flag = os.O_WRONLY | os.O_CREATE | os.O_APPEND
		case "r+":
			flag = os.O_RDWR
		}
		f, err := os.OpenFile(parts[1], flag, 0o644)
		return f, nil, err
	}
	return nil, nil, os.ErrInvalid
}

// wait waits for the child once and records how it ended.
func (p *Process) wait() {
	if p.finished {
		return
	}
	err := p.cmd.Wait()
	p.finished = true
	if err == nil {
		return
	}
	if ee, ok := err.(*exec.ExitError); ok {
		if ws, ok := ee.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
			p.signaled, p.termsig = true, int(ws.Signal())
			p.exitCode = -1
			return
		}
		p.exitCode = ee.ExitCode()
		return
	}
	p.exitCode, _ = strconv.Atoi("1")
}
