package modules

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type contact struct {
	name  string
	email string
	phone string
	path  string
}

// ContactsSearch finds local vCard contacts.
func ContactsSearch(query string) []Result {
	q := strings.TrimSpace(query)
	lower := strings.ToLower(q)
	if lower != "contact" && lower != "contacts" && !strings.HasPrefix(lower, "contact ") && !strings.HasPrefix(lower, "contacts ") {
		return nil
	}

	filter := ""
	if strings.HasPrefix(lower, "contact ") {
		filter = strings.TrimSpace(q[len("contact "):])
	} else if strings.HasPrefix(lower, "contacts ") {
		filter = strings.TrimSpace(q[len("contacts "):])
	}
	if filter == "sync" || filter == "carddav" {
		return contactSyncResults()
	}

	contacts := loadContacts(filter)
	if len(contacts) == 0 {
		return []Result{{
			Type:   "contact",
			Title:  "No Contacts Found",
			Desc:   "Looks for .vcf files in local contact folders",
			Icon:   "x-office-address-book",
			Action: func() {},
		}}
	}

	var results []Result
	for _, c := range contacts {
		contact := c
		if contact.email != "" {
			results = append(results, Result{
				Type:  "contact",
				Title: contact.name,
				Desc:  contact.email,
				Icon:  "internet-mail",
				Action: func() {
					copyText(contact.email)
				},
			})
			results = append(results, Result{
				Type:  "contact",
				Title: "Email " + contact.name,
				Desc:  contact.email,
				Icon:  "internet-mail",
				Action: func() {
					SendEmailFull(contact.email, "", "", nil)
				},
			})
		}
		if contact.phone != "" {
			results = append(results, Result{
				Type:  "contact",
				Title: contact.name,
				Desc:  contact.phone,
				Icon:  "phone",
				Action: func() {
					copyText(contact.phone)
				},
			})
		}
		if len(results) >= 50 {
			break
		}
		if contact.path != "" {
			results = append(results, Result{
				Type:  "contact",
				Title: "Open Contact: " + contact.name,
				Desc:  shortenPath(contact.path),
				Icon:  "x-office-address-book",
				Action: func() {
					exec.Command("xdg-open", contact.path).Start()
				},
			})
		}
	}
	return results
}

func contactSyncResults() []Result {
	results := []Result{{
		Type:   "contact",
		Title:  "CardDAV via vdirsyncer",
		Desc:   "Run vdirsyncer sync, then contact <name>",
		Icon:   "emblem-synchronizing",
		Action: func() { exec.Command("vdirsyncer", "sync").Start() },
	}}
	if _, err := exec.LookPath("khal"); err == nil {
		results = append(results, Result{
			Type:   "contact",
			Title:  "Open khal Contacts",
			Desc:   "khal interactive contact backend",
			Icon:   "x-office-address-book",
			Action: func() { exec.Command("foot", "khal", "interactive").Start() },
		})
	}
	results = append(results, Result{
		Type:  "contact",
		Title: "Open Contact Data Folder",
		Desc:  "~/.local/share/contacts",
		Icon:  "folder-open",
		Action: func() {
			exec.Command("xdg-open", filepath.Join(os.Getenv("HOME"), ".local/share/contacts")).Start()
		},
	})
	return results
}

func loadContacts(filter string) []contact {
	dirs := []string{
		filepath.Join(os.Getenv("HOME"), ".local/share/contacts"),
		filepath.Join(os.Getenv("HOME"), ".local/share/evolution/addressbook"),
		filepath.Join(os.Getenv("HOME"), ".local/share/kaddressbook"),
		filepath.Join(os.Getenv("HOME"), ".local/share/akonadi"),
		filepath.Join(os.Getenv("HOME"), ".local/share/vdirsyncer"),
		filepath.Join(os.Getenv("HOME"), ".local/share/khal"),
		filepath.Join(os.Getenv("HOME"), ".local/share/carddav"),
		filepath.Join(os.Getenv("HOME"), ".cache/evolution/addressbook"),
		filepath.Join(os.Getenv("HOME"), ".contacts"),
		filepath.Join(os.Getenv("HOME"), "Contacts"),
	}

	var contacts []contact
	lowerFilter := strings.ToLower(filter)
	for _, dir := range dirs {
		filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".vcf") {
				return nil
			}
			c := parseVCard(path)
			c.path = path
			if c.name == "" {
				c.name = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
			}
			if c.email == "" && c.phone == "" {
				return nil
			}
			haystack := strings.ToLower(c.name + " " + c.email + " " + c.phone)
			if lowerFilter != "" && !strings.Contains(haystack, lowerFilter) {
				return nil
			}
			contacts = append(contacts, c)
			if len(contacts) >= 50 {
				return filepath.SkipAll
			}
			return nil
		})
	}
	return contacts
}

func FindContactEmail(filter string) string {
	for _, c := range loadContacts(filter) {
		if c.email != "" {
			return c.email
		}
	}
	return ""
}

func parseVCard(path string) contact {
	data, err := os.ReadFile(path)
	if err != nil {
		return contact{}
	}
	var c contact
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		upper := strings.ToUpper(line)
		switch {
		case strings.HasPrefix(upper, "FN:"):
			c.name = strings.TrimSpace(line[3:])
		case strings.HasPrefix(upper, "EMAIL"):
			if idx := strings.Index(line, ":"); idx >= 0 {
				c.email = strings.TrimSpace(line[idx+1:])
			}
		case strings.HasPrefix(upper, "TEL"):
			if idx := strings.Index(line, ":"); idx >= 0 {
				c.phone = strings.TrimSpace(line[idx+1:])
			}
		}
	}
	return c
}
