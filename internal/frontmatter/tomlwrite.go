package frontmatter

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const tomlSpace = "  "

var tomlBareKey = regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)

type tomlStep struct {
	key   string
	index int
	named bool
}

type tomlWriter struct {
	lines       []string
	path        []tomlStep
	level       int
	inTableList bool
}

func EmitTOML(m *Map) ([]byte, error) {
	writer := &tomlWriter{}
	if err := writer.traverse(tomlStep{named: true}, m); err != nil {
		return nil, err
	}
	return []byte(strings.Join(writer.lines, "\n")), nil
}

func tomlKey(key string) string {
	if tomlBareKey.MatchString(key) {
		return key
	}
	return jsString(key)
}

func (w *tomlWriter) header() string {
	var parts []string
	for _, step := range w.path {
		if step.named {
			parts = append(parts, tomlKey(step.key))
		}
	}
	return strings.Join(parts, ".")
}

func (w *tomlWriter) where() string {
	var parts []string
	for _, step := range w.path {
		if step.named {
			parts = append(parts, step.key)
		} else {
			parts = append(parts, fmt.Sprintf("[%d]", step.index))
		}
	}
	if len(parts) == 0 {
		return "<root>"
	}
	return strings.Join(parts, ".")
}

func tomlIndent(text string, level int) string {
	if level == 0 {
		return text
	}
	padding := strings.Repeat(tomlSpace, level)
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = padding + line
		}
	}
	return strings.Join(lines, "\n")
}

func containsTables(items []any) bool {
	for _, item := range items {
		if _, ok := item.(*Map); ok {
			return true
		}
	}
	return false
}

func (w *tomlWriter) keyValue(key string, value any) error {
	if key == "" {
		return errors.New(w.where() + ": Key cannot be an empty string.")
	}
	text, ok, err := w.value(value, false)
	if err != nil || !ok {
		return err
	}
	w.lines = append(w.lines, tomlIndent(tomlKey(key)+" = "+text, w.level))
	return nil
}

func (w *tomlWriter) traverse(step tomlStep, value any) error {
	switch typed := value.(type) {
	case *Map:
		if !(step.named && step.key == "" && len(w.path) == 0) {
			if len(w.lines) > 0 {
				w.lines = append(w.lines, "")
			}
			line := "[" + w.header() + "]"
			if w.inTableList {
				line = "[" + line + "]"
			}
			w.lines = append(w.lines, tomlIndent(line, w.level))
		}
		outer := w.inTableList
		w.inTableList = false
		type later struct {
			key    string
			value  any
			indent bool
		}
		var tables []later
		for _, key := range typed.emitOrder() {
			item, _ := typed.Get(key)
			list, isList := item.([]any)
			_, isTable := item.(*Map)
			if (isList && containsTables(list)) || isTable {
				tables = append(tables, later{key: key, value: item, indent: len(w.path) > 0})
				continue
			}
			w.path = append(w.path, tomlStep{key: key, named: true})
			err := w.traverse(tomlStep{key: key, named: true}, item)
			w.path = w.path[:len(w.path)-1]
			if err != nil {
				return err
			}
		}
		for _, table := range tables {
			w.path = append(w.path, tomlStep{key: table.key, named: true})
			if table.indent {
				w.level++
			}
			err := w.traverse(tomlStep{key: table.key, named: true}, table.value)
			if table.indent {
				w.level--
			}
			w.path = w.path[:len(w.path)-1]
			if err != nil {
				return err
			}
		}
		w.inTableList = outer
		return nil
	case []any:
		if !step.named {
			return nil
		}
		if !containsTables(typed) {
			return w.keyValue(step.key, typed)
		}
		outer := w.inTableList
		w.inTableList = true
		for i, item := range typed {
			w.path = append(w.path, tomlStep{index: i})
			err := w.traverse(tomlStep{index: i}, item)
			w.path = w.path[:len(w.path)-1]
			if err != nil {
				return err
			}
		}
		w.inTableList = outer
		return nil
	default:
		if step.named {
			return w.keyValue(step.key, value)
		}
		return nil
	}
}

func jsKind(value any) string {
	switch value.(type) {
	case bool:
		return "boolean"
	case string:
		return "string"
	case int64, float64:
		return "number"
	case []any:
		return "array"
	case time.Time:
		return "date"
	case nil:
		return "null"
	default:
		return "object"
	}
}

func (w *tomlWriter) value(value any, inPair bool) (string, bool, error) {
	switch typed := value.(type) {
	case nil:
		return "", false, nil
	case bool:
		return strconv.FormatBool(typed), true, nil
	case string:
		return jsString(typed), true, nil
	case int64:
		return tomlNumber(float64(typed)), true, nil
	case float64:
		if math.IsNaN(typed) || math.IsInf(typed, 0) {
			return "", false, errors.New(w.where() + ": Number must be finite.")
		}
		return tomlNumber(typed), true, nil
	case time.Time:
		return typed.UTC().Format("2006-01-02T15:04:05.000Z"), true, nil
	case []any:
		if len(typed) > 1 {
			first := jsKind(typed[0])
			for _, item := range typed[1:] {
				if item != nil && jsKind(item) != first {
					return "", false, errors.New(w.where() + ": Array cannot contain values of different types.")
				}
			}
		}
		inner := inPair || containsTables(typed)
		var lines []string
		for i, item := range typed {
			w.path = append(w.path, tomlStep{index: i})
			text, ok, err := w.value(item, inner)
			w.path = w.path[:len(w.path)-1]
			if err != nil {
				return "", false, err
			}
			if ok {
				lines = append(lines, text)
			}
		}
		if !inner && len(lines) > 0 {
			return "[\n" + tomlIndent(strings.Join(lines, ",\n"), 1) + "\n]", true, nil
		}
		return "[" + strings.Join(lines, ", ") + "]", true, nil
	case *Map:
		var pairs []string
		for _, key := range typed.emitOrder() {
			item, _ := typed.Get(key)
			if item == nil {
				continue
			}
			if key == "" {
				return "", false, errors.New(w.where() + ": Key cannot be an empty string.")
			}
			w.path = append(w.path, tomlStep{key: key, named: true})
			text, ok, err := w.value(item, true)
			w.path = w.path[:len(w.path)-1]
			if err != nil {
				return "", false, err
			}
			if ok {
				pairs = append(pairs, tomlKey(key)+" = "+text)
			}
		}
		return "{" + strings.Join(pairs, ", ") + "}", true, nil
	default:
		return "", false, unsupported(w.where(), value)
	}
}

var tomlExponent = regexp.MustCompile(`([eE])(\d)`)

func tomlNumber(value float64) string {
	if value == 0 {
		value = 0
	}
	var text string
	if value == math.Trunc(value) {
		if math.Abs(value) >= 1e21 {
			text = jsNumber(value)
		} else {
			text = strconv.FormatFloat(value, 'f', 1, 64)
		}
	} else {
		text = jsNumber(value)
	}
	text = tomlExponent.ReplaceAllString(text, "$1+$2")
	if !strings.ContainsAny(text, ".eE") {
		text += ".0"
	}
	return text
}
