package sanitizer

import (
	"html"
	"regexp"
	"strings"
)

var (
	sqlInjectionPattern = regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute|script|javascript|<script|onerror|onload)`)
	xssPattern          = regexp.MustCompile(`(?i)(<script|javascript:|onerror|onload|<iframe|<object|<embed)`)
)

func SanitizeString(input string) string {
	sanitized := strings.TrimSpace(input)
	sanitized = html.EscapeString(sanitized)
	return sanitized
}

func SanitizeHTML(input string) string {
	return html.EscapeString(input)
}

func RemoveHTMLTags(input string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(input, "")
}

func SanitizeEmail(email string) string {
	email = strings.TrimSpace(email)
	email = strings.ToLower(email)
	return email
}

func ContainsSQLInjection(input string) bool {
	return sqlInjectionPattern.MatchString(input)
}

func ContainsXSS(input string) bool {
	return xssPattern.MatchString(input)
}

func IsSafe(input string) bool {
	return !ContainsSQLInjection(input) && !ContainsXSS(input)
}

func SanitizeMap(data map[string]string) map[string]string {
	sanitized := make(map[string]string)
	for k, v := range data {
		sanitized[SanitizeString(k)] = SanitizeString(v)
	}
	return sanitized
}

func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
