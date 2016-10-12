package main

import (
    "github.com/forj-oss/goforjj"
    "github.com/forj-oss/forjj-modules/trace"
    "log"
    "os"
    "fmt"
    "gopkg.in/yaml.v2"
    "io/ioutil"
)

const forjj_repo_file = "forjj-repos.yml"

// Load defaults, then load repos from source then add from cli.
func (a *Forj)BuildReposList(action string) error{
    // Set forjj-options defaults for new repositories.
    a.SetDefault(action)

    // Read Repos list from infra-repo/forjj-repos.yaml
    if err := a.RepoCodeLoad() ; err != nil {
        return err
    }

    // Add cli repos list.
    a.AddReposFromCli(action)

    return nil
}

// Identify number of repository requested for an instance.
func (a *Forj)GetReposRequestedFor(instance, action string) (num int) {
    if instance == "" || action == "" {
        gotrace.Trace("Internal error: instance and action cannot be empty.")
        return
    }
    if _, found := a.Actions[action] ; ! found {
        gotrace.Trace("Internal error: action '%s' not found.", action)
        return
    }
    for _, rd := range a.Actions[action].repoList.Repos {
        if rd.Instance == instance || (rd.Instance == "" && instance == a.o.Defaults["instance"]) {
            num ++
        }
    }
    return
}

// Function providing a PluginRepoData content for the instance given.
func (a *Forj)GetReposData(instance string) (ret map[string]goforjj.PluginRepoData) {
    gotrace.Trace("Forjj managed %d repositories (forjj-repos.yml)", len(a.r.Repos))
    ret = make(map[string]goforjj.PluginRepoData)
    for n, d := range a.r.Repos {
        if d.Instance != instance {
            continue
        }
        ret[n] = *d
    }
    gotrace.Trace("%d repositories identified for instance %s", len(ret), instance)
    return
}


// Stored Repositories managed by the plugin in the list of repos (forjj-repos.yaml)
func (a *Forj)SaveManagedRepos(d *Driver, instance string) {
    for name, repo := range a.r.Repos {
        if _, found := d.plugin.Result.Data.Repos[name] ; found  {
            // Saving infra repository information to the workspace
            repo.Instance = instance
        }
    }
}

// Update a Repolist from another list.
// If new, added. If both exist, update from source.
func (r *ReposList)UpdateFromList(source *ReposList, defaults map[string]string) {
    for name, repo := range source.Repos {
        if d, found := r.Repos[name] ; found {
            repo.SetDefaults(defaults)
            d.UpdateFrom(repo)
        } else {
            repo.SetDefaults(defaults)
            r.Repos[name] = repo
        }
    }
}

func NumDisplay(num int, format, elements, element string) string{

    if num > 1 {
        return fmt.Sprintf(format, num, elements)
    }
    return fmt.Sprintf(format, num, element)
}

func NumReposDisplay(num int) string {
    return NumDisplay(num, "%d repositor%s", "ies", "y")
}

// Function to update Forjj Repos list. Use RepodSave to Save it as code ie in forjj-repos.yml.
// NOTE: a repo can be only created. Never updated or deleted. A repo has his own lifecycle not managed by forjj.
func (a *Forj)AddReposFromCli(action string) {
    gotrace.Trace("Forjj managed %s.",  NumReposDisplay(len(a.r.Repos)))

    a.r.UpdateFromList(a.Actions[action].repoList, a.o.Defaults)
    gotrace.Trace("Now, Forjj manages %s. cli added %s.", NumReposDisplay(len(a.r.Repos)), NumReposDisplay(len(a.Actions[action].repoList.Repos)))
}

//Function to save forjj list of Repositories.
func (a *Forj)RepoCodeSave() (err error) {
    if yd, err := yaml.Marshal(a.r) ; err == nil {
        if err := ioutil.WriteFile(forjj_repo_file, yd, 0644) ; err != nil {
            return fmt.Errorf("Unable to write '%s'. %s", forjj_repo_file, err)
        }
    } else {
        return fmt.Errorf("Unable to encode to yaml '%s'. %s", forjj_repo_file, err)
    }

    gotrace.Trace("%s written with %s.", forjj_repo_file, NumReposDisplay(len(a.r.Repos)))

    git("add", forjj_repo_file)
    return nil
}

// Read the collection of repositories managed by forjj.
func (a *Forj)RepoCodeLoad() (error) {
    a.r.Repos = make(map[string]*goforjj.PluginRepoData)

    if _, err := os.Stat(forjj_repo_file) ; err != nil {
        gotrace.Trace("%s not found. %s.", forjj_repo_file, err)
        return nil
    }
    if d, err := ioutil.ReadFile(forjj_repo_file) ; err == nil {
        if err := yaml.Unmarshal(d, a.r) ; err != nil {
            return fmt.Errorf("Unable to decode '%s'. %s", forjj_repo_file, err)
        }
    } else {
        return fmt.Errorf("Unable to read '%s'. %s", forjj_repo_file, err)
    }

    gotrace.Trace("%s loaded from forjj-repos.yml", NumReposDisplay(len(a.r.Repos)))

    return  nil
}

func (a *Forj)DriverGet(instance string) (d *Driver){
    var found bool

    if d, found = a.drivers[instance] ; found {
            return
        }

    if d, found = a.o.Drivers[instance] ; ! found {
        a.drivers[instance] = &Driver{
            Name:          d.Name,
            DriverType:    d.DriverType,
            InstanceName:  d.InstanceName,
        }
        return
    }
    return nil
}

// Function to create missing repositories in the upstream defined.
// It should do:
// - create missing repos
// - set appropriate local repo config (upstream) depending on flows definition.
func (a *Forj)RepoMaintain(){
    for name, repo := range a.r.Repos {
        // Build the tree and save then (git).
        // mkdir in repos/{repo}/repo.yaml
        gotrace.Trace("Create Repo %s on instance %s", name, repo.Instance)
        d := a.DriverGet(repo.Instance)
        if d == nil {
            log.Printf("Unable to create code for Repo '%s'. Instance '%s' not found. Ignored.", name, repo.Instance)
            continue
        }
        // Ask upstream driver to create the repo. Except if the driver is none
        // Expect flow to be used

        // Create local repo

        // Sync with upstream if not "none"
    }

}
