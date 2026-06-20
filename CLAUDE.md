# CLAUDE.md

Ce fichier guide les futures sessions de Claude Code (et toi !) qui travaillent sur ce projet.

> ⚠️ Document **vivant** : il est complété à chaque phase de développement.

## 🎯 Objectif du projet

Application de quiz de connaissances à choix multiples, **par thèmes**.
Projet **pédagogique** : apprendre Go (backend) et Angular (frontend).
Tout le code et les commentaires sont **en français**.

## 🏗️ Architecture générale

Deux applications indépendantes qui communiquent par une API REST HTTP :

- **`backend/`** — serveur **Go**. Charge les questions depuis des fichiers JSON
  (`backend/data/`) et les expose via une API REST. Volontairement **sans état**
  (stateless) : il ne mémorise pas la progression d'un joueur.
- **`frontend/`** — application **Angular**. Affiche l'interface et contient toute la
  **logique de session de quiz** : mélange aléatoire, file d'attente des questions ratées,
  suivi de la progression.

### Pourquoi ce découpage ?
La logique « une question ratée revient plus tard » est propre à *une partie en cours*
(un état de session côté joueur). Elle est donc naturellement gérée côté **frontend**.
Le backend reste simple : il sert des données. C'est un bon découpage pour apprendre
les deux technos sans surcharge inutile.

> Note : pour une vraie app d'examen, on validerait les réponses **côté serveur** pour ne
> pas exposer les bonnes réponses au client. Ici, on privilégie la simplicité pédagogique.

## 📁 Structure des fichiers

```
backend/
  go.mod                      → module Go + version
  main.go                     → point d'entrée : configure et démarre le serveur HTTP
  internal/quiz/models.go     → types de données (Theme, Question...)
  internal/quiz/store.go      → chargement des fichiers JSON en mémoire
  internal/api/handlers.go    → handlers HTTP (les endpoints de l'API)
  data/themes.json            → liste des thèmes (métadonnées)
  data/questions/*.json       → un fichier de questions par thème
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
cd backend && go run .          # lancer le serveur de dev
cd backend && go build -o quiz-server.exe .   # compiler un binaire

# Frontend
cd frontend && npm install      # installer les dépendances (1re fois)
cd frontend && npm start        # lancer le serveur de dev Angular (port 4200)
cd frontend && npm run build    # build de production
```

## ➕ Ajouter un questionnaire

1. Créer `backend/data/questions/<id>.json` (voir le format dans un fichier existant).
2. Ajouter une entrée correspondante dans `backend/data/themes.json`.
3. Redémarrer le backend. C'est tout — aucune recompilation de code nécessaire.

## 🌳 Convention git

Un commit par **phase** de développement, pour pouvoir étudier l'évolution du projet :

- Phase 0 — structure du projet, git, documentation initiale
- Phase 1 — backend Go (API + données)
- Phase 2 — frontend Angular (scaffold + sélection de thème)
- Phase 3 — moteur de quiz (mélange + file des erreurs)
- Phase 4 — design & finitions
- Phase 5 — enrichissement des questionnaires + doc finale
