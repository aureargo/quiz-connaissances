# CLAUDE.md

Ce fichier guide les futures sessions de Claude Code (et toi !) qui travaillent sur ce projet.

## 🎯 Objectif du projet

Application de quiz de connaissances à choix multiples, **par thèmes**.
Projet **pédagogique** : apprendre Go (backend) et Angular (frontend).
Tout le code et les commentaires sont **en français**.

### Règles de jeu implémentées
- 4 choix par question, une seule bonne réponse, + une **explication** après coup.
- Une question **ratée revient dans la file** jusqu'à être réussie, **réinsérée
  à une position aléatoire** parmi les questions restantes (et pas juste après,
  sauf si c'est la seule question encore en jeu).
- **Ordre des questions** mélangé à chaque partie.
- **Ordre des 4 choix** re-mélangé à chaque affichage d'une question.

## 🏗️ Architecture générale

Deux applications indépendantes qui communiquent par une API REST HTTP, plus un
dossier de **contenu** qui n'appartient à aucune des deux :

- **`data/`** — le **contenu** du quiz (thèmes + questions), en JSON. C'est la
  « base de données » du projet : elle vit **à la racine**, hors de `backend/` et
  de `frontend/`. Voir « Pourquoi `data/` est à part ? » ci-dessous.
- **`backend/`** — serveur **Go**. Charge les questions depuis `data/` et les
  expose via une API REST. Volontairement **sans état** (stateless) : il ne
  mémorise pas la progression d'un joueur.
- **`frontend/`** — application **Angular**. Affiche l'interface et contient toute la
  **logique de session de quiz** (mélange, file des erreurs, progression).

### Pourquoi `data/` est à part ?

Le contenu et le code évoluent à des **rythmes différents** : ajouter 40 thèmes ne
change pas une ligne de Go. Les garder mélangés donnait un `backend/` composé à
95 % de JSON, et des commits « +1230 questions » qui touchaient le dossier du
serveur sans rien y modifier.

Séparés, chacun a son rôle : `backend/` est un **moteur de quiz générique** (il
sert n'importe quel jeu de questions bien formé), `data/` est **un** jeu de
questions, avec ses règles propres — documentées dans
[`QUESTIONNAIRE_TEMPLATE.md`](data/QUESTIONNAIRE_TEMPLATE.md) et **vérifiées
automatiquement** par le validateur (`go test ./internal/quiz -run TestDonnees`).

> Le chemin du dossier n'est plus codé en dur : il se règle avec l'option
> `-data <dossier>` ou la variable d'environnement `QUIZ_DATA_DIR`
> (défaut : `../data`, puisqu'on lance le serveur depuis `backend/`).

### Pourquoi ce découpage ?
La logique « une question ratée revient plus tard » est propre à *une partie en cours*
(un état de session côté joueur). Elle est donc naturellement gérée côté **frontend**
(classe `MoteurQuiz`). Le backend reste simple : il sert des données.

> Note : pour une vraie app d'examen, on validerait les réponses **côté serveur** pour
> ne pas exposer les bonnes réponses au client. Ici, on privilégie la simplicité.

### Parcours de navigation (3 écrans)

Le joueur traverse **trois pages**, et les données sont chargées **progressivement**
(une page = un appel API ciblé), pour ne jamais tout charger d'un coup :

1. **Accueil** (`/`, `theme-selection`) — uniquement les **tuiles de thèmes**, chacune
   étant un lien. Appelle `GET /api/themes` (liste **légère**, sans les niveaux).
2. **Sélection du niveau** (`/themes/:id`, `niveau-selection`) — atteinte au clic sur
   une tuile. C'est **seulement ici** qu'on demande les niveaux de **ce** thème via
   `GET /api/themes/{id}`, puis qu'on affiche les boutons facile/moyen/expert/tous.
3. **Quiz** (`/quiz/:themeId/:niveau`, `quiz`) — charge les questions du niveau choisi.

> Pourquoi cet « étage » intermédiaire ? (1) la tuile mène quelque part — finis les
> « 4 petits boutons » sous chaque tuile que l'utilisateur ne voyait pas ; (2) l'accueil
> ne charge plus les niveaux de **tous** les thèmes, donc moins de charge au démarrage.

> 🪶 **Chargement paresseux des questions.** Le serveur ne lit AUCUN fichier de
> questions au démarrage : il se contente de scanner les **noms** de fichiers
> (`data/questions/<id>/*.json`) pour connaître les niveaux disponibles. Le contenu
> d'un niveau n'est lu (et mis en **cache** mémoire) qu'au premier `GET .../questions/{niveau}`,
> c.-à-d. quand le joueur lance réellement ce quiz. Conséquence assumée : on n'affiche
> **plus le nombre de questions** par niveau, et un JSON mal formé n'est détecté qu'à
> l'ouverture du quiz concerné (et non plus au boot).

## 📁 Structure des fichiers

### Données (`data/`, à la racine)
```
data/
  themes.json                        → liste des thèmes (métadonnées)
  categories.json                    → ordre d'affichage des catégories
  questions/<id>/<niveau>.json       → un fichier par thème ET par niveau
  QUESTIONNAIRE_TEMPLATE.md          → le format + les règles de contenu
```

### Backend (Go)
```
backend/
  go.mod                      → module Go + version
  main.go                     → point d'entrée : options (-data, -addr) et démarrage
  internal/quiz/models.go     → types de données (Theme, Question)
  internal/quiz/store.go      → scan des niveaux au boot + lecture paresseuse (cache) des questions
  internal/quiz/store_test.go → tests unitaires du store (données synthétiques)
  internal/quiz/donnees_test.go → VALIDATEUR : vérifie le vrai contenu de data/
  internal/api/handlers.go    → handlers HTTP, routeur, middleware CORS
  internal/api/handlers_test.go → tests des routes HTTP
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
  pages/theme-selection/            → page d'accueil : tuiles de thèmes (lien)
  pages/niveau-selection/           → sous-page : choix du niveau d'un thème
  pages/quiz/                       → page de jeu (pilote le MoteurQuiz)
```

## 🔌 API REST (backend Go)

| Méthode | Route                                  | Description                                         |
|---------|----------------------------------------|-----------------------------------------------------|
| GET     | `/api/themes`                          | Liste **légère** des thèmes (métadonnées, sans niveaux) |
| GET     | `/api/themes/{id}`                      | **Un** thème + ses niveaux (chargé à la demande)    |
| GET     | `/api/themes/{id}/questions/{niveau}`   | Questions d'un thème à un niveau donné              |

> Le champ `niveaux` d'un thème est la simple **liste des noms** de niveaux
> disponibles (ex : `["facile","moyen","expert","tous"]`), déduite des noms de
> fichiers — **sans** nombre de questions. Il n'est renvoyé QUE par
> `/api/themes/{id}` (tag JSON `omitempty`), pas par la liste légère.
> La route `.../questions/{niveau}` lit le fichier à la demande puis le met en cache.

Le serveur écoute sur **`http://localhost:8080`** et autorise le CORS pour le
frontend Angular (`http://localhost:4200`).

## ▶️ Commandes utiles

### Lanceurs « tout-en-un » (pour démarrer l'app sans rien connaître)

Des scripts à la racine démarrent **les deux serveurs** d'un coup, vérifient que
Go et Node.js sont installés, puis ouvrent le navigateur automatiquement dès que
le frontend répond. Un lanceur par famille d'OS :

| OS            | Lanceur          | Comment                                                |
|---------------|------------------|--------------------------------------------------------|
| Windows       | `quiz.bat`       | double-clic (il appelle `quiz.ps1`)                    |
| Linux / macOS | `quiz.sh`        | `chmod +x quiz.sh` (1re fois) puis `./quiz.sh`         |

> Pourquoi plusieurs fichiers ? `.bat`/`.ps1` ne tournent que sous Windows (les
> popups dépendent de `System.Windows.Forms`) ; Bash n'est pas natif sous Windows.
> Aucun format de script n'est universel, d'où **un lanceur par OS**. Les deux
> partagent la même logique (vérif. des outils → lancement → attente du port 4200
> en IPv4 **et** IPv6 → ouverture du navigateur).

### Commandes manuelles (par serveur)

```bash
# Backend
cd backend && go run .                  # lancer le serveur de dev (data = ../data)
cd backend && go run . -data ../data -addr :8080   # idem, options explicites
cd backend && go build -o quiz-server.exe .   # compiler un binaire
cd backend && go vet ./...              # analyse statique
cd backend && go test ./...             # tests unitaires + VALIDATION du contenu

# Frontend
cd frontend && npm install              # installer les dépendances (1re fois)
cd frontend && npm start                # serveur de dev Angular (port 4200)
cd frontend && npm run build            # build de production
cd frontend && npm test -- --watch=false  # lancer les tests unitaires une fois
```

> 🔧 **Options du backend** : `-data <dossier>` (défaut `../data`) et `-addr <hôte:port>`
> (défaut `:8080`). Équivalents en variables d'environnement : `QUIZ_DATA_DIR` et
> `QUIZ_ADDR`. Priorité : option de ligne de commande > variable d'env > défaut.
> Utile pour lancer le serveur depuis un autre répertoire, ou pour tester un jeu
> de données alternatif sans toucher au code.

> ⚠️ Il faut **lancer les DEUX** serveurs (Go puis Angular) pour utiliser l'app.

## ➕ Ajouter un questionnaire

> 📄 **Référence complète** : [`data/QUESTIONNAIRE_TEMPLATE.md`](data/QUESTIONNAIRE_TEMPLATE.md)
> — format JSON, convention des IDs, règle anti-déduction, checklist. **Lire ce
> fichier avant de générer ou modifier des questions.**

Étapes résumées :

1. Créer un dossier `data/questions/<id>/` puis y déposer **un fichier `.json`
   par niveau**. Le **nom du fichier = le nom du niveau**, et il est **libre** :
   `facile.json`, `debutant.json`, `histoire.json`… (le nom `tous` est réservé).
   Le **nombre de niveaux est libre** lui aussi. Format d'un fichier : tableau d'objets
   `{ id, enonce, choix, bonneReponse, explication }`.
2. Ajouter l'entrée correspondante dans `data/themes.json`
   (`id`, `nom`, `emoji`, `categorie`, `description`). Si la `categorie` est
   nouvelle, l'ajouter aussi à `data/categories.json` (qui fixe l'ordre d'affichage).
3. **Vérifier** : `cd backend && go test ./internal/quiz -run TestDonnees`. Le
   validateur contrôle les règles dures ET les règles de qualité (§3 du template).
4. Redémarrer le backend. Aucune recompilation de code nécessaire.

> ⚠️ Chaque thème listé dans `themes.json` DOIT avoir au moins un fichier de
> questions (`data/questions/<id>/<niveau>.json`), sinon le backend refuse de
> démarrer (c'est volontaire : fail-fast).

> 💡 Le niveau synthétique **`tous`** (qui agrège **tous** les niveaux réels du
> thème, fichiers ajoutés compris) est généré automatiquement dès qu'un thème a
> **au moins 2** niveaux réels. Les noms connus `facile`/`moyen`/`expert` ont un
> ordre d'affichage et une couleur dédiés ; un niveau au nom libre s'affiche après
> eux, avec une couleur neutre (cf. `QUESTIONNAIRE_TEMPLATE.md`, §3.2).

> ⚠️ **Équilibre des choix (anti-déduction)** : les 4 `choix` doivent avoir une
> longueur et un niveau de détail comparables — voir le template pour les détails
> et contre-exemples.

## 🗂️ Thèmes actuels (59)

État du **contenu** actuel (ce ne sont pas des contraintes du code, cf.
`QUESTIONNAIRE_TEMPLATE.md`) : chaque thème a aujourd'hui trois fichiers
(`facile.json`, `moyen.json`, `expert.json`) de 10 questions, soit 30 par thème
(plus le niveau synthétique **`tous`** = 30). Total : **1770 questions**.

- **Programmation** : Go, Python, JavaScript, TypeScript, Angular, Java, Kotlin, C, C++, Rust, Git, Lignes de commande Linux, HTML, CSS, Architecture logicielle, Algorithmes & structures de données
- **DevOps & Conteneurs** : Docker, Kubernetes
- **Matériel informatique** : Processeurs (CPU), Cartes graphiques (GPU)
- **Bases de données** : SQL, NoSQL
- **Crypto & Web3** : Crypto
- **Intelligence artificielle** : IA (généralités), LLM (connaissances arrêtées à 2026)
- **Cybersécurité** : Cybersécurité
- **Réseaux & Internet** : Réseaux
- **Culture geek** : Mangas, Animés, Jeux vidéo, Bande dessinée
- **Jeux & stratégie** : Échecs
- **Sport** : Cyclisme
- **Culture générale** : Pays, Géographie, Villes de France, Mythologie
- **Cinéma** : Films des années 1980, 1990, 2000, 2010
- **Littérature** : Science-fiction (littérature), Littérature d'aventure
- **Histoire** : Histoire de France, Antiquité
- **Sciences** : Mathématiques, Chimie, Biologie, Physique, Astronomie, Géologie, Écologie & climat
- **Économie & Finance** : Finance
- **Automobile** : Automobile
- **Systèmes d'exploitation** : Windows, Android
- **Langue française** : Grammaire, Conjugaison, Orthographe

## 📚 Concepts à retenir (pour apprendre)

**Côté Go** : packages & visibilité (majuscule = exporté), structs + tags JSON,
`(valeur, error)` comme convention d'erreur, `defer`, middleware HTTP, routeur
`net/http` (Go 1.22+) avec `{id}` et `r.PathValue`, lecture de dossier
(`os.ReadDir`), **chargement paresseux + cache** protégé par `sync.Mutex`
(les handlers HTTP tournant en parallèle, accéder à une map partagée sans verrou
= *data race*).

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
