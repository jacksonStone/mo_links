package main

import (
	"errors"
	"strings"
	"unicode"

	_ "github.com/mattn/go-sqlite3"
)

func getMatchingLinks(organizationId int64, name string) ([]string, error) {
	if organizationId == 0 {
		return []string{}, errors.New("organizationId must be defined")
	}
	if name == "" {
		return []string{}, errors.New("name must be defined")
	}
	if len(name) > 255 {
		return []string{}, errors.New("name must be 255 characters or less")
	}
	links, err := dbGetMatchingLinks(organizationId, name)
	if err != nil {
		return []string{}, err
	}
	if len(links) == 0 {
		return []string{}, nil
	}
	if len(links) > 1 {
		return []string{}, errors.New("multiple matches")
	}

	match := links[0]
	go dbIncrementViewCountOfLink(organizationId, name)
	return []string{match}, nil
}

func addLink(url string, name string, userId int64, activeOrganizationId int64) error {
	err := validName(name)
	if err != nil {
		return err
	}
	err = validUrl(url)
	if err != nil {
		return err
	}
	if !strings.Contains(url, "//") {
		url = "https://" + url
	}
	links, err := dbGetMatchingLinks(activeOrganizationId, name)
	if err != nil {
		return err
	}
	if len(links) > 0 {
		return errors.New("link already exists for that organization")
	}
	return dbAddLink(url, name, userId, activeOrganizationId)
}

func validUrl(url string) error {
	if url == "" {
		return errors.New("url must not be empty")
	}
	// can't be longer than 2048 charecters
	if len(url) > 2048 {
		return errors.New("url must be 2048 characters or less")
	}
	return nil
}

func validName(name string) error {
	// Name must be 1-255 characters long
	if len(name) == 0 || len(name) > 255 {
		return errors.New("name must be 1-255 characters long")
	}
	for _, char := range name {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' && char != '-' {
			return errors.New("name must only contain letters, digits, underscores, and hyphens")
		}
	}
	if name == "____reserved" {
		return errors.New("name must not be '____reserved'")
	}

	return nil
}
