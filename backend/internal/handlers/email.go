package handlers

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/assylkhan/invisionu-backend/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EmailService struct {
	cfg  *config.Config
	pool *pgxpool.Pool
}

func NewEmailService(cfg *config.Config, pool *pgxpool.Pool) *EmailService {
	return &EmailService{cfg: cfg, pool: pool}
}

func (e *EmailService) Enabled() bool {
	return e.cfg.SMTPHost != "" && e.cfg.SMTPUser != ""
}

// SendApplicationReceived — подтверждение получения заявки
func (e *EmailService) SendApplicationReceived(toEmail, fullName string, candidateID int) error {
	subject := "Your inVision U Application Has Been Received"
	body := fmt.Sprintf(`Dear %s,

Thank you for applying to inVision U — the 100%% scholarship university by inDrive.

We have received your application and our AI-powered admissions system will review it shortly. You will receive updates via email as your application progresses through the stages.

Application Reference: #%d

What happens next:
1. AI analysis of your essay and background
2. Committee review and scoring
3. If you qualify (score 65+), you will be invited to a Stage 2 Telegram interview
4. Final decision notification

If you have any questions, please contact us at admissions@invisionu.kz

Best regards,
inVision U Admissions Team`, fullName, candidateID)

	return e.sendEmail(toEmail, subject, body, "application_received", candidateID)
}

// SendShortlistNotification — уведомление о попадании в шортлист
func (e *EmailService) SendShortlistNotification(toEmail, fullName string, candidateID int) error {
	subject := "Congratulations! You Have Been Shortlisted — inVision U"
	body := fmt.Sprintf(`Dear %s,

We are thrilled to inform you that you have been SHORTLISTED for inVision U!

This is a significant achievement. Our admissions committee has reviewed your application and selected you as one of our top candidates.

Next Steps:
- You will receive further instructions about the enrollment process
- Please ensure your email and Telegram contact are up to date
- Watch for updates regarding required documents and deadlines

We look forward to welcoming you to inVision U.

Best regards,
inVision U Admissions Team`, fullName)

	return e.sendEmail(toEmail, subject, body, "shortlisted", candidateID)
}

// SendInterviewInvite — приглашение на Telegram-интервью
func (e *EmailService) SendInterviewInvite(toEmail, fullName, deepLink string, candidateID int) error {
	subject := "Stage 2 Interview Invitation — inVision U"
	body := fmt.Sprintf(`Dear %s,

Congratulations! Based on your Stage 1 essay analysis, you have been invited to the Stage 2 Interview.

The interview will be conducted via our Telegram bot. Please click the link below or copy it into Telegram to begin:

%s

Important notes:
- The interview is conducted in English
- You may respond via text or voice message
- The interview consists of 8-15 questions and takes approximately 20-30 minutes
- Please find a quiet place without distractions

The invitation link expires in 7 days.

Good luck!

Best regards,
inVision U Admissions Team`, fullName, deepLink)

	return e.sendEmail(toEmail, subject, body, "interview_invite", candidateID)
}

// SendRejectionNotification — уведомление об отказе
func (e *EmailService) SendRejectionNotification(toEmail, fullName string, candidateID int) error {
	subject := "inVision U Application Update"
	body := fmt.Sprintf(`Dear %s,

Thank you for your interest in inVision U and for taking the time to submit your application.

After careful review by our admissions committee, we regret to inform you that we are unable to offer you a place in the current cohort.

We encourage you to:
- Continue developing your skills and achievements
- Consider applying again in the future
- Stay connected with inVision U on social media for upcoming opportunities

We appreciate your dedication and wish you all the best in your future endeavors.

Best regards,
inVision U Admissions Team`, fullName)

	return e.sendEmail(toEmail, subject, body, "rejected", candidateID)
}

func (e *EmailService) sendEmail(to, subject, body string, template string, candidateID int) error {
	if !e.Enabled() {
		return fmt.Errorf("SMTP not configured")
	}

	msg := buildMIMEMessage(e.cfg.SMTPFrom, to, subject, body)

	addr := fmt.Sprintf("%s:%d", e.cfg.SMTPHost, e.cfg.SMTPPort)
	auth := smtp.PlainAuth("", e.cfg.SMTPUser, e.cfg.SMTPPassword, e.cfg.SMTPHost)

	err := smtp.SendMail(addr, auth, e.cfg.SMTPFrom, []string{to}, []byte(msg))

	// Лог в БД, некритично
	if e.pool != nil {
		logStatus := "sent"
		if err != nil {
			logStatus = "failed"
		}
		ctx := context.Background()
		e.pool.Exec(ctx,
			`INSERT INTO email_logs (candidate_id, recipient_email, subject, template, status) VALUES ($1, $2, $3, $4, $5)`,
			candidateID, to, subject, template, logStatus)
	}

	return err
}

func buildMIMEMessage(from, to, subject, body string) string {
	var sb strings.Builder
	sb.WriteString("From: inVision U Admissions <" + from + ">\r\n")
	sb.WriteString("To: " + to + "\r\n")
	sb.WriteString("Subject: " + subject + "\r\n")
	sb.WriteString("MIME-Version: 1.0\r\n")
	sb.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	sb.WriteString("\r\n")
	sb.WriteString(body)
	return sb.String()
}
