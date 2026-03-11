package appservice

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/nullstone-io/deployment-sdk/logging"
	"golang.org/x/term"
)

// KuduSsh establishes an interactive SSH session to an Azure App Service container
// via the Kudu websocket endpoint (wss://{app}.scm.azurewebsites.net/webssh/host).
func KuduSsh(ctx context.Context, infra Outputs, osWriters logging.OsWriters) error {
	token, err := infra.Deployer.GetToken(ctx)
	if err != nil {
		return fmt.Errorf("error getting Azure token: %w", err)
	}

	wsURL := fmt.Sprintf("wss://%s.scm.azurewebsites.net/webssh/host", infra.SiteName)

	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+token)

	dialer := websocket.DefaultDialer
	conn, _, err := dialer.DialContext(ctx, wsURL, headers)
	if err != nil {
		return fmt.Errorf("error connecting to Kudu SSH websocket: %w", err)
	}
	defer conn.Close()

	// Put the terminal into raw mode for interactive use
	stdinFd := int(os.Stdin.Fd())
	if !term.IsTerminal(stdinFd) {
		return fmt.Errorf("stdin is not a terminal; interactive SSH requires a TTY")
	}

	oldState, err := term.MakeRaw(stdinFd)
	if err != nil {
		return fmt.Errorf("error setting terminal to raw mode: %w", err)
	}
	defer term.Restore(stdinFd, oldState)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	stdout := osWriters.Stdout()
	var wg sync.WaitGroup

	// websocket → stdout
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					// Only report unexpected errors
					if ctx.Err() == nil {
						fmt.Fprintf(os.Stderr, "\r\nconnection closed: %v\r\n", err)
					}
				}
				return
			}
			stdout.Write(message)
		}
	}()

	// stdin → websocket
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()
		buf := make([]byte, 1024)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				if err != io.EOF && ctx.Err() == nil {
					fmt.Fprintf(os.Stderr, "\r\nstdin read error: %v\r\n", err)
				}
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, buf[:n]); err != nil {
				if ctx.Err() == nil {
					fmt.Fprintf(os.Stderr, "\r\nwebsocket write error: %v\r\n", err)
				}
				return
			}
		}
	}()

	// Wait for context cancellation (either goroutine exits or external cancel)
	<-ctx.Done()
	conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	wg.Wait()

	return nil
}
