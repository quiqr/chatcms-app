package relay

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Run struct {
	Cmd  []string          `json:"cmd"`
	Path []string          `json:"path"`
	Env  map[string]string `json:"env"`
}

type Plan struct {
	Run          string
	Site         string
	Home         string
	Sockets      string
	PreviewPort  int
	AgentPort    int
	GatewayPort  int
	Token        string
	OpenCode     string
	Shell        string
	AgentWaitFor time.Duration
}

func DecodeRun(encoded string) (Run, error) {
	data, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(encoded, "="))
	if err != nil {
		return Run{}, err
	}
	var run Run
	if err := json.Unmarshal(data, &run); err != nil {
		return Run{}, err
	}
	if len(run.Cmd) == 0 {
		return Run{}, errors.New("relay: the run names no command")
	}
	return run, nil
}

func previewCommand(ctx context.Context, plan Plan, run Run, log io.Writer) *exec.Cmd {
	home := filepath.Join(plan.Home, "preview")
	os.MkdirAll(filepath.Join(home, "resources"), 0o700)
	command := exec.CommandContext(ctx, run.Cmd[0], run.Cmd[1:]...)
	command.Dir = plan.Site
	command.Env = []string{
		"HOME=" + home,
		"HUGO_RESOURCEDIR=" + filepath.Join(home, "resources"),
		"TZ=UTC",
		"PATH=" + strings.Join(run.Path, ":"),
	}
	for name, value := range run.Env {
		command.Env = append(command.Env, name+"="+value)
	}
	command.Stdout = log
	command.Stderr = log
	return command
}

func agentEnvironment(ctx context.Context, plan Plan, within time.Duration) ([]string, error) {
	url := fmt.Sprintf("http://127.0.0.1:%d/agent/env", plan.GatewayPort)
	deadline := time.Now().Add(within)
	for {
		request, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		request.Header.Set("Authorization", "Bearer "+plan.Token)
		answer, err := http.DefaultClient.Do(request)
		if err == nil {
			body, _ := io.ReadAll(answer.Body)
			answer.Body.Close()
			if answer.StatusCode == http.StatusOK {
				return parseEnv(string(body)), nil
			}
		}
		if time.Now().After(deadline) || ctx.Err() != nil {
			return nil, fmt.Errorf("relay: the agent's settings did not come: %v", err)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func parseEnv(text string) []string {
	var env []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
			continue
		}
		name, value, _ := strings.Cut(line, "=")
		if unquoted, ok := unquote(value); ok {
			value = unquoted
		}
		env = append(env, name+"="+value)
	}
	return env
}

func unquote(value string) (string, bool) {
	if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
		return strings.ReplaceAll(value[1:len(value)-1], `'\''`, "'"), true
	}
	return value, false
}

func agentCommand(ctx context.Context, plan Plan, env []string, log io.Writer) *exec.Cmd {
	command := exec.CommandContext(ctx, plan.OpenCode, "serve", "--hostname", "127.0.0.1", "--port", fmt.Sprint(plan.AgentPort))
	command.Dir = plan.Site
	command.Env = env
	command.Stdout = log
	command.Stderr = log
	return command
}

func Launch(ctx context.Context, plan Plan, log io.Writer) error {
	run, err := DecodeRun(plan.Run)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if _, err := Forward(ctx, "unix", filepath.Join(plan.Sockets, "preview.sock"), "tcp", fmt.Sprintf("127.0.0.1:%d", plan.PreviewPort)); err != nil {
		return err
	}
	if plan.Token != "" {
		if _, err := Forward(ctx, "unix", filepath.Join(plan.Sockets, "agent.sock"), "tcp", fmt.Sprintf("127.0.0.1:%d", plan.AgentPort)); err != nil {
			return err
		}
		if _, err := Forward(ctx, "tcp", fmt.Sprintf("127.0.0.1:%d", plan.GatewayPort), "unix", filepath.Join(plan.Sockets, "ai.sock")); err != nil {
			return err
		}
	}

	var (
		once  sync.Once
		first error
		wg    sync.WaitGroup
	)
	stop := func(err error) {
		once.Do(func() {
			first = err
			cancel()
		})
	}

	preview := previewCommand(ctx, plan, run, log)
	if err := preview.Start(); err != nil {
		return err
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		stop(fmt.Errorf("relay: the preview ended: %v", preview.Wait()))
	}()

	if plan.Token != "" && plan.OpenCode != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			env, err := agentEnvironment(ctx, plan, plan.AgentWaitFor)
			if err != nil {
				fmt.Fprintln(log, err)
				return
			}
			agent := agentCommand(ctx, plan, env, log)
			if err := agent.Start(); err != nil {
				fmt.Fprintln(log, err)
				return
			}
			fmt.Fprintf(log, "relay: the agent ended: %v\n", agent.Wait())
		}()
	}

	<-ctx.Done()
	wg.Wait()
	return first
}
