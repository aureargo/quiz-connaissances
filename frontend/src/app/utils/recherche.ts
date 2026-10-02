/* =========================================================================
   recherche.ts — filtrage et classement des thèmes selon un texte saisi.

   Fonction PURE (aucune dépendance Angular) : facile à tester, et la page
   d'accueil n'a qu'à l'appeler dans un computed().
   ========================================================================= */

import { Theme } from '../models/quiz.models';

/**
 * normaliser : met un texte sous une forme « comparable » :
 * minuscules, sans accents, espaces superflus retirés.
 *
 * Astuce des accents : la forme Unicode NFD décompose « é » en « e » + un
 * accent combinant (plage U+0300 à U+036F) ; il suffit ensuite de supprimer
 * ces accents. Ainsi « ecologie » trouve « Écologie & climat ».
 */
export function normaliser(texte: string): string {
  return texte
    .normalize('NFD')
    .replace(/[̀-ͯ]/g, '')
    .toLowerCase()
    .trim();
}

/** Découpe un texte normalisé en mots (tout ce qui n'est ni lettre ni chiffre sépare). */
function mots(texte: string): string[] {
  return texte.split(/[^a-z0-9]+/).filter((mot) => mot.length > 0);
}

/**
 * score : plus il est PETIT, plus le thème est pertinent. `null` = pas de
 * correspondance. On regarde d'abord le nom, puis la catégorie.
 */
function score(theme: Theme, requete: string): number | null {
  const nom = normaliser(theme.nom);
  if (nom === requete) return 0;                                  // « c » → C
  if (nom.startsWith(requete)) return 1;                          // « ja » → Java, JavaScript
  if (mots(nom).some((mot) => mot.startsWith(requete))) return 2; // « linux » → Lignes de commande Linux
  if (nom.includes(requete)) return 3;                            // « script » → TypeScript
  if (mots(normaliser(theme.categorie)).some((mot) => mot.startsWith(requete))) {
    return 4;                                                     // « sport » → Cyclisme
  }
  return null;
}

/**
 * rechercherThemes : renvoie les thèmes qui correspondent à `recherche`,
 * classés du plus pertinent au moins pertinent.
 *
 * - Recherche vide → tableau vide (c'est à l'appelant d'afficher tout).
 * - À score égal, on garde l'ordre d'origine (Array.prototype.sort est
 *   STABLE depuis ES2019), donc l'ordre des catégories est respecté.
 */
export function rechercherThemes(themes: readonly Theme[], recherche: string): Theme[] {
  const requete = normaliser(recherche);
  if (requete === '') return [];

  return themes
    .map((theme) => ({ theme, score: score(theme, requete) }))
    .filter((r): r is { theme: Theme; score: number } => r.score !== null)
    .sort((a, b) => a.score - b.score)
    .map((r) => r.theme);
}
