package validation

import (
	"encoding/xml"
	"fmt"
	"io"
)

type JUnitReporter struct{}

func NewJUnitReporter() *JUnitReporter {
	return &JUnitReporter{}
}

func (r *JUnitReporter) Format() string {
	return "junit"
}

func (r *JUnitReporter) Write(report *ValidationReport, w io.Writer) error {
	suites := convertToJUnit(report)
	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	if err := encoder.Encode(suites); err != nil {
		return err
	}
	_, err := fmt.Fprint(w, "\n")
	return err
}

func convertToJUnit(report *ValidationReport) JUnitTestSuites {
	suites := JUnitTestSuites{
		Name:  "ade-validate",
		Tests: report.NumChecks(),
		Time:  fmtDuration(report.Duration),
	}

	for _, m := range report.Modules {
		suite := JUnitTestSuite{
			Name: m.ModuleName,
			Time: fmtDuration(m.Duration),
		}

		for _, c := range m.Checks {
			tc := JUnitTestCase{
				Name:      c.Name,
				Classname: m.ModuleName,
				Time:      fmtDuration(c.Duration),
			}

			switch c.Status {
			case StatusFailed:
				suite.Failures++
				suites.Failures++
				msg := c.Message
				if msg == "" {
					msg = "check échoué"
				}
				tc.Failure = &JUnitFailure{
					Message: msg,
					Type:    "failure",
					Content: c.Details,
				}
			case StatusError:
				suite.Errors++
				suites.Errors++
				msg := c.Message
				if msg == "" {
					msg = "erreur technique"
				}
				tc.Error = &JUnitError{
					Message: msg,
					Type:    "error",
					Content: c.Details,
				}
			case StatusWarning, StatusSkipped:
				suite.Skipped++
				tc.Skipped = &JUnitSkipped{
					Message: c.Message,
				}
			}

			suite.Tests++
			suite.TestCases = append(suite.TestCases, tc)
		}

		suites.TestSuites = append(suites.TestSuites, suite)
	}

	return suites
}

type JUnitTestSuites struct {
	XMLName    struct{}         `xml:"testsuites"`
	Name       string           `xml:"name,attr"`
	Tests      int              `xml:"tests,attr"`
	Failures   int              `xml:"failures,attr"`
	Errors     int              `xml:"errors,attr"`
	Time       string           `xml:"time,attr"`
	TestSuites []JUnitTestSuite `xml:"testsuite"`
}

type JUnitTestSuite struct {
	XMLName   struct{}        `xml:"testsuite"`
	Name      string          `xml:"name,attr"`
	Tests     int             `xml:"tests,attr"`
	Failures  int             `xml:"failures,attr"`
	Errors    int             `xml:"errors,attr"`
	Skipped   int             `xml:"skipped,attr"`
	Time      string          `xml:"time,attr"`
	TestCases []JUnitTestCase `xml:"testcase"`
}

type JUnitTestCase struct {
	XMLName   struct{}      `xml:"testcase"`
	Name      string        `xml:"name,attr"`
	Classname string        `xml:"classname,attr"`
	Time      string        `xml:"time,attr"`
	Failure   *JUnitFailure `xml:"failure,omitempty"`
	Error     *JUnitError   `xml:"error,omitempty"`
	Skipped   *JUnitSkipped `xml:"skipped,omitempty"`
}

type JUnitFailure struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Content string `xml:",chardata"`
}

type JUnitError struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Content string `xml:",chardata"`
}

type JUnitSkipped struct {
	Message string `xml:"message,attr"`
}
