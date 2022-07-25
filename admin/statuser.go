package admin

import "context"

type Statuser interface {
	// Status returns a high-level status report on the specified app env
	Status(ctx context.Context) (StatusReport, error)

	// StatusDetail returns a detailed status report on the specified app env
	StatusDetail(ctx context.Context) (StatusDetailReports, error)
}

type StatusReport struct {
	Fields []string
	Data   map[string]interface{}
}

type StatusDetailReports []StatusDetailReport

type StatusDetailReport struct {
	Name    string
	Records StatusRecords
}

type StatusRecords []StatusRecord

type StatusRecord struct {
	Fields []string
	Data   map[string]interface{}
}
