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

// https://stackoverflow.com/questions/20750843/using-named-matches-from-go-regex
func FindAllNamedMatches(regex *regexp.Regexp, str string) map[string]string {
	match := regex.FindAllStringSubmatch(str, -1)

	results := map[string]string{}
	for _, group := range match {
		for i, name := range group {
			if i == 0 {
				continue
			}
			if name != "" {
				key := regex.SubexpNames()[i]
				value := results[key]
				if value != "" {
					value = value + "," + name
				} else {
					value = name
				}
				results[key] = value
			}
		}
	}
	return results
}
