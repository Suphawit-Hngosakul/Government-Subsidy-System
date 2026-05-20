#!/bin/bash
# run_report.sh: Run backend test suites and generate the premium JMeter-style HTML test & coverage dashboard.

echo "=========================================================================="
echo "🚀 [Antigravity Report Runner] Launching Tests & Generating HTML Dashboard"
echo "=========================================================================="

# Get the script's directory and always run from Backend root
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
BACKEND_DIR="$( cd "$SCRIPT_DIR/.." && pwd )"

cd "$BACKEND_DIR"

# Compile and run the report generator script
go run test/scripts/generate_report.go

if [ $? -eq 0 ]; then
    echo "=========================================================================="
    echo "🎉 SUCCESS: Dashboard generated successfully at: Backend/test/reports/test_report.html"
    echo "=========================================================================="
else
    echo "=========================================================================="
    echo "❌ ERROR: Report generation failed. Please check error outputs above."
    echo "=========================================================================="
    exit 1
fi
