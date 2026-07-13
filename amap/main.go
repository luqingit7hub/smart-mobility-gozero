package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

type service struct {
	name string
	dir  string
}

var services = []service{
	{name: "rpcUser", dir: "rpcUser"},
	{name: "rpcDriver", dir: "rpcDriver"},
	{name: "rpcOrder", dir: "rpcOrder"},
	{name: "rpcMap", dir: "rpcMap"},
	{name: "apiGateway", dir: "apiGateway"},
}

func main() {
	root, err := os.Getwd()
	if err != nil {
		log.Fatalf("get working directory: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	var wg sync.WaitGroup
	cmds := make([]*exec.Cmd, 0, len(services))

	for i, svc := range services {
		if i == len(services)-1 {
			time.Sleep(2 * time.Second)
		}

		dir := filepath.Join(root, svc.dir)
		cmd := exec.CommandContext(ctx, "go", "run", ".")
		cmd.Dir = dir
		cmd.Env = os.Environ()
		cmd.Stdout = prefixedWriter(os.Stdout, svc.name)
		cmd.Stderr = prefixedWriter(os.Stderr, svc.name)

		if err := cmd.Start(); err != nil {
			log.Fatalf("[%s] failed to start: %v", svc.name, err)
		}

		cmds = append(cmds, cmd)
		log.Printf("[%s] started (pid %d)", svc.name, cmd.Process.Pid)

		wg.Add(1)
		go func(s service, c *exec.Cmd) {
			defer wg.Done()
			if err := c.Wait(); err != nil && ctx.Err() == nil {
				log.Printf("[%s] exited with error: %v", s.name, err)
			} else {
				log.Printf("[%s] stopped", s.name)
			}
		}(svc, cmd)
	}

	log.Println("all services started, press Ctrl+C to stop")

	<-sigCh
	log.Println("shutting down all services...")
	cancel()

	for _, cmd := range cmds {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	}
	wg.Wait()
}

type prefixWriter struct {
	w      io.Writer
	prefix string
}

func prefixedWriter(w io.Writer, name string) io.Writer {
	return &prefixWriter{w: w, prefix: fmt.Sprintf("[%s] ", name)}
}

func (p *prefixWriter) Write(b []byte) (int, error) {
	_, err := fmt.Fprintf(p.w, "%s%s", p.prefix, string(b))
	if err != nil {
		return 0, err
	}
	return len(b), nil
}
