import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { Theme, Question } from '../models/quiz.models';
import { environment } from '../../environments/environment';

/* =========================================================================
   QuizApi — service qui dialogue avec l'API REST en Go.

   Un "service" Angular est une classe réutilisable (souvent pour la logique
   métier ou les accès réseau). Le décorateur @Injectable + providedIn: 'root'
   en fait un SINGLETON : une seule instance partagée dans toute l'application,
   injectable dans n'importe quel composant.
   ========================================================================= */
@Injectable({
  providedIn: 'root'
})
export class QuizApi {
  // inject() est la façon moderne de récupérer une dépendance (ici HttpClient),
  // sans passer par le constructeur.
  private readonly http = inject(HttpClient);

  // URL de base de notre backend Go. Elle vient du fichier d'environnement, PAS
  // du code : en dev c'est http://localhost:8080/api, en production c'est /api
  // (même domaine, derrière un reverse proxy). Angular échange le fichier au
  // moment de compiler — voir src/environments/environment.ts.
  //
  // Coder cette URL en dur, comme c'était le cas, rendait le build de
  // production inutilisable ailleurs que sur la machine du développeur.
  private readonly baseUrl = environment.apiUrl;

  // Récupère la liste des thèmes en version LÉGÈRE (sans les niveaux) : juste de
  // quoi afficher les tuiles de l'accueil. HttpClient.get<T>() renvoie un
  // Observable : un flux asynchrone auquel on s'abonne pour recevoir la réponse.
  getThemes(): Observable<Theme[]> {
    return this.http.get<Theme[]>(`${this.baseUrl}/themes`);
  }

  // Récupère UN thème complet, niveaux compris (ex : themeId="echecs").
  // Appelé par la sous-page de sélection du niveau : on ne charge les niveaux
  // que du thème réellement consulté, pas de tous les thèmes d'un coup.
  getTheme(themeId: string): Observable<Theme> {
    return this.http.get<Theme>(`${this.baseUrl}/themes/${encodeURIComponent(themeId)}`);
  }

  // Récupère les questions d'un thème POUR UN NIVEAU donné
  // (ex : themeId="animes", niveau="expert").
  //
  // encodeURIComponent : les noms de niveaux sont LIBRES (ce sont des noms de
  // fichiers). Un « # », un « ? » ou un « % » glissé tel quel dans l'URL en
  // changerait le sens ; encodé, il arrive intact au serveur Go.
  getQuestions(themeId: string, niveau: string): Observable<Question[]> {
    return this.http.get<Question[]>(
      `${this.baseUrl}/themes/${encodeURIComponent(themeId)}/questions/${encodeURIComponent(niveau)}`
    );
  }
}
