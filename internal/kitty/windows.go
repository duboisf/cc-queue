package kitty

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type kittyWindow struct {
	ID int `json:"id"`
}

type kittyTab struct {
	Windows []kittyWindow `json:"windows"`
}

type kittyOSWindow struct {
	Tabs []kittyTab `json:"tabs"`
}

// ParseWindowIDs extracts all window IDs from kitty @ ls JSON output.
func ParseWindowIDs(lsOutput []byte) (map[string]bool, error) {
	var osWindows []kittyOSWindow
	if err := json.Unmarshal(lsOutput, &osWindows); err != nil {
		return nil, fmt.Errorf("parsing kitty ls output: %w", err)
	}
	ids := make(map[string]bool)
	for _, ow := range osWindows {
		for _, t := range ow.Tabs {
			for _, w := range t.Windows {
				ids[strconv.Itoa(w.ID)] = true
			}
		}
	}
	return ids, nil
}
