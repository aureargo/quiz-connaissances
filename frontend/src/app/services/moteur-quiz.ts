import { computed, signal } from '@angular/core';

import { Question } from '../models/quiz.models';
import { entierAleatoire, melanger } from '../utils/aleatoire';

/* =========================================================================
   MoteurQuiz — le CŒUR de l'application : la logique d'une partie de quiz.

   C'est une simple classe TypeScript (PAS un service injectable) : on en crée
   une instance par partie, avec `new MoteurQuiz(questions)`. La séparer du
   composant rend la logique facile à lire et à tester.

   Elle gère les 3 règles demandées :
     1. l'ordre des questions est mélangé (file de départ aléatoire) ;
     2. les 4 choix sont re-mélangés À CHAQUE présentation d'une question ;
     3. une question ratée est REMISE EN FILE jusqu'à être réussie.
   ========================================================================= */

// Une question "préparée" pour l'affichage : ses choix sont déjà mélangés.
export interface QuestionPreparee {
  question: Question;        // la question d'origine (pour l'énoncé, l'explication...)
  choixMelanges: string[];   // les 4 choix dans un ordre aléatoire
  indexBonneReponse: number; // index de la bonne réponse DANS choixMelanges
}

export class MoteurQuiz {
  // Les questions de départ, conservées pour pouvoir "recommencer".
  private readonly questionsOriginales: Question[];

  // --- État réactif (signals) ---

  // File des questions PAS ENCORE maîtrisées. Quand elle est vide, c'est gagné.
  private readonly file = signal<Question[]>([]);

  // Nombre total de questions uniques du thème (fixé au démarrage).
  private readonly total = signal(0);

  // La question actuellement affichée (ou null si la partie est finie).
  readonly questionCourante = signal<QuestionPreparee | null>(null);

  // Index du choix cliqué par le joueur (null tant qu'il n'a pas répondu).
  readonly reponseSelectionnee = signal<number | null>(null);

  // Nombre total d'erreurs commises sur la partie (pour le bilan final).
  readonly nbErreurs = signal(0);

  // --- Valeurs DÉRIVÉES (computed) : recalculées automatiquement ---

  // A-t-on déjà répondu à la question courante ?
  readonly aRepondu = computed(() => this.reponseSelectionnee() !== null);

  // Nombre de questions maîtrisées = total - celles encore dans la file.
  readonly nbMaitrisees = computed(() => this.total() - this.file().length);

  // Nombre total de questions du thème (exposé en lecture).
  readonly nombreTotal = computed(() => this.total());

  // La partie est terminée quand il n'y a plus aucune question dans la file.
  readonly termine = computed(() => this.total() > 0 && this.file().length === 0);

  // Progression en pourcentage (pour la barre de progression).
  readonly progression = computed(() =>
    this.total() === 0 ? 0 : Math.round((this.nbMaitrisees() / this.total()) * 100)
  );

  constructor(questions: Question[]) {
    this.questionsOriginales = questions;
    this.demarrer();
  }

  /** Le joueur sélectionne un choix (son index dans choixMelanges). */
  repondre(index: number): void {
    if (this.aRepondu()) {
      return; // déjà répondu : on ignore les clics suivants
    }
    this.reponseSelectionnee.set(index);

    const courante = this.questionCourante();
    if (courante && index !== courante.indexBonneReponse) {
      // Mauvaise réponse : on incrémente le compteur d'erreurs.
      // .update() calcule la nouvelle valeur à partir de l'ancienne.
      this.nbErreurs.update((n) => n + 1);
    }
  }

  /**
   * Passe à la question suivante en appliquant la RÈGLE DE LA FILE :
   *  - réponse correcte  → la question est retirée (maîtrisée) ;
   *  - réponse incorrecte → la question est REMISE À UNE PLACE ALÉATOIRE dans
   *    la file restante, pour revenir plus tard de façon imprévisible.
   *
   * Détail important : on évite de la replacer JUSTE après (position 0), pour
   * ne pas la revoir immédiatement... sauf si c'est la seule question qui
   * reste — auquel cas elle n'a, de toute façon, nulle part d'autre où aller.
   */
  questionSuivante(): void {
    const courante = this.questionCourante();
    const choisi = this.reponseSelectionnee();
    if (!courante || choisi === null) {
      return; // sécurité : rien à faire si on n'a pas (encore) répondu
    }

    const estCorrect = choisi === courante.indexBonneReponse;

    this.file.update((file) => {
      const reste = file.slice(1); // tout sauf la question de tête
      if (estCorrect) {
        return reste; // maîtrisée : on l'abandonne
      }

      // Ratée : on choisit une position d'insertion dans `reste`. Les positions
      // valides vont de 0 (juste après) à reste.length (tout à la fin). On
      // démarre à 1 pour ne pas la remettre en tête immédiatement ; mais si
      // `reste` est vide, la seule place possible est 0 (elle reste seule).
      const positionMin = reste.length === 0 ? 0 : 1;
      const position = entierAleatoire(positionMin, reste.length);

      const nouvelle = [...reste];
      nouvelle.splice(position, 0, courante.question);
      return nouvelle;
    });

    this.preparerCourante();
  }

  /** Recommence une partie depuis zéro (nouveau mélange, compteurs remis à 0). */
  recommencer(): void {
    this.demarrer();
  }

  // --- Méthodes privées ---

  /** (Re)démarre une partie : mélange initial des questions. */
  private demarrer(): void {
    this.total.set(this.questionsOriginales.length);
    this.file.set(melanger(this.questionsOriginales)); // ordre aléatoire (règle 1)
    this.nbErreurs.set(0);
    this.preparerCourante();
  }

  /**
   * Prépare la question en tête de file pour l'affichage : on mélange ses 4
   * choix (règle 2) et on recalcule où se trouve la bonne réponse.
   *
   * Astuce : on mélange un tableau d'INDEX [0,1,2,3] plutôt que les textes.
   * Ainsi, même si deux choix avaient le même texte, on retrouve sans
   * ambiguïté la position de la bonne réponse après mélange.
   */
  private preparerCourante(): void {
    this.reponseSelectionnee.set(null); // on réinitialise pour la nouvelle question

    const file = this.file();
    if (file.length === 0) {
      this.questionCourante.set(null); // plus de question → partie terminée
      return;
    }

    const question = file[0];
    const indexMelanges = melanger(question.choix.map((_, i) => i)); // ex: [2, 0, 3, 1]
    const choixMelanges = indexMelanges.map((i) => question.choix[i]);
    const indexBonneReponse = indexMelanges.indexOf(question.bonneReponse);

    this.questionCourante.set({ question, choixMelanges, indexBonneReponse });
  }
}
