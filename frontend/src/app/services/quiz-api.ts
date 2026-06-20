import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { Theme, Question } from '../models/quiz.models';

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

  // URL de base de notre backend Go. Le serveur autorise le CORS, donc le
  // navigateur (port 4200) peut appeler le backend (port 8080) sans souci.
  // 💡 Alternative : configurer un "proxy" Angular pour appeler simplement /api.
  private readonly baseUrl = 'http://localhost:8080/api';

  // Récupère la liste des thèmes.
  // HttpClient.get<T>() renvoie un Observable : un flux asynchrone auquel on
  // s'abonne pour recevoir la réponse quand elle arrive.
  getThemes(): Observable<Theme[]> {
    return this.http.get<Theme[]>(`${this.baseUrl}/themes`);
  }

  // Récupère les questions d'un thème donné.
  getQuestions(themeId: string): Observable<Question[]> {
    return this.http.get<Question[]>(`${this.baseUrl}/themes/${themeId}/questions`);
  }
}
