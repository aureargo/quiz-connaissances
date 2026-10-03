package quiz

// VALIDATEUR DU CONTENU RÉEL (data/).
//
// Les autres tests (store_test.go) vérifient le CODE sur des données inventées.
// Celui-ci fait l'inverse : il vérifie les DONNÉES avec le code comme outil.
//
//	cd backend && go test ./internal/quiz -run TestDonnees -v
//
// Il encode les règles de data/QUESTIONNAIRE_TEMPLATE.md, qui n'étaient jusqu'ici
// qu'un document : rien ne les vérifiait, et le contenu pouvait dériver sans
// qu'on s'en aperçoive. Deux familles, séparées comme dans le template :
//
//	TestDonneesReglesDures   → §2 : ce que le code impose. Un manquement casse
//	                           le jeu (question ingagnable, thème invisible…).
//	TestDonneesQualite       → §3 : conventions de qualité. Rien ne casse, mais
//	                           le quiz devient moins bon (réponse devinable…).
//
// C'est aussi ce qui rattrape la contrepartie du chargement paresseux : le
// serveur ne lit plus les fichiers au démarrage, donc un JSON casse ne se voit
// qu'à l'ouverture du quiz concerné. Ici, on les lit TOUS, hors production.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"
)

// nbProblemesAffiches limite le bruit : au-delà, on ne montre qu'un total.
// Un test qui déverse 300 lignes ne se lit pas.
const nbProblemesAffiches = 25

// dossierDonnees localise le vrai data/ depuis ce paquet.
//
// Les tests Go s'exécutent avec le dossier du paquet comme répertoire courant,
// soit backend/internal/quiz : trois crans plus haut, on est à la racine du
// dépôt. QUIZ_DATA_DIR permet de pointer ailleurs (autre jeu de contenu).
func dossierDonnees(t *testing.T) string {
	t.Helper()

	chemin := os.Getenv("QUIZ_DATA_DIR")
	if chemin == "" {
		chemin = filepath.Join("..", "..", "..", "data")
	}

	if _, err := os.Stat(filepath.Join(chemin, "themes.json")); err != nil {
		// t.Skip plutôt que t.Fatal : sans le dossier de contenu (dépôt partiel,
		// backend extrait seul…), il n'y a rien à valider — ce n'est pas un
		// échec du code. Le test s'affiche alors comme "SKIP", donc visible.
		t.Skipf("dossier de données introuvable (%s) : rien à valider", chemin)
	}
	return chemin
}

// rapport accumule les problèmes trouvés puis les restitue proprement.
type rapport struct {
	problemes []string
}

func (r *rapport) ajouter(format string, args ...any) {
	r.problemes = append(r.problemes, fmt.Sprintf(format, args...))
}

// signaler transmet les problèmes au test, en bornant l'affichage.
func (r *rapport) signaler(t *testing.T, titre string) {
	t.Helper()
	if len(r.problemes) == 0 {
		return
	}

	affiches := min(len(r.problemes), nbProblemesAffiches)
	var b strings.Builder
	fmt.Fprintf(&b, "%s — %d problème(s) :\n", titre, len(r.problemes))
	for _, p := range r.problemes[:affiches] {
		fmt.Fprintf(&b, "  • %s\n", p)
	}
	if reste := len(r.problemes) - affiches; reste > 0 {
		fmt.Fprintf(&b, "  … et %d autre(s).\n", reste)
	}
	t.Error(b.String())
}

// contenu est le vrai dossier data/, lu en entier pour l'occasion.
type contenu struct {
	racine     string
	themes     []Theme
	categories []string
	// questions : themeID -> niveau -> questions (niveaux RÉELS uniquement).
	questions map[string]map[string][]Question
}

// lireContenu charge tout data/ en mémoire. Un JSON invalide est signalé ici :
// c'est précisément ce que le serveur ne détecte plus au démarrage.
func lireContenu(t *testing.T) *contenu {
	t.Helper()

	racine := dossierDonnees(t)
	c := &contenu{racine: racine, questions: make(map[string]map[string][]Question)}

	if err := lireJSON(filepath.Join(racine, "themes.json"), &c.themes); err != nil {
		t.Fatalf("themes.json illisible : %v", err)
	}

	// categories.json est optionnel (cf. ordonnerParCategorie).
	chemin := filepath.Join(racine, "categories.json")
	if _, err := os.Stat(chemin); err == nil {
		if err := lireJSON(chemin, &c.categories); err != nil {
			t.Fatalf("categories.json illisible : %v", err)
		}
	}

	for _, theme := range c.themes {
		dossier := filepath.Join(racine, "questions", theme.ID)
		entrees, err := os.ReadDir(dossier)
		if err != nil {
			continue // absence signalée par TestDonneesReglesDures
		}

		c.questions[theme.ID] = make(map[string][]Question)
		for _, e := range entrees {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			niveau := strings.TrimSuffix(e.Name(), ".json")

			var questions []Question
			if err := lireJSON(filepath.Join(dossier, e.Name()), &questions); err != nil {
				// §2.3 : un JSON mal formé rend le niveau injouable. Fatal, car
				// sans lui les vérifications suivantes n'ont plus de sens.
				t.Fatalf("%s/%s.json : JSON invalide — le niveau serait injouable (%v)",
					theme.ID, niveau, err)
			}
			c.questions[theme.ID][niveau] = questions
		}
	}
	return c
}

// reference identifie une question dans les messages d'erreur.
func reference(themeID, niveau string, q Question, index int) string {
	id := q.ID
	if id == "" {
		id = fmt.Sprintf("(sans id, %dᵉ de la liste)", index+1)
	}
	return fmt.Sprintf("%s/%s.json → %s", themeID, niveau, id)
}

// --- §2 : les règles DURES -------------------------------------------------

func TestDonneesReglesDures(t *testing.T) {
	c := lireContenu(t)
	r := &rapport{}

	// --- themes.json ---
	vus := make(map[string]string) // id -> nom, pour repérer les doublons
	for _, theme := range c.themes {
		if theme.ID == "" {
			r.ajouter("themes.json : un thème sans id (%q)", theme.Nom)
			continue
		}
		if precedent, existe := vus[theme.ID]; existe {
			r.ajouter("themes.json : id %q en double (%q et %q)", theme.ID, precedent, theme.Nom)
		}
		vus[theme.ID] = theme.Nom

		// Champs indispensables à l'affichage d'une tuile.
		for nom, valeur := range map[string]string{
			"nom": theme.Nom, "emoji": theme.Emoji,
			"categorie": theme.Categorie, "description": theme.Description,
		} {
			if strings.TrimSpace(valeur) == "" {
				r.ajouter("thème %q : champ %q vide", theme.ID, nom)
			}
		}

		// §4 : catégorie déclarée (c'est le filet anti-faute-de-frappe).
		if len(c.categories) > 0 && !slices.Contains(c.categories, theme.Categorie) {
			r.ajouter("thème %q : catégorie %q absente de categories.json (faute de frappe ?)",
				theme.ID, theme.Categorie)
		}

		// §2.2 : au moins un fichier de questions, sinon le backend ne démarre pas.
		if len(c.questions[theme.ID]) == 0 {
			r.ajouter("thème %q : aucun fichier de questions → le backend refusera de démarrer",
				theme.ID)
		}
	}

	// §2.2 (revers) : un dossier sans entrée dans themes.json est ignoré en
	// silence — le travail de rédaction serait invisible dans l'application.
	if entrees, err := os.ReadDir(filepath.Join(c.racine, "questions")); err == nil {
		for _, e := range entrees {
			// Test de PRÉSENCE (comma ok), pas de valeur : un thème déclaré au
			// nom vide aurait sinon son dossier signalé à tort comme orphelin.
			if _, declare := vus[e.Name()]; e.IsDir() && !declare {
				r.ajouter("questions/%s/ : dossier sans entrée dans themes.json → thème invisible",
					e.Name())
			}
		}
	}

	// --- les questions ---
	for themeID, niveaux := range c.questions {
		idsDuTheme := make(map[string]string) // id de question -> niveau d'origine

		for niveau, questions := range niveaux {
			// §2.1 : "tous" est réservé au niveau synthétique.
			if niveau == NiveauTous {
				r.ajouter("%s/%s.json : nom de fichier réservé, il sera ignoré", themeID, NiveauTous)
			}
			if len(questions) == 0 {
				r.ajouter("%s/%s.json : fichier vide", themeID, niveau)
			}

			for i, q := range questions {
				ref := reference(themeID, niveau, q, i)

				// §2.7 : id présent et unique au sein du thème (le niveau "tous"
				// rassemble tous les niveaux, donc l'unicité s'y joue).
				if q.ID == "" {
					r.ajouter("%s : id manquant", ref)
				} else if origine, existe := idsDuTheme[q.ID]; existe {
					r.ajouter("%s : id déjà utilisé dans %s.json du même thème", ref, origine)
				} else {
					idsDuTheme[q.ID] = niveau
				}

				if strings.TrimSpace(q.Enonce) == "" {
					r.ajouter("%s : énoncé vide", ref)
				}
				if strings.TrimSpace(q.Explication) == "" {
					r.ajouter("%s : explication vide", ref)
				}

				// §2.5 : 2 à 4 choix réellement utilisables.
				if len(q.Choix) < 2 || len(q.Choix) > 4 {
					r.ajouter("%s : %d choix (attendu entre 2 et 4)", ref, len(q.Choix))
				}
				for j, choix := range q.Choix {
					if strings.TrimSpace(choix) == "" {
						r.ajouter("%s : choix n°%d vide", ref, j+1)
					}
				}
				// Deux choix identiques : deux bonnes réponses possibles à l'œil
				// du joueur, alors qu'une seule sera comptée juste.
				for j := range q.Choix {
					for k := j + 1; k < len(q.Choix); k++ {
						if q.Choix[j] == q.Choix[k] {
							r.ajouter("%s : choix n°%d et n°%d identiques (%q)", ref, j+1, k+1, q.Choix[j])
						}
					}
				}

				// §2.4 : LE piège silencieux. Un index hors limites rend la
				// question ingagnable : aucune réponse n'est jamais correcte,
				// donc elle revient indéfiniment dans la file, sans message.
				if q.BonneReponse < 0 || q.BonneReponse >= len(q.Choix) {
					r.ajouter("%s : bonneReponse = %d hors limites (0..%d) → question INGAGNABLE",
						ref, q.BonneReponse, len(q.Choix)-1)
				}
			}
		}
	}

	r.signaler(t, "Règles dures (§2 du template)")
}

// --- §3 : les recommandations de qualité -----------------------------------

func TestDonneesQualite(t *testing.T) {
	c := lireContenu(t)
	r := &rapport{}

	for themeID, niveaux := range c.questions {
		for niveau, questions := range niveaux {
			for i, q := range questions {
				ref := reference(themeID, niveau, q, i)

				// §3.1 : exactement 4 choix. Au-delà, l'affichage des lettres
				// A–D et les raccourcis clavier 1–4 ne suivent plus.
				if len(q.Choix) != 4 {
					r.ajouter("%s : %d choix (4 recommandés)", ref, len(q.Choix))
				}

				if deseq, details := choixDesequilibre(q); deseq {
					r.ajouter("%s : %s", ref, details)
				}
			}
		}
	}

	r.signaler(t, "Qualité du contenu (§3 du template)")
}

// choixDesequilibre applique la règle ANTI-DÉDUCTION (§3.6) : la bonne réponse
// ne doit pas se trahir par sa forme.
//
// Un joueur qui ne connaît pas la réponse repère celle qui est visiblement plus
// détaillée que les autres — sigle développé, exemple entre parenthèses,
// définition plus complète. Il gagne sans savoir, et la question ne mesure plus
// rien.
//
// Le critère retenu combine deux conditions, pour éviter les faux positifs :
//   - la bonne réponse est plus de 1,6× plus longue que la plus longue des
//     autres (rapport, donc insensible à l'échelle) ;
//   - ET l'écart dépasse 12 caractères (garde-fou sur les choix très courts,
//     où « oui » vs « absolument » n'a rien de suspect).
//
// ⚠️ On compte en RUNES (utf8.RuneCountInString), pas en octets : en français,
// « é » occupe 2 octets. len() fausserait la comparaison.
func choixDesequilibre(q Question) (bool, string) {
	if q.BonneReponse < 0 || q.BonneReponse >= len(q.Choix) {
		return false, "" // déjà signalé par les règles dures
	}

	bonne := utf8.RuneCountInString(q.Choix[q.BonneReponse])
	plusLongDesAutres := 0
	for i, choix := range q.Choix {
		if i == q.BonneReponse {
			continue
		}
		plusLongDesAutres = max(plusLongDesAutres, utf8.RuneCountInString(choix))
	}
	if plusLongDesAutres == 0 {
		return false, ""
	}

	const facteur = 1.6
	const ecartMini = 12
	if float64(bonne) > float64(plusLongDesAutres)*facteur && bonne-plusLongDesAutres > ecartMini {
		return true, fmt.Sprintf(
			"la bonne réponse fait %d caractères contre %d au plus long des autres — devinable sans savoir (§3.6)",
			bonne, plusLongDesAutres)
	}
	return false, ""
}

// --- Cohérence entre les données et le code --------------------------------

// TestDonneesChargeablesParLeStore boucle la boucle : le contenu réel doit
// traverser le vrai NewStore sans erreur, et chaque niveau annoncé doit
// effectivement se charger. C'est ce que ferait un joueur qui ouvrirait TOUS
// les quiz de l'application, l'un après l'autre.
func TestDonneesChargeablesParLeStore(t *testing.T) {
	racine := dossierDonnees(t)

	store, err := NewStore(racine)
	if err != nil {
		t.Fatalf("le backend refuserait de démarrer sur ce contenu : %v", err)
	}

	themes := store.Themes()
	if len(themes) == 0 {
		t.Fatal("aucun thème chargé")
	}

	for _, theme := range themes {
		for _, niveau := range theme.Niveaux {
			questions, ok := store.Questions(theme.ID, niveau)
			if !ok {
				t.Errorf("%s/%s : annoncé par l'API mais impossible à charger", theme.ID, niveau)
				continue
			}
			if len(questions) == 0 {
				t.Errorf("%s/%s : chargé mais vide → le joueur verrait un quiz sans question",
					theme.ID, niveau)
			}
		}
	}

	// Le JSON sortant doit être encodable : un caractère invalide ferait échouer
	// la réponse HTTP en plein milieu, après l'envoi du code 200.
	if _, err := json.Marshal(themes); err != nil {
		t.Errorf("les thèmes ne sont pas encodables en JSON : %v", err)
	}
}
