// Package gitwiki - tbd
package gitpages

import (
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/microcosm-cc/bluemonday"
	blackfriday "gopkg.in/russross/blackfriday.v2"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

var (
	giturl  = os.Getenv("GIT")
	gitdir  = os.Getenv("GIT_DIR")
	gitusr  = os.Getenv("GIT_USER")
	gitpage = "home.md"
)

func GetPage() (string, time.Time) {

	if !exists(gitdir) {
		// Clone a repo, just a first shot... we may want to keep the data in memory instead
		sshAuth, err := ssh.NewPublicKeysFromFile(gitusr, "/secrets/.ssh/id_rsa", "")
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
	}

	info, err := os.Stat(gitdir + gitpage)
	if err != nil {
		log.Println(err)
	}

	file, err := ioutil.ReadFile(gitdir + gitpage)
	if err != nil {
		log.Println(err)
	}

	unsafe := blackfriday.Run(file)
	html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)

	modified := info.ModTime()

	return string(html), modified
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
