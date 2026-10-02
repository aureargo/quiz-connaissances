import { Component, computed, inject, signal } from '@angular/core';
import { NgTemplateOutlet } from '@angular/common';
import { Router, RouterLink } from '@angular/router';

import { QuizApi } from '../../services/quiz-api';
import { Theme } from '../../models/quiz.models';
import { messageErreur } from '../../utils/erreurs';
import { rechercherThemes } from '../../utils/recherche';

/* =========================================================================
   ThemeSelection — la page d'accueil : choix du thème de quiz.

   Elle récupère la liste des thèmes via QuizApi, gère les états de
   chargement / erreur, et affiche les thèmes regroupés par catégorie.
   Un champ de recherche permet de n'afficher QUE les thèmes correspondants,
   remontés en haut de la page et classés par pertinence.
   ========================================================================= */
@Component({
  selector: 'app-theme-selection',
  // imports : ce composant standalone utilise RouterLink dans son template
  // (pour les liens vers les pages de quiz) et NgTemplateOutlet (pour
  // réutiliser le modèle de carte dans les deux vues).
  imports: [RouterLink, NgTemplateOutlet],
  templateUrl: './theme-selection.html',
  styleUrl: './theme-selection.scss'
})
export class ThemeSelection {
  private readonly api = inject(QuizApi);
  private readonly router = inject(Router);

  // --- État de la page, géré avec des SIGNALS ---
  // Un signal est une valeur réactive : quand on la change avec .set(),
  // le template se met à jour automatiquement. C'est le cœur de la
  // réactivité dans l'Angular moderne.
  readonly themes = signal<Theme[]>([]);
  readonly chargement = signal(true);
  readonly erreur = signal<string | null>(null);

  // Texte du champ de recherche, mis à jour à CHAQUE frappe (ajout ou
  // suppression d'une lettre) via l'événement (input) du template.
  readonly recherche = signal('');

  // Les thèmes correspondant à la recherche, du plus au moins pertinent.
  // Comme il dépend de `recherche()` et `themes()`, ce computed se recalcule
  // tout seul à chaque lettre tapée : aucune « actualisation » manuelle.
  readonly resultats = computed(() => rechercherThemes(this.themes(), this.recherche()));

  // Vrai dès que le champ contient autre chose que des espaces.
  readonly rechercheActive = computed(() => this.recherche().trim() !== '');

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

  /** Entrée dans le champ : ouvre directement le premier résultat. */
  ouvrirPremierResultat(): void {
    const premier = this.resultats()[0];
    if (premier) {
      this.router.navigate(['/themes', premier.id]);
    }
  }

  /** Échap dans le champ : on vide la recherche (toutes les tuiles reviennent). */
  effacerRecherche(): void {
    this.recherche.set('');
  }

  constructor() {
    // Au démarrage du composant, on lance la requête HTTP.
    // .subscribe() "abonne" notre code au flux : next = succès, error = échec.
    this.api.getThemes().subscribe({
      next: (themes) => {
        this.themes.set(themes);
        this.chargement.set(false);
      },
      error: (err) => {
        this.erreur.set(
          messageErreur(err, "La liste des thèmes est introuvable sur le serveur.")
        );
        this.chargement.set(false);
      }
    });
  }
}
