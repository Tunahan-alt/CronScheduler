package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const relayURL = "https://qslack-937194857721.europe-west1.run.app?channel=C09TAV1T0AD"

type Block = map[string]interface{}

var client = &http.Client{Timeout: 15 * time.Second}

func post(payload map[string]interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	resp, err := client.Post(relayURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("sending Slack notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack notification failed: %s", resp.Status)
	}
	return nil
}

// SendNotification sends a plain text message.
func SendNotification(message string) error {
	return SendBlocks(message, []Block{
		Divider(), Text(message), Divider(),
	})
}

// SendBlocks sends rich blocks. fallback is shown in push notifications.
func SendBlocks(fallback string, blocks []Block) error {
	return post(map[string]interface{}{
		"text":   fallback,
		"blocks": blocks,
	})
}

// ---- block helpers ----

func Divider() Block { return Block{"type": "divider"} }

func Header(text string) Block {
	return Block{"type": "header", "text": Block{"type": "plain_text", "text": text, "emoji": true}}
}

func Text(text string) Block {
	return Block{"type": "section", "text": Block{"type": "mrkdwn", "text": text}}
}

// Fields renders a two-column grid (Slack allows max 10 fields).
func Fields(fields ...string) Block {
	fs := make([]Block, 0, len(fields))
	for _, f := range fields {
		fs = append(fs, Block{"type": "mrkdwn", "text": f})
	}
	return Block{"type": "section", "fields": fs}
}

func Context(text string) Block {
	return Block{"type": "context", "elements": []Block{{"type": "mrkdwn", "text": text}}}
}