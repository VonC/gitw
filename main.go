package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gitw/internal/shell"
	"gitw/internal/xregexp"

	"gitw/version"

	"github.com/charmbracelet/lipgloss"
	"github.com/erikgeiser/promptkit/selection"
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
	// fmt.Printf("-----------\n%s\n", ub.users.String())
	if asship != "" {
		if verbose {
			fmt.Printf("SSH connection detected: '%s'\n", asship)
		}
		user = ub.getUser(asship)
		if verbose {
			fmt.Printf("User based on sship '%s': '%+v'\n", asship, user)
		}
		if user == nil {
			if check {
				os.Exit(1)
			}
			user = ub.askUserID()
			if verbose {
				fmt.Printf("User after asked '%s', to be recorded in '%s''\n", user, ub.gitusers)
			}
			ub.recordUser(user, asship)
		} else if ub.hasMultipleEntries {
			if verbose {
				fmt.Printf("Multitple entries detected for '%s': record'\n", ub.gitusers)
			}
			ub.recordUser(user, asship)
			ub.hasMultipleEntries = false
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

func (us users) String() string {
	res := fmt.Sprintf("%d users", len(us))
	for _, u := range us {
		res = res + fmt.Sprintf("\n%s, email '%s', ip '%s'", u.name, u.email, u.ip)
	}
	return res
}

func (us users) addUser(u *user, ip sship) users {
	auser := us.getUserFRomEmail(u.email)
	if auser == nil {
		auser = u
		if verbose {
			fmt.Printf("Add user '%s/%s' to %d users\n", auser.name, auser.ip, len(us))
		}
		us = append(us, auser)
	}
	if verbose {
		fmt.Printf("Update user '%s', IP '%s' => new IP '%s'\n", auser.name, auser.ip, ip)
	}
	auser.setSSHIP(ip)
	return us
}

func (us users) users() []string {
	res := make([]string, 0)
	for _, u := range us {
		res = append(res, u.name)
	}
	return res
}

func (us users) getUser(name string) *user {
	for _, u := range us {
		if u.name == name {
			return u
		}
	}
	return nil
}

func (us users) getUserFRomEmail(email string) *user {
	for _, auser := range us {
		if auser.email == email {
			return auser
		}
	}
	return nil
}

func (u *user) setSSHIP(ip sship) {
	if !ip.isNul() || u.ip.isNul() {
		if verbose {
			fmt.Printf("User old ip '%s', new ip '%s'\n", u.ip, ip)
		}
		u.ip = ip
	}
}

func (ip sship) isNul() bool {
	if string(ip) == "" || string(ip) == "0.0.0.0" {
		return true
	}
	return false
}

type usersBase struct {
	gitusers           string
	users              users
	userAsked          bool
	hasMultipleEntries bool
}

var re = regexp.MustCompile(`(?m)^(?P<ip>[^ ~]+)~(?P<name>.*?)~(?P<email>(.*?)@(.*))`)

func newUsersBase(file string) *usersBase {

	sc := &usersBase{
		gitusers:           file,
		users:              []*user{},
		userAsked:          false,
		hasMultipleEntries: false,
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
			l := len(sc.users)
			if verbose {
				fmt.Printf("***\nUsers before\n***\n%s\n", sc.users.String())
			}
			sc.users = sc.users.addUser(u, sship(matches["ip"]))
			if verbose {
				fmt.Printf("*--*--*\nUsers AFTER\n*--*--*\n%s\n", sc.users.String())
			}
			if len(sc.users) == l {
				if verbose {
					fmt.Printf("Multiple user '%s' detected from '%s'\n", u.name, sc.gitusers)
				}
				sc.hasMultipleEntries = true
			}
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
		if string(u.ip) == string(sship) {
			if verbose {
				fmt.Printf("User found for IP '%s': '%s'\n", sship, u)
			}
			return u
		}
	}
	return nil
}

func (ub *usersBase) userSelector(in string) *user {
	if verbose {
		fmt.Println("You selected! " + in)
	}
	if in == "New name" {
		r := bufio.NewReader(os.Stdin)
		fmt.Print("Enter your first name (space) last name: ")
		name, err := r.ReadString('\n')
		if err != nil {
			fmt.Printf("Error when entering firstname lastname: '%+v'\n", err)
			return nil
		}
		name = strings.TrimSpace(name)
		space := regexp.MustCompile(`\s+`)
		name = space.ReplaceAllString(name, " ")
		if !strings.Contains(name, " ") {
			fmt.Printf("Expect firstname (space) lastname, instead of '%s'\n", name)
			return nil
		}
		in = name

		fmt.Print("Enter your email: ")
		email, err := r.ReadString('\n')
		if err != nil {
			fmt.Printf("Error when entering email: '%+v'\n", err)
			return nil
		}
		email = strings.TrimSpace(email)
		if !strings.Contains(email, "@") {
			fmt.Printf("Expect xxx@yyy email address, instead of '%s'\n", email)
			return nil
		}
		if strings.Contains(email, " ") {
			fmt.Printf("Expect no space in an email address, instead of '%s'\n", email)
			return nil
		}
		u := &user{name: in, email: email}
		ub.recordUser(u, "")
	}
	return ub.userSelector(in)
}

type resetConsoleMode func()

func (ub *usersBase) askUserID() *user {

	con, originalConsoleMode := getOriginalConsoleMode()
	defer resetConsoleMore(con, originalConsoleMode)

	users := ub.users.users()
	users = append(users, "New name")
	sp := selection.New("Chose a VSCode Worskpace to open:",
		selection.Choices(users))
	fmt.Printf("sp colorProfile before: %d\n", sp.ColorProfile)
	sp.ColorProfile = GetColorProfile()
	fmt.Printf("sp colorProfile after: %d\n", sp.ColorProfile)
	sp.FilterInputPlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	sp.AbortFunc = func() error {
		ccon, ccor := getOriginalConsoleMode()
		err := fmt.Errorf("Restore con %d (vs. current %d), orig %d (vs. current %d)", con, ccon, originalConsoleMode, ccor)
		resetConsoleMore(con, originalConsoleMode)
		return err
	}

	choice, err := sp.RunPrompt()
	if err != nil {
		fmt.Printf("Error: %v\n", err)

		os.Exit(1)
	}

	if verbose {
		fmt.Println("Ask user")
	}
	time.Sleep(500 * time.Millisecond)

	ub.userAsked = true
	return ub.users.getUser(choice.String)
}

func (ub *usersBase) recordUser(u *user, ip sship) {
	if u == nil {
		return
	}
	if string(ip) == "" {
		ip = sship("0.0.0.0")
	}
	ub.users = ub.users.addUser(u, ip)
	// https://stackoverflow.com/questions/31050656/can-not-replace-the-content-of-a-csv-file-in-go
	fi, err := os.Create(ub.gitusers)
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
		if verbose {
			fmt.Printf("Write user '%s' for ip '%s' to '%s'\n", u.name, u.ip, ub.gitusers)
		}
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
	if verbose {
		fmt.Printf("Get user from tmp file '%s'\n", b.file)
	}
	fi, err := os.OpenFile(b.file, os.O_RDONLY, 0660)
	if err != nil {
		var pathError *fs.PathError
		if errors.As(err, &pathError) {
			return nil
		}
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
			if verbose {
				fmt.Printf("Error when reading line '%s' of '%s': '%+v'\n", line, b.file, err)
			}
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
				fmt.Printf("BASH found, but wrong date: '%s' vs. '%s'\n", sdate, b.dateFormatted())
			}
			if verbose {
				fmt.Printf("sdate='%s', user='%s'\n", sdate, u)
			}
			break
		} else if verbose {
			fmt.Printf("BASH line '%s', No match for regex '%s'\n", string(line), rebash.String())
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
		file:    fmt.Sprintf("%s%cbash.%s", shell.TempPath(), filepath.Separator, s.pid),
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
