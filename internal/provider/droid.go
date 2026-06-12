package provider

import (
	"github.com/zhenninglang/mantis/internal/session"
)

type DroidAdapter struct{}

func (DroidAdapter) ID() ID              { return Droid }
func (DroidAdapter) DisplayName() string { return "Droid" }
func (DroidAdapter) Capabilities() Capabilities {
	return Capabilities{Resume: true, Fork: true, Rename: true, Delete: true, Inspect: true}
}

func (DroidAdapter) Discover() ([]session.Session, []Diagnostic) {
	sessions, err := session.LoadDroid()
	if err != nil {
		return nil, []Diagnostic{{Provider: Droid, Message: err.Error()}}
	}
	for i := range sessions {
		sessions[i].Provider = string(Droid)
		sessions[i].ProviderName = "Droid"
		sessions[i].ResumeRef = sessions[i].Meta.ID
		sessions[i].ForkRef = sessions[i].Meta.ID
	}
	return sessions, nil
}

func (DroidAdapter) Resume(s session.Session) error {
	return runCommand("droid", "-r", s.ResumeRef)
}

func (DroidAdapter) Fork(s session.Session) error {
	return runCommand("droid", "--fork", s.ForkRef)
}
