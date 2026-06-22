// Package api contient la couche HTTP : les "handlers" (fonctions qui
// répondent aux requêtes), le routeur, et le middleware CORS.
package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	// On importe notre propre package quiz. Le chemin commence par le nom du
	// module défini dans go.mod ("quizconnaissances").
	"quizconnaissances/internal/quiz"
)

// Server regroupe les "dépendances" dont les handlers ont besoin.
// Ici, un seul : le store de données. Les regrouper dans une struct évite
// d'utiliser des variables globales (mauvaise pratique).
type Server struct {
	store *quiz.Store
}

// NewServer construit un Server. On lui injecte le store : c'est de
// l'"injection de dépendances", un patron très courant et testable.
func NewServer(store *quiz.Store) *Server {
	return &Server{store: store}
}

// Routes assemble toutes les routes de l'API et renvoie un http.Handler prêt
// à être passé à http.ListenAndServe.
func (s *Server) Routes() http.Handler {
	// http.ServeMux est le routeur standard de Go. Depuis Go 1.22, on peut
	// préciser la MÉTHODE HTTP et des variables de chemin ({id}) directement
	// dans le motif. Plus besoin de bibliothèque externe.
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/themes", s.handleThemes)
	mux.HandleFunc("GET /api/themes/{id}", s.handleTheme)
	mux.HandleFunc("GET /api/themes/{id}/questions/{niveau}", s.handleQuestions)

	// On enveloppe le routeur dans le middleware CORS (voir plus bas).
	return cors(mux)
}

// handleThemes répond à GET /api/themes : renvoie la liste des thèmes en JSON,
// en version LÉGÈRE (sans les niveaux). Suffisant pour les tuiles de l'accueil ;
// les niveaux sont chargés à la demande via handleTheme.
func (s *Server) handleThemes(w http.ResponseWriter, r *http.Request) {
	repondreJSON(w, http.StatusOK, s.store.ResumesThemes())
}

// handleTheme répond à GET /api/themes/{id} : renvoie UN thème complet, avec ses
// niveaux. C'est la sous-page de sélection du niveau qui l'appelle, au clic sur
// une tuile : on ne charge alors que les niveaux du thème consulté.
func (s *Server) handleTheme(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	theme, ok := s.store.Theme(id)
	if !ok {
		repondreErreur(w, http.StatusNotFound,
			fmt.Sprintf("thème %q introuvable", id))
		return
	}
	repondreJSON(w, http.StatusOK, theme)
}

// handleQuestions répond à GET /api/themes/{id}/questions/{niveau}.
func (s *Server) handleQuestions(w http.ResponseWriter, r *http.Request) {
	// r.PathValue("...") récupère les variables {id} et {niveau} du motif de route.
	id := r.PathValue("id")
	niveau := r.PathValue("niveau")

	questions, ok := s.store.Questions(id, niveau)
	if !ok {
		repondreErreur(w, http.StatusNotFound,
			fmt.Sprintf("aucune question pour le thème %q au niveau %q", id, niveau))
		return
	}
	repondreJSON(w, http.StatusOK, questions)
}

// --- Fonctions utilitaires (privées au package) ---

// repondreJSON écrit `donnees` au format JSON dans la réponse, avec le code
// HTTP voulu. On centralise ça pour ne pas répéter les en-têtes partout.
func repondreJSON(w http.ResponseWriter, code int, donnees any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	// json.NewEncoder écrit directement dans le flux de réponse (efficace).
	if err := json.NewEncoder(w).Encode(donnees); err != nil {
		// Si l'encodage échoue, on ne peut plus changer le code HTTP (déjà
		// envoyé) : on se contente de logguer côté serveur.
		log.Printf("erreur d'encodage JSON : %v", err)
	}
}

// repondreErreur renvoie un message d'erreur au format JSON { "erreur": "..." }.
func repondreErreur(w http.ResponseWriter, code int, message string) {
	repondreJSON(w, code, map[string]string{"erreur": message})
}

// cors est un "middleware" : une fonction qui enveloppe un handler pour ajouter
// un comportement commun. Ici, on ajoute les en-têtes CORS autorisant le
// frontend Angular (servi sur un autre port, 4200) à appeler cette API.
func cors(suivant http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Le navigateur envoie parfois une requête "preflight" OPTIONS avant
		// la vraie requête. On y répond immédiatement, sans aller plus loin.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Sinon, on passe la main au handler suivant (le routeur).
		suivant.ServeHTTP(w, r)
	})
}
