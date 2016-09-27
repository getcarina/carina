package version

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// SemVer is a semantic version
type SemVer struct {
	Major    int
	Minor    int
	Patch    int
	Leftover string
}

// ExtractSemver parses a semantic version from a string
func ExtractSemver(semi string) (*SemVer, error) {
	var err error

	if len(semi) < 5 { // 1.3.5
		return nil, errors.New("Invalid semver")
	}
	// Allow a v in front
	if semi[0] == 'v' {
		semi = semi[1:]
	}
	semVerStrings := strings.SplitN(semi, ".", 3)

	if len(semVerStrings) < 3 {
		return nil, errors.New("Could not parse semver")
	}

	parsedSemver := new(SemVer)

	digitError := errors.New("Could not parse digits of semver")
	if parsedSemver.Major, err = strconv.Atoi(semVerStrings[0]); err != nil {
		return nil, digitError
	}
	if parsedSemver.Minor, err = strconv.Atoi(semVerStrings[1]); err != nil {
		return nil, digitError
	}

	var ps []rune

	// Now to extract the patch and any follow on
	for i, char := range semVerStrings[2] {
		if !unicode.IsDigit(char) {
			parsedSemver.Leftover = semVerStrings[2][i:]
			break
		}
		ps = append(ps, char)
	}

	if parsedSemver.Patch, err = strconv.Atoi(string(ps)); err != nil {
		return nil, digitError
	}

	return parsedSemver, nil

}

// Greater checks if the specified version is greater than the current version
func (s *SemVer) Greater(s2 *SemVer) bool {
	switch {
	case s.Major == s2.Major && s.Minor == s2.Minor:
		return s.Patch > s2.Patch
	case s.Major == s2.Major:
		return s.Minor > s2.Minor
	}

	return s.Major > s2.Major
}

// String returns the string representation of the semver
func (s *SemVer) String() string {
	return fmt.Sprintf("%d.%d.%d", s.Major, s.Minor, s.Patch)
}
