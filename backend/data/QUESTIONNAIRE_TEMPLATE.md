# Template — Créer un questionnaire

Ce fichier sert de référence pour ajouter un thème ou un niveau de questions.

Il s'adresse à **deux publics**, d'où deux sections bien distinctes :

- **§2 — Règles DURES** : ce que le **code** impose. Ne pas les respecter
  provoque un bug (au démarrage, à l'ouverture du quiz, ou en jeu).
- **§3 — Recommandations** : des **conventions de qualité/cohérence** (utiles
  notamment quand on demande à une IA de générer un quiz). Le code n'en dépend
  pas ; les enfreindre ne casse rien, mais le résultat sera moins propre.

> 🧭 En résumé : un humain qui crée son quiz à la main n'a **que la §2** à
> respecter. La §3 est là pour produire des quiz homogènes et équilibrés.

---

## 1. Le format de base

Un **niveau** = un fichier `<nom>.json` déposé dans le dossier d'un thème.
Un fichier = un **tableau JSON** d'objets question.

```
backend/data/
  themes.json                        ← métadonnées des thèmes
  questions/
    <id-du-theme>/
      <niveau>.json                  ← un fichier = un niveau (nom libre)
```

```json
[
  {
    "id": "identifiant-unique",
    "enonce": "La question posée au joueur ?",
    "choix": ["Option A", "Option B", "Option C", "Option D"],
    "bonneReponse": 0,
    "explication": "Pourquoi c'est la bonne réponse, avec un peu de contexte."
  }
]
```

| Champ          | Type    | Rôle                                              |
|----------------|---------|---------------------------------------------------|
| `id`           | string  | Identifiant de la question                        |
| `enonce`       | string  | Le texte de la question                           |
| `choix`        | array   | Les réponses possibles                            |
| `bonneReponse` | integer | **Index** (commence à 0) du bon choix dans `choix`|
| `explication`  | string  | Texte affiché après la réponse                    |

---

## 2. ✅ Règles DURES (imposées par le code)

Chacune correspond à un comportement réel du backend Go ou du frontend Angular.

### 2.1 — Emplacement et nom de fichier
- Le fichier doit être dans `backend/data/questions/<id-du-theme>/`, et `<id-du-theme>`
  doit correspondre exactement à un `id` présent dans `themes.json`.
- Le **nom du niveau est libre** : `facile.json`, `debutant.json`, `histoire.json`,
  `saison-1.json`… Le backend découvre **tous** les fichiers `.json` du dossier.
- Le **nombre de niveaux est libre** : 1, 2, 5… autant de fichiers que voulu.
- ⛔ **`tous` est un nom réservé** : un fichier `tous.json` serait **ignoré**.
  (« tous » est un niveau synthétique généré automatiquement, voir §2.6.)

### 2.2 — Chaque thème doit avoir au moins un niveau
Un thème déclaré dans `themes.json` mais sans **aucun** fichier de questions fait
**échouer le démarrage** du backend (fail-fast volontaire). Inversement, un dossier
de questions sans entrée dans `themes.json` est ignoré (le thème n'apparaît pas).

### 2.3 — JSON valide, sinon le niveau est injouable
Le fichier doit être un tableau JSON bien formé. S'il est invalide (virgule en trop,
guillemet manquant…), l'ouverture de ce niveau renvoie une erreur HTTP et le joueur
voit un message d'échec. ⚠️ Comme les fichiers sont lus **paresseusement** (au premier
accès), une erreur n'est détectée qu'à l'ouverture du quiz concerné, **pas** au boot.

### 2.4 — `bonneReponse` doit être un index VALIDE
`bonneReponse` est l'**indice** du bon choix (0 = premier, 1 = deuxième…), **pas** le
texte de la réponse. Il doit être compris entre `0` et `longueur(choix) - 1`.
> 🐞 Piège silencieux : un index hors limites (ex. `bonneReponse: 5` avec 4 choix)
> rend la question **ingagnable** — aucune réponse n'est jamais « correcte », donc la
> question revient indéfiniment dans la file. Aucun message d'erreur n'est affiché.

### 2.5 — Nombre de `choix` : 1 à 4 réellement utilisables
Le code accepte un nombre quelconque de choix, **mais** :
- l'affichage des lettres (`A`, `B`, `C`, `D`) et les raccourcis clavier (`1`–`4` /
  `A`–`D`) ne couvrent que **4 choix** ;
- un 5ᵉ choix (ou plus) s'affiche **sans lettre ni raccourci clavier**.

Donc en pratique : **2 à 4 choix**. (Voir §3 : 4 est recommandé.)

### 2.6 — Le niveau « tous » est automatique
Dès qu'un thème a **au moins 2** niveaux réels, un niveau synthétique **`tous`** est
généré : il agrège (et le frontend mélange) les questions de **tous** les niveaux du
thème — y compris tout nouveau fichier que tu ajoutes. On ne le crée donc **jamais à
la main**.

### 2.7 — `id` de question : présent et unique « par partie »
- L'`id` sert de clé d'animation côté frontend.
- Il doit être **unique au sein de ce qui est joué ensemble** : un niveau donné, et
  le niveau `tous` (qui rassemble tous les niveaux **du même thème**).
- Deux `id` identiques → au pire un **micro-défaut d'animation** quand deux questions
  homonymes s'enchaînent. C'est **cosmétique**, ça ne casse pas le jeu.
- Pas besoin d'unicité **globale** : des `id` identiques entre **thèmes différents**
  ne se croisent jamais.

---

## 3. 💡 Recommandations (qualité & cohérence)

Le code ne les vérifie pas, mais elles rendent les quiz homogènes. **À suivre en
priorité quand on demande à une IA de générer un questionnaire.**

### 3.1 — Exactement 4 choix
C'est le « point idéal » : couvert par les lettres A–D et les raccourcis clavier (§2.5).

### 3.2 — Noms de niveau : `facile` / `moyen` / `expert` quand c'est pertinent
- Ces trois noms ont une **couleur dédiée** et un **ordre d'affichage canonique**
  (facile → moyen → expert).
- Un niveau au nom personnalisé fonctionne, mais s'affiche avec une **couleur neutre**
  et **après** les niveaux connus (les noms inconnus sont triés alphabétiquement).
  > Pour donner une couleur à un niveau perso, ajouter une variable CSS `--<nom>` dans
  > `frontend/src/styles.scss` et la classe `.niveau--<nom>` (voir les niveaux existants).

### 3.3 — ~10 questions par niveau
Aucune obligation (1 question suffit techniquement), mais une dizaine par niveau donne
un quiz équilibré et une progression lisible.

### 3.4 — `id` structurés et lisibles
Convention du projet : `<theme>-<lettre-niveau><numéro>` — `go-f1` (facile), `go-m3`
(moyen), `go-e10` (expert). Pratique pour s'y retrouver, **non requis par le code**.

### 3.5 — Une explication qui apprend quelque chose
L'explication doit donner le **pourquoi** et un peu de contexte, pas seulement répéter
la bonne réponse.

### 3.6 — Équilibre des choix (anti-déduction) ⚠️
Les 4 choix doivent avoir une **longueur et un niveau de détail comparables**, pour que
la bonne réponse ne soit pas devinable à sa forme.

**À éviter :**
```json
"choix": ["le Bitcoin", "l'Ether (ETH)", "le Solana", "le Dogecoin"]
```
→ Seul `"l'Ether (ETH)"` porte une précision entre parenthèses : ça le trahit.

**À faire :**
```json
"choix": ["le Bitcoin", "l'Ether", "le Solana", "le Dogecoin"]
```
→ Déplacer `(ETH)` dans l'`explication`.

S'applique à : sigles développés, symboles, exemples entre parenthèses, définitions
plus complètes, adjectifs présents sur un seul choix.

---

## 4. Entrée dans `themes.json`

Pour qu'un thème apparaisse, il faut une entrée dans `themes.json` **et** au moins un
fichier de niveau (§2.2).

```json
{
  "id": "mon-theme",
  "nom": "Mon Thème",
  "emoji": "🎯",
  "categorie": "Ma Catégorie",
  "description": "Une phrase courte qui décrit le thème."
}
```

Catégories existantes : `Programmation`, `Bases de données`, `Crypto & Web3`,
`Intelligence artificielle`, `Culture geek`, `Jeux & stratégie`, `Culture générale`.

---

## 5. Exemple complet — `backend/data/questions/python/facile.json`

```json
[
  {
    "id": "python-f1",
    "enonce": "Quel mot-clé définit une fonction en Python ?",
    "choix": ["func", "function", "def", "fn"],
    "bonneReponse": 2,
    "explication": "`def` introduit la définition d'une fonction : `def ma_fonction():`. C'est l'un des mots-clés réservés du langage."
  },
  {
    "id": "python-f2",
    "enonce": "Comment affiche-t-on du texte dans la console en Python 3 ?",
    "choix": ["echo(\"texte\")", "console.log(\"texte\")", "println(\"texte\")", "print(\"texte\")"],
    "bonneReponse": 3,
    "explication": "`print()` est une fonction intégrée de Python 3. En Python 2, `print` était un mot-clé sans parenthèses."
  }
]
```

> Après création, **redémarrer le backend** (`go run .`). Aucune recompilation de code
> n'est nécessaire : les données sont relues au démarrage.

---

## 6. Checklist

**Bloquant (§2 — sinon ça casse) :**
- [ ] Fichier dans `questions/<id-theme>/`, avec `<id-theme>` présent dans `themes.json`
- [ ] Le niveau n'est pas nommé `tous.json`
- [ ] JSON valide (tableau d'objets)
- [ ] `bonneReponse` est un **index** valide (0 … longueur des `choix` − 1)
- [ ] 2 à 4 `choix` par question
- [ ] `id` de question présents et uniques au sein du niveau (et entre niveaux du thème)

**Qualité (§3 — recommandé) :**
- [ ] Exactement 4 choix
- [ ] 10 questions par niveau
- [ ] Choix équilibrés (règle anti-déduction)
- [ ] Explication qui apporte du contexte
- [ ] 3 niveaux "facile", "moyen" et "expert"
- [ ] `id` structurés (`<theme>-<f|m|e><n>`)
