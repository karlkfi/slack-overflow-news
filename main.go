package main // import "github.com/karlkfi/slackstack"

import (
	log "github.com/Sirupsen/logrus"
	"github.com/laktek/Stack-on-Go/stackongo"
)

func main() {
	site := "stackoverflow"
	session := stackongo.NewSession(site)

	_, err := session.Info()
	if err != nil {
		log.Fatalf("Failed to retrieve %s session info: %v", site, err)
	}
}
