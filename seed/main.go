// seed loads app scores from scores.json into the PostgreSQL database.
// Usage: DATABASE_URL=... go run ./seed scores.json
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type AppScore struct {
	Num                int     `json:"num"`
	Name               string  `json:"name"`
	Folder             string  `json:"folder"`
	Category           string  `json:"category"`
	MainLanguage       string  `json:"main_language"`
	Blurb              string  `json:"blurb"`
	MonthlyEquivCost   int     `json:"monthly_equivalent_cost"`
	DeploymentComplexity string `json:"deployment_complexity"`

	Scalability       float64 `json:"scalability"`
	ClientServer      float64 `json:"client_server_balance"`
	DataSafety        float64 `json:"data_safety"`
	Container         float64 `json:"container_friendliness"`
	Security          float64 `json:"security"`
	LocalTestability  float64 `json:"local_testability"`
	Cost              float64 `json:"cost_to_run"`
	Issues            float64 `json:"issue_resolution"`
	Engineering       float64 `json:"engineering_quality"`
	CommercialValue   float64 `json:"commercial_value"`

	TotalScore float64 `json:"total_score"`
	Grade      string  `json:"grade"`

	StackDescription string `json:"stack_description"`
	NexlayerNotes    string `json:"nexlayer_notes"`
	ScalingAnalysis  string `json:"scaling_analysis"`
	ImprovementNote  string `json:"improvement_note"`
}

// Folder name overrides from linkedin-published.md (num → folder)
var folderOverrides = map[int]string{}

// LinkedIn URLs from published log (num → url)
var linkedinURLs = map[int]string{}

func computeTotal(a AppScore) float64 {
	return a.Scalability*2.0 +
		a.ClientServer*1.5 +
		a.DataSafety*1.5 +
		a.Security*1.5 +
		a.Container*1.0 +
		a.LocalTestability*0.5 +
		a.Cost*0.5 +
		a.Issues*0.5 +
		a.Engineering*0.5 +
		a.CommercialValue*0.5
}

func computeGrade(total float64) string {
	switch {
	case total >= 88:
		return "A+"
	case total >= 78:
		return "A"
	case total >= 68:
		return "B+"
	case total >= 58:
		return "B"
	case total >= 48:
		return "C+"
	case total >= 38:
		return "C"
	case total >= 25:
		return "D"
	default:
		return "F"
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: go run ./seed scores.json [linkedin-published.tsv]")
	}

	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalf("read scores: %v", err)
	}

	var apps []AppScore
	if err := json.Unmarshal(data, &apps); err != nil {
		log.Fatalf("parse scores: %v", err)
	}

	// Load folder overrides if TSV provided
	if len(os.Args) >= 3 {
		tsvData, err := os.ReadFile(os.Args[2])
		if err == nil {
			parseOverrides(string(tsvData))
		}
	}

	// Recompute totals and grades (in case JSON has stale values)
	for i := range apps {
		t := computeTotal(apps[i])
		t = math.Round(t*10) / 10
		apps[i].TotalScore = t
		apps[i].Grade = computeGrade(t)

		// Apply folder override
		if f, ok := folderOverrides[apps[i].Num]; ok {
			apps[i].Folder = f
		}
	}

	// Sort by score for rank assignment
	sorted := make([]AppScore, len(apps))
	copy(sorted, apps)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].TotalScore > sorted[j].TotalScore
	})
	rankMap := make(map[int]int)
	for i, a := range sorted {
		rankMap[a.Num] = i + 1
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/america250?sslmode=disable"
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	// Run migrations inline
	db.Exec(`CREATE TABLE IF NOT EXISTS apps (
    id SERIAL PRIMARY KEY,
    num INTEGER UNIQUE NOT NULL,
    name VARCHAR(200) NOT NULL,
    folder VARCHAR(200) NOT NULL,
    category VARCHAR(50) DEFAULT '',
    main_language VARCHAR(100) DEFAULT '',
    github_url TEXT DEFAULT '',
    live_url TEXT DEFAULT '',
    linkedin_url TEXT DEFAULT '',
    blurb TEXT DEFAULT '',
    score_scalability NUMERIC(4,1) DEFAULT 0,
    score_client_server NUMERIC(4,1) DEFAULT 0,
    score_data_safety NUMERIC(4,1) DEFAULT 0,
    score_container NUMERIC(4,1) DEFAULT 0,
    score_security NUMERIC(4,1) DEFAULT 0,
    score_local_testability NUMERIC(4,1) DEFAULT 0,
    score_cost NUMERIC(4,1) DEFAULT 0,
    score_issues NUMERIC(4,1) DEFAULT 0,
    score_engineering NUMERIC(4,1) DEFAULT 0,
    score_commercial_value NUMERIC(4,1) DEFAULT 0,
    total_score NUMERIC(5,1) DEFAULT 0,
    grade VARCHAR(3) DEFAULT '',
    rank_position INTEGER DEFAULT 0,
    stack_description TEXT DEFAULT '',
    nexlayer_notes TEXT DEFAULT '',
    scaling_analysis TEXT DEFAULT '',
    deployment_complexity VARCHAR(20) DEFAULT 'Medium',
    monthly_equivalent_cost INTEGER DEFAULT 0,
    improvement_note TEXT DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW()
)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_apps_num ON apps(num)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_apps_score ON apps(total_score DESC)`)
	db.Exec(`ALTER TABLE apps ADD COLUMN IF NOT EXISTS improvement_note TEXT DEFAULT ''`)

	inserted, updated := 0, 0
	for _, a := range apps {
		linkedinURL := linkedinURLs[a.Num]
		_, err := db.Exec(`
			INSERT INTO apps (num, name, folder, category, main_language, linkedin_url, blurb,
				score_scalability, score_client_server, score_data_safety, score_container, score_security,
				score_local_testability, score_cost, score_issues, score_engineering, score_commercial_value,
				total_score, grade, rank_position,
				stack_description, nexlayer_notes, scaling_analysis,
				deployment_complexity, monthly_equivalent_cost, improvement_note)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26)
			ON CONFLICT (num) DO UPDATE SET
				name=EXCLUDED.name, folder=EXCLUDED.folder, category=EXCLUDED.category,
				main_language=EXCLUDED.main_language, linkedin_url=COALESCE(NULLIF(EXCLUDED.linkedin_url,''), apps.linkedin_url),
				blurb=EXCLUDED.blurb,
				score_scalability=EXCLUDED.score_scalability, score_client_server=EXCLUDED.score_client_server,
				score_data_safety=EXCLUDED.score_data_safety, score_container=EXCLUDED.score_container,
				score_security=EXCLUDED.score_security, score_local_testability=EXCLUDED.score_local_testability,
				score_cost=EXCLUDED.score_cost, score_issues=EXCLUDED.score_issues,
				score_engineering=EXCLUDED.score_engineering, score_commercial_value=EXCLUDED.score_commercial_value,
				total_score=EXCLUDED.total_score, grade=EXCLUDED.grade, rank_position=EXCLUDED.rank_position,
				stack_description=EXCLUDED.stack_description, nexlayer_notes=EXCLUDED.nexlayer_notes,
				scaling_analysis=EXCLUDED.scaling_analysis,
				deployment_complexity=EXCLUDED.deployment_complexity, monthly_equivalent_cost=EXCLUDED.monthly_equivalent_cost,
				improvement_note=EXCLUDED.improvement_note`,
			a.Num, a.Name, a.Folder, a.Category, a.MainLanguage, linkedinURL, a.Blurb,
			a.Scalability, a.ClientServer, a.DataSafety, a.Container, a.Security,
			a.LocalTestability, a.Cost, a.Issues, a.Engineering, a.CommercialValue,
			a.TotalScore, a.Grade, rankMap[a.Num],
			a.StackDescription, a.NexlayerNotes, a.ScalingAnalysis,
			a.DeploymentComplexity, a.MonthlyEquivCost, a.ImprovementNote,
		)
		if err != nil {
			log.Printf("insert #%d %s: %v", a.Num, a.Name, err)
		} else {
			if rankMap[a.Num] > 0 {
				updated++
			} else {
				inserted++
			}
		}
	}

	fmt.Printf("Seeded %d apps (%d new, %d updated)\n", len(apps), inserted, updated)
	fmt.Printf("Grade distribution:\n")
	grades := map[string]int{}
	for _, a := range apps {
		grades[a.Grade]++
	}
	for _, g := range []string{"A+", "A", "B+", "B", "C+", "C", "D", "F"} {
		if n := grades[g]; n > 0 {
			fmt.Printf("  %s: %d\n", g, n)
		}
	}
}

func parseOverrides(tsv string) {
	for _, line := range strings.Split(tsv, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 4 {
			continue
		}
		num, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		linkedinURL := strings.TrimSpace(parts[2])
		folder := strings.TrimSpace(parts[3])
		if linkedinURL != "" {
			linkedinURLs[num] = linkedinURL
		}
		if folder != "" {
			folderOverrides[num] = folder
		}
	}
}
