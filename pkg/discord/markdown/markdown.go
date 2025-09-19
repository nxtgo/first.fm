package markdown

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

// getLongerStr returns the longer of two strings based on rune count.
func getLongerStr(a, b string) string {
	if utf8.RuneCountInString(a) > utf8.RuneCountInString(b) {
		return a
	}
	return b
}

// GenerateTable creates a key-value aligned table from a slice of [2]string.
func GenerateTable(input [][2]string) string {
	if len(input) == 0 {
		return ""
	}

	// Find the longest key
	longest := input[0][0]
	for _, pair := range input {
		longest = getLongerStr(longest, pair[0])
	}
	longestLen := utf8.RuneCountInString(longest)

	var b strings.Builder
	for _, pair := range input {
		key, value := pair[0], pair[1]
		padding := longestLen - utf8.RuneCountInString(key)
		b.WriteString(strings.Repeat(" ", padding))
		b.WriteString(key)
		b.WriteString(": ")
		b.WriteString(value)
		b.WriteByte('\n')
	}
	return b.String()
}

// GenerateList creates a table-like list with headers.
func GenerateList(keyName, valueName string, values [][2]string) string {
	return GenerateListFixedDelim(keyName, valueName, values, utf8.RuneCountInString(keyName), utf8.RuneCountInString(valueName))
}

// GenerateListFixedDelim creates a table-like list with headers and custom delimiter lengths.
func GenerateListFixedDelim(keyName, valueName string, values [][2]string, keyDelimLen, valueDelimLen int) string {
	if len(values) == 0 {
		return ""
	}

	// Find the longest between header and keys
	longest := getLongerStr(keyName, values[0][0])
	for _, pair := range values {
		longest = getLongerStr(longest, pair[0])
	}
	longestLen := utf8.RuneCountInString(longest)

	var b strings.Builder

	// Header
	b.WriteString(" ")
	b.WriteString(strings.Repeat(" ", longestLen-utf8.RuneCountInString(keyName)))
	b.WriteString(keyName)
	b.WriteByte('\t')
	b.WriteString(valueName)
	b.WriteByte('\n')

	// Delimiter row
	b.WriteString(" ")
	b.WriteString(strings.Repeat(" ", longestLen-utf8.RuneCountInString(keyName)))
	b.WriteString(strings.Repeat("-", keyDelimLen))
	b.WriteByte('\t')
	b.WriteString(strings.Repeat("-", valueDelimLen))

	// Values
	for _, pair := range values {
		key, value := pair[0], pair[1]
		b.WriteByte('\n')
		b.WriteString(" ")
		b.WriteString(strings.Repeat(" ", longestLen-utf8.RuneCountInString(key)))
		b.WriteString(key)
		b.WriteByte('\t')
		b.WriteString(value)
	}

	return b.String()
}

type TimestampStyle int

const (
	FullLong TimestampStyle = iota
	FullShort
	DateLong
	DateShort
	TimeLong
	TimeShort
	Relative
)

func (s TimestampStyle) String() string {
	switch s {
	case FullLong:
		return "F"
	case FullShort:
		return "f"
	case DateLong:
		return "D"
	case DateShort:
		return "d"
	case TimeLong:
		return "T"
	case TimeShort:
		return "t"
	case Relative:
		return "R"
	}
	return ""
}

type MD string

func cut(s string, to int) string {
	if utf8.RuneCountInString(s) <= to {
		return s
	}
	var b strings.Builder
	n := 0
	for _, r := range s {
		if n >= to {
			break
		}
		b.WriteRune(r)
		n++
	}
	return b.String()
}

func (m MD) EscapeItalics() string {
	s := cut(string(m), 1998)
	var b strings.Builder
	for _, r := range s {
		if r == '_' || r == '*' {
			b.WriteRune('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}

func (m MD) EscapeBold() string       { return strings.ReplaceAll(cut(string(m), 1998), "**", "\\*\\*") }
func (m MD) EscapeCodeString() string { return strings.ReplaceAll(cut(string(m), 1998), "`", "'") }
func (m MD) EscapeCodeBlock(lang string) string {
	return strings.ReplaceAll(cut(string(m), 1988-len(lang)), "```", "`\u200b`\u200b`")
}
func (m MD) EscapeSpoiler() string { return strings.ReplaceAll(cut(string(m), 1996), "__", "||") }
func (m MD) EscapeStrikethrough() string {
	return strings.ReplaceAll(cut(string(m), 1996), "~~", "\\~\\~")
}
func (m MD) EscapeUnderline() string { return strings.ReplaceAll(cut(string(m), 1996), "||", "__") }

func (m MD) Italics() string    { return "_" + m.EscapeItalics() + "_" }
func (m MD) Bold() string       { return "**" + m.EscapeBold() + "**" }
func (m MD) CodeString() string { return "`" + m.EscapeCodeString() + "`" }
func (m MD) CodeBlock(lang string) string {
	return "```" + lang + "\n" + m.EscapeCodeBlock(lang) + "\n```"
}
func (m MD) Spoiler() string       { return "||" + m.EscapeItalics() + "||" }
func (m MD) Strikethrough() string { return "~~" + m.EscapeItalics() + "~~" }
func (m MD) Underline() string     { return "__" + m.EscapeItalics() + "__" }
func (m MD) URL(url string, comment *string) string {
	if comment != nil {
		return "[" + string(m) + "](" + url + " '" + *comment + "')"
	}
	return "[" + string(m) + "](" + url + ")"
}
func (m MD) Timestamp(seconds int, style TimestampStyle) string {
	return "<t:" + strconv.Itoa(seconds) + ":" + style.String() + ">"
}
func (m MD) Subtext() string { return "-# " + string(m) }

func ParseCodeBlock(s string) string {
	t := strings.TrimSpace(s)
	if strings.HasPrefix(t, "```") && strings.HasSuffix(t, "```") {
		r := strings.ReplaceAll(t, "\n", "\n ")
		parts := strings.Split(r, " ")
		if len(parts) > 1 {
			joined := strings.Join(parts[1:], " ")
			return joined[:len(joined)-3]
		}
		return ""
	}
	if strings.HasPrefix(t, "`") && strings.HasSuffix(t, "`") {
		return t[1 : len(t)-1]
	}
	return s
}
