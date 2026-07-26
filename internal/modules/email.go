package modules

import (
	"strings"
)

func EmailSearch(query string) []Result {
	q := strings.TrimSpace(query)
	lower := strings.ToLower(q)
	if lower != "email" && !strings.HasPrefix(lower, "email ") && lower != "mail" && !strings.HasPrefix(lower, "mail ") {
		return nil
	}

	body := ""
	if strings.HasPrefix(lower, "email ") {
		body = strings.TrimSpace(q[len("email "):])
	} else if strings.HasPrefix(lower, "mail ") {
		body = strings.TrimSpace(q[len("mail "):])
	}
	if body == "" {
		return []Result{{
			Type:       TypeEmail,
			Title:      "Open Email Composer",
			Desc:       "Type: email person@example.com subject",
			Icon:       "internet-mail",
			ActionSpec: EmailWindowAction("", "", ""),
		}}
	}

	to, subject, mailBody := splitEmailBody(body)
	if to != "" && !strings.Contains(to, "@") {
		if email := FindContactEmail(to); email != "" {
			to = email
		}
	}
	title := "Compose Email"
	if to != "" {
		title = "Email " + to
	}
	return []Result{
		{
			Type:       TypeEmail,
			Title:      "Open Email Composer",
			Desc:       to + " | " + subject,
			Icon:       "internet-mail",
			ActionSpec: EmailWindowAction(to, subject, mailBody),
		},
		{
			Type:       TypeEmail,
			Title:      title,
			Desc:       subject,
			Icon:       "internet-mail",
			ActionSpec: EmailAction(to, subject, mailBody),
		},
	}
}

func splitEmailBody(body string) (string, string, string) {
	if strings.Contains(body, "|") {
		parts := strings.SplitN(body, "|", 3)
		to := strings.TrimSpace(parts[0])
		subject := ""
		mailBody := ""
		if len(parts) > 1 {
			subject = strings.TrimSpace(parts[1])
		}
		if len(parts) > 2 {
			mailBody = strings.TrimSpace(parts[2])
		}
		return to, subject, mailBody
	}
	parts := strings.Fields(body)
	if len(parts) == 0 {
		return "", "", ""
	}
	to := parts[0]
	subject := strings.TrimSpace(strings.TrimPrefix(body, to))
	if !strings.Contains(to, "@") {
		return to, subject, ""
	}
	return to, subject, ""
}
