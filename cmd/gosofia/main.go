package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/dimitardanailov/SofiaGoWorkshop/internal/diagnostics"

	"github.com/gorilla/mux"
)

type serverConf struct {
	port   string
	router http.Handler
	name   string
}

func main() {
	log.Printf("Starting the application ...")

	blPort := os.Getenv("PORT")
	if len(blPort) == 0 {
		log.Fatal("The application port should be set")
	}

	diagPort := os.Getenv("DIAG_PORT")
	if len(diagPort) == 0 {
		log.Fatal("The health checker port should be set")
	}

	router := mux.NewRouter()
	router.HandleFunc("/", hello)

	configurations := []serverConf{
		{
			port:   blPort,
			router: router,
			name:   "application server",
		},
		{
			port:   diagPort,
			router: diagnostics.NewDiagnostics(),
			name:   "diagnostics server",
		},
	}

	possibleErrors := make(chan error, 2)
	servers := make([]*http.Server, 2)

	for i, c := range configurations {
		go func(conf serverConf, i int) {
			log.Printf("The %s is preparing to handle connections...", conf.name)
			servers[i] = &http.Server{
				Addr:    ":" + conf.port,
				Handler: conf.router,
			}
			err := servers[i].ListenAndServe()
			if err != nil {
				possibleErrors <- err
			}
		}(c, i)
	}

	select {
	case err := <-possibleErrors:
		for _, s := range servers {
			s.Shutdown(context.Background())
		}
		log.Fatal(err)
	}
}

func hello(w http.ResponseWriter, r *http.Request) {
	log.Printf("The application is server is ready to handle connection ...")
	fmt.Fprint(w, http.StatusText(http.StatusOK))
}