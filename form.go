package main

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

type Errors map[string][]string

// not a pointer receiver, map is a reference type
// A map variable does not store the data itself.
// It stores a pointer to an internal data structure.
func (e Errors) Add(field, message string) {
	e[field] = append(e[field], message)
}

func (e Errors) Get(field string) string {
	er := e[field]
	if len(er) == 0 {
		return ""
	}
	// returning only the first error if there are more than 1 Errors
	return er[0]
}

type Form struct {
	url.Values
	Errors Errors
}

func NewForm(fields url.Values) *Form {
	return &Form{
		Values: fields,
		Errors: Errors(make(map[string][]string)), // Convert map[string][]string → Errors type.
	}
}

func (f *Form) isValid() bool {
	return len(f.Errors) == 0
}

func (f *Form) Required(fields ...string) *Form {
	for _, field := range fields {
		value := f.Get(field)
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, fmt.Sprintf("%s field is required!", field))
		}
	}
	// returning form enables chain validation
	return f
}

func (f *Form) MaxLength(field string, maxLength int) *Form {
	value := f.Get(field)
	if value == "" {
		return f
	}
	if utf8.RuneCountInString(value) > maxLength {
		f.Errors.Add(field, fmt.Sprintf("%s too long. Max length is %d", field, maxLength))
	}
	return f
}

func (f *Form) MinLength(field string, minLength int) *Form {
	value := f.Get(field)
	if value == "" {
		return f
	}
	if utf8.RuneCountInString(value) < minLength {
		f.Errors.Add(field, fmt.Sprintf("%s too short. Min length is %d", field, minLength))
	}
	return f
}

func (f *Form) validateMail(field string) *Form {
	email := f.Get(field)
	if !EmailRX.MatchString(email) {
		f.Errors.Add(field, "Email not valid!")
	}
	return f
}

func (f *Form) validateUrl(field string) *Form {
	urlString := f.Get(field)
	_, err := url.ParseRequestURI(urlString)
	if err != nil {
		f.Errors.Add(field, "URL not valid!")
	}
	return f
}
