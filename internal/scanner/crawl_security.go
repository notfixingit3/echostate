package scanner

import (
	"context"
	"strings"
)

var securityTxtPaths = []string{
	"/.well-known/security.txt",
	"/security.txt",
}

func fetchSecurityTxt(ctx context.Context, host string) (map[string]any, error) {
	for _, path := range securityTxtPaths {
		body, err := fetchHostResource(ctx, host, path, 64*1024)
		if err != nil || len(body) == 0 {
			continue
		}
		parsed := parseSecurityTxt(string(body))
		if len(parsed) == 0 {
			continue
		}
		parsed["source"] = path
		return parsed, nil
	}
	return nil, nil
}

func parseSecurityTxt(content string) map[string]any {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}

	result := map[string]any{}
	var contacts []string
	var acknowledgments []string
	var hiring []string
	var policies []string

	for _, line := range strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}

		switch strings.ToLower(key) {
		case "contact":
			contacts = append(contacts, stripMailto(value))
		case "expires":
			result["expires"] = value
		case "canonical":
			result["canonical"] = value
		case "encryption":
			result["encryption"] = value
		case "preferred-languages":
			result["preferred_languages"] = value
		case "acknowledgments":
			acknowledgments = append(acknowledgments, value)
		case "hiring":
			hiring = append(hiring, value)
		case "policy":
			policies = append(policies, value)
		default:
			result[strings.ToLower(strings.ReplaceAll(key, "-", "_"))] = value
		}
	}

	if len(contacts) > 0 {
		result["contacts"] = contacts
	}
	if len(acknowledgments) > 0 {
		result["acknowledgments"] = acknowledgments
	}
	if len(hiring) > 0 {
		result["hiring"] = hiring
	}
	if len(policies) > 0 {
		result["policies"] = policies
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

func stripMailto(value string) string {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(strings.ToLower(value), "mailto:") {
		return strings.TrimSpace(value[7:])
	}
	return value
}