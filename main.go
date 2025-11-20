package main

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"MrRSS/internal/database"
	"MrRSS/internal/feed"
	"MrRSS/internal/handlers"
	"MrRSS/internal/translation"
)

//go:embed frontend/dist/*
var frontendFiles embed.FS

type CombinedHandler struct {
	apiMux     *http.ServeMux
	fileServer http.Handler
}

func (h *CombinedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		h.apiMux.ServeHTTP(w, r)
		return
	}
	h.fileServer.ServeHTTP(w, r)
}

func main() {
	f, _ := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()
	log.SetOutput(f)

	log.Println("Starting application...")

	// Initialize database
	log.Println("Initializing Database...")
	db, err := database.NewDB("rss.db")
	if err != nil {
		log.Printf("Error initializing database: %v", err)
		log.Fatal(err)
	}
	log.Println("Database initialized successfully")

	translator := translation.NewGoogleFreeTranslator()
	fetcher := feed.NewFetcher(db, translator)
	h := handlers.NewHandler(db, fetcher)

	// API Routes
	log.Println("Setting up API routes...")
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/api/feeds", h.HandleFeeds)
	apiMux.HandleFunc("/api/feeds/add", h.HandleAddFeed)
	apiMux.HandleFunc("/api/feeds/delete", h.HandleDeleteFeed)
	apiMux.HandleFunc("/api/feeds/update", h.HandleUpdateFeed)
	apiMux.HandleFunc("/api/articles", h.HandleArticles)
	apiMux.HandleFunc("/api/articles/read", h.HandleMarkRead)
	apiMux.HandleFunc("/api/articles/favorite", h.HandleToggleFavorite)
	apiMux.HandleFunc("/api/settings", h.HandleSettings)
	apiMux.HandleFunc("/api/refresh", h.HandleRefresh)
	apiMux.HandleFunc("/api/progress", h.HandleProgress)
	apiMux.HandleFunc("/api/opml/import", h.HandleOPMLImport)
	apiMux.HandleFunc("/api/opml/export", h.HandleOPMLExport)

	// Static Files
	log.Println("Setting up static files...")
	frontendFS, err := fs.Sub(frontendFiles, "frontend/dist")
	if err != nil {
		log.Fatal(err)
	}

	fileServer := http.FileServer(http.FS(frontendFS))

	combinedHandler := &CombinedHandler{
		apiMux:     apiMux,
		fileServer: fileServer,
	}

	// Start background scheduler
	log.Println("Starting background scheduler...")
	bgCtx, bgCancel := context.WithCancel(context.Background())

	log.Println("Starting Wails...")
	err = wails.Run(&options.App{
		Title:            "MrRSS",
		Width:            1024,
		Height:           768,
		WindowStartState: options.Maximised,
		LogLevel:         logger.DEBUG,
		AssetServer: &assetserver.Options{
			Assets:  frontendFS,
			Handler: combinedHandler,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		OnShutdown: func(ctx context.Context) {
			log.Println("Shutting down...")

			// Stop background tasks first
			bgCancel()
			// Give some time for tasks to finish
			time.Sleep(500 * time.Millisecond)

			// Close DB with timeout
			done := make(chan struct{})
			go func() {
				if err := db.Close(); err != nil {
					log.Printf("Error closing database: %v", err)
				}
				close(done)
			}()

			select {
			case <-done:
				log.Println("Database closed")
			case <-time.After(2 * time.Second):
				log.Println("Database close timed out")
			}
		},
		OnStartup: func(ctx context.Context) {
			log.Println("App started")

			// Initialize DB in background
			go func() {
				log.Println("Running DB migrations...")
				if err := db.Init(); err != nil {
					log.Printf("Error initializing database schema: %v", err)
				}
				log.Println("DB migrations finished")
			}()

			go func() {
				time.Sleep(2 * time.Second)
				h.StartBackgroundScheduler(bgCtx)
			}()
		},
	})

	if err != nil {
		log.Printf("Error running Wails: %v", err)
		log.Fatal(err)
	}
	log.Println("Application finished")
}
