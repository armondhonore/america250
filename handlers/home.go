package handlers

import (
	"bytes"
	"database/sql"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

type App struct {
	ID                   int
	Num                  int
	Name                 string
	Folder               string
	Category             string
	MainLanguage         string
	GithubURL            string
	LiveURL              string
	LinkedinURL          string
	Blurb                string
	ScoreScalability     float64
	ScoreClientServer    float64
	ScoreDataSafety      float64
	ScoreContainer       float64
	ScoreSecurity        float64
	ScoreLocalTest       float64
	ScoreCost            float64
	ScoreIssues          float64
	ScoreEngineering     float64
	ScoreCommercialValue float64
	TotalScore           float64
	Grade                string
	RankPosition         int
	StackDescription     string
	NexlayerNotes        string
	ScalingAnalysis      string
	DeploymentComplexity string
	MonthlyEquivCost     int
	ImprovementNote      string
}

// NeedsImprovement reports whether this app's grade is below C (D or F) —
// these are shown as "Needs Improvement" with specifics instead of a letter grade.
func (a App) NeedsImprovement() bool {
	return a.Grade == "D" || a.Grade == "F"
}

type HomeData struct {
	Apps       []App
	Categories []string
	Grades     []string
	Sort       string
	Filter     string
	GradeFilter string
	TotalCount int
}

func HomeHandler(db *sql.DB, tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		sort := r.URL.Query().Get("sort")
		if sort == "" {
			sort = "score"
		}
		filter := r.URL.Query().Get("q")
		gradeFilter := r.URL.Query().Get("grade")
		categoryFilter := r.URL.Query().Get("cat")

		orderBy := "total_score DESC, num ASC"
		switch sort {
		case "num":
			orderBy = "num ASC"
		case "name":
			orderBy = "name ASC"
		case "complexity":
			orderBy = "CASE deployment_complexity WHEN 'Easy' THEN 1 WHEN 'Medium' THEN 2 WHEN 'Hard' THEN 3 WHEN 'Expert' THEN 4 ELSE 5 END ASC"
		case "cost":
			orderBy = "monthly_equivalent_cost DESC"
		}

		args := []interface{}{}
		wheres := []string{}
		argIdx := 1

		if filter != "" {
			wheres = append(wheres, `(name ILIKE $`+strconv.Itoa(argIdx)+` OR blurb ILIKE $`+strconv.Itoa(argIdx)+` OR category ILIKE $`+strconv.Itoa(argIdx)+`)`)
			args = append(args, "%"+filter+"%")
			argIdx++
		}
		if gradeFilter != "" {
			wheres = append(wheres, `grade = $`+strconv.Itoa(argIdx))
			args = append(args, gradeFilter)
			argIdx++
		}
		if categoryFilter != "" {
			wheres = append(wheres, `category = $`+strconv.Itoa(argIdx))
			args = append(args, categoryFilter)
			argIdx++
		}

		where := ""
		if len(wheres) > 0 {
			where = "WHERE " + strings.Join(wheres, " AND ")
		}

		q := `SELECT id, num, name, folder, category, main_language, github_url, live_url, linkedin_url, blurb,
			score_scalability, score_client_server, score_data_safety, score_container, score_security,
			score_local_testability, score_cost, score_issues, score_engineering, score_commercial_value,
			total_score, grade, rank_position, deployment_complexity, monthly_equivalent_cost, improvement_note
			FROM apps ` + where + ` ORDER BY ` + orderBy

		rows, err := db.QueryContext(r.Context(), q, args...)
		if err != nil {
			http.Error(w, "DB error: "+err.Error(), 500)
			return
		}
		defer rows.Close()

		var apps []App
		for rows.Next() {
			var a App
			err := rows.Scan(
				&a.ID, &a.Num, &a.Name, &a.Folder, &a.Category, &a.MainLanguage,
				&a.GithubURL, &a.LiveURL, &a.LinkedinURL, &a.Blurb,
				&a.ScoreScalability, &a.ScoreClientServer, &a.ScoreDataSafety,
				&a.ScoreContainer, &a.ScoreSecurity, &a.ScoreLocalTest,
				&a.ScoreCost, &a.ScoreIssues, &a.ScoreEngineering, &a.ScoreCommercialValue,
				&a.TotalScore, &a.Grade, &a.RankPosition,
				&a.DeploymentComplexity, &a.MonthlyEquivCost, &a.ImprovementNote,
			)
			if err != nil {
				continue
			}
			apps = append(apps, a)
		}

		// Get categories for filter
		catRows, _ := db.QueryContext(r.Context(), `SELECT DISTINCT category FROM apps WHERE category != '' ORDER BY category`)
		var cats []string
		if catRows != nil {
			defer catRows.Close()
			for catRows.Next() {
				var c string
				catRows.Scan(&c)
				cats = append(cats, c)
			}
		}

		data := HomeData{
			Apps:        apps,
			Categories:  cats,
			Grades:      []string{"A+", "A", "B+", "B", "C+", "C", "D", "F"},
			Sort:        sort,
			Filter:      filter,
			GradeFilter: gradeFilter,
			TotalCount:  len(apps),
		}

		var buf bytes.Buffer
		if err := tmpl.ExecuteTemplate(&buf, "base-home", data); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		buf.WriteTo(w)
	}
}

func itoa(i int) string {
	return strconv.Itoa(i)
}
