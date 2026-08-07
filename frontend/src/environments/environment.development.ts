/* =========================================================================
   environment.development.ts — configuration de DÉVELOPPEMENT.

   Ce fichier remplace environment.ts pendant `npm start` (voir le
   "fileReplacements" de la configuration "development" dans angular.json).
   ========================================================================= */
export const environment = {
  production: false,

  // En dev, les deux serveurs sont distincts : Angular sur 4200, Go sur 8080.
  // L'URL doit donc être ABSOLUE, et c'est le middleware CORS du backend qui
  // autorise le navigateur à faire l'appel d'un port à l'autre.
  //
  // 💡 Si tu changes le port du backend (`go run . -addr :9090`), c'est ici
  // qu'il faut le répercuter.
  apiUrl: 'http://localhost:8080/api'
};
