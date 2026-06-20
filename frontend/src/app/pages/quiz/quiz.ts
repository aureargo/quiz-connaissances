import { Component, inject, input } from '@angular/core';
import { RouterLink } from '@angular/router';

/* =========================================================================
   Quiz — page de quiz (PLACEHOLDER en Phase 2).

   Le vrai moteur de quiz (mélange, file des erreurs, explications) arrivera
   en Phase 3. Pour l'instant, on se contente d'afficher le thème choisi pour
   vérifier que la navigation et le passage de paramètre d'URL fonctionnent.
   ========================================================================= */
@Component({
  selector: 'app-quiz',
  imports: [RouterLink],
  templateUrl: './quiz.html',
  styleUrl: './quiz.scss'
})
export class Quiz {
  // input() lié à un paramètre de route : grâce à "withComponentInputBinding",
  // Angular injecte automatiquement le :themeId de l'URL dans cette entrée.
  // (Activé en Phase 3 ; en attendant on le récupère aussi via le routeur.)
  readonly themeId = input<string>('');
}
