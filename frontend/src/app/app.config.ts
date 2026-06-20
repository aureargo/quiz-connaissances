import { ApplicationConfig, provideBrowserGlobalErrorListeners } from '@angular/core';
import { provideRouter, withComponentInputBinding } from '@angular/router';
import { provideHttpClient, withFetch } from '@angular/common/http';

import { routes } from './app.routes';

// appConfig liste les "providers" : les services globaux disponibles dans toute
// l'application (système d'injection de dépendances d'Angular).
export const appConfig: ApplicationConfig = {
  providers: [
    provideBrowserGlobalErrorListeners(),

    // provideRouter rend le système de routes disponible (navigation entre pages).
    // withComponentInputBinding() relie automatiquement les paramètres d'URL
    // (ex: :themeId) aux input() des composants de page.
    provideRouter(routes, withComponentInputBinding()),

    // provideHttpClient active HttpClient, le service qui fait les appels HTTP
    // vers notre API Go. withFetch() lui demande d'utiliser l'API fetch moderne.
    provideHttpClient(withFetch())
  ]
};
