package middlewares

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/mcuadros/ofelia/core"
)

var (
	slackUsername   = "Ofelia"
	slackAvatarURL  = "https://raw.githubusercontent.com/mcuadros/ofelia/master/static/avatar.png"
	slackPayloadVar = "payload"
)

type SlackConfig struct {
	URL     string `gcfg:"slack-webhook"`
	OnError bool   `gcfg:"slack-on-error"`
}

func NewSlack(c *SlackConfig) core.Middleware {
	var m core.Middleware
	if !IsEmpty(c) {
		m = &Slack{*c}
	}

	return m
}

type Slack struct {
	SlackConfig
}

// ContinueOnStop return allways true, we want alloways report the final status
func (m *Slack) ContinueOnStop() bool {
	return true
}

// Run sends a message to the slack channel, its close stop the exection to
// collect the metrics
func (m *Slack) Run(ctx *core.Context) error {
	err := ctx.Next()
	ctx.Stop(err)

	if ctx.Execution.Failed || !m.OnError {
		m.pushMessage(ctx)
	}

	return err
}

func (m *Slack) pushMessage(ctx *core.Context) {
	values := make(url.Values, 0)
	content, _ := json.Marshal(m.buildMessage(ctx))
	values.Add(slackPayloadVar, string(content))

	r, err := http.PostForm(m.URL, values)
	if err != nil {
		ctx.Logger.Error("Slack error calling %q error: %q", m.URL, err)
	} else if r.StatusCode != 200 {
		ctx.Logger.Error("Slack error non-200 status code calling %q", m.URL)
	}
}

func (m *Slack) buildMessage(ctx *core.Context) *slackMessage {
	msg := &slackMessage{
		Username: slackUsername,
		IconURL:  slackAvatarURL,
	}

	msg.Text = fmt.Sprintf(
		"Job *%s* finished in *%s*, command _%q_",
		ctx.Job.GetName(), ctx.Execution.Duration, ctx.Job.GetCommand(),
	)

	if ctx.Execution.Failed {
		msg.Attachments = append(msg.Attachments, slackAttachment{
			Title: "Execution failed",
			Text:  ctx.Execution.Error.Error(),
			Color: "#F35A00",
		})
	} else {
		msg.Attachments = append(msg.Attachments, slackAttachment{
			Title: "Execution successful",
			Color: "#7CD197",
		})
	}

	return msg
}

type slackMessage struct {
	Text        string            `json:"text"`
	Username    string            `json:"username"`
	Attachments []slackAttachment `json:"attachments"`
	IconURL     string            `json:"icon_url"`
}

type slackAttachment struct {
	Color string `json:"color,omitempty"`
	Title string `json:"title,omitempty"`
	Text  string `json:"text"`
}
