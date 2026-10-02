import { HttpErrorResponse } from '@angular/common/http';
import { describe, expect, it } from 'vitest';

import { messageErreur } from './erreurs';

/* =========================================================================
   Tests de la traduction des erreurs HTTP en messages pour le joueur.
   ========================================================================= */

const INTROUVABLE = 'Ce thème n\'existe pas.';

describe('messageErreur', () => {
  it('parle du serveur éteint UNIQUEMENT quand la requête n\'aboutit pas (status 0)', () => {
    const message = messageErreur(new HttpErrorResponse({ status: 0 }), INTROUVABLE);

    expect(message).toContain('go run');
  });

  it('renvoie le message spécifique sur un 404, sans accuser le serveur', () => {
    const message = messageErreur(new HttpErrorResponse({ status: 404 }), INTROUVABLE);

    expect(message).toBe(INTROUVABLE);
    // C'était le bug : un 404 affichait « le serveur est-il démarré ? »
    // alors que le serveur venait justement de répondre.
    expect(message).not.toContain('go run');
  });

  it('signale une panne serveur sur un 500', () => {
    const message = messageErreur(new HttpErrorResponse({ status: 500 }), INTROUVABLE);

    expect(message).toContain('500');
    expect(message).not.toBe(INTROUVABLE);
  });

  it('reste robuste face à une erreur qui ne vient pas d\'HttpClient', () => {
    const message = messageErreur(new Error('boum'), INTROUVABLE);

    expect(message).toContain('inattendue');
  });
});
