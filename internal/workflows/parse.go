package workflows

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
)

const Directory = "quiqr/workflows"

type fileShape struct {
	On      string        `yaml:"on"`
	Steps   []stepShape   `yaml:"steps"`
	Outputs []outputShape `yaml:"outputs"`
}

type outputShape struct {
	Path string `yaml:"path"`
	To   string `yaml:"to"`
}

type stepShape struct {
	Name    string `yaml:"name"`
	Action  string `yaml:"action"`
	Run     string `yaml:"run"`
	Timeout string `yaml:"timeout"`
}

func Parse(file string, content []byte) (Workflow, error) {
	var shape fileShape
	if err := yaml.UnmarshalWithOptions(content, &shape, yaml.DisallowUnknownField()); err != nil {
		return Workflow{}, &Error{File: file, Reason: unknownKeyOr(err)}
	}

	workflow := Workflow{Trigger: Trigger(shape.On), File: file}
	if shape.On == "" {
		return Workflow{}, &Error{File: file, Key: "on", Reason: "a workflow has to say when it runs: " + listed(Triggers)}
	}
	if !workflow.Trigger.Known() {
		return Workflow{}, &Error{
			File:   file,
			Key:    "on",
			Reason: fmt.Sprintf("%q is not something that happens; the triggers are %s", shape.On, listed(Triggers)),
		}
	}

	for index, step := range shape.Steps {
		read, err := readStep(file, index+1, step)
		if err != nil {
			return Workflow{}, err
		}
		workflow.Steps = append(workflow.Steps, read)
	}

	if err := order(workflow); err != nil {
		return Workflow{}, err
	}
	for index, shape := range shape.Outputs {
		output, err := readOutput(workflow, index+1, shape)
		if err != nil {
			return Workflow{}, err
		}
		workflow.Outputs = append(workflow.Outputs, output)
	}
	return workflow, nil
}

func readOutput(workflow Workflow, number int, shape outputShape) (Output, error) {
	where := func(key, reason string) error {
		return &Error{File: workflow.File, Key: fmt.Sprintf("outputs[%d].%s", number, key), Reason: reason}
	}
	output := Output{Path: strings.TrimPrefix(shape.Path, "./"), To: Destination(shape.To)}
	if output.Path == "" {
		output.Path = "."
	}
	if shape.Path == "" {
		return Output{}, where("path", "an output has to say which part of the build it is; . is all of it")
	}
	if cleaned := path.Clean(output.Path); cleaned != output.Path || strings.HasPrefix(cleaned, "../") || cleaned == ".." || strings.HasPrefix(cleaned, "/") {
		return Output{}, where("path", fmt.Sprintf("%q has to be a path inside the build's output", shape.Path))
	}
	if _, err := path.Match(output.Path, ""); err != nil {
		return Output{}, where("path", fmt.Sprintf("%q is not a pattern", shape.Path))
	}
	if !output.To.Known() {
		return Output{}, where("to", fmt.Sprintf("%q is not a destination; the destinations are %s", shape.To, listed(Destinations)))
	}
	if !workflow.Does(Build) {
		return Output{}, where("path", "outputs come from a build, and this workflow does not build")
	}
	if output.To.Publishes() && !workflow.Does(Publish) {
		return Output{}, where("to", fmt.Sprintf("%q goes to publish, but this workflow does not publish", shape.Path))
	}
	return output, nil
}

func readStep(file string, number int, shape stepShape) (Step, error) {
	step := Step{Name: shape.Name, Action: Action(shape.Action), Run: shape.Run}

	switch {
	case shape.Action == "":
		return Step{}, &Error{File: file, Step: number, Key: "action", Reason: "a step has to do something: " + listed(Actions)}
	case !step.Action.Known():
		return Step{}, &Error{
			File:   file,
			Step:   number,
			Key:    "action",
			Reason: fmt.Sprintf("%q is not an action; the actions are %s", shape.Action, listed(Actions)),
		}
	case step.Action == Check && shape.Run == "":
		return Step{}, &Error{File: file, Step: number, Key: "run", Reason: "a check has to say what to run"}
	case step.Action != Check && shape.Run != "":
		return Step{}, &Error{
			File:   file,
			Step:   number,
			Key:    "run",
			Reason: fmt.Sprintf("only a %q step runs a command; %q is done by ChatCMS", Check, step.Action),
		}
	}

	if shape.Timeout != "" {
		limit, err := time.ParseDuration(shape.Timeout)
		if err != nil || limit <= 0 {
			return Step{}, &Error{File: file, Step: number, Key: "timeout", Reason: fmt.Sprintf("%q is not a length of time", shape.Timeout)}
		}
		step.Timeout = limit
	}
	return step, nil
}

func order(workflow Workflow) error {
	seen := map[Action]int{}
	for index, step := range workflow.Steps {
		number := index + 1
		if step.Action == Publish && number != len(workflow.Steps) {
			return &Error{File: workflow.File, Step: number, Reason: "nothing runs after publishing; put it last"}
		}
		if first, already := seen[step.Action]; already && step.Action != Check {
			return &Error{
				File:   workflow.File,
				Step:   number,
				Reason: fmt.Sprintf("%q already happened at step %d; it can only happen once", step.Action, first),
			}
		}
		seen[step.Action] = number
	}
	return nil
}

func unknownKeyOr(err error) string {
	said := err.Error()
	if strings.Contains(said, "unknown field") {
		return said + " (a key that is not part of the format is a mistake rather than something to ignore)"
	}
	return said
}

type Set map[Trigger]Workflow

func Load(site fs.FS) (Set, error) {
	entries, err := fs.ReadDir(site, Directory)
	if errors.Is(err, fs.ErrNotExist) {
		return Set{}, nil
	}
	if err != nil {
		return nil, &Error{File: Directory, Reason: err.Error()}
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !isYAML(entry.Name()) {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)

	set := Set{}
	for _, name := range names {
		file := path.Join(Directory, name)
		content, err := fs.ReadFile(site, file)
		if err != nil {
			return nil, &Error{File: file, Reason: err.Error()}
		}
		workflow, err := Parse(file, content)
		if err != nil {
			return nil, err
		}
		if already, taken := set[workflow.Trigger]; taken {
			return nil, &Error{
				File:   file,
				Key:    "on",
				Reason: fmt.Sprintf("%s already says what happens on %q; a trigger can only have one workflow", already.File, workflow.Trigger),
			}
		}
		set[workflow.Trigger] = workflow
	}
	return set, nil
}

func isYAML(name string) bool {
	extension := strings.ToLower(path.Ext(name))
	return extension == ".yaml" || extension == ".yml"
}
