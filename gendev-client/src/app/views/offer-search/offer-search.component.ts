import { Component, computed, OnInit, OnDestroy } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { CommonModule } from '@angular/common';
import { Address } from '../../models/address.model';
import { State } from '../../services/state';
import { SearchService } from '../../services/search.service';
import { NdjsonResponse } from '../../models/response.model';
import { Subscription } from 'rxjs';

import { OfferResultsComponent } from '../offer-results/offer-results.component';

@Component({
  selector: 'app-offer-search',
  imports: [CommonModule, OfferResultsComponent],
  templateUrl: './offer-search.component.html',
  styleUrl: './offer-search.component.css',
  standalone: true,
})
export class OfferSearchComponent implements OnInit, OnDestroy {
  private searchApiSubscription?: Subscription;

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    protected state: State,
    private searchService: SearchService
  ) {}

  ngOnInit() {
    // Subscribe to query parameter changes (for regular searches)
    this.route.queryParams.subscribe((queryParams) => {
      if (
        queryParams['street'] &&
        queryParams['houseNumber'] &&
        queryParams['city'] &&
        queryParams['zipCode']
      ) {
        const address: Address = {
          street: queryParams['street'],
          houseNumber: queryParams['houseNumber'],
          city: queryParams['city'],
          zipCode: queryParams['zipCode'],
        };

        // Start the search
        this.startSearch(address);
      } else {
        // If no address and no shared UUID, redirect to home
        console.error('No address redirecting to home');
        this.router.navigate(['/']);
      }
    });
  }

  ngOnDestroy() {
    this.searchApiSubscription?.unsubscribe();
  }

  startSearch(address: Address) {
    const sessionId = this.state.sessionId();

    if (!sessionId) {
      this.router.navigate(['/']);
      return;
    }

    this.searchApiSubscription?.unsubscribe()
    this.state.resetState();
    this.state.setLoading(true);

    this.searchApiSubscription = this.searchService
      .searchOffers(address, sessionId)
      .subscribe({
        next: (response: NdjsonResponse) => {
          this.state.handleNdjsonResponse(response);
        },
        error: (error) => {
          console.error('Search error:', error);
          this.state.setError('Failed to fetch offers. Please try again.');
          this.state.setLoading(false);
        },
        complete: () => {
          this.state.setLoading(false);
        },
      });
  }
}
