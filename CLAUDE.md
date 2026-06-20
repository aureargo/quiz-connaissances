# CLAUDE.md

Ce fichier guide les futures sessions de Claude Code (et toi !) qui travaillent sur ce projet.

## 🎯 Objectif du projet

Application de quiz de connaissances à choix multiples, **par thèmes**.
Projet **pédagogique** : apprendre Go (backend) et Angular (frontend).
Tout le code et les commentaires sont **en français**.

### Règles de jeu implémentées
- 4 choix par question, une seule bonne réponse, + une **explication** après coup.
- Une question **ratée revient dans la file** jusqu'à être réussie.
- **Ordre des questions** mélangé à chaque partie.
- **Ordre des 4 choix** re-mélangé à chaque affichage d'une question.

## 🏗️ Architecture générale

Deux applications indépendantes qui communiquent par une API REST HTTP :

- **`backend/`** — serveur **Go**. Charge les questions depuis des fichiers JSON
  (`backend/data/`) et les expose via une API REST. Volontairement **sans état**
  (stateless) : il ne mémorise pas la progression d'un joueur.
- **`frontend/`** — application **Angular**. Affiche l'interface et contient toute la
  **logique de session de quiz** (mélange, file des erreurs, progression).

### Pourquoi ce découpage ?
La logique « une question ratée revient plus tard » est propre à *une partie en cours*
(un état de session côté joueur). Elle est donc naturellement gérée côté **frontend**
(classe `MoteurQuiz`). Le backend reste simple : il sert des données.

> Note : pour une vraie app d'examen, on validerait les réponses **côté serveur** pour
> ne pas exposer les bonnes réponses au client. Ici, on privilégie la simplicité.

## 📁 Structure des fichiers

### Backend (Go)
```
backend/
  go.mod                      → module Go + version
  main.go                     → point d'entrée : configure et démarre le serveur
  internal/quiz/models.go     → types de données (Theme, Question)
  internal/quiz/store.go      → chargement des fichiers JSON en mémoire
  internal/api/handlers.go    → handlers HTTP, routeur, middleware CORS
  data/themes.json            → liste des thèmes (métadonnées)
  data/questions/*.json       → un fichier de questions par thème
```

### Frontend (Angular)
```
frontend/src/app/
  app.ts / app.html / app.scss      → composant racine (layout : en-tête + outlet)
  app.config.ts                     → providers globaux (router, HttpClient)
  app.routes.ts                     → routes (lazy loading)
  models/quiz.models.ts             → interfaces TS (miroir des structs Go)
  services/quiz-api.ts              → service d'appels REST (HttpClient)
  services/moteur-quiz.ts           → CŒUR : logique de partie (file, mélange…)
  services/moteur-quiz.spec.ts      → tests unitaires du moteur (vitest)
  utils/aleatoire.ts                → mélange Fisher-Yates
  pages/theme-selection/            → page d'accueil : choix du thème
  pages/quiz/                       → page de jeu (pilote le MoteurQuiz)
```

## 🔌 API REST (backend Go)

| Méthode | Route                          | Description                          |
|---------|--------------------------------|--------------------------------------|
| GET     | `/api/themes`                  | Liste des thèmes (+ nb de questions) |
| GET     | `/api/themes/{id}/questions`   | Questions d'un thème donné           |

Le serveur écoute sur **`http://localhost:8080`** et autorise le CORS pour le
frontend Angular (`http://localhost:4200`).

## ▶️ Commandes utiles

```bash
# Backend
cd backend && go run .                  # lancer le serveur de dev
cd backend && go build -o quiz-server.exe .   # compiler un binaire
cd backend && go vet ./...              # analyse statique

# Frontend
cd frontend && npm install              # installer les dépendances (1re fois)
cd frontend && npm start                # serveur de dev Angular (port 4200)
cd frontend && npm run build            # build de production
cd frontend && npm test -- --watch=false  # lancer les tests unitaires une fois
```

> ⚠️ Il faut **lancer les DEUX** serveurs (Go puis Angular) pour utiliser l'app.

## ➕ Ajouter un questionnaire

1. Créer `backend/data/questions/<id>.json` (voir le format dans un fichier existant :
   `enonce`, `choix` [4], `bonneReponse` [index 0-3], `explication`).
2. Ajouter l'entrée correspondante dans `backend/data/themes.json`
   (`id`, `nom`, `emoji`, `categorie`, `description`).
3. Redémarrer le backend. Aucune recompilation de code nécessaire.

> ⚠️ Chaque thème listé dans `themes.json` DOIT avoir son fichier de questions,
> sinon le backend refuse de démarrer (c'est volontaire : fail-fast).

## 🗂️ Thèmes actuels (12, 6 par questionnaire)

- **Programmation** : Go, Angular, Java, C++, Rust
- **Bases de données** : SQL
- **Crypto & Web3** : Crypto
- **Culture geek** : Mangas, Animés, Jeux vidéo
- **Jeux & stratégie** : Échecs
- **Culture générale** : Pays

## 📚 Concepts à retenir (pour apprendre)

**Côté Go** : packages & visibilité (majuscule = exporté), structs + tags JSON,
`(valeur, error)` comme convention d'erreur, `defer`, middleware HTTP, routeur
`net/http` (Go 1.22+) avec `{id}` et `r.PathValue`.

**Côté Angular** : composants **standalone**, **signals** (`signal`, `computed`),
nouveau control-flow (`@if`, `@for`, `@let`), **injection** (`inject()`), `HttpClient`
+ Observables (`forkJoin`), routing avec **lazy loading** et binding des paramètres
d'URL sur les `input()`, encapsulation des styles par composant.

## 🌳 Historique git (phases de développement)

1. **Phase 0** — structure du projet, git, documentation initiale
2. **Phase 1** — backend Go (API REST + données)
3. **Phase 2** — frontend Angular (scaffold + sélection de thème)
4. **Phase 3** — moteur de quiz (mélange + file des erreurs) + tests
5. **Phase 4** — design & polish (animations, clavier, accessibilité)
6. **Phase 5** — enrichissement des questionnaires + doc finale

Utilise `git log --oneline` et `git show <commit>` pour étudier chaque étape.
