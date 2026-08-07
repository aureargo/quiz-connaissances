// Le package "main" est spécial en Go : c'est le seul qui produit un
// programme EXÉCUTABLE. Il doit contenir une fonction main(), point d'entrée
// du programme.
package main

import (
	"flag" // lecture des options de ligne de commande (-data, -addr)
	"log"
	"net/http"
	"os" // os.Getenv : lecture des variables d'environnement

	"quizconnaissances/internal/api"
	"quizconnaissances/internal/quiz"
)

// Valeurs par défaut. Le dossier de données vit à la RACINE du dépôt (et non
// dans backend/) : vu qu'on lance le serveur depuis backend/, on remonte d'un
// cran avec "../data".
const (
	adresseParDefaut = ":8080"   // ":8080" = écoute sur le port 8080, toutes interfaces
	donneesParDefaut = "../data" // dossier contenant themes.json et questions/
)

func main() {
	// --- Configuration ---
	//
	// Trois sources, de la moins prioritaire à la plus prioritaire :
	//   1. les constantes ci-dessus (défaut) ;
	//   2. les variables d'environnement QUIZ_DATA_DIR / QUIZ_ADDR ;
	//   3. les options de ligne de commande -data / -addr.
	//
	// Pourquoi ne PAS coder le chemin en dur ? Parce qu'un chemin relatif dépend
	// du répertoire depuis lequel on lance le programme : `go run ./backend`
	// depuis la racine et `go run .` depuis backend/ ne voient pas le même
	// "../data". Une option explicite lève l'ambiguïté.
	//
	// flag.String renvoie un POINTEUR vers la valeur ; elle n'est réellement
	// remplie qu'après l'appel à flag.Parse(), d'où l'étoile au moment de lire.
	dossierDonnees := flag.String("data", valeurEnv("QUIZ_DATA_DIR", donneesParDefaut),
		"chemin du dossier de données (themes.json + questions/)")
	adresse := flag.String("addr", valeurEnv("QUIZ_ADDR", adresseParDefaut),
		"adresse d'écoute du serveur HTTP")
	flag.Parse()

	// 1) Charger les métadonnées (thèmes + niveaux disponibles).
	store, err := quiz.NewStore(*dossierDonnees)
	if err != nil {
		// log.Fatalf affiche le message PUIS arrête le programme (code 1).
		// Inutile de continuer si les données n'ont pas pu être lues.
		log.Fatalf("impossible de charger les données depuis %q : %v\n"+
			"→ vérifie le chemin, ou précise-le avec -data <dossier>",
			*dossierDonnees, err)
	}
	log.Printf("✅ %d thème(s) chargé(s) depuis le dossier %q", len(store.Themes()), *dossierDonnees)

	// 2) Construire le serveur et son routeur.
	serveur := api.NewServer(store)

	// 3) Démarrer le serveur HTTP. ListenAndServe est BLOQUANT : il tourne
	//    tant que le serveur fonctionne. S'il renvoie une erreur, c'est qu'il
	//    s'est arrêté anormalement.
	log.Printf("🚀 serveur démarré sur http://localhost%s", *adresse)
	if err := http.ListenAndServe(*adresse, serveur.Routes()); err != nil {
		log.Fatalf("erreur du serveur : %v", err)
	}
}

// valeurEnv renvoie la variable d'environnement `cle` si elle est définie et
// non vide, sinon `defaut`. Pratique pour chaîner les sources de configuration.
func valeurEnv(cle, defaut string) string {
	if v := os.Getenv(cle); v != "" {
		return v
	}
	return defaut
}
