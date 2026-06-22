package quiz

import (
	"encoding/json" // encodage/décodage JSON
	"fmt"           // formatage de chaînes (ici pour des messages d'erreur)
	"log"           // journalisation (erreurs de lecture à la demande)
	"os"            // accès au système de fichiers
	"path/filepath" // construction de chemins (portable Windows/Linux/Mac)
	"slices"        // slices.Contains, slices.Index, slices.SortStableFunc
	"strings"       // manipulation de chaînes (suffixe ".json")
	"sync"          // sync.Mutex : protège le cache des accès concurrents
)

// Un "niveau" est tout simplement un fichier `<niveau>.json` déposé dans le
// dossier d'un thème. Les NOMS sont LIBRES : "facile", "debutant", "histoire",
// "saison-1"… On ne les fixe pas dans le code, et leur NOMBRE est libre aussi
// (un thème peut avoir 1, 2 ou 10 niveaux). Le backend se contente de découvrir
// les fichiers présents.
//
// ordrePrefere ne sert QU'À L'AFFICHAGE : il donne un ordre canonique aux noms
// courants (du plus simple au plus dur). Tout niveau absent de cette liste est
// classé APRÈS, par ordre alphabétique. Ce n'est donc PAS un filtre : retirer un
// nom d'ici ne l'empêche pas d'exister, ça change juste sa position d'affichage.
var ordrePrefere = []string{"facile", "moyen", "expert"}

// NiveauTous est un niveau SYNTHÉTIQUE : il ne correspond à aucun fichier, mais
// agrège toutes les questions des niveaux réels d'un thème (cf. méthode
// Questions). Il permet de jouer un quiz "général" mélangeant les difficultés.
// Ce nom est RÉSERVÉ : un éventuel fichier `tous.json` serait ignoré.
const NiveauTous = "tous"

// Store contient les données du quiz. Au démarrage, on ne charge QUE les
// métadonnées : la liste des thèmes (themes.json) et, pour chacun, la liste de
// ses niveaux DISPONIBLES — déduite des seuls NOMS de fichiers, SANS lire leur
// contenu. Les questions, elles, ne sont lues qu'au premier accès (méthode
// Questions) puis gardées en cache.
//
// Pourquoi ? On évite de charger en mémoire des centaines de questions dont le
// joueur ne jouera qu'une petite partie. C'est un compromis : on perd la
// validation du contenu au démarrage (un JSON mal formé n'est détecté qu'à
// l'ouverture du quiz concerné), au profit d'un boot plus léger.
type Store struct {
	dataDir string  // racine des données : sert à relire les fichiers à la demande
	themes  []Theme // la liste des thèmes, dans l'ordre du fichier

	// mu protège `cache`. Les handlers HTTP s'exécutent EN PARALLÈLE (une
	// goroutine par requête) : lire et écrire la map sans verrou provoquerait
	// une "data race". sync.Mutex garantit qu'un seul accès à la fois.
	mu sync.Mutex

	// cache des questions DÉJÀ lues : themeID -> niveau -> questions.
	// Rempli paresseusement (au 1er accès), pas au démarrage.
	cache map[string]map[string][]Question
}

// NewStore lit le dossier `dataDir` et construit un Store prêt à l'emploi.
//
// Convention Go : une fonction qui peut échouer renvoie une valeur ET une
// `error`. Si tout va bien, error vaut nil. On retourne ici un *Store (un
// POINTEUR vers Store) pour éviter de copier toute la structure.
func NewStore(dataDir string) (*Store, error) {
	s := &Store{
		dataDir: dataDir,
		cache:   make(map[string]map[string][]Question),
	}

	// 1) Charger la liste des thèmes depuis data/themes.json.
	themesPath := filepath.Join(dataDir, "themes.json")
	if err := lireJSON(themesPath, &s.themes); err != nil {
		// fmt.Errorf avec %w "emballe" l'erreur d'origine : on garde le détail
		// tout en ajoutant du contexte. C'est la bonne pratique en Go.
		return nil, fmt.Errorf("chargement des thèmes : %w", err)
	}

	// 2) Pour chaque thème, déterminer ses niveaux disponibles à partir des
	//    seuls NOMS de fichiers dans data/questions/<id>/ (aucune lecture de
	//    contenu). On itère avec l'index `i` pour MODIFIER l'élément réel du
	//    slice (une boucle `for _, t := range` n'en donnerait qu'une copie).
	for i := range s.themes {
		theme := &s.themes[i]

		reels, err := niveauxReels(dataDir, theme.ID)
		if err != nil {
			return nil, err
		}

		// Fail-fast : un thème déclaré dans themes.json mais sans AUCUN fichier
		// de questions est forcément une erreur de configuration. On refuse de
		// démarrer pour la rendre visible tout de suite (cf. CLAUDE.md).
		if len(reels) == 0 {
			return nil, fmt.Errorf(
				"le thème %q n'a aucun fichier de questions (attendu : data/questions/%s/<niveau>.json)",
				theme.ID, theme.ID,
			)
		}

		theme.Niveaux = reels

		// On ajoute le niveau synthétique "tous" UNIQUEMENT s'il y a au moins
		// deux niveaux réels (sinon "tous" ferait doublon avec l'unique niveau).
		if len(reels) >= 2 {
			theme.Niveaux = append(theme.Niveaux, NiveauTous)
		}
	}

	return s, nil
}

// niveauxReels scanne data/questions/<id>/ et renvoie les niveaux RÉELS présents :
// CHAQUE fichier `<niveau>.json` devient un niveau, quel que soit son nom. On NE
// LIT PAS le contenu des fichiers : seuls leurs noms nous intéressent ici. Le
// résultat est trié pour un affichage stable (voir trierNiveaux).
func niveauxReels(dataDir, themeID string) ([]string, error) {
	dossier := filepath.Join(dataDir, "questions", themeID)

	// os.ReadDir liste le contenu d'un dossier sans ouvrir les fichiers.
	entrees, err := os.ReadDir(dossier)
	if err != nil {
		return nil, fmt.Errorf("lecture du dossier de questions de %q : %w", themeID, err)
	}

	// Chaque fichier `.json` est un niveau (son nom, sans l'extension).
	var niveaux []string
	for _, e := range entrees {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		nom := strings.TrimSuffix(e.Name(), ".json")
		// "tous" est réservé au niveau synthétique : on ignore un éventuel
		// fichier tous.json pour éviter toute ambiguïté.
		if nom == NiveauTous {
			continue
		}
		niveaux = append(niveaux, nom)
	}

	trierNiveaux(niveaux)
	return niveaux, nil
}

// trierNiveaux ordonne les niveaux pour un affichage stable : d'abord les noms
// connus dans l'ordre de `ordrePrefere` (facile, moyen, expert), puis tous les
// autres par ordre alphabétique. C'est purement cosmétique.
func trierNiveaux(niveaux []string) {
	slices.SortStableFunc(niveaux, func(a, b string) int {
		ra, rb := rangNiveau(a), rangNiveau(b)
		if ra != rb {
			return ra - rb // rangs différents → on respecte l'ordre préféré
		}
		return strings.Compare(a, b) // même rang (deux inconnus) → alphabétique
	})
}

// rangNiveau renvoie la position d'un niveau dans `ordrePrefere`, ou une valeur
// plus grande que tous les rangs connus si le nom n'y figure pas (les niveaux
// "libres" sont donc affichés après les niveaux courants).
func rangNiveau(nom string) int {
	if i := slices.Index(ordrePrefere, nom); i != -1 {
		return i
	}
	return len(ordrePrefere)
}

// Themes renvoie la liste de tous les thèmes.
// (Méthode "getter" : permet de lire un champ privé depuis l'extérieur.)
func (s *Store) Themes() []Theme {
	return s.themes
}

// ResumesThemes renvoie la liste des thèmes en version LÉGÈRE : chaque thème
// SANS ses niveaux. C'est ce que sert la page d'accueil, qui n'a besoin que des
// métadonnées (nom, emoji, description) pour afficher les tuiles.
//
// On travaille sur une COPIE de chaque thème : mettre Niveaux à nil ne touche
// pas aux données réelles du store (utilisées par Theme() ci-dessous).
func (s *Store) ResumesThemes() []Theme {
	resumes := make([]Theme, len(s.themes))
	for i, theme := range s.themes {
		theme.Niveaux = nil // la copie locale perd ses niveaux ; le store, non
		resumes[i] = theme
	}
	return resumes
}

// Theme renvoie UN thème complet (avec la liste de ses niveaux) à partir de son
// id. Le booléen vaut false si l'id est inconnu (idiome "comma ok"). C'est ce
// que sert la sous-page de sélection du niveau.
func (s *Store) Theme(id string) (Theme, bool) {
	for _, theme := range s.themes {
		if theme.ID == id {
			return theme, true
		}
	}
	return Theme{}, false
}

// Questions renvoie les questions d'un thème POUR UN NIVEAU donné. Les fichiers
// ne sont LUS qu'au PREMIER accès, puis gardés en cache pour les parties
// suivantes. Le second retour (un booléen) vaut false si le couple
// (thème, niveau) n'existe pas, ou si la lecture du fichier a échoué.
func (s *Store) Questions(themeID, niveau string) ([]Question, bool) {
	// On valide d'abord le couple (thème, niveau) à partir des niveaux
	// découverts au démarrage — sans toucher au disque.
	theme, ok := s.Theme(themeID)
	if !ok || !slices.Contains(theme.Niveaux, niveau) {
		return nil, false
	}

	// Verrou : on s'apprête à lire/écrire le cache, partagé entre goroutines.
	// defer s.mu.Unlock() garantit le déverrouillage quel que soit le chemin de
	// sortie de la fonction.
	s.mu.Lock()
	defer s.mu.Unlock()

	// Déjà en cache ? On renvoie directement.
	if questions, ok := s.cache[themeID][niveau]; ok {
		return questions, true
	}

	// Sinon, on lit le(s) fichier(s) MAINTENANT (lecture paresseuse).
	questions, err := s.lireNiveau(themeID, niveau)
	if err != nil {
		// Fichier illisible ou JSON mal formé : on log côté serveur et on
		// signale "pas de questions" au handler (qui renverra une erreur HTTP).
		log.Printf("lecture des questions %q/%s : %v", themeID, niveau, err)
		return nil, false
	}

	// Mémorisation pour ne pas relire le disque aux prochaines parties.
	if s.cache[themeID] == nil {
		s.cache[themeID] = make(map[string][]Question)
	}
	s.cache[themeID][niveau] = questions
	return questions, true
}

// lireNiveau lit les questions d'un niveau DEPUIS LE DISQUE. Pour le niveau
// synthétique "tous", il lit et concatène tous les niveaux réels du thème, dans
// l'ordre renvoyé par niveauxReels. (Le mélange final est fait côté frontend.)
func (s *Store) lireNiveau(themeID, niveau string) ([]Question, error) {
	if niveau != NiveauTous {
		chemin := filepath.Join(s.dataDir, "questions", themeID, niveau+".json")
		var questions []Question
		if err := lireJSON(chemin, &questions); err != nil {
			return nil, err
		}
		return questions, nil
	}

	// "tous" : on relit la liste des niveaux réels (ceux qui existent vraiment
	// sur le disque) puis on concatène leurs questions.
	reels, err := niveauxReels(s.dataDir, themeID)
	if err != nil {
		return nil, err
	}
	var toutes []Question
	for _, n := range reels {
		chemin := filepath.Join(s.dataDir, "questions", themeID, n+".json")
		var questions []Question
		if err := lireJSON(chemin, &questions); err != nil {
			return nil, err
		}
		toutes = append(toutes, questions...)
	}
	return toutes, nil
}

// lireJSON est une petite fonction utilitaire (privée) : elle lit un fichier
// puis décode son contenu JSON dans `dest`.
//
// `dest any` accepte n'importe quel type (any = alias de interface{}). L'appelant
// passe un POINTEUR (&maVariable) pour que json.Unmarshal puisse y écrire.
func lireJSON(chemin string, dest any) error {
	data, err := os.ReadFile(chemin)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}
