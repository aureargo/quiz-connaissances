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

	// NbQuestions n'existe PAS dans le fichier themes.json : il est calculé
	// au chargement (voir store.go). On l'ajoute quand même au JSON renvoyé
	// au frontend, qui aime bien afficher "12 questions".
	NbQuestions int `json:"nbQuestions"`
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
