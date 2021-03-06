/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2018 Zachary Schneider
 */

package provider

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"github.com/chuckpreslar/emission"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"

	"context"

	"github.com/fezz-io/zps/action"
)

type TemplateDefault struct {
	*emission.Emitter
	template *action.Template

	phaseMap map[string]string
}

func NewTemplateDefault(template action.Action, phaseMap map[string]string, emitter *emission.Emitter) Provider {
	return &TemplateDefault{emitter, template.(*action.Template), phaseMap}
}

func (t *TemplateDefault) Realize(ctx context.Context) error {
	switch t.phaseMap[Phase(ctx)] {
	case "configure":
		return t.configure(ctx)
	default:
		t.Emit("action.info", fmt.Sprintf("%s %s", t.template.Type(), t.template.Key()))
		return nil
	}
}

func (t *TemplateDefault) configure(ctx context.Context) error {
	options := Opts(ctx)

	// Process template
	configBytes, err := ioutil.ReadFile(filepath.Join(options.TargetPath, t.template.Source))
	if err != nil {
		return err
	}

	expr, diags := hclsyntax.ParseTemplate(configBytes, t.template.Source, hcl.Pos{})
	if diags.HasErrors() {
		return diags
	}

	// TODO build eval context upstream and pass via context
	val, diags := expr.Value(ctx.Value("hclCtx").(*hcl.EvalContext))
	if diags.HasErrors() {
		return diags
	}

	if t.template.Output != "" {
		output := filepath.Join(options.TargetPath, t.template.Output)

		modeString := t.template.Mode
		if modeString == "" {
			modeString = "0640"
		}

		mode, err := strconv.ParseUint(modeString, 0, 0)
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(output, []byte(val.AsString()), os.FileMode(mode))
		if err != nil {
			return err
		}
		// Silent failures are fine, only a super user can chown to another user
		// Also a given user may not exist on a system though we should catch
		// that elsewhere

		owner, _ := user.Lookup(t.template.Owner)
		grp, _ := user.LookupGroup(t.template.Group)
		var uid int64
		var gid int64

		if owner != nil && grp != nil {
			uid, _ = strconv.ParseInt(owner.Uid, 0, 0)
			gid, _ = strconv.ParseInt(grp.Gid, 0, 0)
		}

		os.Chown(output, int(uid), int(gid))

		t.Emit("action.info", fmt.Sprintf(
			"%s %s %s => %s",
			t.template.Type(),
			t.template.Key(),
			filepath.Join(options.TargetPath, t.template.Source),
			filepath.Join(options.TargetPath, t.template.Output),
		))
	} else {
		_, err = os.Stdout.Write([]byte(val.AsString()))
	}

	return err
}
