import { HttpErrorResponse } from '@angular/common/http';

/* =========================================================================
   erreurs.ts — traduire une erreur HTTP en message lisible par le joueur.

   Avant, les trois pages affichaient toutes le même texte quoi qu'il arrive :
   « Le serveur Go est-il bien démarré sur le port 8080 ? ». C'est trompeur dès
   que le serveur répond : sur l'URL /quiz/nimportequoi/facile, il renvoie un
   404 parfaitement correct, et on accusait quand même le serveur d'être éteint.

   On distingue donc les cas, parce qu'ils appellent des actions différentes :
   relancer le serveur, corriger l'URL, ou simplement réessayer.
   ========================================================================= */

/**
 * Construit le message à afficher.
 *
 * @param erreur      l'erreur reçue par le callback `error` d'un Observable.
 * @param introuvable le message spécifique au cas 404 — seul l'appelant sait
 *                    ce qui était demandé (« ce thème », « ce quiz »…).
 */
export function messageErreur(erreur: unknown, introuvable: string): string {
  // `unknown` plutôt que `any` : TypeScript nous OBLIGE à vérifier le type
  // avant d'utiliser la valeur. C'est exactement ce qu'on veut ici.
  if (erreur instanceof HttpErrorResponse) {
    // status 0 = la requête n'a jamais abouti : serveur éteint, DNS en échec,
    // ou requête bloquée par le navigateur (CORS). C'est LE seul cas où il est
    // pertinent de demander si le backend tourne.
    if (erreur.status === 0) {
      return 'Impossible de joindre le serveur. Est-il bien démarré ? '
        + '(dans un terminal : cd backend && go run .)';
    }

    // 404 : le serveur va très bien, c'est la ressource qui n'existe pas.
    if (erreur.status === 404) {
      return introuvable;
    }

    // Tout le reste (500, 503…) : le serveur a un problème de son côté.
    return `Le serveur a répondu une erreur ${erreur.status}. Réessaie dans un instant.`;
  }

  // Erreur qui ne vient pas d'HttpClient (bug de notre côté, par exemple).
  return "Une erreur inattendue s'est produite.";
}
