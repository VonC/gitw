package xregexp

import "regexp"

// https://stackoverflow.com/questions/20750843/using-named-matches-from-go-regex
func FindNamedMatches(regex *regexp.Regexp, str string) map[string]string {
	match := regex.FindStringSubmatch(str)

	results := map[string]string{}
	for i, name := range match {
		results[regex.SubexpNames()[i]] = name
	}
	return results
}
