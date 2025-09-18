package markdown

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

func Table(headers []string, rows [][]string) string {
	colWidths := make([]int, len(headers))
	for i, h := range headers {
		colWidths[i] = utf8.RuneCountInString(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if w := utf8.RuneCountInString(cell); w > colWidths[i] {
				colWidths[i] = w
			}
		}
	}

	pad := func(s string, width int) string {
		l := utf8.RuneCountInString(s)
		if l >= width {
			return s
		}
		return s + strings.Repeat(" ", width-l)
	}

	var b strings.Builder

	for i, h := range headers {
		if i > 0 {
			b.WriteString(" | ")
		}
		b.WriteString(pad(h, colWidths[i]))
	}
	b.WriteString("\n")

	for i := range headers {
		if i > 0 {
			b.WriteString(" | ")
		}
		b.WriteString(strings.Repeat("-", colWidths[i]))
	}
	b.WriteString("\n")

	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				b.WriteString(" | ")
			}
			b.WriteString(pad(cell, colWidths[i]))
		}
		b.WriteString("\n")
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
