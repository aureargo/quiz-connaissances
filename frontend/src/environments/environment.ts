/* =========================================================================
   environment.ts — configuration de PRODUCTION (fichier par défaut).

   Convention Angular, un peu contre-intuitive : ce fichier-ci est celui qui
   part dans le build normal (`npm run build`). C'est en DÉVELOPPEMENT qu'il est
   remplacé par environment.development.ts, grâce au "fileReplacements" déclaré
   dans angular.json (configuration "development").

   Pourquoi ce mécanisme ? Parce que l'URL de l'API n'est pas la même selon
   l'endroit où tourne l'application, et qu'un `import` doit être résolu à la
   COMPILATION : on ne peut pas décider au dernier moment. Angular échange donc
   le fichier avant de compiler.
   ========================================================================= */
export const environment = {
  production: true,

  // En production, on suppose que le frontend et l'API sont servis derrière le
  // MÊME nom de domaine, un reverse proxy (nginx, Caddy…) redirigeant /api vers
  // le serveur Go. D'où une URL RELATIVE : elle marche quel que soit le domaine,
  // et évite complètement le CORS puisqu'il n'y a plus qu'une seule origine.
  apiUrl: '/api'
};
