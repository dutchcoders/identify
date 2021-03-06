package app

import (
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path"
	"sort"
	"strings"

	"github.com/cheggaaa/pb"
	"github.com/fatih/color"
	version "github.com/hashicorp/go-version"
	_ "github.com/minio/cli"
	_ "github.com/op/go-logging"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"

	"gopkg.in/src-d/go-billy.v2/osfs"

	"bytes"

	"encoding/hex"

	yaml "gopkg.in/yaml.v2"
)

type identify struct {
	config

	client *http.Client

	debug bool

	hashes   map[string]*Result
	versions []string

	application *Application
	db          *DB

	cachePath string

	proxyURL *url.URL

	r *git.Repository
}

func Download(src string, dest string) error {
	resp, err := http.Get(src)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	f, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func New(options ...OptionFn) (*identify, error) {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial:  net.Dial,
	}

	cachePath := ""
	if usr, err := user.Current(); err != nil {
		return nil, err
	} else {
		cachePath = path.Join(usr.HomeDir, ".identify")
	}

	b := &identify{
		client: &http.Client{
			Transport: transport,
		},
		hashes:    map[string]*Result{},
		versions:  []string{},
		cachePath: cachePath,
	}

	for _, optionFunc := range options {
		if err := optionFunc(b); err != nil {
			return nil, err
		}
	}

	if _, err := os.Stat(b.cachePath); err == nil {
	} else if !os.IsNotExist(err) {
		return nil, err
	} else if err = os.Mkdir(b.cachePath, 0700); err != nil {
		return nil, err
	}

	dbPath := path.Join(b.cachePath, "db.yaml")

	if _, err := os.Stat(dbPath); err == nil {
	} else if !os.IsNotExist(err) {
		return nil, err
	} else if err := Download("https://raw.githubusercontent.com/dutchcoders/identify/master/db.yaml", dbPath); err != nil {
		return nil, err
	} else {
	}

	// load database with application sets
	data, err := ioutil.ReadFile(dbPath)
	if err != nil {
		return nil, err
	}

	// load configuration database
	var db DB

	err = yaml.Unmarshal(data, &db.Application)
	if err != nil {
		return nil, err
	}

	b.db = &db

	if application, ok := b.db.Application[b.targetApplication]; !ok {
		return nil, fmt.Errorf("Application not found in rule set")
	} else {
		b.application = &application
	}

	return b, nil
}

func hashStr(s string) string {
	h := sha1.New()
	io.WriteString(h, s)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func CalcHash(r io.ReadCloser) []byte {
	h := sha1.New()

	_, err := io.Copy(h, r)
	if err != nil {
		fmt.Println(err.Error())
	}

	r.Close()

	return h.Sum(nil)
}

func normalize(name plumbing.ReferenceName) string {
	s := name.String()
	s = strings.Replace(s, "refs/tags/", "", -1)
	s = strings.Replace(s, "refs/heads/", "", -1)
	return s
}

func (b *identify) WorkReference(ref *plumbing.Reference) error {
	b.versions = append(b.versions, normalize(ref.Name()))

	var tree *object.Tree
	if c, err := b.r.CommitObject(ref.Hash()); err == nil {
		tree, _ = c.Tree()
	} else if c, err := b.r.TagObject(ref.Hash()); err == nil {
		tree, _ = c.Tree()
	} else if err != nil {
		fmt.Println(color.RedString("Could not find commit or tag for %s: %s %s", ref.Name(), ref.Hash().String(), err.Error()))
		return nil
	}

	if tree == nil {
		return fmt.Errorf("Could not find tree for commit or tag")
	}

	for fileName, hash := range b.hashes {
		f, err := tree.File(path.Join(b.application.Root, fileName))
		if err != nil {
			// verify error
			continue
		}

		rdr, err := f.Reader()
		if err != nil {
			return err
		}

		h := CalcHash(rdr)
		if bytes.Compare(hash.Hash, h) != 0 {
			continue
		}

		hash.AddRef(ref)
	}

	return nil
}

func (b *identify) Identify() error {
	fmt.Printf("| Application: %s\n", b.application.Name)
	fmt.Printf("| Target URL: %s\n", b.targetURL.String())

	if b.proxyURL != nil {
		fmt.Printf("| Using proxy: %s\n", b.proxyURL.String())
	}

	fmt.Println("")

	fmt.Println(color.YellowString("[+] Calculating hashes for remote files"))

	bar := pb.New(len(b.application.Files))
	bar.SetWidth(40)
	bar.SetMaxWidth(40)
	bar.Format("[## ]")
	bar.ShowCounters = true
	bar.ShowFinalTime = false
	bar.ShowPercent = false
	bar.Start()

	for _, file := range b.application.Files {
		rel, err := url.Parse(file)
		if err != nil {
			fmt.Println(color.RedString("[!] Could not parse url %s: %s", file, err.Error()))
			continue
		}

		abs := b.targetURL.ResolveReference(rel)

		resp, err := b.client.Get(abs.String())
		if err != nil {
			fmt.Println(color.RedString("[!] Could not download url %s: %s", rel, err.Error()))
			continue
		}

		bar.Increment()

		if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		} else {
			fmt.Println(color.RedString("[!] Error downloading %s got status code: %d", abs.String(), resp.StatusCode))
			continue
		}

		hash := CalcHash(resp.Body)

		if b.debug {
			fmt.Printf("[ ] Downloaded %s (%d): %x\n", abs.String(), resp.StatusCode, hash)
		}

		b.hashes[file] = &Result{
			Hash: hash,
			Refs: []*plumbing.Reference{},
		}
	}

	bar.Finish()

	repoCachePath := path.Join(b.cachePath, hashStr(b.application.Repository))

	storage, err := filesystem.NewStorage(osfs.New(repoCachePath))
	if err != nil {
		return err
	}

	fmt.Println(color.YellowString("[+] Cloning repository to cache"))

	r, err := git.Open(storage, nil)
	if err == nil {
	} else if err.Error() != "repository not exists" {
		// unknown open error
		return err
	} else if r, err = git.Clone(storage, nil, &git.CloneOptions{
		URL:      b.application.Repository,
		Progress: os.Stdout,
	}); err != nil {
		return err
	}

	fmt.Println(color.YellowString("[+] Pulling latest changes from remote repository"))

	err = r.Fetch(&git.FetchOptions{
		Progress: os.Stdout,
	})
	if err == nil {
	} else if err.Error() == "already up-to-date" {
		fmt.Println(" |  Repository already up-to-date")
	} else {
		return err
	}

	b.r = r

	if b.noBranches {
	} else if ri, err := r.Branches(); err != nil {
		return err
	} else if err := ri.ForEach(b.WorkReference); err != nil {
		return err
	}

	if b.noTags {
	} else if ri, err := r.Tags(); err != nil {
		return err
	} else if err := ri.ForEach(b.WorkReference); err != nil {
		return err
	}

	// convert refs to versions
	Setify := func(refs []*plumbing.Reference) []string {
		vals := make([]string, len(refs))

		for i, _ := range refs {
			vals[i] = normalize(refs[i].Name())
		}

		return vals
	}

	for fileName, hash := range b.hashes {
		versions := Setify(hash.Refs)

		if b.debug {
			fmt.Printf("-> file: %s (%s): versions: %s\n", fileName, hex.EncodeToString(hash.Hash), strings.Join(versions, ", "))
		}
	}

	// count all refs
	counts := map[string]int{}

	for _, hash := range b.hashes {
		for _, ref := range hash.Refs {
			version := normalize(ref.Name())

			if _, ok := counts[version]; !ok {
				counts[version] = 0
			}

			counts[version]++
		}
	}

	if len(counts) == 0 {
		fmt.Println(color.RedString("Could not identify web application"))
		return nil
	}

	/*
		versionsRaw := s.List()

		versions := make([]*version.Version, len(versionsRaw))
		for i, raw := range versionsRaw {
			v, err := version.NewVersion(raw)
			if err != nil {
				fmt.Println(color.RedString("Could not identify version: %s: %s", raw, err.Error()))
				continue
			}

			versions[i] = v
		}

		sort.Sort(version.Collection(versions))

	*/
	fmt.Printf("\n")

	// print identification summary
	fmt.Printf(color.GreenString("[+] Web application has been identified as one of the following versions: \n"))

	n := map[int][]string{}

	var a []int
	for k, v := range counts {
		n[v] = append(n[v], k)
	}

	for k := range n {
		a = append(a, k)
	}

	sort.Sort(sort.Reverse(sort.IntSlice(a)))

	for _, k := range a {
		s2 := []string{}

		for _, s := range n[k] {
			v, _ := version.NewVersion(s)
			if v != nil {
				s2 = append(s2, v.String())
			} else {
				s2 = append(s2, s)
			}
		}

		fmt.Println(color.GreenString(" |  %3.0f%% %s", ((float64(k) * 100) / float64(len(b.application.Files))), strings.Join(s2, ", ")))
	}

	fmt.Printf("\n")
	return nil
}
