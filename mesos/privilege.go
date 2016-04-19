package mesos

import (
	log "github.com/sirupsen/logrus"
)

type Privilege struct {
	WhiteList *RegexList
	BlackList *RegexList
}

func NewPrivilege(w []string, b []string) *Privilege {
	return &Privilege{
		WhiteList: NewRegexList(w),
		BlackList: NewRegexList(b),
	}
}

func (p *Privilege) Allowed(name string) bool {
	if !p.WhiteList.MatchString(name, true) {
		log.WithField("name", name).Debug("Not on whitelist")
		return false
	}

	if p.BlackList.MatchString(name, false) {
		log.WithField("name", name).Debug("On blacklist")
		return false
	}

	return true
}
