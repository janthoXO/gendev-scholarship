import { Component, OnDestroy, OnInit, signal } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { CommonModule } from '@angular/common';
import { State } from '../../services/state';
import { SearchService } from '../../services/search.service';
import { OfferResultsComponent } from '../offer-results/offer-results.component';
import { Subscription } from 'rxjs';
import { NdjsonResponse } from '../../models/response.model';

@Component({
  selector: 'app-offer-shared',
  imports: [CommonModule, OfferResultsComponent],
  templateUrl: './offer-shared.component.html',
  styleUrl: './offer-shared.component.css',
  standalone: true,
})
export class OfferSharedComponent implements OnInit, OnDestroy {
  private shareApiSubscription?: Subscription;

  isLoading = signal<boolean>(false);
  error = signal<string | null>(null);

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    protected state: State,
    private searchService: SearchService
  ) {}

  ngOnInit() {
    // Subscribe to route parameter changes (for shared routes)
    this.route.params.subscribe((params) => {
      if (params['shareId']) {
        this.loadSharedOffers(params['shareId']);
      } else {
        // If no address and no shared UUID, redirect to home
        console.error('No shared UUID provided, redirecting to home');
        this.router.navigate(['/']);
      }
    });
  }

  ngOnDestroy() {
    // Clean up subscription to avoid memory leaks
    this.shareApiSubscription?.unsubscribe();
  }

  loadSharedOffers(shareId: string) {
    if (!shareId) return;

    this.state.resetState();
    this.isLoading.set(true);

    this.shareApiSubscription = this.searchService
      .getSharedOffers(shareId)
      .subscribe({
        next: (response: NdjsonResponse) => {
          this.state.handleNdjsonResponse(response);
        },
        error: (error) => {
          console.error('Shared offers error:', error);
          this.error.set('Failed to fetch shared offers. Please try again.');
          this.isLoading.set(false);
        },
        complete: () => {
          this.isLoading.set(false);
        },
      });
  }
}
