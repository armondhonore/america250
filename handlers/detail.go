package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

type DetailData struct {
	App    App
	Prev   *App
	Next   *App
	Scores []ScoreDim
}

type ScoreDim struct {
	Label  string
	Score  float64
	Weight int
	Pct    int
}

func DetailHandler(db *sql.DB, tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		numStr := strings.TrimPrefix(r.URL.Path, "/app/")
		numStr = strings.Trim(numStr, "/")
		num, err := strconv.Atoi(numStr)
		if err != nil || num < 1 {
			http.NotFound(w, r)
			return
		}

		var a App
		err = db.QueryRowContext(r.Context(), `
			SELECT id, num, name, folder, category, main_language, github_url, live_url, linkedin_url, blurb,
				score_scalability, score_client_server, score_data_safety, score_container, score_security,
				score_local_testability, score_cost, score_issues, score_engineering, score_commercial_value,
				total_score, grade, rank_position, stack_description, nexlayer_notes, scaling_analysis,
				deployment_complexity, monthly_equivalent_cost, improvement_note
			FROM apps WHERE num = $1`, num).Scan(
			&a.ID, &a.Num, &a.Name, &a.Folder, &a.Category, &a.MainLanguage,
			&a.GithubURL, &a.LiveURL, &a.LinkedinURL, &a.Blurb,
			&a.ScoreScalability, &a.ScoreClientServer, &a.ScoreDataSafety,
			&a.ScoreContainer, &a.ScoreSecurity, &a.ScoreLocalTest,
			&a.ScoreCost, &a.ScoreIssues, &a.ScoreEngineering, &a.ScoreCommercialValue,
			&a.TotalScore, &a.Grade, &a.RankPosition,
			&a.StackDescription, &a.NexlayerNotes, &a.ScalingAnalysis,
			&a.DeploymentComplexity, &a.MonthlyEquivCost, &a.ImprovementNote,
		)
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		scores := []ScoreDim{
			{Label: "Scalability", Score: a.ScoreScalability, Weight: 20, Pct: int(a.ScoreScalability * 10)},
			{Label: "Client/Server Balance", Score: a.ScoreClientServer, Weight: 15, Pct: int(a.ScoreClientServer * 10)},
			{Label: "Data Safety", Score: a.ScoreDataSafety, Weight: 15, Pct: int(a.ScoreDataSafety * 10)},
			{Label: "Security", Score: a.ScoreSecurity, Weight: 15, Pct: int(a.ScoreSecurity * 10)},
			{Label: "Container Friendliness", Score: a.ScoreContainer, Weight: 10, Pct: int(a.ScoreContainer * 10)},
			{Label: "Local Testability", Score: a.ScoreLocalTest, Weight: 5, Pct: int(a.ScoreLocalTest * 10)},
			{Label: "Cost to Self-Host", Score: a.ScoreCost, Weight: 5, Pct: int(a.ScoreCost * 10)},
			{Label: "Issue Resolution", Score: a.ScoreIssues, Weight: 5, Pct: int(a.ScoreIssues * 10)},
			{Label: "Engineering Quality", Score: a.ScoreEngineering, Weight: 5, Pct: int(a.ScoreEngineering * 10)},
			{Label: "Commercial Value", Score: a.ScoreCommercialValue, Weight: 5, Pct: int(a.ScoreCommercialValue * 10)},
		}

		// Adjacent apps (by score rank)
		var prev, next *App
		var pn App
		if err := db.QueryRowContext(r.Context(),
			`SELECT num, name, folder, grade, total_score FROM apps WHERE total_score > $1 ORDER BY total_score ASC LIMIT 1`,
			a.TotalScore).Scan(&pn.Num, &pn.Name, &pn.Folder, &pn.Grade, &pn.TotalScore); err == nil {
			prev = &pn
		}
		var nn App
		if err := db.QueryRowContext(r.Context(),
			`SELECT num, name, folder, grade, total_score FROM apps WHERE total_score < $1 ORDER BY total_score DESC LIMIT 1`,
			a.TotalScore).Scan(&nn.Num, &nn.Name, &nn.Folder, &nn.Grade, &nn.TotalScore); err == nil {
			next = &nn
		}

		data := DetailData{App: a, Prev: prev, Next: next, Scores: scores}
		var buf bytes.Buffer
		if err := tmpl.ExecuteTemplate(&buf, "base-detail", data); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		buf.WriteTo(w)
	}
}

func ApiAppsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.QueryContext(r.Context(), `
			SELECT num, name, folder, category, grade, total_score, blurb, deployment_complexity, monthly_equivalent_cost
			FROM apps ORDER BY total_score DESC, num ASC`)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()

		type Slim struct {
			Num        int     `json:"num"`
			Name       string  `json:"name"`
			Folder     string  `json:"folder"`
			Category   string  `json:"category"`
			Grade      string  `json:"grade"`
			Score      float64 `json:"score"`
			Blurb      string  `json:"blurb"`
			Complexity string  `json:"complexity"`
			Cost       int     `json:"cost"`
		}

		var apps []Slim
		for rows.Next() {
			var a Slim
			rows.Scan(&a.Num, &a.Name, &a.Folder, &a.Category, &a.Grade, &a.Score, &a.Blurb, &a.Complexity, &a.Cost)
			apps = append(apps, a)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=300")
		json.NewEncoder(w).Encode(apps)
	}
}
