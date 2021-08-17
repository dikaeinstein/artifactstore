package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dikaeinstein/artifactstore/artifactstore"
	"github.com/dikaeinstein/artifactstore/pkg/store"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
)

func handleArtifact(artifactStore *artifactstore.ArtifactStore, prefix string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		artifactURL := vars["artifactURL"]
		fmt.Println(artifactURL)
		ctx, cancelFunc := context.WithTimeout(r.Context(), time.Duration(60*time.Second))
		defer cancelFunc()

		artifact, err := artifactStore.Get(ctx, prefix, artifactURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		info, err := artifact.File.Stat()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Disposition", "Attachment")
		http.ServeContent(w, r, artifact.Name, info.ModTime(), artifact.File.(io.ReadSeeker))
	}
}

func main() {
	artifactStore := artifactstore.New(http.DefaultClient, ".", os.DirFS("."),
		store.NewInMemStore(make(map[string]*artifactstore.Artifact)))

	r := mux.NewRouter()
	r.HandleFunc("/3rdparty/{artifactURL}", handleArtifact(artifactStore, "3rdparty")).Methods(http.MethodGet)
	r.HandleFunc("/internal/{artifactURL}", handleArtifact(artifactStore, "internal")).Methods(http.MethodGet)
	http.Handle("/", r)

	log.Fatal(http.ListenAndServe(":9050", nil))
}
