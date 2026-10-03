import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { of } from 'rxjs';
import { describe, expect, it } from 'vitest';

import { Quiz } from './quiz';
import { QuizApi } from '../../services/quiz-api';
import { Question, Theme } from '../../models/quiz.models';

/* =========================================================================
   Tests des raccourcis clavier de la page de quiz.

   On remplace QuizApi par un faux service (useValue) qui renvoie des données
   fixes : pas de backend, pas de réseau, le test ne dépend que du composant.
   ========================================================================= */

const THEME: Theme = { id: 'go', nom: 'Go', emoji: '🐹', categorie: 'Programmation', description: '' };
const QUESTIONS: Question[] = [
  { id: 'q1', enonce: 'Q1', choix: ['a', 'b', 'c', 'd'], bonneReponse: 0, explication: '' },
  { id: 'q2', enonce: 'Q2', choix: ['a', 'b', 'c', 'd'], bonneReponse: 1, explication: '' }
];

function creerQuiz(): Quiz {
  TestBed.configureTestingModule({
    imports: [Quiz],
    providers: [
      provideRouter([]),
      { provide: QuizApi, useValue: { getTheme: () => of(THEME), getQuestions: () => of(QUESTIONS) } }
    ]
  });
  const fixture = TestBed.createComponent(Quiz);
  fixture.detectChanges(); // déclenche ngOnInit → chargement → création du moteur
  return fixture.componentInstance;
}

function touche(key: string, options: KeyboardEventInit = {}): KeyboardEvent {
  return new KeyboardEvent('keydown', { key, bubbles: true, cancelable: true, ...options });
}

describe('Quiz — clavier', () => {
  it('répond avec la touche « c »', () => {
    const quiz = creerQuiz();

    quiz.gererClavier(touche('c'));

    expect(quiz.moteur()!.reponseSelectionnee()).toBe(2);
  });

  it('ignore Ctrl+C (copier l\'énoncé ne doit pas répondre « C »)', () => {
    const quiz = creerQuiz();

    quiz.gererClavier(touche('c', { ctrlKey: true }));
    quiz.gererClavier(touche('c', { metaKey: true })); // Cmd+C sur Mac

    expect(quiz.moteur()!.aRepondu()).toBe(false);
  });

  it('passe à la question suivante avec Entrée, une fois la réponse donnée', () => {
    const quiz = creerQuiz();
    const premiere = quiz.moteur()!.questionCourante()!.affichage;

    quiz.gererClavier(touche('1'));
    quiz.gererClavier(touche('Enter'));

    expect(quiz.moteur()!.aRepondu()).toBe(false);
    expect(quiz.moteur()!.questionCourante()!.affichage).not.toBe(premiere);
  });

  it('laisse Entrée activer un lien qui a le focus, sans passer à la suite', () => {
    const quiz = creerQuiz();
    quiz.gererClavier(touche('1'));
    const lien = document.createElement('a');
    document.body.appendChild(lien);

    // dispatchEvent : l'événement remonte jusqu'au document, où écoute le
    // @HostListener, avec le lien pour cible (evenement.target).
    lien.dispatchEvent(touche('Enter'));

    expect(quiz.moteur()!.aRepondu()).toBe(true); // toujours sur la même question
    lien.remove();
  });
});
