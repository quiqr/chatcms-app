package workspace

import (
	"bytes"
	"strings"
	"sync"
)

type Step string

const (
	StepStarting Step = "starting-workspace"
	StepFetching Step = "fetching-site"
	StepBuilding Step = "building-preview"
)

var steps = []Step{StepStarting, StepFetching, StepBuilding}

type StepState string

const (
	StepWaiting StepState = "waiting"
	StepRunning StepState = "running"
	StepDone    StepState = "done"
	StepFailed  StepState = "failed"
)

type StepReport struct {
	Step  Step      `json:"step"`
	State StepState `json:"state"`
}

const logLines = 200

type progress struct {
	mu      sync.Mutex
	states  map[Step]StepState
	lines   []string
	partial bytes.Buffer
}

func newProgress() *progress {
	p := &progress{states: map[Step]StepState{}}
	for _, step := range steps {
		p.states[step] = StepWaiting
	}
	return p
}

func (p *progress) begin(step Step) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, earlier := range steps {
		if earlier == step {
			break
		}
		if p.states[earlier] == StepRunning || p.states[earlier] == StepWaiting {
			p.states[earlier] = StepDone
		}
	}
	p.states[step] = StepRunning
}

func (p *progress) finish(failed bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, step := range steps {
		switch {
		case p.states[step] == StepRunning && failed:
			p.states[step] = StepFailed
		case p.states[step] == StepRunning:
			p.states[step] = StepDone
		}
	}
	if p.partial.Len() > 0 {
		p.add(p.partial.String())
		p.partial.Reset()
	}
}

func (p *progress) Write(data []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.partial.Write(data)
	for {
		line, err := p.partial.ReadString('\n')
		if err != nil {
			p.partial.Reset()
			p.partial.WriteString(line)
			break
		}
		p.add(strings.TrimRight(line, "\r\n"))
	}
	return len(data), nil
}

func (p *progress) add(line string) {
	p.lines = append(p.lines, line)
	if len(p.lines) > logLines {
		p.lines = p.lines[len(p.lines)-logLines:]
	}
}

func (p *progress) logf(line string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.add(line)
}

func (p *progress) report() ([]StepReport, []string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	reports := make([]StepReport, 0, len(steps))
	for _, step := range steps {
		reports = append(reports, StepReport{Step: step, State: p.states[step]})
	}
	return reports, append([]string{}, p.lines...)
}
