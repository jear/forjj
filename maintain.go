package main

import (
	"fmt"
	"github.com/forj-oss/forjj-modules/trace"
	"forjj/git"
)

// Call docker to create the Solution source code from scratch with validated parameters.
// This container do the real stuff (git/call drivers)
// I would expect to have this go tool to have a do_create to replace the shell script.
// But this would be a next version and needs to be validated before this decision is made.
// Workspace information already loaded by the cli context.
func (a *Forj) Maintain() error {
	if _, err := a.w.Check_exist(); err != nil {
		return fmt.Errorf("Invalid workspace. %s. Please create it with 'forjj create'", err)
	}

	a.ScanAndSetObjectData()

	gotrace.Trace("Infra upstream selected: '%s'", a.w.Instance)

	if  err := a.get_infra_repo(); err != nil {
		return fmt.Errorf("Invalid workspace. %s. Please create it with 'forjj create'", err)
	}

	// Now, we are in the infra repo root directory and at least, the 1st commit exist.

	// Load drivers from forjj-options.yml
	// loop from options/Repos and keep them in a.drivers

	return a.do_maintain()
}

func (a *Forj) do_maintain() error {
	// Loop on instances to maintain them
	instances := a.define_drivers_execution_order()
	for _, instance := range instances {
		if err := a.do_driver_maintain(instance); err != nil {
			return fmt.Errorf("Unable to maintain requested resources of %s. %s", instance, err)
		}
	}
	return nil
}

func (a *Forj) do_driver_maintain(instance string) error {
	if instance == "none" {
		return nil
	}

	gotrace.Trace("Start maintaining instance '%s'", instance)
	if err := a.driver_start(instance); err != nil {
		return err
	}
	d := a.CurrentPluginDriver

	// Ensure remote upstream exists - calling upstream driver - maintain
	// This will create/update the upstream service
	if err, _ := a.driver_do(d, instance, "maintain"); err != nil {
		return fmt.Errorf("Driver issue. %s.", err)
	}

	if a.f.GetInfraInstance() == instance {
		// Update git remote and 'master' branch to infra repository.
		var infra_name string
		if i, found, err := a.GetPrefs(infra_name_f) ; err != nil {
			return err
		} else {
			if !found {
				return nil
			}
			infra_name = i
		}
		if r, found := d.Plugin.Result.Data.Repos[infra_name] ; found {
			for name, remote := range r.Remotes {
				a.i.EnsureGitRemote(remote.Ssh, name)
			}
			for branch, remote := range r.BranchConnect {
				status, err := a.i.EnsureBranchConnected(branch, remote)
				if err != nil { return err }
				switch status {
				case "-1" :
					return fmt.Errorf("Warning! Remote branch is most recent than your local branch. " +
						"Do a git pull and restart 'forjj maintain'")
				case "+1" :
					git.Push()
				case "-1+1" :
					return fmt.Errorf("Local and remote branch has diverged. You must fix it before going on.")
				}
			}
		}
	}
	return nil
}

// get_infra_repo detect in the path given contains the infra repository.
func (a *Forj) get_infra_repo() error {
	return a.i.Use(a.f.InfraPath())
}
