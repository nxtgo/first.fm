package components

import "fmt"

type ComponentType int
type ButtonStyle int

const (
	TypeActionRow   ComponentType = 1
	TypeButton      ComponentType = 2
	TypeTextDisplay ComponentType = 10
	TypeThumbnail   ComponentType = 11
	TypeSection     ComponentType = 9
	TypeDivider     ComponentType = 14
	TypeContainer   ComponentType = 17
)
const (
	ButtonStylePrimary ButtonStyle = iota + 1
	ButtonStyleSecondary
	ButtonStyleSuccess
	ButtonStyleDanger
	ButtonStyleLink
)

type Component interface {
	componentType() ComponentType
}

type ContainerComponent interface {
	Component
	isContainer()
}

type Container struct {
	Type        ComponentType `json:"type"`
	AccentColor int           `json:"accent_color,omitempty"`
	Components  []Component   `json:"components"`
}

func (c *Container) componentType() ComponentType { return TypeContainer }
func (c *Container) isContainer()                 {}

type Section struct {
	Type       ComponentType `json:"type"`
	Components []Component   `json:"components"`
	Accessory  Component     `json:"accessory,omitempty"`
}

func (s *Section) componentType() ComponentType { return TypeSection }
func (s *Section) isContainer()                 {}

type TextDisplay struct {
	Type    ComponentType `json:"type"`
	Content string        `json:"content"`
}

func (t *TextDisplay) componentType() ComponentType { return TypeTextDisplay }

type Thumbnail struct {
	Type  ComponentType `json:"type"`
	Media Media         `json:"media"`
}

func (t *Thumbnail) componentType() ComponentType { return TypeThumbnail }

type Media struct {
	URL string `json:"url"`
}

type Divider struct {
	Type    ComponentType `json:"type"`
	Divider bool          `json:"divider"`
}

func (d *Divider) componentType() ComponentType { return TypeDivider }

type ActionRow struct {
	Type       ComponentType `json:"type"`
	Components []Component   `json:"components"`
}

func (a *ActionRow) componentType() ComponentType { return TypeActionRow }

type Button struct {
	Type     ComponentType `json:"type"`
	Style    ButtonStyle   `json:"style"`
	Label    string        `json:"label"`
	CustomID *string       `json:"custom_id,omitempty"`
	Emoji    *Emoji        `json:"emoji,omitempty"`
	URL      *string       `json:"url,omitempty"`
}

func (b *Button) componentType() ComponentType { return TypeButton }

type Emoji struct {
	Name     *string `json:"name,omitempty"`
	ID       string  `json:"id"`
	Animated bool    `json:"animated"`
}

func NewContainer(accent int, children ...Component) *Container {
	return &Container{Type: TypeContainer, AccentColor: accent, Components: children}
}

func NewSection(children ...Component) *Section {
	return &Section{Type: TypeSection, Components: children}
}

func (s *Section) WithAccessory(accessory Component) *Section {
	s.Accessory = accessory
	return s
}

func NewTextDisplay(content string) *TextDisplay {
	return &TextDisplay{Type: TypeTextDisplay, Content: content}
}

func NewTextDisplayf(content string, args ...any) *TextDisplay {
	return &TextDisplay{Type: TypeTextDisplay, Content: fmt.Sprintf(content, args...)}
}

func NewThumbnail(url string) *Thumbnail {
	return &Thumbnail{Type: TypeThumbnail, Media: Media{URL: url}}
}

func NewDivider() *Divider {
	return &Divider{Type: TypeDivider, Divider: true}
}

func NewActionRow(children ...Component) *ActionRow {
	return &ActionRow{Type: TypeActionRow, Components: children}
}

func NewButton(style ButtonStyle, label string, customID *string) *Button {
	return &Button{Type: TypeButton, Style: style, Label: label, CustomID: customID}
}

func (b *Button) WithURL(url string) *Button {
	b.URL = &url
	return b
}

func (b *Button) WithEmoji(id string) *Button {
	b.Emoji = &Emoji{ID: id}
	return b
}
