package utils

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"

	"gitlab.bbdev.team/vh/vh-srv-events/common"
)

func SendConfirmationEmail(ctx context.Context, firstName, lastName, email, language string) error {
	var templateId string
	switch language {
	case "ru":
		templateId = "d-91e781b2d85d4bbea36b5726e43379fd"
	case "es":
		templateId = "d-57743abff2fa46db8273ec5e3f387ed9"
	case "en":
		templateId = "d-7bc0a43341cb481aa1432d599bbbd1f6"
	case "he":
		templateId = "d-6352bc18e9e14e7f8db77de98c3d804c"
	}

	if templateId == "" {
		return fmt.Errorf("no template found for user email language %s", language)
	}

	return SendEmail(ctx, nil, nil, templateId, email, firstName, lastName)
}

func SendEmail(ctx context.Context, fromName *string, fromEmail *string, templateId string, email string, firstname string, lastname string) error {
	var frEmail, frName string

	if fromEmail == nil {
		frEmail = "help@kli.one"
	} else {
		frEmail = *fromEmail
	}

	if fromName == nil {
		frName = "Bnei Baruch"
	} else {
		frName = *fromName
	}

	from := mail.NewEmail(frName, frEmail)
	to := mail.NewEmail(firstname+" "+lastname, email)
	subject := "A notification for you"

	m := mail.NewV3MailInit(from, subject, to)

	m.SetTemplateID(templateId)

	client := sendgrid.NewSendClient(common.Config.SendGridApiKey)
	response, err := client.SendWithContext(ctx, m)
	if err != nil {
		return fmt.Errorf("sendgrid.Send: %w", err)
	}

	if response.StatusCode > 202 {
		LogFor(ctx).Warn("sendgrid.Send",
			slog.Attr{Key: "mail", Value: slog.GroupValue(
				slog.Int("status_code", response.StatusCode),
				slog.String("body", response.Body),
				slog.Any("headers", response.Headers),
			)})
		return fmt.Errorf("sendgrid http error: %d", response.StatusCode)
	}

	return nil
}
