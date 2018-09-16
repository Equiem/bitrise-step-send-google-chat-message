package main

import (
	"encoding/json"
	"strings"
)

// Message to post to a google chat room.
type Message struct {
	Cards []Card `json:"cards,omitempty"`
}

// Card format of a google chat message.
// See also: https://developers.google.com/hangouts/chat/reference/message-formats/cards
type Card struct {
	Header   Header    `json:"header,omitempty"`
	Sections []Section `json:"sections,omitempty"`
}

// Header section
type Header struct {
	Title    string `json:"title,omitempty"`
	SubTitle string `json:"subtitle,omitempty"`
	IconURL  string `json:"imageUrl,omitempty"`
}

// Section of a card
type Section struct {
	Text    string   `json:"header,omitempty"`
	Widgets []Widget `json:"widgets,omitempty"`
}

// Widget abstract
type Widget struct {
	Buttons []TextButton
	Field   Field
	Image   Image
}

// Field Widget
type Field struct {
	Label   string `json:"topLabel"`
	Content string `json:"content"`
}

// Image Widget
type Image struct {
	URL string `json:"imageUrl,omitempty"`
}

// TextButton Widget
type TextButton struct {
	Button Button `json:"textButton"`
}

// Button Widget
type Button struct {
	Text    string `json:"text"`
	OnClick struct {
		OpenLink struct {
			URL string `json:"url"`
		} `json:"openLink"`
	} `json:"onClick"`
}

// Equal checks if a Image is equal to antoher Image
func (s Image) Equal(o Image) bool {
	if s.URL != o.URL {
		return false
	}
	return true
}

// IsEmpty checks if a Image is empty
func (s Image) IsEmpty() bool {
	return s.Equal(Image{})
}

// Equal checks if a Field is equal to antoher Field
func (s Field) Equal(o Field) bool {
	if s.Label != o.Label {
		return false
	}
	if s.Content != o.Content {
		return false
	}
	return true
}

// IsEmpty checks if a Field is empty
func (s Field) IsEmpty() bool {
	return s.Equal(Field{})
}

// MarshalJSON implements json.Marshaler.MarshalJSON.
func (f Widget) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{})
	if !f.Image.IsEmpty() {
		m["image"] = f.Image
	} else if !f.Field.IsEmpty() {
		m["keyValue"] = f.Field
	} else {
		m["buttons"] = f.Buttons
	}
	return json.Marshal(m)
}

// parseFields parses a string into Field widgets
func parseFields(s string) (fs []Widget) {
	for _, p := range pairs(s) {
		fs = append(fs, Widget{Field: Field{Label: p[0], Content: p[1]}})
	}
	return
}

// parseButtons parses a string into a TextButton widget
func parseButtons(s string) (bs []Widget) {
	buttons := []TextButton{}
	for _, p := range pairs(s) {
		button := Button{
			Text: p[0],
		}
		button.OnClick.OpenLink.URL = p[1]
		buttons = append(buttons, TextButton{Button: button})
	}
	bs = append(bs, Widget{Buttons: buttons})
	return
}

// pairs slices every lines in s into two substrings separated by the first pipe
// character and returns a slice of those pairs.
func pairs(s string) [][2]string {
	var ps [][2]string
	for _, line := range strings.Split(s, "\n") {
		a := strings.SplitN(line, "|", 2)
		if len(a) == 2 && a[0] != "" && a[1] != "" {
			ps = append(ps, [2]string{a[0], a[1]})
		}
	}
	return ps
}
