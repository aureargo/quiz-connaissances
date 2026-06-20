import { Component, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';

import { QuizApi } from '../../services/quiz-api';
import { Theme } from '../../models/quiz.models';

/* =========================================================================
   ThemeSelection — la page d'accueil : choix du thème de quiz.

   Elle récupère la liste des thèmes via QuizApi, gère les états de
   chargement / erreur, et affiche les thèmes regroupés par catégorie.
   ========================================================================= */
@Component({
  selector: 'app-theme-selection',
  // imports : ce composant standalone utilise RouterLink dans son template
  // (pour les liens vers les pages de quiz).
  imports: [RouterLink],
  templateUrl: './theme-selection.html',
  styleUrl: './theme-selection.scss'
})
export class ThemeSelection {
  private readonly api = inject(QuizApi);

  // --- État de la page, géré avec des SIGNALS ---
  // Un signal est une valeur réactive : quand on la change avec .set(),
  // le template se met à jour automatiquement. C'est le cœur de la
  // réactivité dans l'Angular moderne.
  readonly themes = signal<Theme[]>([]);
  readonly chargement = signal(true);
  readonly erreur = signal<string | null>(null);

  // computed() dérive une nouvelle valeur à partir d'autres signals. Ici, on
  // regroupe les thèmes par catégorie. Le calcul se relance tout seul quand
  // `themes()` change.
  readonly themesParCategorie = computed(() => {
    const groupes = new Map<string, Theme[]>();
    for (const theme of this.themes()) {
      const liste = groupes.get(theme.categorie) ?? [];
      liste.push(theme);
      groupes.set(theme.categorie, liste);
    }
    // On transforme la Map en tableau pour pouvoir l'itérer avec @for.
    return Array.from(groupes, ([categorie, themes]) => ({ categorie, themes }));
  });

  constructor() {
    // Au démarrage du composant, on lance la requête HTTP.
    // .subscribe() "abonne" notre code au flux : next = succès, error = échec.
    this.api.getThemes().subscribe({
      next: (themes) => {
        this.themes.set(themes);
        this.chargement.set(false);
      },
      error: () => {
        this.erreur.set(
          "Impossible de charger les thèmes. Le serveur Go est-il bien démarré sur le port 8080 ?"
        );
        this.chargement.set(false);
      }
    });
  }
}
