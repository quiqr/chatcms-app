package content

import (
	"errors"
	"io/fs"
	"path"

	"github.com/quiqr/chatcms-app/internal/model"
	"github.com/quiqr/chatcms-app/internal/runtime"
)

type ShortcodeShape struct {
	Key       string
	Title     string
	Inner     string
	Controls  []Control
	Positions map[string]int
}

func ShortcodeDialect(m *model.Model, site fs.FS) string {
	if m.ShortcodeDialect != "" {
		return m.ShortcodeDialect
	}
	program := RuntimeOf(site)
	if model.IsShortcodeDialect(program) {
		return program
	}
	return ""
}

func RuntimeOf(site fs.FS) string {
	data, err := fs.ReadFile(site, runtime.File)
	if errors.Is(err, fs.ErrNotExist) {
		return "hugo"
	}
	if err != nil {
		return ""
	}
	definition, err := runtime.Parse(data)
	if err != nil || len(definition.Preview.Command) == 0 {
		return ""
	}
	return path.Base(definition.Preview.Command[0])
}

func ShortcodeShapes(declared []model.Shortcode) []ShortcodeShape {
	shapes := make([]ShortcodeShape, 0, len(declared))
	for _, shortcode := range declared {
		fields := make([]model.Field, 0, len(shortcode.Fields))
		positions := map[string]int{}
		for _, field := range shortcode.Fields {
			fields = append(fields, field.Field)
			if field.Position != nil {
				positions[field.Key] = *field.Position
			}
		}
		shape := ShortcodeShape{
			Key:      shortcode.Key,
			Title:    titleOr(shortcode.Title, shortcode.Key),
			Inner:    shortcode.Inner,
			Controls: shapeOf(fields),
		}
		if len(positions) > 0 {
			shape.Positions = positions
		}
		shapes = append(shapes, shape)
	}
	return shapes
}
