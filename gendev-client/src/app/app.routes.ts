import { Routes } from '@angular/router';

export const routes: Routes = [
  {
    path: '',
    loadComponent: () => import('./views/home/home.component').then(m => m.HomeComponent)
  },
  {
    path: 'offers',
    loadComponent: () => import('./views/offer-results/offer-results.component').then(m => m.OfferResultsComponent)
  },
  {
    path: 'offers/shared/:uuid',
    loadComponent: () => import('./views/offer-results/offer-results.component').then(m => m.OfferResultsComponent)
  },
  {
    path: '**',
    redirectTo: ''
  }
];
