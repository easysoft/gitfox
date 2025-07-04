// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package githook

import (
	"fmt"
	"time"

	"github.com/easysoft/gitfox/app/services/codeowners"
	"github.com/easysoft/gitfox/git"
	"github.com/easysoft/gitfox/git/hook"

	"github.com/fatih/color"
	"github.com/gotidy/ptr"
)

var (
	colorScanHeader            = color.New(color.FgHiWhite, color.Underline)
	colorScanSummary           = color.New(color.FgHiRed, color.Bold)
	colorScanSummaryNoFindings = color.New(color.FgHiGreen, color.Bold)
)

func printScanSecretsFindings(
	output *hook.Output,
	findings []secretFinding,
	multipleRefs bool,
	duration time.Duration,
) {
	findingsCnt := len(findings)

	// no results? output success and continue
	if findingsCnt == 0 {
		output.Messages = append(
			output.Messages,
			colorScanSummaryNoFindings.Sprintf("No secrets found")+
				fmt.Sprintf(" in %s", duration.Round(time.Millisecond)),
			"", "", // add two empty lines for making it visually more consumable
		)
		return
	}

	output.Messages = append(
		output.Messages,
		colorScanHeader.Sprintf(
			"Push contains %s:",
			singularOrPlural("secret", findingsCnt > 1),
		),
		"", // add empty line for making it visually more consumable
	)

	for _, finding := range findings {
		headerTxt := fmt.Sprintf("%s in %s:%d", finding.RuleID, finding.File, finding.StartLine)
		if finding.StartLine != finding.EndLine {
			headerTxt += fmt.Sprintf("-%d", finding.EndLine)
		}
		if multipleRefs {
			headerTxt += fmt.Sprintf(" [%s]", finding.Ref)
		}

		output.Messages = append(
			output.Messages,
			fmt.Sprintf("  %s", headerTxt),
			fmt.Sprintf("      Secret:       %s", finding.Secret),
			fmt.Sprintf("      Commit:       %s", finding.Commit),
			fmt.Sprintf("      Details:      %s", finding.Description),
			fmt.Sprintf("      Fingerprint:  %s", finding.Fingerprint),
			"", // add empty line for making it visually more consumable
		)
	}

	output.Messages = append(
		output.Messages,
		colorScanSummary.Sprintf(
			"%d %s found",
			findingsCnt,
			singularOrPlural("secret", findingsCnt > 1),
		)+fmt.Sprintf(" in %s", FMTDuration(time.Millisecond)),
		"", "", // add two empty lines for making it visually more consumable
	)
}

func FMTDuration(d time.Duration) string {
	const secondsRounding = time.Second / time.Duration(10)
	switch {
	case d <= time.Millisecond:
	// keep anything under a millisecond untouched
	case d < time.Second:
		d = d.Round(time.Millisecond) // round under a second to millisecondss
	case d < time.Minute:
		d = d.Round(secondsRounding) // round under a minute to .1 precision
	default:
		d = d.Round(time.Second) // keep rest at second precision
	}
	return d.String()
}

func printOversizeFiles(
	output *hook.Output,
	oversizeFiles []git.FileInfo,
	sizeLimit int64,
) {
	output.Messages = append(
		output.Messages,
		colorScanHeader.Sprintf(
			"Push contains files exceeding the size limit:",
		),
		"", // add empty line for making it visually more consumable
	)

	for _, file := range oversizeFiles {
		output.Messages = append(
			output.Messages,
			fmt.Sprintf("  %s", file.SHA),
			fmt.Sprintf("      Size: %dB", file.Size),
			"", // add empty line for making it visually more consumable
		)
	}

	total := len(oversizeFiles)
	output.Messages = append(
		output.Messages,
		colorScanSummary.Sprintf(
			"%d %s found exceeding the size limit of %dB",
			total, singularOrPlural("file", total > 1), sizeLimit,
		),
		"", "", // add two empty lines for making it visually more consumable
	)
}

func singularOrPlural(noun string, plural bool) string {
	if plural {
		return noun + "s"
	}
	return noun
}

func printCodeOwnerPush(
	output *hook.Output,
	entries []codeowners.Entry,
) {
	output.Messages = append(
		output.Messages,
		colorScanHeader.Sprintf(
			"Push is denied by CODEOWNERS rule",
		),
	)
	for _, entry := range entries {
		output.Messages = append(output.Messages, fmt.Sprintf("  - %s", entry.Pattern))
	}
	output.Error = ptr.String("push  is denied by CODEOWNERS rule")
}
