package model

import "fmt"

var ShortcodeDialects = []string{"hugo", "quarto"}

var ShortcodeFieldTypes = []string{
	"string", "markdown", "easymde", "number", "slider", "boolean",
	"date", "select", "color", "image-select",
}

func IsShortcodeDialect(name string) bool {
	return contains(ShortcodeDialects, name)
}

func contains(list []string, name string) bool {
	for _, one := range list {
		if one == name {
			return true
		}
	}
	return false
}

func decodeShortcode(file string, index int, tree map[string]any) (Shortcode, error) {
	where := pathOf("shortcodes", index)
	r := newReader(file, where, tree)

	shortcode := Shortcode{
		Key:   r.requiredStr("key"),
		Title: r.str("title"),
		Inner: r.str("inner"),
	}
	raw := r.list("fields")
	if r.err != nil {
		return Shortcode{}, r.err
	}
	if shortcode.Inner != "" && shortcode.Inner != "markdown" && shortcode.Inner != "text" {
		return Shortcode{}, atf(file, join(where, "inner"), "inner is %q; it can be markdown or text", shortcode.Inner)
	}

	fields, err := decodeFields(file, join(where, "fields"), raw)
	if err != nil {
		return Shortcode{}, err
	}

	taken := map[int]string{}
	for at, field := range fields {
		path := fmt.Sprintf("%s[%d]", join(where, "fields"), at)
		if !contains(ShortcodeFieldTypes, field.Type) {
			return Shortcode{}, atf(file, path, "a shortcode field holds one value; %q cannot be one", field.Type)
		}
		one := ShortcodeField{Field: field}
		if value, ok := field.Extra["position"]; ok {
			position, ok := wholeNumber(value)
			if !ok || position < 0 {
				return Shortcode{}, atf(file, join(path, "position"), "a position is a whole number from 0")
			}
			if other, ok := taken[position]; ok {
				return Shortcode{}, atf(file, join(path, "position"), "%q and %q both take position %d", other, field.Key, position)
			}
			taken[position] = field.Key
			one.Position = &position
			delete(one.Extra, "position")
			if len(one.Extra) == 0 {
				one.Extra = nil
			}
		}
		shortcode.Fields = append(shortcode.Fields, one)
	}
	for position := range taken {
		if _, ok := taken[position-1]; position > 0 && !ok {
			return Shortcode{}, atf(file, join(where, "fields"), "positions start at 0 and leave no gaps; %d has nothing before it", position)
		}
	}

	shortcode.Extra = r.extra()
	return shortcode, nil
}

func wholeNumber(value any) (int, bool) {
	switch number := value.(type) {
	case int:
		return number, true
	case int64:
		return int(number), true
	case uint64:
		return int(number), true
	case float64:
		if number == float64(int(number)) {
			return int(number), true
		}
	}
	return 0, false
}
