package k8s

import (
	"fmt"
	"io"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/util/interrupt"
	"k8s.io/kubectl/pkg/util/term"
)

type ExecOptions struct {
	In              io.Reader
	Out             io.Writer
	ErrOut          io.Writer
	TTY             bool
	InterruptParent *interrupt.Handler
}

func (o *ExecOptions) CreateTTY() (term.TTY, remotecommand.TerminalSizeQueue, error) {
	tty := term.TTY{
		In:  o.In,
		Out: o.Out,
		Raw: o.TTY,
	}
	if !tty.IsTerminalIn() {
		return term.TTY{}, nil, fmt.Errorf("unable to use a TTY - input is not a terminal or the right kind of file")
	}
	var sizeQueue remotecommand.TerminalSizeQueue
	if tty.Raw {
		// this call spawns a goroutine to monitor/update the terminal size
		sizeQueue = tty.MonitorSize(tty.GetSize())

		// unset p.Err if it was previously set because both stdout and stderr go over p.Out when tty is
		// true
		o.ErrOut = nil
	}
	return tty, sizeQueue, nil
}
