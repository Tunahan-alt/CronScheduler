package main

import (
	"fmt"
	"strings"
	"time"

	"ssl-monitor/slack"
)

func statusLabel(s string) string {
	switch {
	case s == "valid":
		return "✅ valid"
	case s == "":
		return "❔ missing"
	case strings.Contains(strings.ToLower(s), "pending"):
		return "⏳ " + s
	default:
		return "❌ " + s
	}
}

func orDash(s string) string {
	if s == "" {
		return "–"
	}
	return s
}

func humanAge(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	return fmt.Sprintf("%dh", hours)
}

func buildAlert(channelID, clusterID int64, insertDate string, v channelValue) (string, []slack.Block) {
	title := "🔴 " + orDash(v.ShopID)

	footer := fmt.Sprintf("Channel #%d", channelID)
	if t, err := time.Parse("2006-01-02 15:04:05", insertDate); err == nil {
		footer += " · pending for " + humanAge(time.Since(t))
	}

	blocks := []slack.Block{
		slack.Header(title),
		slack.Text(fmt.Sprintf("*Domain:* %s\n*DNS:* %s   *SSL:* %s",
			orDash(v.DelegatedDomain),
			statusLabel(v.DelegationStateDNS),
			statusLabel(v.DelegationStateSSL))),
		slack.Context(footer),
		slack.Divider(),
	}
	return title, blocks
}

func buildBadJSONAlert(channelID, clusterID int64, err error) (string, []slack.Block) {
	title := fmt.Sprintf("🟠 Channel #%d: unreadable ChannelValue", channelID)
	return title, []slack.Block{
		slack.Header(title),
		slack.Divider(),
	}
}