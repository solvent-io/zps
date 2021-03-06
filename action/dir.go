/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2018 Zachary Schneider
 */

package action

import (
	"fmt"
	"strings"
)

type Dir struct {
	Path  string `json:"path" hcl:"path,label"`
	Owner string `json:"owner" hcl:"owner,optional"`
	Group string `json:"group" hcl:"group,optional"`
	Mode  string `json:"mode" hcl:"mode,optional"`
}

func NewDir() *Dir {
	return &Dir{}
}

func (d *Dir) Key() string {
	return d.Path
}

func (d *Dir) Type() string {
	return "Dir"
}

func (d *Dir) Columns() string {
	return strings.Join([]string{
		strings.ToUpper(d.Type()),
		d.Mode,
		d.Owner + ":" + d.Group,
		d.Path,
	}, "|")
}

func (d *Dir) Id() string {
	return fmt.Sprint(d.Type(), ".", d.Key())
}

func (d *Dir) Condition() *bool {
	return nil
}

func (d *Dir) MayFail() bool {
	return false
}

func (d *Dir) IsValid() bool {
	if d.Path != "" && d.Owner != "" && d.Group != "" && d.Mode != "" {
		return true
	}

	return false
}
