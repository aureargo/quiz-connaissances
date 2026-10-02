import { describe, expect, it } from 'vitest';

import { Theme } from '../models/quiz.models';
import { normaliser, rechercherThemes } from './recherche';

/* =========================================================================
   Tests de la recherche de thèmes (champ de l'accueil).
   ========================================================================= */

function theme(id: string, nom: string, categorie: string): Theme {
  return { id, nom, emoji: '', categorie, description: '' };
}

const THEMES: Theme[] = [
  theme('javascript', 'JavaScript', 'Programmation'),
  theme('typescript', 'TypeScript', 'Programmation'),
  theme('java', 'Java', 'Programmation'),
  theme('c', 'C', 'Programmation'),
  theme('cpp', 'C++', 'Programmation'),
  theme('linux', 'Lignes de commande Linux', 'Programmation'),
  theme('ecologie', 'Écologie & climat', 'Sciences'),
  theme('cyclisme', 'Cyclisme', 'Sport')
];

const ids = (themes: Theme[]) => themes.map((t) => t.id);

describe('normaliser', () => {
  it('retire les accents, la casse et les espaces autour', () => {
    expect(normaliser('  Écologie ')).toBe('ecologie');
  });
});

describe('rechercherThemes', () => {
  it('renvoie un tableau vide pour une recherche vide ou blanche', () => {
    expect(rechercherThemes(THEMES, '')).toEqual([]);
    expect(rechercherThemes(THEMES, '   ')).toEqual([]);
  });

  it('ignore la casse et les accents', () => {
    expect(ids(rechercherThemes(THEMES, 'ECOLO'))).toEqual(['ecologie']);
  });

  it('place le nom exact, puis les débuts de nom, avant le reste', () => {
    // « c » : C (exact), puis C++ / Cyclisme (le nom débute par c), puis
    // les noms dont un MOT débute par c (« commande », « climat »), et enfin
    // ceux qui contiennent juste la lettre (« javasCript », « typesCript »).
    expect(ids(rechercherThemes(THEMES, 'c'))).toEqual(
      ['c', 'cpp', 'cyclisme', 'linux', 'ecologie', 'javascript', 'typescript']
    );
  });

  it('classe un début de nom avant une simple inclusion', () => {
    // « java » : Java (nom exact) passe devant JavaScript (début de nom).
    expect(ids(rechercherThemes(THEMES, 'java'))).toEqual(['java', 'javascript']);
    // « script » n'est qu'inclus dans les deux noms.
    expect(ids(rechercherThemes(THEMES, 'script'))).toEqual(['javascript', 'typescript']);
  });

  it('trouve un mot au milieu du nom', () => {
    expect(ids(rechercherThemes(THEMES, 'linux'))).toEqual(['linux']);
  });

  it('cherche aussi dans la catégorie, en dernier', () => {
    expect(ids(rechercherThemes(THEMES, 'sport'))).toEqual(['cyclisme']);
  });

  it('renvoie un tableau vide quand rien ne correspond', () => {
    expect(rechercherThemes(THEMES, 'zzz')).toEqual([]);
  });
});
