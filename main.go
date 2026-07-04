package main

import (
	"database/sql"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/armondhonore/america250/handlers"
	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static
var staticFS embed.FS

//go:embed screenshots
var screenshotsFS embed.FS

func main() {
	var err error
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/america250?sslmode=disable"
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err = db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}
	log.Println("DB connected")

	if err = runMigrations(db); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	funcMap := template.FuncMap{
		"gradeColor": func(grade string) string {
			switch grade {
			case "A+":
				return "#00ff88"
			case "A":
				return "#00d4a0"
			case "B+":
				return "#00b4d8"
			case "B":
				return "#0096c7"
			case "C+":
				return "#f4a261"
			case "C":
				return "#e07b39"
			case "D":
				return "#e63946"
			default:
				return "#666"
			}
		},
		"scoreColor": func(score float64) string {
			switch {
			case score >= 8.5:
				return "#00ff88"
			case score >= 7:
				return "#00d4a0"
			case score >= 5.5:
				return "#f4a261"
			case score >= 4:
				return "#e07b39"
			default:
				return "#e63946"
			}
		},
		"scorePct": func(s float64) int { return int(s * 10) },
		"scoreDash": func(s float64) float64 { return s / 100.0 * 263.9 },
		"formatScore": func(s float64) string { return fmt.Sprintf("%.1f", s) },
		"printf":      fmt.Sprintf,
		"lower":       strings.ToLower,
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFS(templateFS, "templates/*.html")
	if err != nil {
		log.Fatalf("parse templates: %v", err)
	}

	mux := http.NewServeMux()

	// Static assets (embedded)
	mux.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		http.FileServer(http.FS(staticFS)).ServeHTTP(w, r)
	})

	// Screenshots (embedded in the binary — no external volume dependency)
	mux.HandleFunc("/screenshots/", func(w http.ResponseWriter, r *http.Request) {
		folder := strings.TrimPrefix(r.URL.Path, "/screenshots/")
		folder = filepath.Clean(folder)
		if strings.Contains(folder, "..") {
			http.NotFound(w, r)
			return
		}
		embeddedPath := "screenshots/" + folder + "/screenshot.png"
		if data, readErr := screenshotsFS.ReadFile(embeddedPath); readErr == nil {
			w.Header().Set("Content-Type", "image/png")
			w.Header().Set("Cache-Control", "public, max-age=86400")
			w.Write(data)
			return
		}
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		fmt.Fprintf(w, `<svg xmlns="http://www.w3.org/2000/svg" width="400" height="250" viewBox="0 0 400 250"><rect width="400" height="250" fill="#1a1a2e"/><text x="200" y="130" fill="#444" text-anchor="middle" font-family="system-ui" font-size="13">No screenshot</text></svg>`)
	})

	mux.HandleFunc("/", handlers.HomeHandler(db, tmpl))
	mux.HandleFunc("/app/", handlers.DetailHandler(db, tmpl))
	mux.HandleFunc("/api/apps", handlers.ApiAppsHandler(db))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	})

	addr := os.Getenv("BIND_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	log.Printf("Listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func runMigrations(db *sql.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS apps (
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
);
CREATE INDEX IF NOT EXISTS idx_apps_num ON apps(num);
CREATE INDEX IF NOT EXISTS idx_apps_score ON apps(total_score DESC);
CREATE INDEX IF NOT EXISTS idx_apps_grade ON apps(grade);
CREATE INDEX IF NOT EXISTS idx_apps_category ON apps(category);
ALTER TABLE apps ADD COLUMN IF NOT EXISTS improvement_note TEXT DEFAULT '';
`)
	return err
}
