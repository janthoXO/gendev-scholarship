import { Routes } from '@angular/router';

export const routes: Routes = [
  {
    path: '',
    loadComponent: () => import('./views/home/home.component').then(m => m.HomeComponent)
  },
  {
    path: 'offers',
    loadComponent: () => import('./views/offer-search/offer-search.component').then(m => m.OfferSearchComponent)
  },
  {
    path: 'offers/shared/:uuid',
    loadComponent: () => import('./views/offer-shared/offer-shared.component').then(m => m.OfferSharedComponent)
  },
  {
    path: '**',
    redirectTo: ''
  }
];
