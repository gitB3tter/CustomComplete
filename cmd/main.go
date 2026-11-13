package main

import (
	"context"
	"customcomplete/internal/admin"
	"customcomplete/internal/catalog"
	"customcomplete/internal/docs"
	"customcomplete/internal/home"
	"customcomplete/internal/progress"
	"customcomplete/internal/store"
	"customcomplete/internal/suggestions"
	"customcomplete/internal/textprocessing"
	"customcomplete/internal/texts"
	"customcomplete/internal/transport"
	"customcomplete/internal/ui"
	"customcomplete/internal/weblinking"
	"customcomplete/static"
	_ "embed"
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

//go:embed favicon-light.ico
var faviconLight []byte

//go:embed favicon-dark.ico
var faviconDark []byte

//go:embed robots.txt
var robotsTxt []byte

func main() {

	var err error
	var readWrite store.ReadWrite
	var read store.Read

	reset := flag.Bool("reset", false, "if the db should be reset on server start")
	addr := flag.String("addr", ":8080", "listen address")
	dbPath := flag.String("db", "customcomplete.db", "SQLite database file (created if missing)")
	numWorkers := flag.Int("workers", 2, "Number of Async background workers")

	flag.Parse()

	if *reset {
		err = store.ResetDB(*dbPath)
		if err != nil {
			log.Fatal(err)
		}
	}

	origin, err := url.Parse("http://" + *addr)
	if err != nil {
		log.Fatalf("invalid address: %v", err)
	}

	weblinking.Init(origin)

	if readWrite, err = store.GetReadWrite(*dbPath); err != nil {
		log.Fatal(err)
	}
	if err = readWrite.Ping(); err != nil {
		log.Fatal(err)
	}
	if read, err = store.GetRead(*dbPath); err != nil {
		log.Fatal(err)
	}
	if err = read.Ping(); err != nil {
		log.Fatal(err)
	}

	// async job workers, worker channel
	broker := progress.NewBroker()
	jobs := make(chan int64, 100)
	var wg sync.WaitGroup
	if err = textprocessing.Init(readWrite); err != nil {
		log.Fatal(err)
	}
	ctxWorker, cancelWorkers := context.WithCancel(context.Background())
	defer cancelWorkers()
	for workerID := 1; workerID <= *numWorkers; workerID++ {
		wg.Add(1)
		go textprocessing.Worker(ctxWorker, workerID, readWrite, read, broker, jobs, &wg)
	}

	rootMux := http.NewServeMux()

	//admin endpoints (only reachable from localhost)
	adminMux := admin.NewMux()
	rootMux.Handle("/admin/", adminMux)

	mainMux := http.NewServeMux()
	rootMux.Handle("/", transport.Log(transport.NegotiateLanguague(mainMux)))

	var baseTemplates *ui.BaseTemplates
	baseTemplates, err = ui.NewBaseTemplates()
	if err != nil {
		log.Fatal(err)
	}

	//home
	// /
	homeHandler, err := home.NewHandler(baseTemplates.MustFork())
	if err != nil {
		log.Fatal(err)
	}
	homeHandler.Register(mainMux)

	// http resources
	// /texts and below
	textservice := catalog.NewService(readWrite, read, broker, jobs)
	textsHandler, err := texts.NewHandler(textservice, baseTemplates.MustFork())
	if err != nil {
		log.Fatal(err)
	}
	textsHandler.Register(mainMux)
	// docs (link-rel definitions, json profiles, ...)
	// /docs and below
	docsHandler, err := docs.NewHandler(baseTemplates.MustFork())
	if err != nil {
		log.Fatal(err)
	}
	docsHandler.Register(mainMux)

	//suggestions
	// /suggest
	suggestHandler, err := suggestions.NewHandler(read, baseTemplates)
	if err != nil {
		log.Fatal(err)
	}
	suggestHandler.Register(mainMux)

	// static assets
	rootMux.Handle("/static/", withCacheControlAssets(http.StripPrefix("/static/", http.FileServer(http.FS(static.FS)))))

	rootMux.HandleFunc("GET /favicon-light.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/x-icon")
		transport.SetCachePublic(w.Header(), time.Hour, time.Hour*24)
		w.Write(faviconLight)
	})

	rootMux.HandleFunc("GET /favicon-dark.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/x-icon")
		transport.SetCachePublic(w.Header(), time.Hour, time.Hour*24)
		w.Write(faviconDark)
	})

	// Fallback / standard route
	rootMux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/x-icon")
		transport.SetCachePublic(w.Header(), time.Hour, time.Hour*24)
		w.Write(faviconLight)
	})

	rootMux.HandleFunc("GET /robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		transport.SetCachePublic(w.Header(), time.Hour, time.Hour*24)
		w.Write(robotsTxt)
	})

	server := &http.Server{
		Addr:         *addr,
		Handler:      rootMux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		log.Printf("Server listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("Shutdown signal received. Trapping state...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = server.Shutdown(ctx)

	close(jobs)

	cancelWorkers()

	wg.Wait()

	read.Close()
	readWrite.Close()

	log.Println("Clean shutdown complete. System exits safely.")

}

func withCacheControlAssets(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// cache for 1 day (86400s), stale-while-revalidate for 1 week (604800s)
		w.Header().Set("Cache-Control", "public, max-age=86400, stale-while-revalidate=604800")

		next.ServeHTTP(w, r)
	})
}
