package main // import "github.com/karlkfi/slackstack"

import (
	log "github.com/Sirupsen/logrus"
	"github.com/laktek/Stack-on-Go/stackongo"

	"fmt"
	"time"
)

func main() {
	site := "stackoverflow"
	session := stackongo.NewSession(site)

	_, err := session.Info()
	if err != nil {
		log.Fatalf("Failed to retrieve %s session info: %v", site, err)
	}

	fmt.Printf("No. of DC/OS Questions (Last 30 days):\n")

	from_date := time.Now().Unix() - (60 * 60 * 24 * 30)

	//set the common params
	params := make(stackongo.Params)
	params.Add("filter", "total")
	params.Add("fromdate", from_date)
	params.Add("tagged", "dcos")

	results, err := session.AllQuestions(params)
	if err != nil {
		log.Fatalf("Failed to query %s: %v", site, err)
	}

	fmt.Printf("%v\n", results.Total)
}
