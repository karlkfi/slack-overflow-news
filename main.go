package main // import "github.com/karlkfi/slack-overflow-news"

import (
	log "github.com/Sirupsen/logrus"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"github.com/laktek/Stack-on-Go/stackongo"
	"github.com/nlopes/slack"
	"github.com/kelseyhightower/envconfig"

	"fmt"
	"time"
	"os"
	"regexp"
	"strings"
	"strconv"
)

const envPrefix = "SS"

type Config struct {
	StackSite string	`required:"true" envconfig:"STACK_SITE"`
	StackTags string	`required:"true" envconfig:"STACK_TAGS"`
	StackPoll time.Duration	`required:"false" envconfig:"STACK_POLL" default:"30s"`
	StackHistory int	`required:"false" envconfig:"STACK_HISTORY" default:"30"`

	SlackToken string	`required:"true" envconfig:"SLACK_TOKEN"`
	SlackUserName string	`required:"true" envconfig:"SLACK_USERNAME"`
	SlackChannel string	`required:"true" envconfig:"SLACK_CHANNEL"` // channel name, not ID
	SlackDebug bool		`required:"false" envconfig:"SLACK_DEBUG" default:"false"`

	LogLevel string		`required:"false" envconfig:"LOG_LEVEL" default:"INFO"`
}

const timePattern = "2006-01-02 15:04:05 MST"
const msgPattern = "[%s] %s: %s"
var (
	msgMatcher = regexp.MustCompile(`\[(.*)\] (.*): (.*)`)
)

func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
		TimestampFormat: time.RFC3339,
	})

	// support Beep Boop's env var
	value, found := os.LookupEnv("SLACK_TOKEN")
	if found {
		os.Setenv("SS_SLACK_TOKEN", value)
	}

	var config Config
	err := envconfig.Process(envPrefix, &config)
	if err != nil {
		log.Errorf("Failed to parse config: %v", err)
		exit(2)
	}

	level, err := log.ParseLevel(config.LogLevel)
	if err != nil {
		log.Errorf("Failed to parse log level '%s': %v", config.LogLevel, err)
		exit(2)
	}
	log.SetLevel(level)

	config.SlackChannel = normalizeChannelName(config.SlackChannel)

	configCopy := config
	configCopy.SlackToken = "<redacted>" // don't log secrets
	log.Infof("Config: %+v", configCopy)

	stackClient := stackongo.NewSession(config.StackSite)
	slackClient := slack.New(config.SlackToken)
	slackClient.SetDebug(config.SlackDebug)

	slackAuth, err := slackClient.AuthTest()
	if err != nil {
		log.Errorf("Failed to get user ID: %v", err)
		exit(1)
	}

	channelId, err := findChannelID(slackClient, config.SlackChannel)
	if err != nil {
		log.Errorf("Failed to find channel '%s': %v", config.SlackChannel, err)
		exit(1)
	}

	latestReport, err := latestReportQuestionTime(slackClient, channelId, slackAuth.UserID)
	if err != nil {
		log.Warnf("Failed to lookup last reported question: %v", err)
		// use configured history duration
		latestReport = time.Now().AddDate(0, 0, -config.StackHistory)
	} else {
		log.Infof("Found last reported question: %v", fmtTime(latestReport))
	}

	for {
		reqParams := make(stackongo.Params)
		reqParams.Add("fromdate", latestReport.Add(time.Second).Unix()) // inclusive
		reqParams.Add("sort", "creation")
		reqParams.Add("order", "asc")
		reqParams.Add("tagged", config.StackTags)

		results, err := stackClient.AllQuestions(reqParams)
		if err != nil {
			log.Errorf("Failed to query %s: %v", config.StackSite, err)
			exit(1)
		}

		log.Infof("Questions since %v: %d", fmtTime(latestReport), results.Total)

		for _, question := range results.Items {
			creation := time.Unix(question.Creation_date, 0)
			msgText := fmt.Sprintf(msgPattern, fmtTime(creation), question.Owner.Display_name, question.Link)
			log.Infof("> %s", msgText)
			msgParams := slack.PostMessageParameters{
				Username: config.SlackUserName,
				AsUser: true,
				//Markdown: true,
			}
			_, _, err = slackClient.PostMessage(channelId, msgText, msgParams)
			if err != nil {
				log.Errorf("Failed to post message: %v", err)
				exit(1)
			}
			// bump the timestamp after successful report
			latestReport = time.Unix(question.Creation_date, 0)
		}

		log.Debugf("Sleeping %v", config.StackPoll)
		time.Sleep(config.StackPoll)
	}
}

func fmtTime(t time.Time) string {
	return t.Local().Format(timePattern)
}

func exit(exitCode int) {
	log.Infof("Exit (%d)", exitCode)
	os.Exit(exitCode)
}

// lastReportQuestionTime returns the creation time of the last question reported to the configured channel
// before the specified 'latest' time.
func latestReportQuestionTime(slackClient *slack.Client, channelId, userId string) (time.Time, error) {
	history, err := slackClient.GetChannelHistory(channelId, slack.HistoryParameters{
		Count: 100,
	})
	if err != nil {
		return time.Time{}, bosherr.WrapErrorf(err, "Getting history of channel")
	}

	log.Infof("Reading chat channel history...")

	for _, message := range history.Messages {
		if message.User == userId {
			logSlackMessage(message)

			// parse bot messages
			match := msgMatcher.FindStringSubmatch(message.Text)
			if match != nil {
				lastReported, err := time.Parse(timePattern, match[1])
				if err != nil {
					return time.Time{}, bosherr.WrapErrorf(err, "Parsing last reported question timestamp")
				}
				return lastReported, nil
			}
		}
	}

	return time.Time{}, bosherr.Error("No matching message found (in the last 100 messages)")
}

func findChannelID(slackClient *slack.Client, channelName string) (channelId string, err error) {
	channels, err := slackClient.GetChannels(true)
	if err != nil {
		return "", bosherr.WrapError(err, "Getting list of channels")
	}

	for _, channel := range channels {
		if channel.Name == channelName {
			return channel.ID, nil
		}
	}

	return "", bosherr.Errorf("No channel found with name '%s'", channelName)
}

func normalizeChannelName(channelName string) (string) {
	if strings.HasPrefix(channelName, "#") {
		channelName = channelName[1:]
	}
	return channelName
}

func parseSlackTimestamp(timestamp string) (t time.Time, uid string, err error) {
	match := strings.Split(timestamp, ".")
	if len(match) != 2 {
		return time.Time{}, "", bosherr.Errorf("Parsing Slack timestamp '%s'", timestamp)
	}
	unixTime, err := strconv.ParseInt(match[0], 10, 0)
	if err != nil {
		return time.Time{}, "", bosherr.Errorf("Parsing Slack timestamp '%s'", timestamp)
	}
	return time.Unix(unixTime, 0), match[1], nil
}

func logSlackMessage(message slack.Message) {
	created, _, err := parseSlackTimestamp(message.Timestamp)
	if err != nil {
		log.Errorf("Parsing timestamp of history message: %+v", message)
		return
	}

	user := message.Username
	if user == "" {
		user = message.User
	}

	log.Infof("< [%s] %s: %s", fmtTime(created), user, message.Text)
}
