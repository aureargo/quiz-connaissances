import { Component, OnInit, inject, input, signal } from '@angular/core';
import { RouterLink } from '@angular/router';

import { QuizApi } from '../../services/quiz-api';
import { Theme } from '../../models/quiz.models';

/* =========================================================================
   NiveauSelection — l'étage intermédiaire entre l'accueil et le quiz.

   On y arrive en cliquant la tuile d'un thème sur l'accueil. La page charge
   alors, À LA DEMANDE, le détail de CE seul thème (via getTheme) pour afficher
   ses niveaux (facile / moyen / expert / tous). Avantages :
     - on ne charge plus les niveaux de TOUS les thèmes au démarrage ;
     - le clic sur la tuile mène quelque part (plus de confusion "ça bug ?").
   ========================================================================= */
@Component({
  selector: 'app-niveau-selection',
  imports: [RouterLink],
  templateUrl: './niveau-selection.html',
  styleUrl: './niveau-selection.scss'
})
export class NiveauSelection implements OnInit {
  private readonly api = inject(QuizApi);

  // Paramètre d'URL :themeId, lié automatiquement (withComponentInputBinding).
  // Le nom DOIT correspondre au paramètre déclaré dans la route ('themeId').
  readonly themeId = input<string>('');

  // --- État de la page (signals) ---
  readonly theme = signal<Theme | null>(null);
  readonly chargement = signal(true);
  readonly erreur = signal<string | null>(null);

  // ngOnInit s'exécute une fois les input() prêts.
  ngOnInit(): void {
    this.api.getTheme(this.themeId()).subscribe({
      next: (theme) => {
        this.theme.set(theme);
        this.chargement.set(false);
      },
      error: () => {
        this.erreur.set(
          "Impossible de charger ce thème. Le serveur Go est-il bien démarré sur le port 8080 ?"
        );
        this.chargement.set(false);
      }
    });
  }
}
