package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gitw/internal/shell"
	"gitw/internal/xregexp"

	"gitw/version"

	"github.com/c-bata/go-prompt"
)

var verbose bool

func main() {
	if len(os.Args) > 2 {
		usage()
	}
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	version.ExeDir = filepath.Dir(ex)
	check := false
	if len(os.Args) == 2 && os.Args[1] == "check" {
		check = true
	}
	if len(os.Args) == 2 && os.Args[1] == "version" {
		fmt.Println(version.String())
		os.Exit(0)
	}
	verbose = len(os.Getenv("verbose")) > 0
	err = shell.CleanOldBashFiles(verbose)
	if err != nil {
		log.Fatalf("Unable to cleanup old /tmp/bash.* files: '%+v'", err)
	}
	asship := getSSHConnection()
	if verbose {
		fmt.Printf("sship='%s'\n", asship)
	}
	var user *user
	ub := newUsersBase(gitusersPath())
	if asship != "" {
		user = ub.getUser(asship)
		if user == nil {
			if check {
				os.Exit(1)
			}
			user = ub.askUserID()
			ub.recordUser(user, asship)
		}
	} else {
		s, err := getBashSession()
		if err != nil {
			if check {
				os.Exit(1)
			}
			log.Fatalf("No session found (%+v)", err)
		}
		b := newBash(s)
		user = b.getUser()
		if user == nil {
			if check {
				os.Exit(1)
			}
			user = ub.askUserID()
			b.recordUser(user)
		}
	}
	if user != nil {
		if check {
			os.Exit(0)
		}
		if !ub.userAsked {
			fmt.Printf("%s/%s", user.name, user.email)
		}
	} else {
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("gitw <check>")
	log.Fatalf("Incorrect parameters: '%+v'", os.Args[:])
}

func gitusersPath() string {
	home := os.Getenv("HOME")
	info, err := os.Stat(home + "/.gitusers")
	if err == nil && !info.IsDir() {
		return home + "/.gitusers"
	}
	return version.ExeDir + "/.gitusers"
}

type sship string

func getSSHConnection() sship {
	res := os.Getenv("SSH_CONNECTION")
	if len(res) > 0 {
		fields := strings.Fields(res)
		if len(fields) > 0 {
			res = fields[0]
		}
	}
	if verbose {
		fmt.Printf("getSSHConnection res='%s'\n", res)
	}
	//return sship("")
	return sship(res)
}

type user struct {
	name  string
	email string
	ip    sship
}

type users []*user

func (us users) addUser(u *user, ip sship) users {
	if !us.hasUser(u) {
		us = append(us, u)
	}
	u.setSSHIP(ip)
	return us
}

func (us users) hasUser(u *user) bool {
	for _, auser := range us {
		if auser.email == u.email {
			return true
		}
	}
	return false
}

func (u *user) setSSHIP(ip sship) {
	if !ip.isNul() || u.ip.isNul() {
		u.ip = ip
	}
}

func (ip sship) isNul() bool {
	if ip == "" || ip == "0.0.0.0" {
		return true
	}
	return false
}

type usersBase struct {
	gitusers  string
	users     users
	userAsked bool
}

type choice struct {
	user     *user
	ub       *usersBase
	mustExit bool
}

var re = regexp.MustCompile(`(?m)^(?P<ip>(\d+\.?)+)~(?P<name>.*?)~(?P<email>(.*?)@(.*))`)

func newUsersBase(file string) *usersBase {

	sc := &usersBase{
		gitusers:  file,
		users:     []*user{},
		userAsked: false,
	}

	fi, err := os.OpenFile(sc.gitusers, os.O_RDONLY|os.O_CREATE, 0660)
	if err != nil {
		log.Fatalf("Unable to open file '%s': '%+v'", sc.gitusers, err)
	}
	// close fi on exit and check for its returned error
	defer func() {
		if err := fi.Close(); err != nil {
			panic(err)
		}
	}()

	reader := bufio.NewReader(fi)
	var line []byte
	var prefix bool
	for {
		if line, prefix, err = reader.ReadLine(); err != nil {
			break
		}
		matches := xregexp.FindNamedMatches(re, string(line))
		if len(matches) > 0 && !prefix {
			// fmt.Printf("line '%s', matches '%+v'\n", string(line), matches)
			// fmt.Printf("ip '%s', name '%s', email '%s'\n", matches["ip"], matches["name"], matches["email"])
			u := &user{name: matches["name"], email: matches["email"]}
			sc.users = sc.users.addUser(u, sship(matches["ip"]))
		}
	}
	if err == io.EOF {
		err = nil
	}
	if err != nil {
		panic(err)
	}
	return sc
}

func (ub *usersBase) getUser(sship sship) *user {
	for _, u := range ub.users {
		if verbose {
			fmt.Printf("ip '%s' vs sship '%s'\n", u.ip, sship)
		}
		if u.ip == sship {
			if verbose {
				fmt.Printf("User found for IP '%s': '%s'\n", sship, u)
			}
			return u
		}
	}
	return nil
}

func (c *choice) exit(_ *prompt.Buffer) {
	// return
	// os.Exit(0)
	c.mustExit = true
}

func completer(ub *usersBase) prompt.Completer {
	return func(d prompt.Document) []prompt.Suggest {
		s := []prompt.Suggest{}
		for _, u := range ub.users {
			s = append(s, prompt.Suggest{Text: u.name})
		}
		s = append(s, prompt.Suggest{Text: "New name", Description: "Firstname Lastname"})
		return prompt.FilterFuzzy(s, d.GetWordBeforeCursor(), true)
	}
}

func (c *choice) exitIfUserSelected(in string, breakline bool) bool {
	//fmt.Println("Check " + in)
	return c.user != nil || c.mustExit
}

func (c *choice) userSelector(in string) {
	if verbose {
		fmt.Println("You selected! " + in)
	}
	if in == "New name" {
		r := bufio.NewReader(os.Stdin)
		fmt.Print("Enter your first name (space) last name: ")
		name, err := r.ReadString('\n')
		if err != nil {
			fmt.Printf("Error when entering firstname lastname: '%+v'\n", err)
			return
		}
		name = strings.TrimSpace(name)
		space := regexp.MustCompile(`\s+`)
		name = space.ReplaceAllString(name, " ")
		if !strings.Contains(name, " ") {
			fmt.Printf("Expect firstname (space) lastname, instead of '%s'\n", name)
			return
		}
		in = name

		fmt.Print("Enter your email: ")
		email, err := r.ReadString('\n')
		if err != nil {
			fmt.Printf("Error when entering email: '%+v'\n", err)
			return
		}
		email = strings.TrimSpace(email)
		if !strings.Contains(email, "@") {
			fmt.Printf("Expect xxx@yyy email address, instead of '%s'\n", email)
			return
		}
		if strings.Contains(email, " ") {
			fmt.Printf("Expect no space in an email address, instead of '%s'\n", email)
			return
		}
		u := &user{name: in, email: email}
		c.ub.recordUser(u, "")
	}
	for _, u := range c.ub.users {
		if u.name == in {
			c.user = u
			break
		}
	}
}

func (ub *usersBase) askUserID() *user {
	if verbose {
		fmt.Println("Ask user")
	}
	time.Sleep(500 * time.Millisecond)

	c := &choice{ub: ub}
	quitOnCtrlC := prompt.KeyBind{
		Key: prompt.ControlC,
		Fn:  c.exit,
	}
	quitOnEscape := prompt.KeyBind{
		Key: prompt.Escape,
		Fn:  c.exit,
	}
	fsuggestions := completer(ub)
	suggestions := fsuggestions(prompt.Document{})
	p := prompt.New(
		c.userSelector,
		fsuggestions,
		prompt.OptionPrefix(">>> "),
		prompt.OptionAddKeyBind(quitOnEscape),
		prompt.OptionAddKeyBind(quitOnCtrlC),
		prompt.OptionShowCompletionAtStart(),
		prompt.OptionMaxSuggestion(uint16(len(suggestions))),
		prompt.OptionCompletionOnDown(),
		prompt.OptionSetExitCheckerOnInput(c.exitIfUserSelected),
	)
	fmt.Println("Please select Sync command.")
	p.Run()
	ub.userAsked = true
	return c.user
}

func (ub *usersBase) recordUser(u *user, ip sship) {
	if u == nil {
		return
	}
	if ip == "" {
		ip = sship("0.0.0.0")
	}
	ub.users = ub.users.addUser(u, ip)
	// https://stackoverflow.com/questions/31050656/can-not-replace-the-content-of-a-csv-file-in-go
	fi, err := os.OpenFile(ub.gitusers, os.O_WRONLY|os.O_CREATE, 0775)
	if err != nil {
		log.Fatalf("Unable to open write file '%s': '%+v'", ub.gitusers, err)
	}
	// close fi on exit and check for its returned error
	defer func() {
		if err := fi.Close(); err != nil {
			log.Fatalf("Unable to close '%s': err='%+v'\n", fi.Name(), err)
		}
	}()
	for _, u := range ub.users {
		line := fmt.Sprintf("%s~%s~%s\n", u.ip, u.name, u.email)
		l, err := fi.WriteString(line)
		if l == 0 || err != nil {
			log.Printf("Unable to write line '%s' to '%s': (%d) err='%+v'\n", line, fi.Name(), l, err)
		}
	}
	// nop
}

var rebash = regexp.MustCompile(`(?m)^(?P<date>[^~]+)~(?P<name>.*?)~(?P<email>(.*?)@(.*))`)

func (b *bash) getUser() *user {
	fi, err := os.OpenFile(b.file, os.O_RDONLY|os.O_CREATE, 0660)
	if err != nil {
		log.Fatalf("Unable to open bash file '%s': '%+v'", b.file, err)
	}
	// close fi on exit and check for its returned error
	defer func() {
		if err := fi.Close(); err != nil {
			panic(err)
		}
	}()

	reader := bufio.NewReader(fi)
	var line []byte
	var prefix bool
	var u *user
	for {
		if line, prefix, err = reader.ReadLine(); err != nil {
			break
		}
		matches := xregexp.FindNamedMatches(rebash, string(line))
		if len(matches) > 0 && !prefix {
			if verbose {
				fmt.Printf("BASH line '%s', matches '%+v'\n", string(line), matches)
				fmt.Printf("BASH date '%s' (vs '%s'), name '%s', email '%s'\n", matches["date"], b.dateFormatted(), matches["name"], matches["email"])
			}
			sdate := matches["date"]
			if sdate == b.dateFormatted() {
				u = &user{name: matches["name"], email: matches["email"]}
			} else {
				fmt.Printf("BASH found, but wrong date: '%s' vs. '%s", sdate, b.dateFormatted())
			}
			if verbose {
				fmt.Printf("sdate='%s', user='%s'\n", sdate, u)
			}
			break
		}
	}
	if err == io.EOF {
		err = nil
	}
	if err != nil {
		panic(err)
	}
	return u
}

type bash struct {
	file    string
	user    *user
	session *session
}

type session struct {
	date time.Time
	pid  shell.Pid
}

func getBashSession() (*session, error) {
	pid, err := shell.GetBashPID(verbose)
	if err != nil {
		return nil, err
	}
	date, err := shell.GetPSStartDate(pid, verbose)
	if err != nil {
		return nil, err
	}
	if verbose {
		fmt.Printf("date='%+v'\n", date)
	}
	s := &session{pid: pid, date: *date}
	return s, nil
}

func newBash(s *session) *bash {
	b := &bash{
		file:    fmt.Sprintf("/tmp/bash.%s", s.pid),
		session: s,
	}
	return b
}

func (b *bash) recordUser(u *user) {
	if u == nil {
		return
	}
	b.user = u
	// https://stackoverflow.com/questions/31050656/can-not-replace-the-content-of-a-csv-file-in-go
	fi, err := os.OpenFile(b.file, os.O_WRONLY|os.O_CREATE, 0775)
	if err != nil {
		log.Fatalf("Unable to open write file '%s': '%+v'", b.file, err)
	}
	// close fi on exit and check for its returned error
	defer func() {
		if err := fi.Close(); err != nil {
			log.Fatalf("Unable to close '%s': err='%+v'\n", fi.Name(), err)
		}
	}()
	sdate := b.dateFormatted()
	line := fmt.Sprintf("%s~%s~%s", sdate, u.name, u.email)
	l, err := fi.WriteString(line)
	if l == 0 || err != nil {
		log.Printf("Unable to write line '%s' to '%s': (%d) err='%+v'\n", line, fi.Name(), l, err)
	}
}

func (b *bash) dateFormatted() string {
	return b.session.date.Format("Mon Jan 02 15:04:05 2006")
}
