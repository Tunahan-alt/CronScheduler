job "ssl-monitor" {
  datacenters = ["dc1"]
  type        = "service"

  group "monitor" {
    task "ssl-monitor" {
      driver = "docker"

      artifact {
        source      = "git::https://github.com/YOUR_ORG/YOUR_REPO.git"
        destination = "local/app"
      }

      config {
        image          = "golang:1.23-alpine"
        auth_soft_fail = true
        command        = "sh"
        args           = ["-c", "cd local/app && go build -o /tmp/ssl-monitor . && /tmp/ssl-monitor"]
      }

      env {
        CHECK_INTERVAL = "1h"
        NOTIFY_DELAY   = "5s"
        DB_DSN         = "user:password@tcp(db_datafy.g.quanticfy.io:3306)/houston_datafy"
      }

      resources {
        cpu    = 100
        memory = 128
      }

      restart {
        attempts = 3
        delay    = "30s"
        interval = "5m"
        mode     = "delay"
      }
    }
  }
}