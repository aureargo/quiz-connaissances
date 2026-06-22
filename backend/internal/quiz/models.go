// Package quiz contient le "cœur métier" de l'application :
// les types de données (ce fichier) et leur chargement depuis le disque (store.go).
//
// En Go, un "package" est un regroupement de fichiers .go partageant le même
// espace de noms. Tous les fichiers de ce dossier commencent par `package quiz`.
package quiz

// Theme représente une catégorie de questions (ex : "Java", "Échecs").
//
// Les "tags" entre backquotes (`json:"..."`) indiquent à l'encodeur JSON
// comment nommer chaque champ. En Go, un champ qui commence par une MAJUSCULE
// est "exporté" (visible en dehors du package) ; c'est obligatoire pour que
// le package encoding/json puisse le lire/écrire.
type Theme struct {
	ID          string `json:"id"`          // identifiant unique, ex: "java"
	Nom         string `json:"nom"`         // libellé affiché, ex: "Java"
	Emoji       string `json:"emoji"`       // petit emoji décoratif, ex: "☕"
	Categorie   string `json:"categorie"`   // regroupement, ex: "Programmation"
	Description string `json:"description"` // courte phrase de présentation

	// Niveaux n'existe PAS dans le fichier themes.json : il est déduit au
	// chargement (voir store.go) des seuls NOMS de fichiers présents dans
	// data/questions/<id>/ (ex : "facile.json" → "facile"), sans lire leur
	// contenu. On y ajoute le niveau synthétique "tous". Le frontend l'affiche
	// comme un sélecteur de difficulté.
	//
	// `omitempty` : ce champ est OMIS du JSON quand la slice est vide. C'est
	// volontaire : la route /api/themes renvoie une liste LÉGÈRE (sans niveaux,
	// pour ne pas tout charger sur la page d'accueil), tandis que la route
	// /api/themes/{id} renvoie UN thème complet, niveaux compris.
	Niveaux []string `json:"niveaux,omitempty"`
}

// Question représente une question à choix multiples.
type Question struct {
	ID     string   `json:"id"`     // identifiant unique de la question
	Enonce string   `json:"enonce"` // le texte de la question
	Choix  []string `json:"choix"`  // les 4 réponses possibles (une slice de chaînes)

	// BonneReponse est l'INDICE (commençant à 0) de la bonne réponse dans Choix.
	// Exemple : si Choix = ["A","B","C","D"] et BonneReponse = 2, la bonne
	// réponse est "C".
	//
	// ⚠️ Pédagogie : on envoie ici la bonne réponse au frontend pour rester
	// simple. Dans une vraie app d'examen, on la garderait secrète côté serveur.
	BonneReponse int    `json:"bonneReponse"`
	Explication  string `json:"explication"` // texte affiché APRÈS avoir répondu
}
