package main

import "strings"

func normalizeRoomName(s string) string {
	return strings.ToLower(s)
}

func truncateString(s string, maxLength int) string {
	r := []rune(s)
	if len(r) <= maxLength {
		return s
	}
	return string(r[:maxLength])
}
