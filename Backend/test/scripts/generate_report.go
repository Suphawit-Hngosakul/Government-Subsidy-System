package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type TestEvent struct {
	Time    time.Time `json:"Time"`
	Action  string    `json:"Action"`
	Package string    `json:"Package"`
	Test    string    `json:"Test"`
	Elapsed float64   `json:"Elapsed"`
	Output  string    `json:"Output"`
}

type TestCaseInfo struct {
	Name      string    `json:"Name"`
	Package   string    `json:"Package"`
	ElapsedMs float64   `json:"ElapsedMs"`
	Status    string    `json:"Status"`
	Output    string    `json:"Output"`
}

type FunctionInfo struct {
	Name            string  `json:"Name"`
	Line            int     `json:"Line"`
	CoveragePercent float64 `json:"CoveragePercent"`
}

type FileCoverage struct {
	FileName          string         `json:"FileName"`
	CoveredStatements int            `json:"CoveredStatements"`
	TotalStatements   int            `json:"TotalStatements"`
	CoveragePercent   float64        `json:"CoveragePercent"`
	Functions         []FunctionInfo `json:"Functions"`
}

type OverallStats struct {
	TotalTests        int     `json:"TotalTests"`
	PassRate          float64 `json:"PassRate"`
	Failures          int     `json:"Failures"`
	TotalDurationMs   float64 `json:"TotalDurationMs"`
	AvgDurationMs     float64 `json:"AvgDurationMs"`
	MinDurationMs     float64 `json:"MinDurationMs"`
	MaxDurationMs     float64 `json:"MaxDurationMs"`
	TotalStatements   int     `json:"TotalStatements"`
	CoveredStatements int     `json:"CoveredStatements"`
	CoveragePercent   float64 `json:"CoveragePercent"`
	ApdexScore        float64 `json:"ApdexScore"`
}

type ReportData struct {
	Overall       OverallStats   `json:"Overall"`
	Files         []FileCoverage `json:"Files"`
	TestCases     []TestCaseInfo `json:"TestCases"`
	GeneratedTime string         `json:"GeneratedTime"`
}

func main() {
	fmt.Println("🚀 Initializing JMeter-Style Test & Coverage Dashboard Generator...")

	// 1. Run tests and collect JSON output
	testCases, err := runTestsAndParse()
	if err != nil {
		fmt.Printf("⚠️  Test execution completed with warnings or failures: %v\n", err)
	}

	// 2. Parse coverage profile
	overall, files, err := parseCoverageProfile(testCases)
	if err != nil {
		fmt.Printf("❌ Failed to parse coverage data: %v\n", err)
		os.Exit(1)
	}

	// 3. Prepare full report data
	reportData := ReportData{
		Overall:       overall,
		Files:         files,
		TestCases:     testCases,
		GeneratedTime: time.Now().Format("Mon, 02 Jan 2006 15:04:05 MST"),
	}

	// Sort files by coverage percent (highest to lowest) or alphabetically
	sort.Slice(reportData.Files, func(i, j int) bool {
		return reportData.Files[i].FileName < reportData.Files[j].FileName
	})

	// Ensure test/reports directory exists
	err = os.MkdirAll("test/reports", 0755)
	if err != nil {
		fmt.Printf("❌ Failed to create reports directory: %v\n", err)
		os.Exit(1)
	}

	// 4. Generate HTML file
	outputPath := "test/reports/test_report.html"
	err = writeHTMLReport(outputPath, reportData)
	if err != nil {
		fmt.Printf("❌ Failed to write HTML report: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✨ SUCCESS: JMeter-Style Interactive Report has been generated successfully!\n")
	fmt.Printf("📂 Report Path: %s\n", filepath.Clean(outputPath))
	fmt.Printf("📊 Summary Details:\n")
	fmt.Printf("   - Total Test Cases: %d\n", overall.TotalTests)
	fmt.Printf("   - Pass Rate: %.1f%%\n", overall.PassRate)
	fmt.Printf("   - Statement Coverage: %.1f%% (Target Range: 90-100%%)\n", overall.CoveragePercent)
	fmt.Printf("   - Apdex Score: %.1f%%\n\n", overall.ApdexScore)
}

func runTestsAndParse() ([]TestCaseInfo, error) {
	fmt.Println("📦 Running 'go test -json -coverprofile=test/reports/coverage.out ./service/...'")
	cmd := exec.Command("go", "test", "-json", "-coverprofile=test/reports/coverage.out", "./service/...")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run command
	runErr := cmd.Run()

	scanner := bufio.NewScanner(&stdout)
	testLogs := make(map[string]*TestCaseInfo)
	var testOrder []string
	var outputsMap = make(map[string][]string)

	for scanner.Scan() {
		var event TestEvent
		line := scanner.Bytes()
		if err := json.Unmarshal(line, &event); err != nil {
			continue
		}

		if event.Test == "" {
			continue
		}

		testKey := event.Package + "/" + event.Test

		// Track outputs for the test
		if event.Action == "output" {
			outputsMap[testKey] = append(outputsMap[testKey], event.Output)
		}

		// Track the lifecycle
		if _, exists := testLogs[testKey]; !exists {
			testLogs[testKey] = &TestCaseInfo{
				Name:    event.Test,
				Package: event.Package,
				Status:  "RUNNING",
			}
			testOrder = append(testOrder, testKey)
		}

		if event.Action == "pass" || event.Action == "fail" || event.Action == "skip" {
			testLogs[testKey].Status = strings.ToUpper(event.Action)
			testLogs[testKey].ElapsedMs = event.Elapsed * 1000.0
		}
	}

	var results []TestCaseInfo
	for _, key := range testOrder {
		info := testLogs[key]
		if info.Status == "RUNNING" {
			info.Status = "UNKNOWN"
		}
		info.Output = strings.Join(outputsMap[key], "")
		results = append(results, *info)
	}

	if len(results) == 0 && runErr != nil {
		return nil, fmt.Errorf("error running tests: %v, stderr: %s", runErr, stderr.String())
	}

	return results, nil
}

func parseCoverageProfile(results []TestCaseInfo) (OverallStats, []FileCoverage, error) {
	fmt.Println("🔍 Analyzing code coverage from 'test/reports/coverage.out'...")
	file, err := os.Open("test/reports/coverage.out")
	if err != nil {
		return OverallStats{}, nil, fmt.Errorf("unable to open coverage.out: %v", err)
	}
	defer file.Close()

	type stmtCount struct {
		total   int
		covered int
	}
	fileStats := make(map[string]*stmtCount)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "mode:") {
			continue
		}

		// e.g. government-subsidy-system/backend/service/auth_service.go:38.125,40.2 1 1
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		pathAndRange := parts[0]
		numStmts, err1 := strconv.Atoi(parts[1])
		count, err2 := strconv.Atoi(parts[2])
		if err1 != nil || err2 != nil {
			continue
		}

		colonIdx := strings.Index(pathAndRange, ":")
		if colonIdx == -1 {
			continue
		}
		fullPath := pathAndRange[:colonIdx]
		fileName := filepath.Base(fullPath)

		if _, exists := fileStats[fileName]; !exists {
			fileStats[fileName] = &stmtCount{}
		}

		fileStats[fileName].total += numStmts
		if count > 0 {
			fileStats[fileName].covered += numStmts
		}
	}

	fmt.Println("📊 Fetching detailed function coverage using 'go tool cover'...")
	cmd := exec.Command("go", "tool", "cover", "-func=test/reports/coverage.out")
	var coverFuncOut bytes.Buffer
	cmd.Stdout = &coverFuncOut
	if err := cmd.Run(); err != nil {
		return OverallStats{}, nil, fmt.Errorf("error running go tool cover -func: %v", err)
	}

	funcMap := make(map[string][]FunctionInfo)
	var overallCoverage float64

	funcScanner := bufio.NewScanner(&coverFuncOut)
	for funcScanner.Scan() {
		line := funcScanner.Text()
		if strings.Contains(line, "total:") {
			percentStr := strings.TrimSpace(line[strings.LastIndex(line, ")")+1:])
			percentStr = strings.TrimSuffix(percentStr, "%")
			if val, err := strconv.ParseFloat(strings.TrimSpace(percentStr), 64); err == nil {
				overallCoverage = val
			}
			continue
		}

		// e.g. government-subsidy-system/backend/service/auth_service.go:38:		NewAuthService		100.0%
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		pathLine := parts[0]
		funcName := parts[1]
		pctStr := strings.TrimSuffix(parts[2], "%")
		pct, errPct := strconv.ParseFloat(pctStr, 64)
		if errPct != nil {
			continue
		}

		colons := strings.Split(pathLine, ":")
		if len(colons) < 2 {
			continue
		}
		fullPath := colons[0]
		lineNo, _ := strconv.Atoi(colons[1])
		fileName := filepath.Base(fullPath)

		funcMap[fileName] = append(funcMap[fileName], FunctionInfo{
			Name:            funcName,
			Line:            lineNo,
			CoveragePercent: pct,
		})
	}

	var filesCoverages []FileCoverage
	var totalCovered, totalStmts int

	for fName, counts := range fileStats {
		pct := 0.0
		if counts.total > 0 {
			pct = (float64(counts.covered) / float64(counts.total)) * 100.0
		}

		// Sort functions within each file by line number
		functions := funcMap[fName]
		sort.Slice(functions, func(i, j int) bool {
			return functions[i].Line < functions[j].Line
		})

		filesCoverages = append(filesCoverages, FileCoverage{
			FileName:          fName,
			CoveredStatements: counts.covered,
			TotalStatements:   counts.total,
			CoveragePercent:   pct,
			Functions:         functions,
		})

		totalCovered += counts.covered
		totalStmts += counts.total
	}

	totalTests := len(results)
	failures := 0
	var totalDuration float64
	minDuration := -1.0
	maxDuration := 0.0

	for _, tc := range results {
		if tc.Status == "FAIL" {
			failures++
		}
		totalDuration += tc.ElapsedMs
		if minDuration == -1.0 || tc.ElapsedMs < minDuration {
			minDuration = tc.ElapsedMs
		}
		if tc.ElapsedMs > maxDuration {
			maxDuration = tc.ElapsedMs
		}
	}
	if minDuration == -1.0 {
		minDuration = 0.0
	}

	passRate := 100.0
	if totalTests > 0 {
		passRate = float64(totalTests-failures) / float64(totalTests) * 100.0
	}

	avgDuration := 0.0
	if totalTests > 0 {
		avgDuration = totalDuration / float64(totalTests)
	}

	calculatedCoverage := 0.0
	if totalStmts > 0 {
		calculatedCoverage = (float64(totalCovered) / float64(totalStmts)) * 100.0
	}
	if overallCoverage == 0 {
		overallCoverage = calculatedCoverage
	}

	// JMeter-style APDEX (Coverage & Passing combined)
	apdexScore := passRate * (overallCoverage / 100.0)

	overall := OverallStats{
		TotalTests:        totalTests,
		PassRate:          passRate,
		Failures:          failures,
		TotalDurationMs:   totalDuration,
		AvgDurationMs:     avgDuration,
		MinDurationMs:     minDuration,
		MaxDurationMs:     maxDuration,
		TotalStatements:   totalStmts,
		CoveredStatements: totalCovered,
		CoveragePercent:   overallCoverage,
		ApdexScore:        apdexScore,
	}

	return overall, filesCoverages, nil
}

func writeHTMLReport(filePath string, data ReportData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	tmpl, err := template.New("report").Funcs(template.FuncMap{
		"sub": func(a, b int) int {
			return a - b
		},
	}).Parse(htmlTemplate)
	if err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	type ViewContext struct {
		JSONData template.JS
		Data     ReportData
	}

	return tmpl.Execute(file, ViewContext{
		JSONData: template.JS(jsonData),
		Data:     data,
	})
}

// Premium dashboard HTML & CSS & JS template
const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Antigravity Government Subsidy System - Test & Coverage Report</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700;800&family=Outfit:wght@400;500;600;700;800&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg-color: #0f172a;
            --bg-gradient: radial-gradient(circle at 50% 0%, #1e1b4b 0%, #0f172a 70%);
            --card-bg: rgba(30, 41, 59, 0.45);
            --card-border: rgba(255, 255, 255, 0.08);
            --card-hover-border: rgba(99, 102, 241, 0.3);
            --text-main: #f8fafc;
            --text-muted: #94a3b8;
            
            --primary: #6366f1;
            --primary-gradient: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
            --primary-glow: rgba(99, 102, 241, 0.25);
            
            --success: #10b981;
            --success-glow: rgba(16, 185, 129, 0.2);
            --warning: #f59e0b;
            --warning-glow: rgba(245, 158, 11, 0.2);
            --danger: #ef4444;
            --danger-glow: rgba(239, 68, 68, 0.2);
            
            --font-outfit: 'Outfit', sans-serif;
            --font-inter: 'Inter', sans-serif;
            --font-mono: 'JetBrains Mono', monospace;
        }

        * {
            box-sizing: border-box;
            margin: 0;
            padding: 0;
        }

        body {
            background-color: var(--bg-color);
            background-image: var(--bg-gradient);
            color: var(--text-main);
            font-family: var(--font-inter);
            min-height: 100vh;
            padding: 2.5rem 2rem;
            line-height: 1.5;
            -webkit-font-smoothing: antialiased;
        }

        .container {
            max-width: 1400px;
            margin: 0 auto;
        }

        /* Header section */
        header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 2.5rem;
            border-bottom: 1px solid var(--card-border);
            padding-bottom: 1.5rem;
        }

        .header-title h1 {
            font-family: var(--font-outfit);
            font-size: 2.25rem;
            font-weight: 800;
            letter-spacing: -0.025em;
            background: linear-gradient(to right, #ffffff, #c7d2fe, #818cf8);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            margin-bottom: 0.25rem;
        }

        .header-title p {
            color: var(--text-muted);
            font-size: 0.95rem;
            font-weight: 400;
        }

        .header-meta {
            text-align: right;
        }

        .timestamp-badge {
            background: rgba(99, 102, 241, 0.1);
            border: 1px solid rgba(99, 102, 241, 0.2);
            color: #a5b4fc;
            padding: 0.35rem 0.85rem;
            border-radius: 9999px;
            font-size: 0.85rem;
            font-weight: 500;
            display: inline-block;
            margin-bottom: 0.5rem;
        }

        .meta-details {
            color: var(--text-muted);
            font-size: 0.8rem;
        }

        /* Widgets Grid */
        .widgets-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
            gap: 1.5rem;
            margin-bottom: 2.5rem;
        }

        .card {
            background: var(--card-bg);
            backdrop-filter: blur(16px);
            -webkit-backdrop-filter: blur(16px);
            border: 1px solid var(--card-border);
            border-radius: 1.25rem;
            padding: 1.75rem;
            transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
            position: relative;
            overflow: hidden;
            box-shadow: 0 4px 30px rgba(0, 0, 0, 0.3);
        }

        .card:hover {
            transform: translateY(-4px);
            border-color: var(--card-hover-border);
            box-shadow: 0 10px 30px rgba(99, 102, 241, 0.1);
        }

        .card::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            width: 100%;
            height: 4px;
            background: transparent;
        }

        .card.primary::before { background: var(--primary-gradient); }
        .card.success::before { background: var(--success); }
        .card.warning::before { background: var(--warning); }
        .card.danger::before { background: var(--danger); }

        .card-label {
            color: var(--text-muted);
            font-size: 0.875rem;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.05em;
            margin-bottom: 0.5rem;
        }

        .card-value {
            font-family: var(--font-outfit);
            font-size: 2rem;
            font-weight: 700;
            line-height: 1.2;
            margin-bottom: 0.5rem;
        }

        .card-subtext {
            color: var(--text-muted);
            font-size: 0.875rem;
        }

        /* Apdex score gauge card specific */
        .apdex-card {
            display: flex;
            align-items: center;
            justify-content: space-between;
        }

        .apdex-gauge-container {
            width: 100px;
            height: 100px;
            position: relative;
        }

        .apdex-gauge-container svg {
            width: 100%;
            height: 100%;
        }

        /* Nav tabs */
        .tabs-container {
            display: flex;
            border-bottom: 1px solid var(--card-border);
            margin-bottom: 2rem;
            gap: 1rem;
        }

        .tab-btn {
            background: transparent;
            border: none;
            color: var(--text-muted);
            font-family: var(--font-outfit);
            font-size: 1.1rem;
            font-weight: 600;
            padding: 0.75rem 1.25rem;
            cursor: pointer;
            border-bottom: 3px fill transparent;
            transition: all 0.2s ease;
            position: relative;
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }

        .tab-btn::after {
            content: '';
            position: absolute;
            bottom: -2px;
            left: 0;
            width: 100%;
            height: 3px;
            background: transparent;
            border-radius: 999px;
            transition: all 0.2s ease;
        }

        .tab-btn:hover {
            color: var(--text-main);
        }

        .tab-btn.active {
            color: #818cf8;
        }

        .tab-btn.active::after {
            background: var(--primary-gradient);
        }

        /* Tab Content panels */
        .tab-panel {
            display: none;
            animation: fadeIn 0.4s cubic-bezier(0.4, 0, 0.2, 1);
        }

        .tab-panel.active {
            display: block;
        }

        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(8px); }
            to { opacity: 1; transform: translateY(0); }
        }

        /* Coverage table styling */
        .table-container {
            overflow: hidden;
            border-radius: 1rem;
            border: 1px solid var(--card-border);
            background: rgba(30, 41, 59, 0.25);
            margin-bottom: 2rem;
        }

        table {
            width: 100%;
            border-collapse: collapse;
            text-align: left;
        }

        th {
            background: rgba(15, 23, 42, 0.6);
            color: var(--text-muted);
            font-weight: 600;
            font-size: 0.85rem;
            text-transform: uppercase;
            letter-spacing: 0.05em;
            padding: 1rem 1.5rem;
            border-bottom: 1px solid var(--card-border);
        }

        td {
            padding: 1.15rem 1.5rem;
            border-bottom: 1px solid var(--card-border);
            font-size: 0.925rem;
            color: var(--text-main);
        }

        tr.table-row-main {
            cursor: pointer;
            transition: background 0.2s ease;
        }

        tr.table-row-main:hover {
            background: rgba(99, 102, 241, 0.04);
        }

        .progress-bar-container {
            width: 100%;
            max-width: 200px;
            background: rgba(255, 255, 255, 0.05);
            height: 8px;
            border-radius: 9999px;
            overflow: hidden;
            display: inline-block;
            vertical-align: middle;
            margin-right: 0.75rem;
        }

        .progress-bar-fill {
            height: 100%;
            border-radius: 9999px;
            transition: width 1s ease-in-out;
        }

        /* Coverage status colors */
        .cov-excellent { color: var(--success); }
        .cov-tolerable { color: #60a5fa; }
        .cov-poor { color: var(--danger); }

        .bg-cov-excellent { background: var(--success); }
        .bg-cov-tolerable { background: #60a5fa; }
        .bg-cov-poor { background: var(--danger); }

        .badge {
            display: inline-flex;
            align-items: center;
            padding: 0.25rem 0.6rem;
            border-radius: 9999px;
            font-size: 0.75rem;
            font-weight: 600;
        }

        .badge-excellent {
            background: rgba(16, 185, 129, 0.1);
            color: #34d399;
            border: 1px solid rgba(16, 185, 129, 0.2);
        }

        .badge-tolerable {
            background: rgba(96, 165, 250, 0.1);
            color: #93c5fd;
            border: 1px solid rgba(96, 165, 250, 0.2);
        }

        .badge-poor {
            background: rgba(239, 68, 68, 0.1);
            color: #fca5a5;
            border: 1px solid rgba(239, 68, 68, 0.2);
        }

        /* Nested detail row */
        .expanded-row {
            background: rgba(15, 23, 42, 0.4);
            display: none;
        }

        .expanded-row.active {
            display: table-row;
        }

        .detail-table-container {
            padding: 1.5rem 2.5rem;
        }

        .detail-table-title {
            font-family: var(--font-outfit);
            font-size: 1rem;
            font-weight: 700;
            margin-bottom: 0.75rem;
            color: #a5b4fc;
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }

        .inner-table {
            width: 100%;
            background: rgba(15, 23, 42, 0.3);
            border-radius: 0.75rem;
            border: 1px solid var(--card-border);
        }

        .inner-table th {
            padding: 0.65rem 1.25rem;
            font-size: 0.775rem;
            background: rgba(15, 23, 42, 0.5);
        }

        .inner-table td {
            padding: 0.65rem 1.25rem;
            font-size: 0.85rem;
            border-bottom: 1px solid rgba(255, 255, 255, 0.03);
        }

        /* Test Run Logs Layout */
        .log-controls {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 1.5rem;
            gap: 1rem;
        }

        .search-bar {
            flex-grow: 1;
            max-width: 450px;
            position: relative;
        }

        .search-input {
            width: 100%;
            background: rgba(30, 41, 59, 0.6);
            border: 1px solid var(--card-border);
            border-radius: 0.75rem;
            padding: 0.65rem 1rem 0.65rem 2.5rem;
            color: var(--text-main);
            font-family: var(--font-inter);
            font-size: 0.9rem;
            transition: all 0.2s ease;
        }

        .search-input:focus {
            outline: none;
            border-color: var(--primary);
            box-shadow: 0 0 12px rgba(99, 102, 241, 0.2);
            background: rgba(30, 41, 59, 0.95);
        }

        .search-icon {
            position: absolute;
            left: 0.85rem;
            top: 50%;
            transform: translateY(-50%);
            color: var(--text-muted);
            pointer-events: none;
            width: 16px;
            height: 16px;
        }

        .filter-buttons {
            display: flex;
            gap: 0.5rem;
        }

        .filter-btn {
            background: rgba(255, 255, 255, 0.05);
            border: 1px solid var(--card-border);
            color: var(--text-muted);
            padding: 0.5rem 1rem;
            border-radius: 0.75rem;
            font-size: 0.85rem;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.2s ease;
        }

        .filter-btn:hover {
            background: rgba(255, 255, 255, 0.08);
            color: var(--text-main);
        }

        .filter-btn.active {
            background: var(--primary-gradient);
            color: #ffffff;
            border-color: transparent;
            box-shadow: 0 0 10px rgba(99, 102, 241, 0.25);
        }

        /* Test suite log rows */
        .status-pill {
            display: inline-flex;
            align-items: center;
            justify-content: center;
            width: 70px;
            padding: 0.2rem 0.5rem;
            border-radius: 0.5rem;
            font-size: 0.75rem;
            font-weight: 700;
            text-transform: uppercase;
        }

        .pill-pass {
            background: rgba(16, 185, 129, 0.1);
            color: #34d399;
            border: 1px solid rgba(16, 185, 129, 0.2);
        }

        .pill-fail {
            background: rgba(239, 68, 68, 0.1);
            color: #fca5a5;
            border: 1px solid rgba(239, 68, 68, 0.2);
        }

        .log-output-container {
            background: #020617;
            border-radius: 0.75rem;
            padding: 1.25rem;
            margin-top: 0.75rem;
            border: 1px solid rgba(255, 255, 255, 0.03);
            position: relative;
        }

        .log-output {
            font-family: var(--font-mono);
            font-size: 0.8rem;
            color: #cbd5e1;
            white-space: pre-wrap;
            overflow-x: auto;
            max-height: 350px;
        }

        .copy-btn {
            position: absolute;
            top: 0.75rem;
            right: 0.75rem;
            background: rgba(255, 255, 255, 0.05);
            border: 1px solid var(--card-border);
            color: var(--text-muted);
            padding: 0.25rem 0.5rem;
            border-radius: 0.35rem;
            font-size: 0.7rem;
            cursor: pointer;
            transition: all 0.2s ease;
        }

        .copy-btn:hover {
            background: rgba(255, 255, 255, 0.1);
            color: var(--text-main);
        }

        /* Coverage Heatmap grid */
        .heatmap-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
            gap: 1.25rem;
            margin-bottom: 2rem;
        }

        .heatmap-node {
            background: var(--card-bg);
            border: 1px solid var(--card-border);
            border-radius: 1rem;
            padding: 1.25rem;
            display: flex;
            flex-direction: column;
            justify-content: space-between;
            min-height: 120px;
            transition: all 0.2s ease;
        }

        .heatmap-node:hover {
            transform: scale(1.02);
            border-color: var(--card-hover-border);
        }

        .node-name {
            font-family: var(--font-outfit);
            font-weight: 700;
            font-size: 1.05rem;
            color: var(--text-main);
            margin-bottom: 0.5rem;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .node-meta {
            display: flex;
            align-items: center;
            justify-content: space-between;
        }

        .node-pct {
            font-family: var(--font-outfit);
            font-size: 1.5rem;
            font-weight: 800;
        }

        .node-bar {
            width: 100%;
            height: 6px;
            background: rgba(255, 255, 255, 0.05);
            border-radius: 999px;
            overflow: hidden;
            margin-top: 0.5rem;
        }

        .node-bar-fill {
            height: 100%;
            border-radius: 999px;
        }

        /* Rotator arrow for table expanding */
        .chevron {
            display: inline-block;
            transition: transform 0.2s ease;
            color: var(--text-muted);
            font-weight: bold;
        }

        tr.table-row-main.open .chevron {
            transform: rotate(90deg);
            color: var(--primary);
        }

        /* Utility classes */
        .flex { display: flex; }
        .items-center { align-items: center; }
        .justify-between { justify-content: space-between; }
        .gap-2 { gap: 0.5rem; }
    </style>
</head>
<body>
    <div class="container">
        <!-- Header -->
        <header>
            <div class="header-title">
                <h1>Government Subsidy System</h1>
                <p>Interactive Service Test Cases & Statement Coverage Dashboard</p>
            </div>
            <div class="header-meta">
                <span class="timestamp-badge">📅 Generated: {{ .Data.GeneratedTime }}</span>
                <div class="meta-details">OS: macOS | Environment: Backend | Go v1.22</div>
            </div>
        </header>

        <!-- Widgets Overview Grid -->
        <section class="widgets-grid">
            <!-- APDEX CARD -->
            <div class="card primary apdex-card">
                <div>
                    <div class="card-label">Overall Apdex Score</div>
                    <div class="card-value" style="font-size: 1.75rem; color: #a5b4fc;">
                        {{ printf "%.1f%%" .Data.Overall.ApdexScore }}
                    </div>
                    <div class="card-subtext">Pass Rate & Coverage Index</div>
                </div>
                <div class="apdex-gauge-container">
                    <svg viewBox="0 0 100 100">
                        <circle cx="50" cy="50" r="40" fill="transparent" stroke="rgba(255, 255, 255, 0.05)" stroke-width="8"></circle>
                        <circle id="apdex-progress" cx="50" cy="50" r="40" fill="transparent" stroke="url(#apdexGrad)" stroke-width="8"
                                stroke-dasharray="251.2" stroke-dashoffset="251.2" stroke-linecap="round"
                                style="transform: rotate(-90deg); transform-origin: 50% 50%; transition: stroke-dashoffset 1.2s cubic-bezier(0.4, 0, 0.2, 1);"></circle>
                        <defs>
                            <linearGradient id="apdexGrad" x1="0%" y1="0%" x2="100%" y2="100%">
                                <stop offset="0%" stop-color="#818cf8"></stop>
                                <stop offset="100%" stop-color="#34d399"></stop>
                            </linearGradient>
                        </defs>
                        <text id="apdex-text" x="50" y="56" text-anchor="middle" font-family="var(--font-outfit)" font-weight="800" font-size="14" fill="#ffffff">0%</text>
                    </svg>
                </div>
            </div>

            <!-- SUCCESS RATE CARD -->
            <div class="card success">
                <div class="card-label">Test Case Pass Rate</div>
                <div class="card-value" style="color: var(--success);">
                    {{ printf "%.1f%%" .Data.Overall.PassRate }}
                </div>
                <div class="card-subtext">
                    {{ printf "%d / %d Passed" (sub .Data.Overall.TotalTests .Data.Overall.Failures) .Data.Overall.TotalTests }} ({{ .Data.Overall.Failures }} Failed)
                </div>
            </div>

            <!-- STATEMENT COVERAGE CARD -->
            <div class="card success">
                <div class="card-label">Statement Coverage</div>
                <div class="card-value" style="color: #34d399;">
                    {{ printf "%.1f%%" .Data.Overall.CoveragePercent }}
                </div>
                <div class="card-subtext">
                    {{ printf "%d / %d Statements Covered" .Data.Overall.CoveredStatements .Data.Overall.TotalStatements }}
                </div>
            </div>

            <!-- EXECUTION TIME CARD -->
            <div class="card warning">
                <div class="card-label">Total Execution Time</div>
                <div class="card-value" style="color: var(--warning); font-size: 1.75rem;">
                    {{ printf "%.2f ms" .Data.Overall.TotalDurationMs }}
                </div>
                <div class="card-subtext">
                    Avg: {{ printf "%.2f ms" .Data.Overall.AvgDurationMs }} | Max: {{ printf "%.1f ms" .Data.Overall.MaxDurationMs }}
                </div>
            </div>
        </section>

        <!-- Navigation Tabs -->
        <div class="tabs-container">
            <button class="tab-btn active" onclick="switchTab('dashboard')">
                <svg width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M4 6a2 2 0 012-2h2a2 2 0 012 2v4a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v4a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z"></path></svg>
                Dashboard Overview
            </button>
            <button class="tab-btn" onclick="switchTab('logs')">
                <svg width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01"></path></svg>
                Test Run Logs
            </button>
            <button class="tab-btn" onclick="switchTab('heatmap')">
                <svg width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M17 14v6m-3-3h6M6 10h2a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v2a2 2 0 002 2zm10 0h2a2 2 0 002-2V6a2 2 0 00-2-2h-2a2 2 0 00-2 2v2a2 2 0 002 2zM6 20h2a2 2 0 002-2v-2a2 2 0 00-2-2H6a2 2 0 00-2 2v2a2 2 0 002 2z"></path></svg>
                Coverage Heatmap
            </button>
        </div>

        <!-- 1. DASHBOARD OVERVIEW PANEL -->
        <div id="panel-dashboard" class="tab-panel active">
            <div class="table-container">
                <table>
                    <thead>
                        <tr>
                            <th width="40"></th>
                            <th>Service File Name</th>
                            <th>Covered Statements</th>
                            <th>Total Statements</th>
                            <th>Statement Coverage %</th>
                            <th>Rating Status</th>
                        </tr>
                    </thead>
                    <tbody id="files-table-body">
                        <!-- Dynamic file list will be injected here -->
                    </tbody>
                </table>
            </div>
        </div>

        <!-- 2. TEST LOGS PANEL -->
        <div id="panel-logs" class="tab-panel">
            <div class="log-controls">
                <div class="search-bar">
                    <svg class="search-icon" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path></svg>
                    <input type="text" id="log-search" class="search-input" placeholder="Search test cases by name or package..." oninput="filterTestCases()">
                </div>
                <div class="filter-buttons">
                    <button class="filter-btn active" id="btn-all" onclick="setTestFilter('ALL')">All</button>
                    <button class="filter-btn" id="btn-pass" onclick="setTestFilter('PASS')" style="color: #34d399;">Passed</button>
                    <button class="filter-btn" id="btn-fail" onclick="setTestFilter('FAIL')" style="color: #fca5a5;">Failed</button>
                </div>
            </div>

            <div class="table-container">
                <table>
                    <thead>
                        <tr>
                            <th width="40"></th>
                            <th>Test Case Name</th>
                            <th>Package / Domain</th>
                            <th>Elapsed (ms)</th>
                            <th width="120">Outcome</th>
                        </tr>
                    </thead>
                    <tbody id="logs-table-body">
                        <!-- Dynamic test case list will be injected here -->
                    </tbody>
                </table>
            </div>
        </div>

        <!-- 3. COVERAGE HEATMAP PANEL -->
        <div id="panel-heatmap" class="tab-panel">
            <div class="heatmap-grid" id="heatmap-container">
                <!-- Heatmap blocks will be injected here -->
            </div>
        </div>
    </div>

    <!-- Inject data into frontend script -->
    <script>
        const reportData = {{ .JSONData }};
        
        console.log("Raw Report Data Loaded:", reportData);

        // Helper to format values
        function getCoverageColorClass(pct) {
            if (pct >= 90.0) return 'cov-excellent';
            if (pct >= 80.0) return 'cov-tolerable';
            return 'cov-poor';
        }

        function getCoverageBgClass(pct) {
            if (pct >= 90.0) return 'bg-cov-excellent';
            if (pct >= 80.0) return 'bg-cov-tolerable';
            return 'bg-cov-poor';
        }

        function getCoverageBadge(pct) {
            if (pct >= 90.0) return '<span class="badge badge-excellent">Excellent</span>';
            if (pct >= 80.0) return '<span class="badge badge-tolerable">Tolerable</span>';
            return '<span class="badge badge-poor">Needs Work</span>';
        }

        // 1. Render Files Table (with expandable function level detail)
        function renderFilesTable() {
            const tbody = document.getElementById('files-table-body');
            tbody.innerHTML = '';

            reportData.Files.forEach((file, index) => {
                const colorClass = getCoverageColorClass(file.CoveragePercent);
                const bgClass = getCoverageBgClass(file.CoveragePercent);
                const badgeHtml = getCoverageBadge(file.CoveragePercent);

                // Main Row
                const mainRow = document.createElement('tr');
                mainRow.className = 'table-row-main';
                mainRow.onclick = function() { toggleRowExpand(index); };
                mainRow.id = 'file-row-' + index;
                mainRow.innerHTML = 
                    '<td align="center"><span class="chevron" id="chevron-' + index + '">▶</span></td>' +
                    '<td style="font-weight: 600; font-family: var(--font-outfit);">' + file.FileName + '</td>' +
                    '<td>' + file.CoveredStatements + '</td>' +
                    '<td>' + file.TotalStatements + '</td>' +
                    '<td>' +
                        '<div class="progress-bar-container">' +
                            '<div class="progress-bar-fill ' + bgClass + '" style="width: 0%"></div>' +
                        '</div>' +
                        '<span class="node-pct-val ' + colorClass + '" style="font-weight: 700;">' + file.CoveragePercent.toFixed(1) + '%</span>' +
                    '</td>' +
                    '<td>' + badgeHtml + '</td>';
                tbody.appendChild(mainRow);

                // Nested Function Detail Row
                const detailRow = document.createElement('tr');
                detailRow.className = 'expanded-row';
                detailRow.id = 'file-detail-' + index;

                let innerRows = '';
                if (file.Functions && file.Functions.length > 0) {
                    file.Functions.forEach(fn => {
                        const fnColorClass = getCoverageColorClass(fn.CoveragePercent);
                        const fnBgClass = getCoverageBgClass(fn.CoveragePercent);
                        innerRows += 
                            '<tr>' +
                                '<td style="font-family: var(--font-mono); color: #c7d2fe;">' + fn.Name + '</td>' +
                                '<td>Line ' + fn.Line + '</td>' +
                                '<td>' +
                                    '<div class="progress-bar-container" style="max-width: 150px; height: 6px;">' +
                                        '<div class="progress-bar-fill ' + fnBgClass + '" style="width: ' + fn.CoveragePercent + '%"></div>' +
                                    '</div>' +
                                    '<span class="' + fnColorClass + '" style="font-weight: 600;">' + fn.CoveragePercent.toFixed(1) + '%</span>' +
                                '</td>' +
                            '</tr>';
                    });
                } else {
                    innerRows = '<tr><td colspan="3" align="center" style="color: var(--text-muted); font-style: italic;">No function statistics available.</td></tr>';
                }

                detailRow.innerHTML = 
                    '<td colspan="6">' +
                        '<div class="detail-table-container">' +
                            '<div class="detail-table-title">' +
                                '<svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"></path></svg>' +
                                ' Detailed Function Coverage: ' + file.FileName +
                            '</div>' +
                            '<table class="inner-table">' +
                                '<thead>' +
                                    '<tr>' +
                                        '<th>Function Name</th>' +
                                        '<th width="150">Start Line</th>' +
                                        '<th width="300">Statement Coverage</th>' +
                                    '</tr>' +
                                '</thead>' +
                                '<tbody>' +
                                    innerRows +
                                '</tbody>' +
                            '</table>' +
                        '</div>' +
                    '</td>';
                tbody.appendChild(detailRow);
            });

            // Animate progress bars on loading
            setTimeout(() => {
                const fills = tbody.querySelectorAll('.table-row-main .progress-bar-fill');
                reportData.Files.forEach((file, index) => {
                    if (fills[index]) {
                        fills[index].style.width = file.CoveragePercent + '%';
                    }
                });
            }, 100);
        }

        function toggleRowExpand(index) {
            const mainRow = document.getElementById('file-row-' + index);
            const detailRow = document.getElementById('file-detail-' + index);
            const chevron = document.getElementById('chevron-' + index);
            
            const isActive = detailRow.classList.contains('active');
            
            if (isActive) {
                detailRow.classList.remove('active');
                mainRow.classList.remove('open');
            } else {
                detailRow.classList.add('active');
                mainRow.classList.add('open');
            }
        }

        // 2. Render Heatmap View
        function renderHeatmap() {
            const container = document.getElementById('heatmap-container');
            container.innerHTML = '';

            reportData.Files.forEach(file => {
                const colorClass = getCoverageColorClass(file.CoveragePercent);
                const bgClass = getCoverageBgClass(file.CoveragePercent);

                const node = document.createElement('div');
                node.className = 'heatmap-node';
                node.innerHTML = 
                    '<div>' +
                        '<div class="node-name" title="' + file.FileName + '">' + file.FileName + '</div>' +
                        '<div style="color: var(--text-muted); font-size: 0.775rem;">' +
                            'Statements: ' + file.CoveredStatements + '/' + file.TotalStatements +
                        '</div>' +
                    '</div>' +
                    '<div class="node-meta">' +
                        '<div class="node-pct ' + colorClass + '">' + file.CoveragePercent.toFixed(0) + '%</div>' +
                        '<div style="width: 50%;">' +
                            '<div class="node-bar">' +
                                '<div class="node-bar-fill ' + bgClass + '" style="width: ' + file.CoveragePercent + '%"></div>' +
                            '</div>' +
                        '</div>' +
                    '</div>';
                container.appendChild(node);
            });
        }

        // 3. Render Test Case Logs (Interactive Filters)
        let currentFilter = 'ALL';

        function renderTestLogs() {
            const tbody = document.getElementById('logs-table-body');
            tbody.innerHTML = '';
            
            const searchText = document.getElementById('log-search').value.toLowerCase();

            let matchedIndex = 0;
            reportData.TestCases.forEach((tc, idx) => {
                // Filter logic
                if (currentFilter === 'PASS' && tc.Status !== 'PASS') return;
                if (currentFilter === 'FAIL' && tc.Status !== 'PASS' && tc.Status !== 'FAIL') {
                    // skip unknown/skipped if filtering failures, unless status is FAIL
                }
                if (currentFilter === 'FAIL' && tc.Status === 'PASS') return;

                const nameMatch = tc.Name.toLowerCase().includes(searchText);
                const pkgMatch = tc.Package.toLowerCase().includes(searchText);
                if (!nameMatch && !pkgMatch) return;

                const statusClass = tc.Status === 'PASS' ? 'pill-pass' : 'pill-fail';
                
                const mainRow = document.createElement('tr');
                mainRow.className = 'table-row-main';
                mainRow.onclick = function() { toggleLogExpand(idx); };
                mainRow.id = 'log-row-' + idx;
                mainRow.innerHTML = 
                    '<td align="center"><span class="chevron" id="log-chevron-' + idx + '">▶</span></td>' +
                    '<td style="font-weight: 600; color: #e2e8f0;">' + tc.Name + '</td>' +
                    '<td style="color: var(--text-muted); font-size: 0.825rem; font-family: var(--font-mono);">' + tc.Package + '</td>' +
                    '<td>' + tc.ElapsedMs.toFixed(2) + ' ms</td>' +
                    '<td><span class="status-pill ' + statusClass + '">' + tc.Status + '</span></td>';
                tbody.appendChild(mainRow);

                // Standard Output expand details
                const detailRow = document.createElement('tr');
                detailRow.className = 'expanded-row';
                detailRow.id = 'log-detail-' + idx;
                
                const cleanOutput = tc.Output ? tc.Output.replace(/</g, "&lt;").replace(/>/g, "&gt;") : "No stdout captures generated for this test.";

                detailRow.innerHTML = 
                    '<td colspan="5">' +
                        '<div class="detail-table-container">' +
                            '<div class="log-output-container">' +
                                '<button class="copy-btn" onclick="copyToClipboard(\'' + idx + '\')">Copy Output</button>' +
                                '<pre class="log-output" id="log-pre-' + idx + '"><code>' + cleanOutput + '</code></pre>' +
                            '</div>' +
                        '</div>' +
                    '</td>';
                tbody.appendChild(detailRow);
                matchedIndex++;
            });

            if (matchedIndex === 0) {
                tbody.innerHTML = '<tr><td colspan="5" align="center" style="padding: 2rem; color: var(--text-muted); font-style: italic;">No test cases found matching filters.</td></tr>';
            }
        }

        function toggleLogExpand(index) {
            const mainRow = document.getElementById('log-row-' + index);
            const detailRow = document.getElementById('log-detail-' + index);
            const chevron = document.getElementById('log-chevron-' + index);
            
            const isActive = detailRow.classList.contains('active');
            
            if (isActive) {
                detailRow.classList.remove('active');
                mainRow.classList.remove('open');
            } else {
                detailRow.classList.add('active');
                mainRow.classList.add('open');
            }
        }

        function copyToClipboard(index) {
            const text = document.getElementById('log-pre-' + index).textContent;
            navigator.clipboard.writeText(text).then(() => {
                const btn = document.querySelector('#log-detail-' + index + ' .copy-btn');
                btn.textContent = 'Copied!';
                setTimeout(() => {
                    btn.textContent = 'Copy Output';
                }, 2000);
            });
        }

        function filterTestCases() {
            renderTestLogs();
        }

        function setTestFilter(status) {
            currentFilter = status;
            
            document.querySelectorAll('.filter-btn').forEach(btn => btn.classList.remove('active'));
            if (status === 'ALL') document.getElementById('btn-all').classList.add('active');
            if (status === 'PASS') document.getElementById('btn-pass').classList.add('active');
            if (status === 'FAIL') document.getElementById('btn-fail').classList.add('active');
            
            renderTestLogs();
        }

        // Tab Switcher
        function switchTab(tabId) {
            document.querySelectorAll('.tab-btn').forEach(btn => btn.classList.remove('active'));
            document.querySelectorAll('.tab-panel').forEach(panel => panel.classList.remove('active'));

            const activeBtn = Array.from(document.querySelectorAll('.tab-btn')).find(btn => btn.getAttribute('onclick').includes(tabId));
            activeBtn.classList.add('active');
            
            document.getElementById('panel-' + tabId).classList.add('active');
        }

        // Global Initialization
        window.addEventListener('DOMContentLoaded', () => {
            renderFilesTable();
            renderTestLogs();
            renderHeatmap();

            // Animate Apdex Gauge
            setTimeout(() => {
                const apdexScore = reportData.Overall.ApdexScore;
                const offset = 251.2 - (apdexScore / 100.0) * 251.2;
                document.getElementById('apdex-progress').style.strokeDashoffset = offset;
                
                // Count up text animation
                let start = 0;
                const duration = 1200; // ms
                const stepTime = 15;
                const totalSteps = duration / stepTime;
                const increment = apdexScore / totalSteps;
                
                const timer = setInterval(() => {
                    start += increment;
                    if (start >= apdexScore) {
                        document.getElementById('apdex-text').textContent = apdexScore.toFixed(0) + '%';
                        clearInterval(timer);
                    } else {
                        document.getElementById('apdex-text').textContent = start.toFixed(0) + '%';
                    }
                }, stepTime);
            }, 300);
        });
    </script>
</body>
</html>
`

// Simple mathematical helper used inside HTML template
func sub(a, b int) int {
	return a - b
}
