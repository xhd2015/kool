package viewer

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/creack/pty"
)

type BashSession struct {
	cmd           *exec.Cmd
	pty           *os.File
	outputChannel chan string
	errorChannel  chan string
	ctx           context.Context
	cancel        context.CancelFunc
	mutex         sync.Mutex
}

func initBashSession(workingDir string) (*BashSession, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Start bash in interactive mode
	cmd := exec.CommandContext(ctx, "bash", "--login", "-i")
	cmd.Dir = workingDir

	// Set environment variables for proper terminal behavior
	cmd.Env = append(os.Environ(),
		"TERM=xterm-256color",
		"SHELL=/bin/bash",
	)

	// Start the command with a PTY
	ptmx, err := pty.Start(cmd)
	if err != nil {
		cancel()
		return nil, err
	}

	session := &BashSession{
		cmd:           cmd,
		pty:           ptmx,
		outputChannel: make(chan string, 100),
		errorChannel:  make(chan string, 100),
		ctx:           ctx,
		cancel:        cancel,
	}

	// Start goroutine to read from PTY
	go session.readPTY()

	return session, nil
}

func (bs *BashSession) readPTY() {
	fmt.Println("Starting readPTY goroutine")
	buf := make([]byte, 1024)
	for {
		select {
		case <-bs.ctx.Done():
			fmt.Println("readPTY context done")
			return
		default:
			n, err := bs.pty.Read(buf)
			if err != nil {
				if err != io.EOF {
					fmt.Printf("PTY read error: %v\n", err)
				}
				return
			}

			if n > 0 {
				output := string(buf[:n])
				fmt.Printf("Read from PTY: %q\n", output)
				select {
				case bs.outputChannel <- output:
					fmt.Printf("Sent to output channel: %q\n", output)
				case <-bs.ctx.Done():
					return
				}
			}
		}
	}
}

func (bs *BashSession) sendInput(input string) error {
	fmt.Printf("sendInput called with: %q\n", input)
	bs.mutex.Lock()
	defer bs.mutex.Unlock()

	_, err := bs.pty.Write([]byte(input))
	if err != nil {
		fmt.Printf("Error writing to PTY: %v\n", err)
	} else {
		fmt.Printf("Successfully wrote to PTY: %q\n", input)
	}
	return err
}

func (bs *BashSession) setSize(cols, rows int) error {
	fmt.Printf("setSize called with cols=%d, rows=%d\n", cols, rows)
	bs.mutex.Lock()
	defer bs.mutex.Unlock()

	winsize := &pty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
	}

	err := pty.Setsize(bs.pty, winsize)
	if err != nil {
		fmt.Printf("Error setting PTY size: %v\n", err)
	} else {
		fmt.Printf("Successfully set PTY size to %dx%d\n", cols, rows)
	}
	return err
}

func (bs *BashSession) close() {
	bs.cancel()
	if bs.pty != nil {
		bs.pty.Close()
	}
	if bs.cmd != nil && bs.cmd.Process != nil {
		bs.cmd.Process.Kill()
	}
}
