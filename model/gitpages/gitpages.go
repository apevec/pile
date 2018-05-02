// Package gitwiki - tbd
package gitpages

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/microcosm-cc/bluemonday"
	blackfriday "gopkg.in/russross/blackfriday.v2"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

var (
	giturl  = os.Getenv("GIT")
	gitdir  = os.Getenv("GIT_DIR")
	gitusr  = os.Getenv("GIT_USER")
	gitssh  = os.Getenv("GIT_SSH")
	gitpage = "home.md"
)

func GetPage() (string, time.Time) {

	input, modified := GetPageRaw()

	unsafe := blackfriday.Run(input)
	html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)

	return string(html), modified
}

func GetPageRaw() ([]byte, time.Time) {
	if !exists(gitdir) {
		clone()
	}

	info, err := os.Stat(gitdir + gitpage)
	if err != nil {
		log.Println(err)
	}

	file, err := ioutil.ReadFile(gitdir + gitpage)
	if err != nil {
		log.Println(err)
	}

	modified := info.ModTime()

	return file, modified
}

func Update(page string, change string) error {

	// dos2unix(page): convert /r/n to /n, markdown needs it to generate <p> properly
	page_unix := bytes.Replace([]byte(page), []byte("\r\n"), []byte("\n"), -1)

	err := ioutil.WriteFile(gitdir+gitpage, []byte(page_unix), 0644)
	if err != nil {
		log.Println(err)
	}

	commit(change)
	push()

	return err
}

func push() error {
	sshAuth, err := ssh.NewPublicKeysFromFile(gitusr, gitssh, "")
	if err != nil {
		log.Println(err)
	}

	r, err := git.PlainOpen(gitdir)
	if err != nil {
		log.Println(err)
	}

	// push using default options
	err = r.Push(&git.PushOptions{
		Auth: sshAuth,
	})
	if err != nil {
		log.Println(err)
	}

	return err
}

func commit(change string) error {
	r, err := git.PlainOpen(gitdir)
	if err != nil {
		log.Println(err)
	}

	w, err := r.Worktree()
	if err != nil {
		log.Println(err)
	}

	_, err = w.Add(gitpage)
	if err != nil {
		log.Println(err)
	}

	status, err := w.Status()
	if err != nil {
		log.Println(err)
	}
	log.Println(status)

	commit, err := w.Commit(change, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "RHOS Roster",
			Email: "rhos@roster.nowhere",
			When:  time.Now(),
		},
	})
	if err != nil {
		log.Println(err)
	}

	obj, err := r.CommitObject(commit)
	if err != nil {
		log.Println(err)
	}
	log.Println(obj)

	return err
}

func clone() error {
	// Clone a repo, just a first shot... we may want to keep the data in memory instead
	sshAuth, err := ssh.NewPublicKeysFromFile(gitusr, gitssh, "")
	if err != nil {
		log.Println(err)
	}
	r, err := git.PlainClone(gitdir, false, &git.CloneOptions{
		URL:               giturl,
		ReferenceName:     "refs/heads/master",
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		Auth:              sshAuth,
	})

	if err != nil {
		log.Println(err)
	}
	log.Println(r)

	return err
}

// Exists reports whether the named file or directory exists.
func exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
