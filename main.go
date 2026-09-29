package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"ssl-monitor/slack"
)

type channelValue struct {
	ShopID             string   `json:"shopId"`
	ParentDomain       string   `json:"parentDomain"`
	DelegatedDomain    string   `json:"delegatedDomain"`
	TrackingDomain     string   `json:"trackingDomain"`
	NameServers        []string `json:"nameServers"`
	DelegationStateDNS string   `json:"delegationStateDNS"`
	DelegationStateSSL string   `json:"delegationStateSSL"`
	InternalZoneName   string   `json:"internalZoneName"`
}

type issue struct {
	ID       int64
	Fallback string
	Blocks   []slack.Block
}

func envDuration(key string, def time.Duration) time.Duration {
	s := os.Getenv(key)
	if s == "" {
		return def
	}
	d, err := time.ParseDuration(s)
	if err != nil || d < 0 {
		log.Fatalf("invalid %s %q (use e.g. 5s, 30m, 1h)", key, s)
	}
	return d
}

func main() {
	db, err := sql.Open("mysql", os.Getenv("DB_DSN"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	checkInterval := envDuration("CHECK_INTERVAL", time.Hour)
	notifyDelay := envDuration("NOTIFY_DELAY", 5*time.Second)
	log.Printf("starting: check every %s, %s between notifications", checkInterval, notifyDelay)

	for {
		run(db, notifyDelay)
		log.Printf("sleeping %s until next check", checkInterval)
		time.Sleep(checkInterval)
	}
}

func run(db *sql.DB, notifyDelay time.Duration) {
	issues, total, err := findIssues(db)
	if err != nil {
		log.Println("check error:", err)
		return
	}
	log.Printf("checked %d channels, %d not valid", total, len(issues))

	for i, is := range issues {
		if err := slack.SendBlocks(is.Fallback, is.Blocks); err != nil {
			log.Printf("slack error for channel %d: %v", is.ID, err)
		}
		if i < len(issues)-1 {
			time.Sleep(notifyDelay)
		}
	}
}

func findIssues(db *sql.DB) ([]issue, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	rows, err := db.QueryContext(ctx,
		`SELECT ClusterChannelID, InternalClusterID, ChannelValue, InsertDate
		 FROM ClusterData
		 WHERE ChannelTypeID = 18`)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var issues []issue
	total := 0

	for rows.Next() {
		var id, clusterID int64
		var raw []byte
		var inserted sql.NullString
		if err := rows.Scan(&id, &clusterID, &raw, &inserted); err != nil {
			log.Println("scan error:", err)
			continue
		}
		total++

		var v channelValue
		if err := json.Unmarshal(raw, &v); err != nil {
			fb, blocks := buildBadJSONAlert(id, clusterID, err)
			issues = append(issues, issue{id, fb, blocks})
			continue
		}

		if v.DelegationStateDNS == "valid" && v.DelegationStateSSL == "valid" {
			continue
		}

		fb, blocks := buildAlert(id, clusterID, inserted.String, v)
		issues = append(issues, issue{id, fb, blocks})
	}
	return issues, total, rows.Err()
}