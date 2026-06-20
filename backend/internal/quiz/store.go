package quiz

import (
	"encoding/json" // encodage/décodage JSON
	"fmt"           // formatage de chaînes (ici pour des messages d'erreur)
	"os"            // accès au système de fichiers
	"path/filepath" // construction de chemins (portable Windows/Linux/Mac)
)

// Store contient TOUTES les données du quiz, chargées une fois en mémoire au
// démarrage du serveur. C'est volontairement simple : pas de base de données,
// juste des fichiers JSON lus au lancement.
//
// Les champs commencent par une minuscule : ils sont donc "privés" au package
// quiz. On y accède de l'extérieur via les méthodes Themes() et Questions().
type Store struct {
	themes    []Theme               // la liste des thèmes, dans l'ordre du fichier
	questions map[string][]Question // questions indexées par ID de thème
}

// NewStore lit le dossier `dataDir` et construit un Store prêt à l'emploi.
//
// Convention Go : une fonction qui peut échouer renvoie une valeur ET une
// `error`. Si tout va bien, error vaut nil. On retourne ici un *Store (un
// POINTEUR vers Store) pour éviter de copier toute la structure.
func NewStore(dataDir string) (*Store, error) {
	// On initialise la map (sinon elle vaut nil et écrire dedans planterait).
	s := &Store{questions: make(map[string][]Question)}

	// 1) Charger la liste des thèmes depuis data/themes.json.
	themesPath := filepath.Join(dataDir, "themes.json")
	if err := lireJSON(themesPath, &s.themes); err != nil {
		// fmt.Errorf avec %w "emballe" l'erreur d'origine : on garde le détail
		// tout en ajoutant du contexte. C'est la bonne pratique en Go.
		return nil, fmt.Errorf("chargement des thèmes : %w", err)
	}

	// 2) Pour chaque thème, charger son fichier de questions associé.
	//    On itère avec l'index `i` pour pouvoir MODIFIER l'élément du slice
	//    (s.themes[i].NbQuestions). Une boucle `for _, t := range` donnerait
	//    une COPIE de chaque thème, et la modification serait perdue.
	for i := range s.themes {
		theme := &s.themes[i] // pointeur vers l'élément réel du slice
		questionsPath := filepath.Join(dataDir, "questions", theme.ID+".json")

		var questions []Question
		if err := lireJSON(questionsPath, &questions); err != nil {
			return nil, fmt.Errorf("chargement des questions du thème %q : %w", theme.ID, err)
		}

		s.questions[theme.ID] = questions
		theme.NbQuestions = len(questions) // on calcule le nombre de questions
	}

	return s, nil
}

// Themes renvoie la liste de tous les thèmes.
// (Méthode "getter" : permet de lire un champ privé depuis l'extérieur.)
func (s *Store) Themes() []Theme {
	return s.themes
}

// Questions renvoie les questions d'un thème.
// Le second retour (un booléen) vaut false si le thème n'existe pas.
// C'est l'idiome Go classique du "comma ok" : `valeur, ok := ...`.
func (s *Store) Questions(themeID string) ([]Question, bool) {
	questions, ok := s.questions[themeID]
	return questions, ok
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
