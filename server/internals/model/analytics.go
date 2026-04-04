package model

import "time"

type IssueOpenTimeRow struct {
	CreatedTime time.Time
	ClosedTime  time.Time
}

type IssuePriorityRow struct {
	Priority string
}
