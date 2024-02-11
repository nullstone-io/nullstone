package version

import (
	"strconv"
	"strings"
)

// FindLatestVersionSequence takes the provided shortSha and finds any versions in the list of artifacts
//
//	a match is any artifact that starts with the shortSha
//	this will return the largest sequence number found
//	if no matches are found, this will return 0
func FindLatestVersionSequence(shortSha string, artifacts []string) int {
	sequence := 0
	for _, artifact := range artifacts {
		// if we find an artifact with the same shortSha, we will increase the sequence if it is the largest
		if strings.HasPrefix(artifact, shortSha) {
			// split the sha and sequence
			parts := strings.Split(artifact, "-")
			// if we don't get exactly 2 parts, this isn't the correct format so we will ignore
			if len(parts) != 2 {
				continue
			}
			sequenceStr := parts[1]
			seq, err := strconv.Atoi(sequenceStr)
			// if the second part isn't a number, this isn't the correct format so we will ignore
			if err != nil {
				continue
			}
			// if the sequence is larger than the current sequence, we will update the sequence
			if seq > sequence {
				sequence = seq
			}
		}
	}

	return sequence
}
