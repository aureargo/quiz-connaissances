package quiz

// Tests unitaires du Store.
//
// Convention Go : un fichier de test se termine par `_test.go`, chaque test est
// une fonction `func TestXxx(t *testing.T)`, et on lance le tout avec `go test`.
// Rien à installer : le framework de test fait partie de la bibliothèque
// standard.
//
// Ces tests travaillent sur des données SYNTHÉTIQUES créées dans un dossier
// temporaire (t.TempDir()), jamais sur le vrai dossier data/ : un test doit être
// reproductible et ne dépendre de rien d'extérieur. La validation du VRAI
// contenu, c'est le rôle de donnees_test.go.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"
	"time"
)

// --- Utilitaires de test -------------------------------------------------

// creerDonnees fabrique un dossier de données temporaire.
//
// `themes` décrit les thèmes à écrire dans themes.json ; `niveaux` associe à
// chaque id de thème les fichiers de niveau à créer, avec leurs questions.
//
// t.TempDir() crée un dossier unique QUE Go supprime automatiquement à la fin du
// test. Aucun nettoyage à écrire, et aucun risque de collision entre tests.
func creerDonnees(t *testing.T, themes []Theme, niveaux map[string]map[string][]Question) string {
	t.Helper() // signale que cette fonction est un utilitaire : en cas d'échec,
	// Go pointe la ligne de l'APPELANT, pas celle d'ici.

	racine := t.TempDir()

	ecrireJSON(t, filepath.Join(racine, "themes.json"), themes)

	for themeID, fichiers := range niveaux {
		dossier := filepath.Join(racine, "questions", themeID)
		if err := os.MkdirAll(dossier, 0o755); err != nil {
			t.Fatalf("création du dossier %s : %v", dossier, err)
		}
		for niveau, questions := range fichiers {
			ecrireJSON(t, filepath.Join(dossier, niveau+".json"), questions)
		}
	}
	return racine
}

func ecrireJSON(t *testing.T, chemin string, valeur any) {
	t.Helper()
	data, err := json.Marshal(valeur)
	if err != nil {
		t.Fatalf("encodage JSON de %s : %v", chemin, err)
	}
	if err := os.WriteFile(chemin, data, 0o644); err != nil {
		t.Fatalf("écriture de %s : %v", chemin, err)
	}
}

// question fabrique une question minimale mais valide, identifiée par son id.
func question(id string) Question {
	return Question{
		ID:           id,
		Enonce:       "Énoncé de " + id + " ?",
		Choix:        []string{"a", "b", "c", "d"},
		BonneReponse: 0,
		Explication:  "Parce que.",
	}
}

// idsDe extrait les ids d'une liste de questions, pour comparer facilement.
func idsDe(questions []Question) []string {
	ids := make([]string, len(questions))
	for i, q := range questions {
		ids[i] = q.ID
	}
	return ids
}

// storeDeTest monte un store à deux thèmes, utilisé par plusieurs tests :
//   - "go"     : trois niveaux (facile, moyen, expert) → aura donc "tous" ;
//   - "solo"   : un seul niveau → n'aura PAS "tous".
func storeDeTest(t *testing.T) (*Store, string) {
	t.Helper()

	racine := creerDonnees(t,
		[]Theme{
			{ID: "go", Nom: "Go", Emoji: "🐹", Categorie: "Programmation", Description: "…"},
			{ID: "solo", Nom: "Solo", Emoji: "1️⃣", Categorie: "Divers", Description: "…"},
		},
		map[string]map[string][]Question{
			"go": {
				"facile": {question("go-f1"), question("go-f2")},
				"moyen":  {question("go-m1")},
				"expert": {question("go-e1")},
			},
			"solo": {
				"unique": {question("solo-1")},
			},
		},
	)

	store, err := NewStore(racine)
	if err != nil {
		t.Fatalf("NewStore : %v", err)
	}
	return store, racine
}

// --- Découverte des niveaux au démarrage ---------------------------------

func TestNewStoreDecouvreLesNiveauxDansLOrdrePrefere(t *testing.T) {
	store, _ := storeDeTest(t)

	theme, ok := store.Theme("go")
	if !ok {
		t.Fatal("thème \"go\" introuvable")
	}

	// facile → moyen → expert (ordre préféré), puis le synthétique "tous".
	attendu := []string{"facile", "moyen", "expert", NiveauTous}
	if !slices.Equal(theme.Niveaux, attendu) {
		t.Errorf("niveaux = %v, attendu %v", theme.Niveaux, attendu)
	}
}

func TestNewStoreNAjoutePasTousSiUnSeulNiveau(t *testing.T) {
	store, _ := storeDeTest(t)

	theme, ok := store.Theme("solo")
	if !ok {
		t.Fatal("thème \"solo\" introuvable")
	}

	// Un seul niveau réel : "tous" ferait doublon, on ne l'ajoute pas.
	attendu := []string{"unique"}
	if !slices.Equal(theme.Niveaux, attendu) {
		t.Errorf("niveaux = %v, attendu %v", theme.Niveaux, attendu)
	}
}

func TestNewStoreEchoueSiUnThemeNaAucuneQuestion(t *testing.T) {
	// Thème déclaré dans themes.json… mais sans dossier de questions.
	racine := creerDonnees(t,
		[]Theme{{ID: "fantome", Nom: "Fantôme"}},
		nil,
	)

	if _, err := NewStore(racine); err == nil {
		t.Fatal("NewStore aurait dû échouer (fail-fast) pour un thème sans questions")
	}
}

func TestNewStoreIgnoreUnFichierTousJson(t *testing.T) {
	// "tous" est un nom RÉSERVÉ : un tous.json déposé à la main doit être ignoré
	// pour éviter toute ambiguïté avec le niveau synthétique.
	racine := creerDonnees(t,
		[]Theme{{ID: "go", Nom: "Go"}},
		map[string]map[string][]Question{
			"go": {
				"facile":   {question("go-f1")},
				"moyen":    {question("go-m1")},
				NiveauTous: {question("go-pirate")},
			},
		},
	)

	store, err := NewStore(racine)
	if err != nil {
		t.Fatalf("NewStore : %v", err)
	}
	theme, _ := store.Theme("go")

	attendu := []string{"facile", "moyen", NiveauTous}
	if !slices.Equal(theme.Niveaux, attendu) {
		t.Errorf("niveaux = %v, attendu %v", theme.Niveaux, attendu)
	}

	// Et surtout : la question du tous.json pirate ne doit apparaître nulle part.
	questions, ok := store.Questions("go", NiveauTous)
	if !ok {
		t.Fatal("le niveau \"tous\" devrait exister")
	}
	if slices.Contains(idsDe(questions), "go-pirate") {
		t.Errorf("le contenu de tous.json a été chargé : %v", idsDe(questions))
	}
}

func TestTrierNiveaux(t *testing.T) {
	// Les noms connus passent devant, dans l'ordre facile → moyen → expert ;
	// les noms libres suivent, par ordre alphabétique.
	cas := []struct {
		nom     string
		entree  []string
		attendu []string
	}{
		{"ordre préféré respecté", []string{"expert", "facile", "moyen"}, []string{"facile", "moyen", "expert"}},
		{"noms libres après", []string{"zebre", "facile", "alpha"}, []string{"facile", "alpha", "zebre"}},
		{"que des noms libres", []string{"saison-2", "saison-1"}, []string{"saison-1", "saison-2"}},
	}

	for _, c := range cas {
		// t.Run crée un SOUS-TEST nommé : en cas d'échec, on sait lequel.
		t.Run(c.nom, func(t *testing.T) {
			entree := slices.Clone(c.entree)
			trierNiveaux(entree)
			if !slices.Equal(entree, c.attendu) {
				t.Errorf("trierNiveaux(%v) = %v, attendu %v", c.entree, entree, c.attendu)
			}
		})
	}
}

// --- Lecture des questions ------------------------------------------------

func TestQuestionsDUnNiveauReel(t *testing.T) {
	store, _ := storeDeTest(t)

	questions, ok := store.Questions("go", "facile")
	if !ok {
		t.Fatal("Questions(go, facile) a échoué")
	}
	if attendu := []string{"go-f1", "go-f2"}; !slices.Equal(idsDe(questions), attendu) {
		t.Errorf("ids = %v, attendu %v", idsDe(questions), attendu)
	}
}

func TestQuestionsTousConcatèneLesNiveauxReels(t *testing.T) {
	store, _ := storeDeTest(t)

	questions, ok := store.Questions("go", NiveauTous)
	if !ok {
		t.Fatal("Questions(go, tous) a échoué")
	}

	// Dans l'ordre des niveaux : facile (2), puis moyen (1), puis expert (1).
	attendu := []string{"go-f1", "go-f2", "go-m1", "go-e1"}
	if !slices.Equal(idsDe(questions), attendu) {
		t.Errorf("ids = %v, attendu %v", idsDe(questions), attendu)
	}
}

func TestQuestionsRefuseLesCouplesInconnus(t *testing.T) {
	store, _ := storeDeTest(t)

	cas := []struct{ theme, niveau string }{
		{"inexistant", "facile"}, // thème inconnu
		{"go", "inexistant"},     // niveau inconnu
		{"solo", NiveauTous},     // "tous" non généré (un seul niveau réel)
	}

	for _, c := range cas {
		if _, ok := store.Questions(c.theme, c.niveau); ok {
			t.Errorf("Questions(%q, %q) aurait dû échouer", c.theme, c.niveau)
		}
	}
}

func TestQuestionsSignaleUnJSONInvalide(t *testing.T) {
	store, racine := storeDeTest(t)

	// On casse le fichier APRÈS le démarrage : le store n'a pas encore lu son
	// contenu (lecture paresseuse), l'erreur doit donc apparaître ici.
	chemin := filepath.Join(racine, "questions", "go", "facile.json")
	if err := os.WriteFile(chemin, []byte("[{ pas du JSON"), 0o644); err != nil {
		t.Fatalf("écriture : %v", err)
	}

	if _, ok := store.Questions("go", "facile"); ok {
		t.Error("Questions aurait dû échouer sur un JSON invalide")
	}
	// Et "tous", qui agrège ce niveau, doit échouer aussi.
	if _, ok := store.Questions("go", NiveauTous); ok {
		t.Error("Questions(tous) aurait dû échouer puisqu'un de ses niveaux est cassé")
	}
}

// --- Cache -----------------------------------------------------------------

func TestQuestionsRelitLeFichierApresModification(t *testing.T) {
	store, racine := storeDeTest(t)

	// 1er accès : met le fichier en cache.
	if _, ok := store.Questions("go", "facile"); !ok {
		t.Fatal("premier accès échoué")
	}

	// On réécrit le fichier avec un contenu différent…
	chemin := filepath.Join(racine, "questions", "go", "facile.json")
	ecrireJSON(t, chemin, []Question{question("go-f99")})

	// …et on force une date de modification franchement postérieure. Sans ça, le
	// test serait fragile : sur certains systèmes de fichiers, deux écritures
	// rapprochées peuvent partager la même date (résolution parfois > 1 s).
	futur := time.Now().Add(2 * time.Second)
	if err := os.Chtimes(chemin, futur, futur); err != nil {
		t.Fatalf("os.Chtimes : %v", err)
	}

	questions, ok := store.Questions("go", "facile")
	if !ok {
		t.Fatal("second accès échoué")
	}
	if attendu := []string{"go-f99"}; !slices.Equal(idsDe(questions), attendu) {
		t.Errorf("ids = %v, attendu %v — le cache n'a pas été invalidé", idsDe(questions), attendu)
	}
}

func TestQuestionsUtiliseLeCacheSiLeFichierNaPasChange(t *testing.T) {
	store, racine := storeDeTest(t)

	if _, ok := store.Questions("go", "facile"); !ok {
		t.Fatal("premier accès échoué")
	}

	// On supprime le fichier SANS toucher au cache. Un os.Stat va échouer, donc
	// l'accès échoue aussi : c'est le comportement voulu (on vérifie toujours le
	// disque). Ce test documente cette limite en la rendant explicite.
	chemin := filepath.Join(racine, "questions", "go", "facile.json")
	if err := os.Remove(chemin); err != nil {
		t.Fatalf("suppression : %v", err)
	}

	if _, ok := store.Questions("go", "facile"); ok {
		t.Error("Questions aurait dû échouer : le fichier n'existe plus")
	}
}

// TestQuestionsEnParallele vérifie qu'un accès concurrent au cache fonctionne.
// Il prend tout son sens avec le détecteur de data races :
//
//	go test -race ./...
//
// Sans -race il passe toujours ; avec, il échoue si le cache est accédé sans
// verrou. C'est le filet de sécurité du sync.RWMutex du store.
//
// ⚠️ Sous Windows, `-race` exige cgo et donc un compilateur C (gcc via MSYS2 ou
// TDM-GCC). Sans lui, `go test -race` refuse de compiler : le test tourne alors
// en mode ordinaire, ce qui reste utile (il détecte un blocage ou une panique)
// mais NE prouve pas l'absence de data race.
func TestQuestionsEnParallele(t *testing.T) {
	store, _ := storeDeTest(t)

	// sync.WaitGroup attend qu'un groupe de goroutines soit terminé.
	var attente sync.WaitGroup
	niveaux := []string{"facile", "moyen", "expert", NiveauTous}

	for i := 0; i < 50; i++ {
		attente.Add(1)
		go func(n string) {
			defer attente.Done()
			if _, ok := store.Questions("go", n); !ok {
				t.Errorf("Questions(go, %s) a échoué en parallèle", n)
			}
		}(niveaux[i%len(niveaux)])
	}
	attente.Wait()
}

// --- Encapsulation ---------------------------------------------------------

func TestThemesRenvoieUneCopie(t *testing.T) {
	store, _ := storeDeTest(t)

	// On récupère les thèmes… et on les saccage.
	themes := store.Themes()
	themes[0].Nom = "PIRATÉ"
	if len(themes[0].Niveaux) > 0 {
		themes[0].Niveaux[0] = "PIRATÉ"
	}

	// Le store, lui, ne doit pas avoir bougé d'un poil.
	theme, _ := store.Theme("go")
	if theme.Nom == "PIRATÉ" {
		t.Error("Themes() expose le nom interne : une modification externe a atteint le store")
	}
	if slices.Contains(theme.Niveaux, "PIRATÉ") {
		t.Error("Themes() expose le slice de niveaux interne")
	}
}

func TestThemeRenvoieUneCopie(t *testing.T) {
	store, _ := storeDeTest(t)

	theme, _ := store.Theme("go")
	theme.Niveaux[0] = "PIRATÉ"

	encore, _ := store.Theme("go")
	if slices.Contains(encore.Niveaux, "PIRATÉ") {
		t.Error("Theme() expose le slice de niveaux interne")
	}
}

func TestResumesThemesNExposePasLesNiveaux(t *testing.T) {
	store, _ := storeDeTest(t)

	for _, theme := range store.ResumesThemes() {
		if len(theme.Niveaux) != 0 {
			t.Errorf("le thème %q ne devrait pas porter ses niveaux dans la liste légère", theme.ID)
		}
	}

	// …et le store doit toujours les connaître.
	if theme, _ := store.Theme("go"); len(theme.Niveaux) == 0 {
		t.Error("ResumesThemes() a effacé les niveaux du store")
	}
}
