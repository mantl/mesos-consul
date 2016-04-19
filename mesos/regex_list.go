package mesos

import (
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

type RegexList struct {
	List []string
	Regex *regexp.Regexp
}

func NewRegexList(l []string) *RegexList {
	rval := &RegexList{
		List: l,
	}

	rval.Compile(l)

	return rval
}

func (rl *RegexList) Compile(l []string) {
	log.WithField("l", l).Debug("Using regex list")
	log.WithField("len(l)", len(l)).Debug("List length")
	if len(l) > 0 {
		lstring := strings.Join(l, "|")
		log.WithField("regex_string", lstring).Debug("Using regex string")
		re, err := regexp.Compile(lstring)
		if err != nil {
			// Emit a warning message stating that the regex didn't compile
			// and leave the Regex member set to the last value
			log.WithField("regex_string", lstring).Warn("Regex failed to compile")
		}

		rl.Regex = re
	}
}

func (rl *RegexList) MatchString(s string, def bool) bool {
	if rl.Regex != nil {
		return rl.Regex.MatchString(s)
	}

	// Return default value if no regex
	return def
}
