import { describe, it, expect } from 'vitest';

import { MoteurQuiz } from './moteur-quiz';
import { Question } from '../models/quiz.models';

/* =========================================================================
   Tests unitaires du MoteurQuiz.

   Un test unitaire vérifie automatiquement qu'un morceau de code se comporte
   comme prévu. On utilise vitest (déjà configuré par Angular) :
     describe(...) regroupe des tests ; it(...) décrit un cas ; expect(...) vérifie.

   Lancement : `npm test` (depuis le dossier frontend).
   ========================================================================= */

// Petite usine à questions factices, pour ne pas dépendre du backend.
function questionsFactices(): Question[] {
  return [
    { id: 'q1', enonce: 'Q1', choix: ['a', 'b', 'c', 'd'], bonneReponse: 0, explication: '' },
    { id: 'q2', enonce: 'Q2', choix: ['a', 'b', 'c', 'd'], bonneReponse: 1, explication: '' },
    { id: 'q3', enonce: 'Q3', choix: ['a', 'b', 'c', 'd'], bonneReponse: 2, explication: '' }
  ];
}

describe('MoteurQuiz', () => {
  it('démarre avec toutes les questions à maîtriser', () => {
    const m = new MoteurQuiz(questionsFactices());
    expect(m.nombreTotal()).toBe(3);
    expect(m.nbMaitrisees()).toBe(0);
    expect(m.termine()).toBe(false);
    expect(m.questionCourante()).not.toBeNull();
  });

  it('mélange les choix tout en gardant la bonne réponse cohérente', () => {
    const m = new MoteurQuiz(questionsFactices());
    const q = m.questionCourante()!;
    // À l'index "indexBonneReponse" (après mélange) doit se trouver le bon texte.
    const texteAttendu = q.question.choix[q.question.bonneReponse];
    expect(q.choixMelanges[q.indexBonneReponse]).toBe(texteAttendu);
    expect(q.choixMelanges).toHaveLength(4);
  });

  it('remet une question ratée dans la file (elle revient plus tard)', () => {
    const m = new MoteurQuiz(questionsFactices());
    const courante = m.questionCourante()!;
    const indexFaux = (courante.indexBonneReponse + 1) % 4; // forcément une mauvaise réponse

    m.repondre(indexFaux);
    expect(m.nbErreurs()).toBe(1);

    m.questionSuivante();
    // La question ratée n'est PAS maîtrisée : on a toujours 0 maîtrisée et la partie continue.
    expect(m.nbMaitrisees()).toBe(0);
    expect(m.termine()).toBe(false);
  });

  it('ignore les clics multiples sur une même question', () => {
    const m = new MoteurQuiz(questionsFactices());
    const q = m.questionCourante()!;
    const indexFaux = (q.indexBonneReponse + 1) % 4;

    m.repondre(indexFaux);
    m.repondre(q.indexBonneReponse); // ce 2e clic doit être ignoré
    expect(m.nbErreurs()).toBe(1); // une seule erreur comptée
    expect(m.reponseSelectionnee()).toBe(indexFaux); // la 1re réponse reste
  });

  it('se termine quand toutes les questions sont réussies (sans erreur)', () => {
    const m = new MoteurQuiz(questionsFactices());
    for (let i = 0; i < 3; i++) {
      const q = m.questionCourante()!;
      m.repondre(q.indexBonneReponse);
      m.questionSuivante();
    }
    expect(m.termine()).toBe(true);
    expect(m.nbMaitrisees()).toBe(3);
    expect(m.nbErreurs()).toBe(0);
    expect(m.questionCourante()).toBeNull();
  });

  it('finit toujours par se terminer en répondant correctement', () => {
    const m = new MoteurQuiz(questionsFactices());
    let securite = 0;
    while (!m.termine() && securite++ < 100) {
      const q = m.questionCourante()!;
      m.repondre(q.indexBonneReponse);
      m.questionSuivante();
    }
    expect(m.termine()).toBe(true);
  });
});
