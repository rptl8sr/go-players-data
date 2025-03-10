package filter

import (
	"strings"
	"time"

	"go-players-data/internal/logger"
	"go-players-data/internal/model"
)

type criteria struct {
	ignoredGroups    []string
	allowedCompanies []string
	maxOffline       time.Duration
}

// Criteria defines an interface for filtering a slice of Player objects based on specific conditions.
// The Filter method returns a filtered list of players and an error if any issues are encountered during the operation.
type Criteria interface {
	Filter(players []*model.Player) ([]*model.Player, error)
}

// New creates a new Filter instance with the specified criteria.
func New(ignoredGroups []string, allowedCompanies []string, maxOffline time.Duration) Criteria {
	return &criteria{
		ignoredGroups:    ignoredGroups,
		allowedCompanies: allowedCompanies,
		maxOffline:       maxOffline,
	}
}

// Filter filters players based on offline duration, group, and company criteria.
// Returns a slice of players that meet the conditions.
func (c *criteria) Filter(players []*model.Player) ([]*model.Player, error) {
	start := time.Now()
	defer func() { logger.Debug("filter.Filter: Time spent", "time", time.Since(start).String()) }()

	var filteredPlayers []*model.Player

	for _, p := range players {
		if c.isIgnored(p) {
			continue
		}

		filteredPlayers = append(filteredPlayers, p)
	}

	logger.Debug("filter.Filter: Total players", "filtered", len(filteredPlayers), "total", len(players))
	return filteredPlayers, nil
}

// isIgnored determines if a player should be ignored based on group, company, and offline duration criteria.
func (c *criteria) isIgnored(p *model.Player) bool {
	groupName := c.extractGroupName(p)

	if c.stringInSlice(c.ignoredGroups, groupName) {
		return true
	}

	if !c.stringInSlice(c.allowedCompanies, p.CompanyName) {
		return true
	}

	if c.hoursDelta(p.LastOnline) <= c.maxOffline.Hours() {
		return true
	}

	return false
}

// extractGroupName extracts and returns the first segment of the GroupName field in the provided Player struct.
func (c *criteria) extractGroupName(player *model.Player) string {
	return strings.Split(player.GroupName, "/")[0]
}

// stringInSlice checks if a given string exists within a slice of strings, returning true if found, otherwise false.
func (c *criteria) stringInSlice(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

// hoursDelta calculates the difference in hours between the current time and the provided time t.
func (c *criteria) hoursDelta(t time.Time) float64 {
	return time.Since(t).Hours()
}
