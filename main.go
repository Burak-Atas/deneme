package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/pion/webrtc/v2"
)

// Prepare the configuration
var peerConnectionConfig = webrtc.Configuration{
	ICEServers: []webrtc.ICEServer{
		{
			URLs: []string{"stun:stun.l.google.com:19302"},
		},
	},
}

func main() {
	rooms := NewRooms()
	router := mux.NewRouter()

	router.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		bytes, err := json.Marshal(rooms.GetStats())
		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}
		w.Write(bytes)
	}).Methods("GET")

	router.HandleFunc("/api/rooms/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		roomID := vars["id"]
		room, err := rooms.Get(roomID)
		if err == errNotFound {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Room not found"})
			return
		}
		bytes, err := json.Marshal(room.Wrap(nil))
		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}
		w.Write(bytes)
	}).Methods("GET")

	router.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		serveWs(rooms, w, r)
	})

	// Serve the HTML file
	router.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./static"))))

	// CORS ayarlarını ekleyin
	headers := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	origins := handlers.AllowedOrigins([]string{"*"})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}
	addr := fmt.Sprintf(":%s", port)
	fmt.Printf("listening on %s\n", addr)

	srv := &http.Server{
		Handler:      handlers.CORS(headers, methods, origins)(router),
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
