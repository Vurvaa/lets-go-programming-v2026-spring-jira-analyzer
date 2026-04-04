package model

import "time"

type IssueOpenTimeRow struct {
	CreatedTime time.Time
	ClosedTime  time.Time
}

type IssuePriorityRow struct {
	Priority string
}

type GraphData struct {
	Categories []string       `json:"categories"`
	Count      map[string]int `json:"count"`
}

type GraphResponse struct {
	Data *GraphData `json:"data"`
}

type IsAnalyzedResponse struct {
	IsAnalyzed bool `json:"isAnalyzed"`
}
