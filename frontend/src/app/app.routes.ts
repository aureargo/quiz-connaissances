import { Routes } from '@angular/router';

/* =========================================================================
   Définition des ROUTES de l'application (quelle URL affiche quel composant).

   On utilise `loadComponent` avec un `import()` dynamique : c'est du
   "lazy loading". Le code d'une page n'est téléchargé par le navigateur que
   lorsqu'on visite cette page → démarrage plus rapide.
   ========================================================================= */
export const routes: Routes = [
  {
    // URL racine "/" → page de sélection du thème
    path: '',
    loadComponent: () =>
      import('./pages/theme-selection/theme-selection').then((m) => m.ThemeSelection),
    title: 'Quiz Connaissances'
  },
  {
    // "/themes/<id>" → sous-page de sélection du NIVEAU pour un thème donné.
    // C'est l'"étage" intermédiaire : on y arrive en cliquant une tuile de
    // l'accueil, et c'est SEULEMENT là qu'on charge les niveaux de ce thème.
    // :themeId est lié à l'input() du composant (withComponentInputBinding).
    path: 'themes/:themeId',
    loadComponent: () =>
      import('./pages/niveau-selection/niveau-selection').then((m) => m.NiveauSelection),
    title: 'Choix du niveau'
  },
  {
    // "/quiz/<id>/<niveau>" → page de quiz pour un thème À UN NIVEAU donné.
    // :themeId et :niveau sont des paramètres d'URL, liés automatiquement aux
    // input() du composant grâce à withComponentInputBinding (voir app.config.ts).
    path: 'quiz/:themeId/:niveau',
    loadComponent: () => import('./pages/quiz/quiz').then((m) => m.Quiz),
    title: 'Quiz en cours'
  },
  {
    // Toute autre URL inconnue → on redirige vers l'accueil.
    path: '**',
    redirectTo: ''
  }
];
