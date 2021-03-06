/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2018 Zachary Schneider
 */

package action

// TODO look into cleaning this up with reflection

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

type Manifest struct {
	Zpkg *Zpkg `hcl:"Zpkg,block" json:"zpkg"`

	Tags []*Tag `hcl:"Tag,block" json:"tag,omitempty"`

	Requirements []*Requirement `hcl:"Requirement,block" json:"requirement,omitempty"`

	Dirs     []*Dir     `hcl:"Dir,block" json:"dir,omitempty"`
	Files    []*File    `hcl:"File,block" json:"file,omitempty"`
	SymLinks []*SymLink `hcl:"SymLink,block" json:"symlink,omitempty"`

	Templates []*Template `hcl:"Template,block" json:"template,omitempty"`

	Services []*Service `hcl:"Service,block" json:"service,omitempty"`

	Signatures []*Signature `hcl:"Signature,block" json:"signature,omitempty"`

	index map[string]int
}

func NewManifest() *Manifest {
	return &Manifest{index: make(map[string]int)}
}

func (m *Manifest) Add(action Action) {
	if len(m.index) == 0 {
		m.Index()
	}

	switch action.Type() {
	case "Tag":
		if m.Exists(action) {
			m.Tags[m.index[action.Id()]] = action.(*Tag)
		} else {
			m.Tags = append(m.Tags, action.(*Tag))
			m.index[action.Id()] = len(m.Tags) - 1
		}
	case "Requirement":
		if m.Exists(action) {
			m.Requirements[m.index[action.Id()]] = action.(*Requirement)
		} else {
			m.Requirements = append(m.Requirements, action.(*Requirement))
			m.index[action.Id()] = len(m.Requirements) - 1
		}
	case "Dir":
		if m.Exists(action) {
			m.Dirs[m.index[action.Id()]] = action.(*Dir)
		} else {
			m.Dirs = append(m.Dirs, action.(*Dir))
			m.index[action.Id()] = len(m.Dirs) - 1
		}
	case "File":
		if m.Exists(action) {
			m.Files[m.index[action.Id()]] = action.(*File)
		} else {
			m.Files = append(m.Files, action.(*File))
			m.index[action.Id()] = len(m.Files) - 1
		}
	case "Template":
		if m.Exists(action) {
			m.Templates[m.index[action.Id()]] = action.(*Template)
		} else {
			m.Templates = append(m.Templates, action.(*Template))
			m.index[action.Id()] = len(m.Templates) - 1
		}
	case "Service":
		if m.Exists(action) {
			m.Services[m.index[action.Id()]] = action.(*Service)
		} else {
			m.Services = append(m.Services, action.(*Service))
			m.index[action.Id()] = len(m.Services) - 1
		}
	case "Signature":
		if m.Exists(action) {
			m.Signatures[m.index[action.Id()]] = action.(*Signature)
		} else {
			m.Signatures = append(m.Signatures, action.(*Signature))
			m.index[action.Id()] = len(m.Signatures) - 1
		}
	case "SymLink":
		if m.Exists(action) {
			m.SymLinks[m.index[action.Id()]] = action.(*SymLink)
		} else {
			m.SymLinks = append(m.SymLinks, action.(*SymLink))
			m.index[action.Id()] = len(m.SymLinks) - 1
		}
	}
}

func (m *Manifest) Exists(action Action) bool {
	if _, ok := m.index[action.Id()]; ok {
		return true
	}

	return false
}

func (m *Manifest) Section(filters ...string) Actions {
	var items []Action

	for _, filter := range filters {
		switch filter {
		case "Tag":
			for _, item := range m.Tags {
				items = append(items, item)
			}
		case "Requirement":
			for _, item := range m.Requirements {
				items = append(items, item)
			}
		case "Dir":
			for _, item := range m.Dirs {
				items = append(items, item)
			}
		case "File":
			for _, item := range m.Files {
				items = append(items, item)
			}
		case "SymLink":
			for _, item := range m.SymLinks {
				items = append(items, item)
			}
		case "Template":
			for _, item := range m.Templates {
				items = append(items, item)
			}
		case "Service":
			for _, item := range m.Services {
				items = append(items, item)
			}
		case "Signature":
			for _, item := range m.Signatures {
				items = append(items, item)
			}
		}
	}

	return items
}

func (m *Manifest) Index() {
	for index, act := range m.Tags {
		m.index[act.Id()] = index
	}

	for index, act := range m.Requirements {
		m.index[act.Id()] = index
	}

	for index, act := range m.Dirs {
		m.index[act.Id()] = index
	}

	for index, act := range m.Files {
		m.index[act.Id()] = index
	}

	for index, act := range m.SymLinks {
		m.index[act.Id()] = index
	}

	for index, act := range m.Templates {
		m.index[act.Id()] = index
	}

	for index, act := range m.Services {
		m.index[act.Id()] = index
	}

	for index, act := range m.Signatures {
		m.index[act.Id()] = index
	}
}

func (m *Manifest) Actions() Actions {
	var actions Actions

	fs := m.Section("Dir", "File", "SymLink")
	sort.Sort(fs)

	actions = append(actions, m.Zpkg)
	actions = append(actions, m.Section("Tag")...)
	actions = append(actions, m.Section("Requirement")...)
	actions = append(actions, m.Section("Template")...)
	actions = append(actions, m.Section("Service")...)
	actions = append(actions, m.Section("Signature")...)
	actions = append(actions, fs...)

	return actions
}

func (m *Manifest) Validate() error {
	var actions Actions

	// Do not allow a requirement to name itself
	reqs := m.Section("Requirement")
	for _, req := range reqs {
		if req.(*Requirement).Name == m.Zpkg.Name {
			return errors.New("Action Requirement: cannot reference itself")
		}
	}

	// Ensure there are no duplicate paths present for FS actions
	actions = m.Section("Dir", "File", "SymLink")

	sort.Sort(actions)
	for index, action := range actions {
		prev := index - 1
		if prev != -1 {
			if action.Key() == actions[prev].Key() {
				return errors.New(fmt.Sprint(
					"Action Conflicts:\n",
					strings.ToUpper(actions[prev].Type()), " => ", actions[prev].Key(), "\n",
					strings.ToUpper(action.Type()), " => ", action.Key()))
			}
		}
	}

	// Ensure that template actions refer to an existing file
	tpls := m.Section("Template")
	for _, tpl := range tpls {
		if _, ok := m.index["File."+tpl.(*Template).Source]; ok {
			continue
		} else {
			return errors.New("Action Template: template does not source packaged file")
		}
	}

	// TODO add a check to ensure that service includes the systemd unit

	return nil
}

func (m *Manifest) ToJson() string {
	out, _ := json.Marshal(m)

	return string(out)
}

func (m *Manifest) ToSigningJson() string {
	s := *m

	s.Signatures = nil

	out, _ := json.Marshal(s)

	return string(out)
}

func (m *Manifest) Load(manifest string) error {
	err := json.Unmarshal([]byte(manifest), m)
	if err != nil {
		return err
	}

	m.Index()

	return nil
}
