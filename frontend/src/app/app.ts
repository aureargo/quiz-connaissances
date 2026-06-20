import { Component } from '@angular/core';
import { RouterOutlet, RouterLink } from '@angular/router';

/* =========================================================================
   App — le composant RACINE de l'application (sélecteur <app-root>).

   Il fournit le "cadre" commun à toutes les pages : l'en-tête, puis une zone
   <router-outlet> où Angular injecte le composant correspondant à l'URL.
   ========================================================================= */
@Component({
  selector: 'app-root',
  imports: [RouterOutlet, RouterLink],
  templateUrl: './app.html',
  styleUrl: './app.scss'
})
export class App {}
