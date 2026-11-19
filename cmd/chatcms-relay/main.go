package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/quiqr/chatcms-app/internal/relay"
)

func number(name string, fallback int) int {
	if value, err := strconv.Atoi(os.Getenv(name)); err == nil {
		return value
	}
	return fallback
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if len(os.Args) < 2 || os.Args[1] != "launch" {
		fmt.Fprintln(os.Stderr, "usage: chatcms-relay launch")
		os.Exit(2)
	}
	plan := relay.Plan{
		Run:          os.Getenv("CHATCMS_RUN"),
		Site:         "/site",
		Home:         "/run",
		Sockets:      "/run/chatcms",
		PreviewPort:  number("CHATCMS_PREVIEW_PORT", 8080),
		AgentPort:    number("CHATCMS_AGENT_PORT", 4096),
		GatewayPort:  number("CHATCMS_GATEWAY_PORT", 8091),
		Token:        os.Getenv("CHATCMS_AI_TOKEN"),
		OpenCode:     os.Getenv("CHATCMS_OPENCODE"),
		AgentWaitFor: 2 * time.Minute,
	}
	os.Unsetenv("CHATCMS_AI_TOKEN")
	if err := relay.Launch(ctx, plan, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
