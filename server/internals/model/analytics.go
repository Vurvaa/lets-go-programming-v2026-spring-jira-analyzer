package model

import (
	"database/sql"
	"time"
)

type IssueOpenTimeRow struct {
	CreatedTime time.Time
	ClosedTime  sql.NullTime
}

type IssuePriorityRow struct {
	Priority string
}

type GraphData struct {
	Categories []string       `json:"data"`
	Count      map[string]int `json:"count"`
}

type CompareGraphData struct {
	Data  []string         `json:"data"`
	Count map[string][]int `json:"count"`
}
type IsAnalyzedResponse struct {
	IsAnalyzed bool `json:"data"`
}
