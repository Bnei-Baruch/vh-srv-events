package util

import (
	"context"
	"fmt"
	"log"
	"os"

	part "vh-srv-event/participant"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/kelseyhightower/envconfig"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type config struct {
	DBUser         string `envconfig:"DB_USER" default:"postgres"`
	DBPass         string `envconfig:"DB_PASSWORD" default:"password"`
	DBName         string `envconfig:"DB_DATABASE" default:"event"`
	DBHost         string `envconfig:"DB_HOST" default:"localhost"`
	DBPort         string `envconfig:"DB_PORT" default:"5432"`
	APP_PORT       string `envconfig:"APP_PORT" default:"8080"`
	SendGridApiKey string `envconfig:"SEND_GRID_KEY" default:"SENDGRID_KEY_REDACTED"`
}

func SendConfirmationEmail(ctx *gin.Context, r *pgxpool.Pool, participationStatusID int) error {
	u := part.Part{}
	if err := r.QueryRow(ctx, `select 
	p.first_name,
	p.last_name,
	p.email,
	p.email_language 
	FROM participation_status as ps, participant as p 
	WHERE ps.id = $1 AND ps.participant_id = p.id`, participationStatusID).Scan(
		&u.FirstName, &u.LastName, &u.Email, &u.EmailLanguage,
	); err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("not found")
		}
		return err
	}

	if u.Email == nil {
		return fmt.Errorf("email not sent to participant")
	}

	var templateId string

	switch *u.EmailLanguage {
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
		return fmt.Errorf("no template found for user email language %s", *u.EmailLanguage)
	}

	mailErr := SendEmail(nil, nil, templateId, *u.Email, *u.FirstName, *u.LastName)

	return mailErr
}

func SendEmail(fromName *string, fromEmail *string, templateId string, email string, firstname string, lastname string) error {

	config, err := GetEnv()

	if err != nil {
		fmt.Println(err)
		return err
	}

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

	client := sendgrid.NewSendClient(config.SendGridApiKey)
	response, _ := client.Send(m)
	fmt.Println("email-StatusCode", response.StatusCode)
	fmt.Println("email-Body", response.Body)
	fmt.Println("email-Headers", response.Headers)
	if response.StatusCode > 202 {
		fmt.Println(response.StatusCode)
		return fmt.Errorf("error while sending email")
	}
	return nil
}

func GetEnv() (config, error) {
	var Config config
	if err := envconfig.Process("LIST", &Config); err != nil {
		log.Fatalln("Error while fetching env file")
		return Config, err
	}
	return Config, nil
}

func GetPgxPoolDBConnection(ctx context.Context) (*pgxpool.Pool, error) {

	config, err := GetEnv()

	if err != nil {
		log.Fatalln("Error while fetching env file")
	}

	databaseURL := "postgres://" + config.DBUser + ":" + config.DBPass + "@" + config.DBHost + ":" + config.DBPort + "/" + config.DBName

	conn, err := pgxpool.Connect(ctx, databaseURL)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		fmt.Fprintf(os.Stderr, "Connection url: %s", databaseURL)
		os.Exit(1)
	}

	return conn, nil
}
