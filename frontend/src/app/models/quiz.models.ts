/* =========================================================================
   quiz.models.ts — les "modèles" de données côté frontend.

   Ce sont des INTERFACES TypeScript : elles décrivent la FORME des objets JSON
   renvoyés par l'API Go. Elles n'existent qu'à la compilation (aucun code
   généré) mais nous offrent l'autocomplétion et la vérification des types.

   ⚠️ Ces interfaces doivent rester synchronisées avec les structs Go
   (backend/internal/quiz/models.go). Les noms de champs correspondent aux
   tags `json:"..."` côté Go.
   ========================================================================= */

// Un thème = une catégorie de questions (ex : "Java", "Échecs").
export interface Theme {
  id: string;            // identifiant unique, ex: "java"
  nom: string;           // libellé affiché, ex: "Java"
  emoji: string;         // emoji décoratif
  categorie: string;     // regroupement, ex: "Programmation"
  description: string;   // courte phrase de présentation

  // niveaux est OPTIONNEL : la liste légère (GET /api/themes, pour l'accueil) ne
  // le renvoie pas, alors que le détail d'un thème (GET /api/themes/{id}) si.
  // C'est juste la liste des NOMS de niveaux disponibles (ex : ["facile",
  // "moyen", "expert", "tous"]) ; le nombre de questions n'est pas calculé.
  niveaux?: string[];
}

// Une question à choix multiples, telle que renvoyée par l'API.
export interface Question {
  id: string;           // identifiant unique de la question
  enonce: string;       // le texte de la question
  choix: string[];      // les 4 réponses possibles
  bonneReponse: number; // INDICE (à partir de 0) de la bonne réponse dans `choix`
  explication: string;  // texte affiché après avoir répondu
}
