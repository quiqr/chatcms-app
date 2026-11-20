package frontmatter

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

func EmitJSON(m *Map) ([]byte, error) {
	var out strings.Builder
	if err := emitJSONValue(&out, m, 0, ""); err != nil {
		return nil, err
	}
	return []byte(out.String()), nil
}

func emitJSONValue(out *strings.Builder, value any, depth int, path string) error {
	switch typed := value.(type) {
	case *Map:
		keys := typed.emitOrder()
		if len(keys) == 0 {
			out.WriteString("{}")
			return nil
		}
		out.WriteString("{\n")
		for i, key := range keys {
			item, _ := typed.Get(key)
			out.WriteString(strings.Repeat("  ", depth+1))
			out.WriteString(jsString(key))
			out.WriteString(": ")
			if err := emitJSONValue(out, item, depth+1, join(path, key)); err != nil {
				return err
			}
			if i < len(keys)-1 {
				out.WriteString(",")
			}
			out.WriteString("\n")
		}
		out.WriteString(strings.Repeat("  ", depth))
		out.WriteString("}")
		return nil
	case []any:
		if len(typed) == 0 {
			out.WriteString("[]")
			return nil
		}
		out.WriteString("[\n")
		for i, item := range typed {
			out.WriteString(strings.Repeat("  ", depth+1))
			if err := emitJSONValue(out, item, depth+1, fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
			if i < len(typed)-1 {
				out.WriteString(",")
			}
			out.WriteString("\n")
		}
		out.WriteString(strings.Repeat("  ", depth))
		out.WriteString("]")
		return nil
	case nil:
		out.WriteString("null")
	case bool:
		out.WriteString(strconv.FormatBool(typed))
	case int64:
		out.WriteString(strconv.FormatInt(typed, 10))
	case float64:
		if math.IsNaN(typed) || math.IsInf(typed, 0) {
			out.WriteString("null")
			return nil
		}
		out.WriteString(jsNumber(typed))
	case time.Time:
		out.WriteString(jsString(typed.UTC().Format("2006-01-02T15:04:05.000Z")))
	case string:
		out.WriteString(jsString(typed))
	default:
		return unsupported(path, value)
	}
	return nil
}

func jsString(text string) string {
	var out strings.Builder
	out.WriteByte('"')
	for _, r := range text {
		switch r {
		case '"':
			out.WriteString(`\"`)
		case '\\':
			out.WriteString(`\\`)
		case '\b':
			out.WriteString(`\b`)
		case '\f':
			out.WriteString(`\f`)
		case '\n':
			out.WriteString(`\n`)
		case '\r':
			out.WriteString(`\r`)
		case '\t':
			out.WriteString(`\t`)
		default:
			if r < 0x20 {
				fmt.Fprintf(&out, `\u%04x`, r)
				continue
			}
			out.WriteRune(r)
		}
	}
	out.WriteByte('"')
	return out.String()
}

func jsNumber(value float64) string {
	if value == 0 {
		return "0"
	}
	magnitude := math.Abs(value)
	if magnitude >= 1e21 || magnitude < 1e-6 {
		text := strconv.FormatFloat(value, 'e', -1, 64)
		mantissa, exponent, _ := strings.Cut(text, "e")
		sign := exponent[:1]
		digits := strings.TrimLeft(exponent[1:], "0")
		if digits == "" {
			digits = "0"
		}
		return mantissa + "e" + sign + digits
	}
	return strconv.FormatFloat(value, 'f', -1, 64)
}
