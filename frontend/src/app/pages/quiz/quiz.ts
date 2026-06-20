import { Component, HostListener, OnInit, inject, input, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { forkJoin } from 'rxjs';

import { QuizApi } from '../../services/quiz-api';
import { MoteurQuiz } from '../../services/moteur-quiz';
import { Theme } from '../../models/quiz.models';

/* =========================================================================
   Quiz — la page de jeu. Elle :
     1. charge le thème + ses questions depuis l'API ;
     2. crée un MoteurQuiz (toute la logique de jeu y est déléguée) ;
     3. affiche la question courante et réagit aux clics du joueur.

   Le composant reste "mince" : il orchestre, mais ne contient pas la logique
   de mélange / file (c'est le rôle du MoteurQuiz).
   ========================================================================= */
@Component({
  selector: 'app-quiz',
  imports: [RouterLink],
  templateUrl: './quiz.html',
  styleUrl: './quiz.scss'
})
export class Quiz implements OnInit {
  private readonly api = inject(QuizApi);

  // Paramètre d'URL :themeId, injecté automatiquement (withComponentInputBinding).
  readonly themeId = input<string>('');

  // --- État de la page ---
  readonly chargement = signal(true);
  readonly erreur = signal<string | null>(null);
  readonly themeInfo = signal<Theme | null>(null);

  // Le moteur de quiz. null tant que les questions ne sont pas chargées.
  readonly moteur = signal<MoteurQuiz | null>(null);

  // ngOnInit s'exécute une fois le composant initialisé (les input() sont prêts).
  ngOnInit(): void {
    const id = this.themeId();

    // forkJoin attend que PLUSIEURS Observables soient TOUS terminés, puis
    // renvoie leurs résultats ensemble. Ici : la liste des thèmes (pour le
    // titre/emoji) ET les questions du thème.
    forkJoin([this.api.getThemes(), this.api.getQuestions(id)]).subscribe({
      next: ([themes, questions]) => {
        this.themeInfo.set(themes.find((t) => t.id === id) ?? null);

        if (questions.length === 0) {
          this.erreur.set('Ce thème ne contient aucune question.');
        } else {
          // On instancie le moteur : la partie démarre (mélange initial).
          this.moteur.set(new MoteurQuiz(questions));
        }
        this.chargement.set(false);
      },
      error: () => {
        this.erreur.set(
          "Impossible de charger le quiz. Le serveur Go est-il démarré sur le port 8080 ?"
        );
        this.chargement.set(false);
      }
    });
  }

  /**
   * Navigation au CLAVIER (bonus confort).
   * @HostListener écoute un événement (ici keydown sur tout le document) et
   * appelle cette méthode. '$event' transmet l'objet KeyboardEvent.
   *   - touches 1–4 (ou A–D) : sélectionne une réponse ;
   *   - Entrée / Espace après réponse : passe à la question suivante.
   */
  @HostListener('document:keydown', ['$event'])
  gererClavier(evenement: KeyboardEvent): void {
    const m = this.moteur();
    if (!m || m.termine()) {
      return;
    }
    const courante = m.questionCourante();
    if (!courante) {
      return;
    }

    if (!m.aRepondu()) {
      const touche = evenement.key.toLowerCase();
      // On cherche l'index correspondant, que ce soit un chiffre ou une lettre.
      const parChiffre = ['1', '2', '3', '4'].indexOf(touche);
      const parLettre = ['a', 'b', 'c', 'd'].indexOf(touche);
      const index = parChiffre !== -1 ? parChiffre : parLettre;

      if (index !== -1 && index < courante.choixMelanges.length) {
        m.repondre(index);
        evenement.preventDefault();
      }
    } else if (evenement.key === 'Enter' || evenement.key === ' ') {
      m.questionSuivante();
      evenement.preventDefault();
    }
  }
}
