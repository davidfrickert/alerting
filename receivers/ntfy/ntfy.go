package ntfy

import (
	"context"
	"encoding/json"

	"github.com/prometheus/alertmanager/types"

	"github.com/grafana/alerting/images"
	"github.com/grafana/alerting/logging"
	"github.com/grafana/alerting/receivers"
	"github.com/grafana/alerting/templates"
)

type ntfyMessage struct {
	Topic   string `json:"topic"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

type Notifier struct {
	*receivers.Base
	log        logging.Logger
	ns         receivers.WebhookSender
	images     images.Provider
	tmpl       *templates.Template
	settings   Config
	appVersion string
}

func New(cfg Config, meta receivers.Metadata, template *templates.Template, sender receivers.WebhookSender, images images.Provider, logger logging.Logger, appVersion string) *Notifier {
	return &Notifier{
		Base:       receivers.NewBase(meta),
		log:        logger,
		ns:         sender,
		images:     images,
		tmpl:       template,
		settings:   cfg,
		appVersion: appVersion,
	}
}

func (n Notifier) Notify(ctx context.Context, as ...*types.Alert) (bool, error) {
	var tmplErr error
	tmpl, _ := templates.TmplText(ctx, n.tmpl, as, n.log, &tmplErr)

	title := templates.DefaultMessageTitleEmbed
	message := templates.DefaultMessageEmbed
	channel := n.settings.Channel

	msg := &ntfyMessage{
		Title:           tmpl(title),
		Message:         tmpl(message),
		Topic: 			 channel,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return false, err
	}

	cmd := &receivers.SendWebhookSettings{
		URL:         n.settings.ServerURL,
		Body:        string(body),
		HTTPMethod:  "POST",
		ContentType: "application/json",
	}

	if err := n.ns.SendWebhook(ctx, cmd); err != nil {
		n.log.Error("failed to send notification to ntfy", "error", err)
		return false, err
	}
	return true, nil
}

func (d Notifier) SendResolved() bool {
	return !d.GetDisableResolveMessage()
}
