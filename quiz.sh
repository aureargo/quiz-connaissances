#!/usr/bin/env bash
# ============================================================================
#  quiz.sh — Lance le projet « quiz-connaissances » (backend Go + frontend Angular)
#            sous Linux et macOS.  (Équivalent de quiz.bat / quiz.ps1 pour Windows.)
# ----------------------------------------------------------------------------
#  Ce script :
#    1. Vérifie que les outils nécessaires sont installés (Go, Node.js/npm).
#    2. Si un outil manque (ou est trop ancien), affiche dans le terminal quoi
#       installer (commande du gestionnaire de paquets + page de téléchargement).
#    3. Si tout est OK, lance le backend (« go run . ») et le frontend
#       (« npm start ») en arrière-plan, attend que le port 4200 réponde, puis
#       ouvre le navigateur sur le site.
#
#  Utilisation :
#       chmod +x quiz.sh     # une seule fois, pour le rendre exécutable
#       ./quiz.sh
#
#  Pour arrêter : Ctrl+C dans ce terminal (les deux serveurs sont alors coupés).
# ============================================================================

set -u  # erreur si on utilise une variable non définie

# Dossier où se trouve ce script = racine du projet.
RACINE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND="$RACINE/backend"
FRONTEND="$RACINE/frontend"

# --- Petites fonctions utilitaires -----------------------------------------

# Renvoie 0 (vrai) si la commande existe dans le PATH (ex : « go », « node »).
outil_present() {
    command -v "$1" >/dev/null 2>&1
}

# Extrait le 1er numéro de version « X.Y(.Z) » d'une chaîne quelconque.
extraire_version() {
    echo "$1" | grep -oE '[0-9]+\.[0-9]+(\.[0-9]+)?' | head -n1
}

# Compare deux versions « X.Y.Z ». Renvoie 0 (vrai) si $1 >= $2.
version_suffisante() {
    # « sort -V » trie en ordre de version ; si la plus petite des deux est la
    # version minimale requise, alors la version trouvée est suffisante.
    local minimale
    minimale="$(printf '%s\n%s\n' "$1" "$2" | sort -V | head -n1)"
    [ "$minimale" = "$2" ]
}

# Ouvre une URL dans le navigateur par défaut (macOS = open, Linux = xdg-open).
ouvrir_navigateur() {
    if outil_present open; then
        open "$1"
    elif outil_present xdg-open; then
        xdg-open "$1" >/dev/null 2>&1 &
    else
        echo "  (Ouvrez manuellement votre navigateur sur $1)"
    fi
}

# Teste si quelque chose écoute sur localhost:$1 (connexion TCP rapide).
port_pret() {
    # Bash sait ouvrir une socket via /dev/tcp ; on essaie IPv4 et IPv6 (le
    # serveur de dev Angular/Vite peut écouter sur l'un ou l'autre).
    (exec 3<>/dev/tcp/127.0.0.1/"$1") >/dev/null 2>&1 && exec 3>&- && return 0
    (exec 3<>/dev/tcp/::1/"$1")       >/dev/null 2>&1 && exec 3>&- && return 0
    return 1
}

# --- Définition des prérequis ----------------------------------------------
# Chaque ligne : « commande|nom affiché|rôle|version mini|install|page ».

PREREQUIS=(
    "go|Go|backend (serveur)|1.26|brew install go  (ou: sudo apt install golang)|https://go.dev/dl/"
    "node|Node.js (inclut npm)|frontend (Angular)|20.19|brew install node  (ou: voir nodesource)|https://nodejs.org/fr/download"
)

# --- Diagnostic de l'environnement -----------------------------------------

manquants=()
trop_anciens=()

for ligne in "${PREREQUIS[@]}"; do
    IFS='|' read -r cmd nom role vmin install page <<< "$ligne"

    if ! outil_present "$cmd"; then
        manquants+=("$ligne")
        continue
    fi

    # L'outil est présent : on vérifie la version (best-effort, non bloquant).
    case "$cmd" in
        go)   brute="$(go version 2>/dev/null)" ;;
        node) brute="$(node -v 2>/dev/null)" ;;
        *)    brute="" ;;
    esac
    trouvee="$(extraire_version "$brute")"
    if [ -n "$trouvee" ] && ! version_suffisante "$trouvee" "$vmin"; then
        trop_anciens+=("$ligne|$trouvee")
    fi
done

# --- Cas 1 : il manque des outils => message, puis on s'arrête ---------------

if [ "${#manquants[@]}" -gt 0 ]; then
    echo ""
    echo "=== Outils manquants ==="
    echo "Il manque des outils pour lancer le projet :"
    echo ""
    for ligne in "${manquants[@]}"; do
        IFS='|' read -r cmd nom role vmin install page <<< "$ligne"
        echo "[X] $nom  -  nécessaire pour le $role"
        echo "      Installation : $install"
        echo "      ou télécharger : $page"
        echo ""
    done
    echo "Angular CLI n'est PAS à installer à la main : il est inclus dans le"
    echo "projet et sera mis en place automatiquement par « npm install »."
    echo ""
    echo "Après l'installation : relancez ce script."
    exit 1
fi

# --- Cas 2 : tout est présent mais une version est trop ancienne (avertissement)

if [ "${#trop_anciens[@]}" -gt 0 ]; then
    echo ""
    echo "Des outils sont installés mais peut-être trop anciens :"
    echo ""
    for entree in "${trop_anciens[@]}"; do
        IFS='|' read -r cmd nom role vmin install page trouvee <<< "$entree"
        echo "[!] $nom : version $trouvee détectée, $vmin+ recommandée."
        echo "      Mise à jour : $install"
    done
    echo ""
    echo "On tente quand même de lancer le projet..."
fi

# --- Cas 3 : environnement OK => on lance les deux serveurs ------------------

echo ""
echo "Environnement OK. Lancement du backend et du frontend..."

# On garde les PID pour pouvoir couper les deux serveurs à la sortie (Ctrl+C).
pids=()

# Backend : « go run . » dans le dossier backend.
( cd "$BACKEND" && exec go run . ) &
pids+=($!)
echo "  Backend  (API Go)  : http://localhost:8080"

# Frontend : « npm install » (seulement si besoin) puis « npm start ».
(
    cd "$FRONTEND" || exit 1
    if [ ! -d node_modules ]; then
        echo "Première installation des dépendances (npm install)..."
        npm install
    fi
    exec npm start
) &
pids+=($!)
echo "  Frontend (Angular) : http://localhost:4200"

# À la sortie du script (Ctrl+C ou fin), on coupe proprement les deux serveurs.
nettoyer() {
    echo ""
    echo "Arrêt des serveurs..."
    for pid in "${pids[@]}"; do
        # On tue le groupe de processus pour emporter go/npm ET leurs enfants.
        kill "$pid" 2>/dev/null
    done
}
trap nettoyer EXIT INT TERM

# On attend que le serveur Angular réponde sur le port 4200, puis on ouvre le
# navigateur. La 1re fois (npm install) ça peut être long, d'où un timeout large.
PORT=4200
TIMEOUT=300   # secondes (5 minutes)

echo ""
echo "Attente du démarrage du frontend (ouverture auto du navigateur)..."
ecoule=0
pret=0
while [ "$ecoule" -lt "$TIMEOUT" ]; do
    if port_pret "$PORT"; then pret=1; break; fi
    sleep 1
    ecoule=$((ecoule + 1))
done

if [ "$pret" -eq 1 ]; then
    echo "Frontend prêt. Ouverture de http://localhost:$PORT"
else
    echo "Délai dépassé, ouverture quand même de http://localhost:$PORT"
fi
ouvrir_navigateur "http://localhost:$PORT"

# On laisse le script vivant tant que les serveurs tournent : Ctrl+C déclenche
# le trap « nettoyer » qui les coupe tous les deux.
echo ""
echo "Les deux serveurs tournent. Laissez ce terminal ouvert."
echo "Pour arrêter : Ctrl+C ici."
wait
