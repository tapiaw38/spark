package modules

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// DevToolsSearch handles encoding/hashing/id helpers:
// b64, b64d, url, urld, hash, uuid, epoch.
func DevToolsSearch(query string) []Result {
	q := strings.TrimSpace(query)
	if q == "" {
		return nil
	}
	fields := strings.SplitN(q, " ", 2)
	cmd := strings.ToLower(fields[0])
	arg := ""
	if len(fields) > 1 {
		arg = fields[1]
	}

	switch cmd {
	case "b64":
		if arg == "" {
			return nil
		}
		return devResult("Base64 encode", base64.StdEncoding.EncodeToString([]byte(arg)))
	case "b64d":
		dec, err := base64.StdEncoding.DecodeString(strings.TrimSpace(arg))
		if err != nil {
			return nil
		}
		return devResult("Base64 decode", string(dec))
	case "url":
		if arg == "" {
			return nil
		}
		return devResult("URL encode", url.QueryEscape(arg))
	case "urld":
		dec, err := url.QueryUnescape(strings.TrimSpace(arg))
		if err != nil {
			return nil
		}
		return devResult("URL decode", dec)
	case "hash":
		if arg == "" {
			return nil
		}
		sum := sha256.Sum256([]byte(arg))
		return devResult("SHA-256", hex.EncodeToString(sum[:]))
	case "uuid":
		return devResult("UUID v4", genUUID())
	case "epoch":
		if arg == "" {
			return devResult("Unix epoch", strconv.FormatInt(time.Now().Unix(), 10))
		}
		n, err := strconv.ParseInt(strings.TrimSpace(arg), 10, 64)
		if err != nil {
			return nil
		}
		return devResult("Epoch → date", time.Unix(n, 0).Format("2006-01-02 15:04:05"))
	}
	return nil
}

func devResult(label, value string) []Result {
	return []Result{{
		Type:    "devtool",
		Title:   value,
		Desc:    label + " — copy to clipboard",
		Icon:    "applications-utilities",
		Preview: value,
		Action:  func() { copyToClipboard(value) },
	}}
}

func genUUID() string {
	var b [16]byte
	rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
