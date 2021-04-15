package main

import "C"
import (
	"strings"

	abi "github.com/nicholasjackson/wasp/go-abi"
)

//export send_channel_message
func sendMessage(channel abi.WasmString, content abi.WasmString) int32

//export delete_channel_message
func deleteMessage(channel abi.WasmString, id abi.WasmString) int32

func main() {}

func sanitize(s string) string {
	s = strings.ToLower(s)
	s = strings.Replace(s, "0", "o", -1)
	s = strings.Replace(s, "1", "i", -1)
	s = strings.Replace(s, "3", "e", -1)
	s = strings.Replace(s, "4", "a", -1)
	s = strings.Replace(s, "5", "s", -1)
	s = strings.Replace(s, "6", "b", -1)
	s = strings.Replace(s, "7", "l", -1)
	s = strings.Replace(s, "8", "b", -1)
	s = strings.Replace(s, "@", "a", -1)
	s = strings.Replace(s, "+", "t", -1)
	s = strings.Replace(s, "$", "s", -1)
	s = strings.Replace(s, "#", "h", -1)
	s = strings.Replace(s, "()", "o", -1)
	s = strings.Replace(s, "_", "", -1)
	s = strings.Replace(s, "-", "", -1)
	s = strings.Replace(s, "*", "", -1)
	s = strings.Replace(s, "'", "", -1)
	s = strings.Replace(s, "?", "", -1)
	s = strings.Replace(s, "!", "", -1)
	s = strings.Replace(s, " ", "", -1)
	return s
}

func isProfane(s string) bool {
	s = sanitize(s)
	// Check for false false positives
	for _, word := range falseNegatives {
		if match := strings.Contains(s, word); match {
			return true
		}
	}
	// Remove false positives
	for _, word := range falsePositives {
		s = strings.Replace(s, word, "", -1)
	}
	// Check for profanities
	for _, word := range profanities {
		if match := strings.Contains(s, word); match {
			return true
		}
	}
	return false
}

//go:export message_create
func ReplaceBadWords(channel abi.WasmString, author abi.WasmString, id abi.WasmString, content abi.WasmString) {
	// get the string from the memory pointer
	a := author.String()
	s := content.String()

	if isProfane(s) {
		err := deleteMessage(channel, id)
		if err != 0 {
			abi.Error("Could not delete message")
		}

		naughty := abi.String(a + " is naughty...")
		err = sendMessage(channel, naughty)
		if err != 0 {
			abi.Error("Could not send message")
		}
	}
}
