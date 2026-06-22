# ============================================================================
#  lancer.ps1 — Lance le projet « quiz-connaissances » (backend Go + frontend Angular)
# ----------------------------------------------------------------------------
#  Ce script :
#    1. Vérifie que les outils nécessaires sont installés (Go, Node.js/npm).
#    2. Si un outil manque (ou est trop ancien), affiche une POPUP qui explique
#       quoi installer, avec :
#         - Méthode 1 : une ligne de commande (winget) pour les habitués.
#         - Méthode 2 : le lien de la page de téléchargement de l'installateur.
#    3. Si tout est OK, lance le backend (« go run . ») et le frontend
#       (« npm start ») chacun dans sa propre fenêtre.
#
#  Astuce : double-cliquez plutôt sur « lancer.bat » (il appelle ce script en
#  contournant la politique d'exécution PowerShell).
# ============================================================================

# On charge les bibliothèques Windows pour afficher des fenêtres (popups).
Add-Type -AssemblyName System.Windows.Forms | Out-Null

# Dossier où se trouve ce script = racine du projet.
$racine   = Split-Path -Parent $MyInvocation.MyCommand.Definition
$backend  = Join-Path $racine 'backend'
$frontend = Join-Path $racine 'frontend'

# --- Petites fonctions utilitaires -----------------------------------------

# Affiche une popup d'information (titre + message).
function Afficher-Popup {
    param(
        [string]$Message,
        [string]$Titre = 'Quiz Connaissances',
        [System.Windows.Forms.MessageBoxIcon]$Icone = [System.Windows.Forms.MessageBoxIcon]::Information,
        [System.Windows.Forms.MessageBoxButtons]$Boutons = [System.Windows.Forms.MessageBoxButtons]::OK
    )
    return [System.Windows.Forms.MessageBox]::Show($Message, $Titre, $Boutons, $Icone)
}

# Renvoie $true si la commande existe dans le PATH (ex : « go », « node »).
function Outil-Present {
    param([string]$Nom)
    return [bool](Get-Command $Nom -ErrorAction SilentlyContinue)
}

# Extrait le 1er numéro de version « X.Y(.Z) » d'une chaîne quelconque.
# Renvoie $null si rien trouvé.
function Extraire-Version {
    param([string]$Texte)
    if ($Texte -match '(\d+)\.(\d+)(?:\.(\d+))?') {
        return [version]("{0}.{1}.{2}" -f $matches[1], $matches[2], ($(if ($matches[3]) { $matches[3] } else { '0' })))
    }
    return $null
}

# --- Définition des prérequis ----------------------------------------------
# Chaque prérequis sait se vérifier lui-même et connaît ses méthodes d'install.

$prerequis = @(
    [pscustomobject]@{
        Nom         = 'Go'
        Role        = 'backend (serveur)'
        Commande    = 'go'
        VersionMin  = [version]'1.26'
        LireVersion = { Extraire-Version (go version 2>$null) }
        Winget      = 'winget install --id GoLang.Go -e'
        Page        = 'https://go.dev/dl/'
    },
    [pscustomobject]@{
        Nom         = 'Node.js (inclut npm)'
        Role        = 'frontend (Angular)'
        Commande    = 'node'
        VersionMin  = [version]'20.19'
        LireVersion = { Extraire-Version (node -v 2>$null) }
        Winget      = 'winget install --id OpenJS.NodeJS.LTS -e'
        Page        = 'https://nodejs.org/fr/download'
    }
)

# --- Diagnostic de l'environnement -----------------------------------------

$manquants = @()      # outils carrément absents (bloquant)
$tropAnciens = @()    # outils présents mais version trop vieille (avertissement)

foreach ($p in $prerequis) {
    if (-not (Outil-Present $p.Commande)) {
        $manquants += $p
        continue
    }
    # L'outil est présent : on vérifie la version (best-effort, non bloquant).
    $version = & $p.LireVersion
    if ($version -and $version -lt $p.VersionMin) {
        $tropAnciens += [pscustomobject]@{ Prerequis = $p; VersionTrouvee = $version }
    }
}

# --- Cas 1 : il manque des outils => popup explicative, puis on s'arrête -----

if ($manquants.Count -gt 0) {

    $lignes = @()
    $lignes += "Il manque des outils pour lancer le projet."
    $lignes += ""
    foreach ($p in $manquants) {
        $lignes += "[X] $($p.Nom)  -  necessaire pour le $($p.Role)"
        $lignes += "      Methode 1 (ligne de commande, rapide) :"
        $lignes += "          $($p.Winget)"
        $lignes += "      Methode 2 (telecharger l'installateur) :"
        $lignes += "          $($p.Page)"
        $lignes += ""
    }
    $lignes += "Angular CLI n'est PAS a installer a la main : il est inclus dans"
    $lignes += "le projet et sera mis en place automatiquement par « npm install »."
    $lignes += ""
    $lignes += "Apres l'installation : fermez puis rouvrez ce script."
    $lignes += ""
    $lignes += "Voulez-vous ouvrir maintenant les pages de telechargement ?"

    $message = $lignes -join "`r`n"

    # On affiche aussi le détail dans la console (texte sélectionnable/copiable).
    Write-Host ""
    Write-Host "=== Outils manquants ===" -ForegroundColor Red
    Write-Host $message
    Write-Host ""

    $reponse = Afficher-Popup -Message $message -Titre 'Outils manquants' `
        -Icone ([System.Windows.Forms.MessageBoxIcon]::Warning) `
        -Boutons ([System.Windows.Forms.MessageBoxButtons]::YesNo)

    if ($reponse -eq [System.Windows.Forms.DialogResult]::Yes) {
        foreach ($p in $manquants) { Start-Process $p.Page }
    }

    exit 1
}

# --- Cas 2 : tout est present mais une version est trop ancienne (avertissement)

if ($tropAnciens.Count -gt 0) {
    $lignes = @()
    $lignes += "Des outils sont installes mais peut-etre trop anciens :"
    $lignes += ""
    foreach ($t in $tropAnciens) {
        $p = $t.Prerequis
        $lignes += "[!] $($p.Nom) : version $($t.VersionTrouvee) detectee, $($p.VersionMin)+ recommandee."
        $lignes += "      Mise a jour : $($p.Winget)"
        $lignes += "      ou : $($p.Page)"
        $lignes += ""
    }
    $lignes += "On tente quand meme de lancer le projet..."
    $message = $lignes -join "`r`n"

    Write-Host $message -ForegroundColor Yellow
    Afficher-Popup -Message $message -Titre 'Versions a verifier' `
        -Icone ([System.Windows.Forms.MessageBoxIcon]::Warning) | Out-Null
}

# --- Cas 3 : environnement OK => on lance les deux serveurs ------------------

Write-Host "Environnement OK. Lancement du backend et du frontend..." -ForegroundColor Green

# Backend : « go run . » dans une nouvelle fenetre PowerShell qui reste ouverte.
Start-Process -FilePath 'powershell' -WorkingDirectory $backend `
    -ArgumentList '-NoExit', '-Command',
        "Write-Host 'BACKEND Go - http://localhost:8080' -ForegroundColor Cyan; go run ."

# Frontend : « npm install » (seulement si besoin) puis « npm start ».
$cmdFront = "Write-Host 'FRONTEND Angular - http://localhost:4200' -ForegroundColor Cyan; " +
            "if (-not (Test-Path 'node_modules')) { Write-Host 'Premiere installation (npm install)...' -ForegroundColor Yellow; npm install }; " +
            "npm start"
Start-Process -FilePath 'powershell' -WorkingDirectory $frontend `
    -ArgumentList '-NoExit', '-Command', $cmdFront

# Petit récapitulatif dans la console (PAS de popup quand tout va bien).
Write-Host ""
Write-Host "Les deux serveurs demarrent dans deux nouvelles fenetres :" -ForegroundColor Green
Write-Host "  Backend  (API Go)  : http://localhost:8080"
Write-Host "  Frontend (Angular) : http://localhost:4200"
Write-Host "Pour arreter : fermez les deux fenetres ou faites Ctrl+C dans chacune."

# On attend que le serveur Angular reponde sur le port 4200, puis on ouvre le
# navigateur directement sur le site. La 1re fois (npm install) ca peut etre long,
# d'ou un timeout genereux.
$port    = 4200
$timeout = [TimeSpan]::FromMinutes(5)
$horloge = [System.Diagnostics.Stopwatch]::StartNew()
$pret    = $false

# Teste si quelque chose ecoute sur localhost:$port (connexion TCP rapide).
function Port-Pret {
    param([int]$Port)
    $client = New-Object System.Net.Sockets.TcpClient
    try {
        $async = $client.BeginConnect('127.0.0.1', $Port, $null, $null)
        if ($async.AsyncWaitHandle.WaitOne(500) -and $client.Connected) { return $true }
        return $false
    } catch {
        return $false
    } finally {
        $client.Close()
    }
}

Write-Host ""
Write-Host "Attente du demarrage du frontend (ouverture auto du navigateur)..." -ForegroundColor Cyan
while ($horloge.Elapsed -lt $timeout) {
    if (Port-Pret -Port $port) { $pret = $true; break }
    Start-Sleep -Seconds 1
}

if ($pret) {
    Write-Host "Frontend pret. Ouverture de http://localhost:$port" -ForegroundColor Green
    Start-Process "http://localhost:$port"
} else {
    # Filet de securite : on n'a pas vu le port s'ouvrir a temps, on ouvre quand meme.
    Write-Host "Delai depasse, ouverture quand meme de http://localhost:$port" -ForegroundColor Yellow
    Start-Process "http://localhost:$port"
}
