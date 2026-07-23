package modules

import (
	"net/url"
	"os/exec"
	"strings"
)

// EmailSearch composes an email.
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
			Type:  "email",
			Title: "Compose Email",
			Desc:  "Type: email person@example.com subject",
			Icon:  "internet-mail",
			Action: func() {
				exec.Command("xdg-email").Start()
			},
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
	return []Result{{
		Type:  "email",
		Title: title,
		Desc:  subject,
		Icon:  "internet-mail",
		Action: func() {
			args := []string{}
			if subject != "" {
				args = append(args, "--subject", subject)
			}
			if mailBody != "" {
				args = append(args, "mailto:"+url.QueryEscape(to)+"?subject="+url.QueryEscape(subject)+"&body="+url.QueryEscape(mailBody))
				exec.Command("xdg-open", args[len(args)-1]).Start()
				return
			}
			if to != "" {
				args = append(args, to)
			}
			exec.Command("xdg-email", args...).Start()
		},
	}}
}

func EmailFile(path string) {
	exec.Command("xdg-email", "--attach", path).Start()
}

func EmailFiles(paths []string) {
	args := make([]string, 0, len(paths)*2)
	for _, path := range paths {
		if path == "" {
			continue
		}
		args = append(args, "--attach", path)
	}
	exec.Command("xdg-email", args...).Start()
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
