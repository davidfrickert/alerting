package notify

import (
	"fmt"

	"github.com/prometheus/alertmanager/notify"
	"github.com/prometheus/alertmanager/types"

	"github.com/davidfrickert/alerting/images"
	"github.com/davidfrickert/alerting/logging"
	"github.com/davidfrickert/alerting/receivers"
	"github.com/davidfrickert/alerting/receivers/alertmanager"
	"github.com/davidfrickert/alerting/receivers/dinding"
	"github.com/davidfrickert/alerting/receivers/discord"
	"github.com/davidfrickert/alerting/receivers/email"
	"github.com/davidfrickert/alerting/receivers/googlechat"
	"github.com/davidfrickert/alerting/receivers/kafka"
	"github.com/davidfrickert/alerting/receivers/line"
	"github.com/davidfrickert/alerting/receivers/ntfy"
	"github.com/davidfrickert/alerting/receivers/oncall"
	"github.com/davidfrickert/alerting/receivers/opsgenie"
	"github.com/davidfrickert/alerting/receivers/pagerduty"
	"github.com/davidfrickert/alerting/receivers/pushover"
	"github.com/davidfrickert/alerting/receivers/sensugo"
	"github.com/davidfrickert/alerting/receivers/slack"
	"github.com/davidfrickert/alerting/receivers/teams"
	"github.com/davidfrickert/alerting/receivers/telegram"
	"github.com/davidfrickert/alerting/receivers/threema"
	"github.com/davidfrickert/alerting/receivers/victorops"
	"github.com/davidfrickert/alerting/receivers/webex"
	"github.com/davidfrickert/alerting/receivers/webhook"
	"github.com/davidfrickert/alerting/receivers/wecom"
	"github.com/davidfrickert/alerting/templates"
)

// BuildReceiverIntegrations creates integrations for each configured notification channel in GrafanaReceiverConfig.
// It returns a slice of Integration objects, one for each notification channel, along with any errors that occurred.
func BuildReceiverIntegrations(
	receiver GrafanaReceiverConfig,
	tmpl *templates.Template,
	img images.Provider,
	logger logging.LoggerFactory,
	newWebhookSender func(n receivers.Metadata) (receivers.WebhookSender, error),
	newEmailSender func(n receivers.Metadata) (receivers.EmailSender, error),
	orgID int64,
	version string,
) ([]*Integration, error) {
	type notificationChannel interface {
		notify.Notifier
		notify.ResolvedSender
	}
	var (
		integrations []*Integration
		errors       types.MultiError
		nl           = func(meta receivers.Metadata) logging.Logger {
			return logger("ngalert.notifier."+meta.Type, "notifierUID", meta.UID)
		}
		ci = func(idx int, cfg receivers.Metadata, n notificationChannel) {
			i := NewIntegration(n, n, cfg.Type, idx, cfg.Name)
			integrations = append(integrations, i)
		}
		nw = func(cfg receivers.Metadata) receivers.WebhookSender {
			w, e := newWebhookSender(cfg)
			if e != nil {
				errors.Add(fmt.Errorf("unable to build webhook client for %s notifier %s (UID: %s): %w ", cfg.Type, cfg.Name, cfg.UID, e))
				return nil // return nil to simplify the construction code. This works because constructor in notifiers do not check the argument for nil.
				// This does not cause misconfigured notifiers because it populates `errors`, which causes the function to return nil integrations and non-nil error.
			}
			return w
		}
	)
	// Range through each notification channel in the receiver and create an integration for it.
	for i, cfg := range receiver.AlertmanagerConfigs {
		ci(i, cfg.Metadata, alertmanager.New(cfg.Settings, cfg.Metadata, img, nl(cfg.Metadata)))
	}
	for i, cfg := range receiver.DingdingConfigs {
		ci(i, cfg.Metadata, dinding.New(cfg.Settings, cfg.Metadata, tmpl, nw(cfg.Metadata), nl(cfg.Metadata)))
	}
	for i, cfg := range receiver.DiscordConfigs {
		ci(i, cfg.Metadata, discord.New(cfg.Settings, cfg.Metadata, tmpl, nw(cfg.Metadata), img, nl(cfg.Metadata), version))
	}
	for i, cfg := range receiver.NtfyConfigs {
		ci(i, cfg.Metadata, ntfy.New(cfg.Settings, cfg.Metadata, tmpl, nw(cfg.Metadata), img, nl(cfg.Metadata), version))
	}
	for i, cfg := range receiver.EmailConfigs {
		mailCli, e := newEmailSender(cfg.Metadata)
		if e != nil {
			errors.Add(fmt.Errorf("unable to build email client for %s notifier %s (UID: %s): %w ", cfg.Type, cfg.Name, cfg.UID, e))
			continue
		}
		ci(i, cfg.Metadata, email.New(cfg.Settings, cfg.Metadata, tmpl, mailCli, img, nl(cfg.Metadata)))
	}
	for i, cfg := range receiver.GooglechatConfigs {
		ci(i, cfg.Metadata, googlechat.New(cfg.Settings, cfg.Metadata, tmpl, nw(cfg.Metadata), img, nl(cfg.Metadata), version))
	}
	for i, cfg := range receiver.KafkaConfigs {
		ci(i, cfg.Metadata, kafka.New(cfg.Settings, cfg.Metadata, tmpl, nw(cfg.Metadata), img, nl(cfg.Metadata)))
	}
	for i, cfg := range receiver.LineConfigs {
		ci(i, cfg.Metadata, line.New(cfg.Settings, cfg.Metadata, tmpl, nw(cfg.Metadata), nl(cfg.Metadata)))
	}
	for i, cfg := range receiver.OnCallConfigs {
		ci(i, cfg.Metadata, oncall.New(cfg.Settings, cfg.Metadata, tmpl, nw(cfg.Metadata), img, nl(cfg.Metadata), orgID))
	}
	for i, cfg := range receiver.OpsgenieConfigs {
		ci(i, cfg.Metadata, opsgenie.New(cfg.Settings, cfg.Metadata, tmpl, nw(cfg.Metadata), img, nl(cfg.Metadata)))
	}
	for i, cfg := range receiver.PagerdutyConfigs {
		ci(i, cfg.Metadata, pagerduty.New(cfg.Settings, cfg.Metadata, tmpl, nw(cfg.Metadata), img, nl(cfg.Metadata)))
	}
	for i, cfg := range receiver.PushoverConfigs {
		ci(i, cfg.Metadata, pushover.New(cfg.Settings, cfg.Metadata, tmpl, nw(cfg.Metadata), img, nl(cfg.Metadata)))
	}
	for i, cfg := range receiver.SensugoConfigs {
		ci(i, cfg.Metadata, sensugo.New(cfg.Settings, cfg.Metadata, tmpl, nw(cfg.Metadata), img, nl(cfg.Metadata)))
	}
	for i, cfg := range receiver.SlackConfigs {
		ci(i, cfg.Metadata, slack.New(cfg.Settings, cfg.Metadata, tmpl, nw(cfg.Metadata), img, nl(cfg.Metadata), version))
	}
	for i, cfg := range receiver.TeamsConfigs {
		ci(i, cfg.Metadata, teams.New(cfg.Settings, cfg.Metadata, tmpl, nw(cfg.Metadata), img, nl(cfg.Metadata)))
	}
	for i, cfg := range receiver.TelegramConfigs {
		ci(i, cfg.Metadata, telegram.New(cfg.Settings, cfg.Metadata, tmpl, nw(cfg.Metadata), img, nl(cfg.Metadata)))
	}
	for i, cfg := range receiver.ThreemaConfigs {
		ci(i, cfg.Metadata, threema.New(cfg.Settings, cfg.Metadata, tmpl, nw(cfg.Metadata), img, nl(cfg.Metadata)))
	}
	for i, cfg := range receiver.VictoropsConfigs {
		ci(i, cfg.Metadata, victorops.New(cfg.Settings, cfg.Metadata, tmpl, nw(cfg.Metadata), img, nl(cfg.Metadata), version))
	}
	for i, cfg := range receiver.WebhookConfigs {
		ci(i, cfg.Metadata, webhook.New(cfg.Settings, cfg.Metadata, tmpl, nw(cfg.Metadata), img, nl(cfg.Metadata), orgID))
	}
	for i, cfg := range receiver.WecomConfigs {
		ci(i, cfg.Metadata, wecom.New(cfg.Settings, cfg.Metadata, tmpl, nw(cfg.Metadata), nl(cfg.Metadata)))
	}
	for i, cfg := range receiver.WebexConfigs {
		ci(i, cfg.Metadata, webex.New(cfg.Settings, cfg.Metadata, tmpl, nw(cfg.Metadata), img, nl(cfg.Metadata), orgID))
	}
	if errors.Len() > 0 {
		return nil, &errors
	}
	return integrations, nil
}
