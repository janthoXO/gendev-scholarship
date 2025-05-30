import {
  Component,
  computed,
  OnInit,
  OnDestroy,
  signal,
  effect,
} from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Address, isAddressEmpty } from '../../models/address.model';
import { State } from '../../services/state';
import { SearchService } from '../../services/search.service';
import { NdjsonResponse } from '../../models/response.model';
import { OfferCardComponent } from '../../components/offer-card/offer-card.component';
import { Subscription } from 'rxjs';
import { FontAwesomeModule } from '@fortawesome/angular-fontawesome';
import { faSearch } from '@fortawesome/free-solid-svg-icons';
import {
  filterOffer,
  FilterOptions,
  isFilterEmpty,
} from '../../models/filterOptions.model';

@Component({
  selector: 'app-offer-results',
  imports: [CommonModule, FormsModule, FontAwesomeModule, OfferCardComponent],
  templateUrl: './offer-results.component.html',
  styleUrl: './offer-results.component.css',
  standalone: true,
})
export class OfferResultsComponent implements OnInit, OnDestroy {
  private subscription?: Subscription;
  protected faSearch = faSearch

  private offers = computed(() => {
    const offersMap = this.state.query()?.offers;
    return offersMap ? Array.from(offersMap.values()) : [];
  });

  filter = signal<FilterOptions>({
    provider: undefined,
    installation: undefined,
    speedMin: undefined,
    costMax: undefined,
    connectionType: undefined,
  });

  addressSearch = signal<Address>({
    street: '',
    houseNumber: '',
    city: '',
    zipCode: '',
  });

  // Computed values for filter options
  availableProviders = computed(() => {
    const providers = this.offers().map((offer) => offer.provider);
    return [...new Set(providers)].sort();
  });

  availableConnectionTypes = computed(() => {
    const types = this.offers().map((offer) => offer.connectionType);
    return [...new Set(types)].sort();
  });

  // Filtered offers
  filteredOffers = computed(() => {
    let offers = this.offers();
    let filter = this.filter();

    if (filter) {
      offers = offers.filter((offer) => {
        return filterOffer(offer, filter);
      });
    }

    return offers;
  });

  // Active filters for display as bubbles
  activeFilters = computed(() => {
    const filters = [];

    if (this.filter()?.provider) {
      filters.push({
        type: 'provider',
        label: `Provider: ${this.filter()?.provider}`,
        value: this.filter()?.provider,
      });
    }

    if (this.filter()?.connectionType) {
      filters.push({
        type: 'connectionType',
        label: `Type: ${this.filter()?.connectionType}`,
        value: this.filter()?.connectionType,
      });
    }

    if (this.filter()?.speedMin) {
      filters.push({
        type: 'minSpeed',
        label: `Min Speed: ${this.filter()?.speedMin} Mbps`,
        value: this.filter()?.speedMin,
      });
    }

    if (this.filter()?.costMax) {
      filters.push({
        type: 'maxCost',
        label: `Max Cost: €${this.filter()?.costMax}`,
        value: this.filter()?.costMax,
      });
    }

    if (this.filter()?.installation) {
      filters.push({
        type: 'installation',
        label: 'Installation Required',
        value: true,
      });
    }

    return filters;
  });

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    protected state: State,
    private searchService: SearchService
  ) {
    effect(() => {
      const query = this.state.query();
      if (isAddressEmpty(this.addressSearch()) && query?.address) {
        this.addressSearch.set({
          street: query.address.street,
          houseNumber: query.address.houseNumber,
          city: query.address.city,
          zipCode: query.address.zipCode,
        });
      }
    });
  }

  ngOnInit() {
    // Check if this is a shared route first
    let sharedUuid: string | null = null;
    let address: Address | null = null;

    this.route.params.subscribe((params) => {
      if (params['uuid']) {
        sharedUuid = params['uuid'];
        console.log('Shared UUID:', sharedUuid);
      }
    });

    // If not shared route, check for query parameters (regular search)
    this.route.queryParams.subscribe((queryParams) => {
      if (
        queryParams['street'] &&
        queryParams['houseNumber'] &&
        queryParams['city'] &&
        queryParams['zipCode']
      ) {
        address = {
          street: queryParams['street'],
          houseNumber: queryParams['houseNumber'],
          city: queryParams['city'],
          zipCode: queryParams['zipCode'],
        };
      }
    });

    if (sharedUuid) {
      this.loadSharedOffers(sharedUuid);
    } else if (address) {
      this.startSearch(address);
    } else {
      // If no address and no shared UUID, redirect to home
      console.error('No address or shared UUID provided, redirecting to home');
      this.router.navigate(['/']);
    }
  }

  ngOnDestroy() {
    this.subscription?.unsubscribe();
  }

  loadSharedOffers(uuid: string) {
    if (!uuid) return;

    this.state.setLoading(true);
    this.state.setError(null);

    // TODO: Implement API call to fetch shared offers by UUID
    // Example: this.searchService.getSharedOffers(this.sharedUuid)
    console.log('Loading shared offers for UUID:', uuid);

    // For now, just set loading to false
    // You'll need to implement this method in your SearchService
    this.state.setLoading(false);
  }

  startSearch(address: Address) {
    const sessionId = this.state.sessionId();

    if (!sessionId) {
      this.router.navigate(['/']);
      return;
    }

    this.state.setLoading(true);
    this.state.setError(null);

    this.subscription = this.searchService
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

  retrySearch() {
    this.router.navigate(['/']);
  }

  clearFilters() {
    this.filter.set({
      provider: undefined,
      installation: undefined,
      speedMin: undefined,
      costMax: undefined,
      connectionType: undefined,
    });
  }

  removeFilter(filterType: string) {
    switch (filterType) {
      case 'provider':
        this.filter.set({
          ...this.filter(),
          provider: undefined,
        });
        break;
      case 'connectionType':
        this.filter.set({
          ...this.filter(),
          connectionType: undefined,
        });
        break;
      case 'minSpeed':
        this.filter.set({
          ...this.filter(),
          speedMin: undefined,
        });
        break;
      case 'maxCost':
        this.filter.set({
          ...this.filter(),
          costMax: undefined,
        });
        break;
      case 'installation':
        this.filter.set({
          ...this.filter(),
          installation: undefined,
        });
        break;
    }
  }

  hasActiveFilters(): boolean {
    return this.filter() !== null && isFilterEmpty(this.filter()!);
  }

  onSearchInput(event: Event, field: string) {
    const target = event.target as HTMLInputElement;
    switch (field) {
      case 'street':
        this.addressSearch.set({
          ...this.addressSearch(),
          street: target.value,
        });
        break;
      case 'houseNumber':
        this.addressSearch.set({
          ...this.addressSearch(),
          houseNumber: target.value,
        });
        break;
      case 'city':
        this.addressSearch.set({
          ...this.addressSearch(),
          city: target.value,
        });
        break;
      case 'zipCode':
        this.addressSearch.set({
          ...this.addressSearch(),
          zipCode: target.value,
        });
        break;
    }
  }

  isSearchFormValid(): boolean {
    return !isAddressEmpty(this.addressSearch());
  }

  onSearchSubmit() {
    if (!this.isSearchFormValid()) {
      console.warn('Please fill in all address fields');
      return;
    }

    const newAddress: Address = {
      street: this.addressSearch().street.trim(),
      houseNumber: this.addressSearch().houseNumber.trim(),
      city: this.addressSearch().city.trim(),
      zipCode: this.addressSearch().zipCode.trim(),
    };

    // Navigate to search with the new address
    this.router.navigate(['/offer-results'], {
      queryParams: {
        street: newAddress.street,
        houseNumber: newAddress.houseNumber,
        city: newAddress.city,
        zipCode: newAddress.zipCode,
      },
    });
  }

  shareQuery() {
    const currentQuery = this.state.query();
    if (!currentQuery) {
      console.error('No query to share');
      return;
    }

    // TODO: Implement share functionality
    // This should call an API to create a share link
    console.log('Sharing query:', currentQuery);

    // For now, just copy the current URL to clipboard
    const currentUrl = window.location.href;
    navigator.clipboard
      .writeText(currentUrl)
      .then(() => {
        console.log('URL copied to clipboard');
        // You could show a toast notification here
      })
      .catch((err) => {
        console.error('Failed to copy URL:', err);
      });
  }

  goBack() {
    this.router.navigate(['/']);
  }
}
