package hierarchy

import (
	"fmt"
	"regexp"
	"strings"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func ValidateCreateLeagueInput(input CreateLeagueInput) error {
	if strings.TrimSpace(input.Name) == "" {
		return fmt.Errorf("%w: league name is required", ErrInvalidInput)
	}
	if !slugPattern.MatchString(input.Slug) {
		return fmt.Errorf("%w: league slug must be lowercase kebab-case", ErrInvalidInput)
	}
	return nil
}

func ValidateCreateFranchiseInput(input CreateFranchiseInput) error {
	if input.LeagueID <= 0 {
		return fmt.Errorf("%w: leagueId must be greater than zero", ErrInvalidInput)
	}
	if strings.TrimSpace(input.Name) == "" {
		return fmt.Errorf("%w: franchise name is required", ErrInvalidInput)
	}
	if !slugPattern.MatchString(input.Slug) {
		return fmt.Errorf("%w: franchise slug must be lowercase kebab-case", ErrInvalidInput)
	}
	return nil
}

func ValidateCreateClubInput(input CreateClubInput) error {
	if input.FranchiseID <= 0 {
		return fmt.Errorf("%w: franchiseId must be greater than zero", ErrInvalidInput)
	}
	if strings.TrimSpace(input.Name) == "" {
		return fmt.Errorf("%w: club name is required", ErrInvalidInput)
	}
	if !slugPattern.MatchString(input.Slug) {
		return fmt.Errorf("%w: club slug must be lowercase kebab-case", ErrInvalidInput)
	}
	return nil
}
