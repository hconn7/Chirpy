package main

import (
	"fmt"
	"strings"
	"testing"
)

func CleanFuncTest(t *testing.T) {
	var tests = []struct {
		chirp string
		want  string
	}{

		{"I had so much fun kerfuffle", "I had so much fun ****"},
		{"I fornax! you so much", "I fornax! you so much"},
		{"you are sharbert", "you are ****"},
	}

	for _, tt := range tests {
		t.Run(tt.chirp, func(t *testing.T) {
			answers := CheckProfanityChirp(tt.chirp)
			result := strings.Join(answers, " ")
			if result != tt.want {
				t.Errorf("got %s, want %s", result , tt.want)
			} else {

		})
	}
}
