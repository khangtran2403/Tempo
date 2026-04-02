package connector

import (
	"context"
	"fmt"
	"net/mail"
	"net/smtp"
	"strings"

	"github.com/brokeboycoding/tempo/internal/config"
	"github.com/brokeboycoding/tempo/pkg/common"
	"github.com/sirupsen/logrus"
)

type EmailConnector struct {
	smtpHost string
	smtpPort string
	username string
	password string
	from     string
	logger   *logrus.Logger
}

func NewEmailConnector(cfg *config.Config) *EmailConnector {
	logger := common.GetLogger()
	logger.Infof("[EMAIL DEBUG] Initializing EmailConnector with SMTP_HOST: '%s' and SMTP_PORT: '%s'", cfg.SMTP.Host, cfg.SMTP.Port)

	return &EmailConnector{
		smtpHost: cfg.SMTP.Host,
		smtpPort: cfg.SMTP.Port,
		username: cfg.SMTP.Username,
		password: cfg.SMTP.Password,
		from:     cfg.SMTP.From,
		logger:   logger,
	}
}
func (e *EmailConnector) Name() string {
	return "email"
}

func (e *EmailConnector) Execute(

	ctx context.Context,

	config map[string]interface{},

	prevResults map[string]interface{},

) (interface{}, error) {

	to, ok := config["to"].(string)

	if !ok || to == "" {

		return nil, fmt.Errorf("email: 'đia chỉ nhận' là bắt buộc")

	}

	parsedTo, err := mail.ParseAddress(to)

	if err != nil {

		return nil, fmt.Errorf("email: định dạng địa chỉ email nhận không hợp lệ: %w", err)

	}

	parsedFrom, err := mail.ParseAddress(e.from)

	if err != nil {

		return nil, fmt.Errorf("email: định dạng địa chỉ email gửi không hợp lệ trong cấu hình SMTP: %w", err)

	}

	subject := config["subject"].(string)

	body, ok := config["body"].(string)

	if !ok || body == "" {

		return nil, fmt.Errorf("email: 'nội dung' là bắt buộc")

	}

	message := e.buildMessage(to, subject, body)

	auth := smtp.PlainAuth("", e.username, e.password, e.smtpHost)

	addr := fmt.Sprintf("%s:%s", e.smtpHost, e.smtpPort)

	e.logger.Infof("[EMAIL DEBUG] Attempting to connect to SMTP server at address: '%s'", addr)

	err = smtp.SendMail(

		addr,

		auth,

		parsedFrom.Address,

		[]string{parsedTo.Address},

		[]byte(message),
	)
	if err != nil {

		return nil, fmt.Errorf("email:gửi thất bại: %w", err)

	}
	return map[string]interface{}{

			"sent": true,

			"to": to,

			"subject": subject,
		},
		nil

}
func (e *EmailConnector) ValidateConfig(config map[string]interface{}) error {

	to, ok := config["to"].(string)
	if !ok || to == "" {
		return fmt.Errorf("email: địa chỉ 'nhận' là bắt buộc")
	}
	if !strings.Contains(to, "@") {
		return fmt.Errorf("email: định dạng địa chỉ email nhận không hợp lệ")
	}

	if body, ok := config["body"].(string); !ok || body == "" {
		return fmt.Errorf("email: 'nội dung' là bắt buộc")
	}

	return nil
}
func (e *EmailConnector) buildMessage(to, subject, body string) string {
	// Sanitize headers to prevent SMTP injection and syntax errors.
	replacer := strings.NewReplacer("\r", "", "\n", "")
	cleanFrom := replacer.Replace(e.from)
	cleanTo := replacer.Replace(to)
	cleanSubject := replacer.Replace(subject)

	return fmt.Sprintf(
		"Từ: %s\r\n"+
			"Đến: %s\r\n"+
			"Tiêu đề: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/plain; charset=UTF-8\r\n"+
			"\r\n"+
			"%s",
		cleanFrom, cleanTo, cleanSubject, body,
	)
}
