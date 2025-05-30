import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { CommonModule } from '@angular/common';
import { State } from '../../services/state';
import { SearchService } from '../../services/search.service';
import { OfferResultsComponent } from '../offer-results/offer-results.component';

@Component({
  selector: 'app-offer-shared',
  imports: [CommonModule, OfferResultsComponent],
  templateUrl: './offer-shared.component.html',
  styleUrl: './offer-shared.component.css',
  standalone: true,
})
export class OfferSharedComponent implements OnInit {
  constructor(
    private route: ActivatedRoute,
    private router: Router,
    protected state: State,
    private searchService: SearchService
  ) {}

  ngOnInit() {
    // Subscribe to route parameter changes (for shared routes)
    this.route.params.subscribe((params) => {
      if (params['uuid']) {
        this.loadSharedOffers(params['uuid']);
      } else {
        // If no address and no shared UUID, redirect to home
        console.error('No shared UUID provided, redirecting to home');
        this.router.navigate(['/']);
      }
    });
  }

  loadSharedOffers(uuid: string) {
    if (!uuid) return;

    this.state.resetState();
    this.state.setLoading(true);

    // TODO: Implement API call to fetch shared offers by UUID
    // Example: this.searchService.getSharedOffers(this.sharedUuid)
    console.log('Loading shared offers for UUID:', uuid);

    // For now, just set loading to false
    // You'll need to implement this method in your SearchService
    this.state.setLoading(false);
  }
}
