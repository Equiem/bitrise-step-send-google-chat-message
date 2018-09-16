package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-tools/go-steputils/stepconf"
)

// Config ...
type Config struct {
	Debug bool `env:"is_debug_mode,opt[yes,no]"`

	// Message
	WebhookURL      stepconf.Secret `env:"webhook_url,required"`
	AppTitle        string          `env:"app_title"`
	ImageURL        string          `env:"image_url"`
	ImageURLOnError string          `env:"image_url_on_error"`
	Title           string          `env:"title"`
	TitleOnError    string          `env:"title_on_error"`
	QRCodeURL       string          `env:"qr_image_url"`
	Fields          string          `env:"fields"`
	Buttons         string          `env:"buttons"`
}

// success is true if the build is successful, false otherwise.
var success = os.Getenv("BITRISE_BUILD_STATUS") == "0"

// selectValue chooses the right value based on the result of the build.
func selectValue(ifSuccess, ifFailed string) string {
	if success || ifFailed == "" {
		return ifSuccess
	}
	return ifFailed
}

// ensureNewlines replaces all \n substrings with newline characters.
func ensureNewlines(s string) string {
	return strings.Replace(s, "\\n", "\n", -1)
}

func newCard(c Config) Card {
	sections := []Section{}

	fields := parseFields(c.Fields)
	if len(fields) > 0 {
		sections = append(sections, Section{Widgets: fields})
	}

	if len(c.QRCodeURL) > 0 {
		widget := Widget{Image: Image{URL: c.QRCodeURL}}
		sections = append(sections, Section{Widgets: []Widget{widget}})
	}

	buttons := parseButtons(c.Buttons)
	if len(buttons) > 0 {
		sections = append(sections, Section{Widgets: buttons})
	}

	card := Card{
		Header: Header{
			Title:    selectValue(c.Title, c.TitleOnError),
			SubTitle: c.AppTitle,
			IconURL:  selectValue(c.ImageURL, c.ImageURLOnError),
		},
		Sections: sections,
	}
	return card
}

func newMessage(c Config) Message {
	msg := Message{
		Cards: []Card{newCard(c)},
	}
	return msg
}

func postMessage(webhookURL string, msg Message) error {
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	log.Debugf("Request to Google Chat: %s\n", b)

	resp, err := http.Post(webhookURL, "application/json", bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("failed to send the request: %s", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); err == nil {
			err = cerr
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("server error: %s, failed to read response: %s", resp.Status, err)
		}
		return fmt.Errorf("server error: %s, response: %s", resp.Status, body)
	}

	return nil
}

func main() {
	var conf Config

	if err := stepconf.Parse(&conf); err != nil {
		log.Errorf("Error: %s\n", err)
		os.Exit(1)
	}
	stepconf.Print(conf)
	log.SetEnableDebugLog(conf.Debug)

	msg := newMessage(conf)
	if err := postMessage(string(conf.WebhookURL), msg); err != nil {
		log.Errorf("Error: %s", err)
		os.Exit(1)
	}

	log.Donef("\nGoogle Chat message successfully sent! 🚀\n")
}
