# 🧠 Quiz Connaissances

Une application de **quiz de connaissances** à choix multiples (4 réponses possibles),
organisée par **thèmes** (Java, C++, Angular, Go, Rust, SQL, Crypto, Mangas, Animés,
Échecs, Jeux vidéo, Pays...).

Projet pédagogique pour apprendre **Go** (backend) et **Angular** (frontend).

## ✨ Fonctionnalités

- ❓ Questions à **4 choix**, une seule bonne réponse.
- 💡 **Explication** affichée après chaque réponse.
- 🔁 Une question **mal répondue revient dans la file** jusqu'à être réussie.
- 🎲 **Ordre des questions aléatoire** à chaque partie.
- 🔀 **Ordre des 4 choix mélangé** à chaque affichage (pas de mémorisation de pattern).
- 🗂️ Questionnaires **extensibles** via de simples fichiers JSON.

## 🏗️ Architecture

```
quiz-connaissances/
├── data/         → LE CONTENU : thèmes et questions en JSON (la « base de données »)
├── backend/      → API REST en Go (sert les thèmes et les questions)
├── frontend/     → Application Angular (interface + logique de quiz)
└── CLAUDE.md     → Documentation technique détaillée
```

Le **backend Go** charge les questions depuis `data/` et les expose via une API REST.
Le **frontend Angular** gère la session de quiz (mélange, file d'attente des erreurs,
progression).

Le dossier `data/` est **volontairement à part** : le contenu évolue à un rythme
différent du code, et le backend est un moteur générique qui sait servir n'importe
quel jeu de questions bien formé. Voir [`data/QUESTIONNAIRE_TEMPLATE.md`](data/QUESTIONNAIRE_TEMPLATE.md)
pour en ajouter.

## 🚀 Démarrage rapide

### Backend (Go)
```bash
cd backend
go run .
# → écoute sur http://localhost:8080, données lues dans ../data
```

> Options : `-data <dossier>` et `-addr <hôte:port>` (ou les variables
> d'environnement `QUIZ_DATA_DIR` / `QUIZ_ADDR`) pour changer le dossier de
> données ou le port sans toucher au code.

### Frontend (Angular)
```bash
cd frontend
npm install
npm start
# → ouvre http://localhost:4200
```

## 📚 Pour apprendre

Le code est **abondamment commenté en français**. Consulte `CLAUDE.md` pour la
documentation technique, et l'historique git (`git log`) pour suivre les phases
de développement étape par étape.
