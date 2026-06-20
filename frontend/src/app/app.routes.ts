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
    // "/quiz/<id>" → page de quiz pour un thème (:themeId est un paramètre d'URL)
    // Le composant Quiz sera créé en Phase 3.
    path: 'quiz/:themeId',
    loadComponent: () => import('./pages/quiz/quiz').then((m) => m.Quiz),
    title: 'Quiz en cours'
  },
  {
    // Toute autre URL inconnue → on redirige vers l'accueil.
    path: '**',
    redirectTo: ''
  }
];
