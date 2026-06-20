/* =========================================================================
   aleatoire.ts — petites fonctions utilitaires liées au hasard.
   ========================================================================= */

/**
 * melanger : renvoie une COPIE mélangée d'un tableau (mélange de Fisher-Yates).
 *
 * Pourquoi Fisher-Yates ? C'est l'algorithme standard pour un mélange
 * vraiment uniforme (chaque permutation est aussi probable). On parcourt le
 * tableau de la fin vers le début et on échange chaque élément avec un autre
 * tiré au hasard parmi ceux pas encore traités.
 *
 * Le `<T>` est un GÉNÉRIQUE : la fonction marche pour un tableau de n'importe
 * quel type (chaînes, nombres, objets Question...) tout en gardant le typage.
 *
 * On travaille sur une copie (`[...source]`) pour ne PAS modifier le tableau
 * d'origine : une fonction sans effet de bord est plus simple à raisonner.
 */
export function melanger<T>(source: readonly T[]): T[] {
  const tableau = [...source];
  for (let i = tableau.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [tableau[i], tableau[j]] = [tableau[j], tableau[i]]; // échange via destructuration
  }
  return tableau;
}
