package validation

import "io"

type ReportWriter interface {
	Format() string
	Write(report *ValidationReport, w io.Writer) error
}

var reportWriters []ReportWriter

func registerReporter(r ReportWriter) {
	reportWriters = append(reportWriters, r)
}

func AvailableReporters() []ReportWriter {
	result := make([]ReportWriter, len(reportWriters))
	copy(result, reportWriters)
	return result
}

func NewReportWriter(format string) ReportWriter {
	for _, r := range reportWriters {
		if r.Format() == format {
			return r
		}
	}
	return nil
}

func init() {
	registerReporter(NewJSONReporter())
	registerReporter(NewJUnitReporter())
}
