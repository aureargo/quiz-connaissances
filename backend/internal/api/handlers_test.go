package api

// Tests de la couche HTTP.
//
// On n'a pas besoin de démarrer un vrai serveur ni d'ouvrir un port :
// `net/http/httptest` fournit une fausse requête (httptest.NewRequest) et un
// faux ResponseWriter qui enregistre la réponse (httptest.NewRecorder). On
// appelle le routeur directement dessus, en mémoire — c'est instantané.

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"quizconnaissances/internal/quiz"
)

// serveurDeTest monte un Server sur des données synthétiques temporaires.
func serveurDeTest(t *testing.T) http.Handler {
	t.Helper()

	racine := t.TempDir()

	ecrire(t, filepath.Join(racine, "themes.json"), []quiz.Theme{
		{ID: "go", Nom: "Go", Emoji: "🐹", Categorie: "Programmation", Description: "…"},
	})

	dossier := filepath.Join(racine, "questions", "go")
	if err := os.MkdirAll(dossier, 0o755); err != nil {
		t.Fatalf("mkdir : %v", err)
	}
	ecrire(t, filepath.Join(dossier, "facile.json"), []quiz.Question{
		{ID: "go-f1", Enonce: "?", Choix: []string{"a", "b", "c", "d"}, BonneReponse: 0, Explication: "…"},
	})
	ecrire(t, filepath.Join(dossier, "expert.json"), []quiz.Question{
		{ID: "go-e1", Enonce: "?", Choix: []string{"a", "b", "c", "d"}, BonneReponse: 3, Explication: "…"},
	})

	store, err := quiz.NewStore(racine)
	if err != nil {
		t.Fatalf("NewStore : %v", err)
	}
	return NewServer(store).Routes()
}

func ecrire(t *testing.T, chemin string, valeur any) {
	t.Helper()
	data, err := json.Marshal(valeur)
	if err != nil {
		t.Fatalf("encodage : %v", err)
	}
	if err := os.WriteFile(chemin, data, 0o644); err != nil {
		t.Fatalf("écriture : %v", err)
	}
}

// appeler joue une requête GET contre le routeur et renvoie la réponse.
func appeler(t *testing.T, routeur http.Handler, url string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	routeur.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, url, nil))
	return rec
}

func TestListeDesThemesEstLegere(t *testing.T) {
	rec := appeler(t, serveurDeTest(t), "/api/themes")

	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d, attendu 200", rec.Code)
	}

	var themes []quiz.Theme
	if err := json.Unmarshal(rec.Body.Bytes(), &themes); err != nil {
		t.Fatalf("réponse illisible : %v", err)
	}
	if len(themes) != 1 {
		t.Fatalf("%d thème(s), attendu 1", len(themes))
	}
	// Le point du "omitempty" : la liste d'accueil ne transporte pas les niveaux.
	if len(themes[0].Niveaux) != 0 {
		t.Errorf("la liste légère ne doit pas contenir les niveaux, reçu %v", themes[0].Niveaux)
	}
}

func TestUnThemePorteSesNiveaux(t *testing.T) {
	rec := appeler(t, serveurDeTest(t), "/api/themes/go")

	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d, attendu 200", rec.Code)
	}

	var theme quiz.Theme
	if err := json.Unmarshal(rec.Body.Bytes(), &theme); err != nil {
		t.Fatalf("réponse illisible : %v", err)
	}
	// facile + expert (ordre préféré) + le synthétique "tous".
	if len(theme.Niveaux) != 3 {
		t.Errorf("niveaux = %v, attendu 3 entrées", theme.Niveaux)
	}
}

func TestQuestionsDUnNiveau(t *testing.T) {
	rec := appeler(t, serveurDeTest(t), "/api/themes/go/questions/facile")

	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d, attendu 200", rec.Code)
	}

	var questions []quiz.Question
	if err := json.Unmarshal(rec.Body.Bytes(), &questions); err != nil {
		t.Fatalf("réponse illisible : %v", err)
	}
	if len(questions) != 1 || questions[0].ID != "go-f1" {
		t.Errorf("questions = %+v, attendu la seule go-f1", questions)
	}
}

func TestRessourcesInconnuesRenvoient404(t *testing.T) {
	routeur := serveurDeTest(t)

	cas := []struct {
		nom string
		url string
	}{
		{"thème inconnu", "/api/themes/inexistant"},
		{"niveau inconnu", "/api/themes/go/questions/inexistant"},
		{"thème inconnu sur les questions", "/api/themes/inexistant/questions/facile"},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			rec := appeler(t, routeur, c.url)
			if rec.Code != http.StatusNotFound {
				t.Errorf("code = %d, attendu 404", rec.Code)
			}
			// L'erreur doit être du JSON exploitable par le frontend.
			var corps map[string]string
			if err := json.Unmarshal(rec.Body.Bytes(), &corps); err != nil {
				t.Fatalf("le corps d'erreur n'est pas du JSON : %v", err)
			}
			if corps["erreur"] == "" {
				t.Errorf("pas de champ \"erreur\" dans %v", corps)
			}
		})
	}
}

func TestEnTetesCORS(t *testing.T) {
	rec := appeler(t, serveurDeTest(t), "/api/themes")

	if origine := rec.Header().Get("Access-Control-Allow-Origin"); origine == "" {
		t.Error("en-tête Access-Control-Allow-Origin absent : le frontend sera bloqué par le navigateur")
	}
}

func TestPreflightOptions(t *testing.T) {
	rec := httptest.NewRecorder()
	serveurDeTest(t).ServeHTTP(rec, httptest.NewRequest(http.MethodOptions, "/api/themes", nil))

	// Une requête préliminaire OPTIONS doit être traitée par le middleware et
	// répondre 204 sans atteindre le routeur.
	if rec.Code != http.StatusNoContent {
		t.Errorf("code = %d, attendu 204", rec.Code)
	}
}
