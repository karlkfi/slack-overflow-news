package main // import "github.com/karlkfi/slack-overflow-news"

import (
	log "github.com/Sirupsen/logrus"
	"github.com/laktek/Stack-on-Go/stackongo"
	"github.com/nlopes/slack"
	"github.com/kelseyhightower/envconfig"

	"fmt"
	"time"
)

const envPrefix = "SS"

type Config struct {
	StackSite string	`required:"true" envconfig:"STACK_SITE"`
	StackTags string	`required:"true" envconfig:"STACK_TAGS"`
	StackPoll time.Duration	`required:"false" envconfig:"STACK_POLL" default:"30s"`

	SlackToken string	`required:"true" envconfig:"SLACK_TOKEN"`
	SlackUserName string	`required:"true" envconfig:"SLACK_USERNAME"`
	SlackChannel string	`required:"true" envconfig:"SLACK_CHANNEL"`
	SlackHistory int	`required:"false" envconfig:"SLACK_HISTORY" default:"30"`
	SlackDebug bool		`required:"false" envconfig:"SLACK_DEBUG" default:"false"`
}

func main() {
	var config Config
	err := envconfig.Process(envPrefix, &config)
	if err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	log.Infof("Config: %+v", config)

	session := stackongo.NewSession(config.StackSite)

	/*
	_, err = session.Info()
	if err != nil {
		log.Fatalf("Failed to retrieve '%s' session info: %v", config.StackSite, err)
	}
	*/

	api := slack.New(config.SlackToken)

	api.SetDebug(config.SlackDebug)

	lastUpdate := time.Now().AddDate(0, 0, -config.SlackHistory)

	for {
		reqParams := make(stackongo.Params)
		reqParams.Add("fromdate", lastUpdate.Unix())
		reqParams.Add("sort", "creation")
		reqParams.Add("order", "asc")
		reqParams.Add("tagged", config.StackTags)

		results, err := session.AllQuestions(reqParams)
		if err != nil {
			log.Fatalf("Failed to query %s: %v", config.StackSite, err)
		}

		log.Infof("Questions since %v: %d", fmtTime(lastUpdate), results.Total)

		lastUpdate = time.Now()

		for _, question := range results.Items {
			creation := time.Unix(question.Creation_date, 0)
			msgText := fmt.Sprintf("[%s] %s: %s", fmtTime(creation), question.Owner.Display_name, question.Link)
			log.Infof("> %s: %s", config.SlackChannel, msgText)
			msgParams := slack.PostMessageParameters{
				Username: config.SlackUserName,
				AsUser: true,
				//Markdown: true,
			}
			_, _, err = api.PostMessage(config.SlackChannel, msgText, msgParams)
			if err != nil {
				log.Fatalf("Failed to post message: %v", err)
			}
		}

		log.Infof("Sleeping %v", config.StackPoll)
		time.Sleep(config.StackPoll)
	}
}

func fmtTime(t time.Time) string {
	return t.Local().Format("2006-01-02 15:04:05 MST")
}
