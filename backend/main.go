// Le package "main" est spécial en Go : c'est le seul qui produit un
// programme EXÉCUTABLE. Il doit contenir une fonction main(), point d'entrée
// du programme.
package main

import (
	"log"
	"net/http"

	"quizconnaissances/internal/api"
	"quizconnaissances/internal/quiz"
)

func main() {
	// Constantes de configuration. Pour un vrai projet, on les lirait plutôt
	// depuis des variables d'environnement ou un fichier de config.
	const adresse = ":8080"     // ":8080" = écoute sur le port 8080, toutes interfaces
	const dossierDonnees = "data" // dossier contenant themes.json et questions/

	// 1) Charger les données (thèmes + questions) en mémoire.
	store, err := quiz.NewStore(dossierDonnees)
	if err != nil {
		// log.Fatalf affiche le message PUIS arrête le programme (code 1).
		// Inutile de continuer si les données n'ont pas pu être lues.
		log.Fatalf("impossible de charger les données : %v", err)
	}
	log.Printf("✅ %d thème(s) chargé(s) depuis le dossier %q", len(store.Themes()), dossierDonnees)

	// 2) Construire le serveur et son routeur.
	serveur := api.NewServer(store)

	// 3) Démarrer le serveur HTTP. ListenAndServe est BLOQUANT : il tourne
	//    tant que le serveur fonctionne. S'il renvoie une erreur, c'est qu'il
	//    s'est arrêté anormalement.
	log.Printf("🚀 serveur démarré sur http://localhost%s", adresse)
	if err := http.ListenAndServe(adresse, serveur.Routes()); err != nil {
		log.Fatalf("erreur du serveur : %v", err)
	}
}
