import { Component, computed, signal, effect, input } from '@angular/core';
import { Router } from '@angular/router';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Address, isAddressEmpty } from '../../models/address.model';
import { State } from '../../services/state';
import { SearchService } from '../../services/search.service';
import { OfferCardComponent } from '../../components/offer-card/offer-card.component';
import { ShareDialogComponent } from '../../components/share-dialog/share-dialog.component';
import { FontAwesomeModule } from '@fortawesome/angular-fontawesome';
import {
  faSearch,
  faShareAlt,
  faClose,
  faRedo,
  faExclamationTriangle,
} from '@fortawesome/free-solid-svg-icons';
import {
  filterOffer,
  FilterOptions,
  isFilterEmpty,
} from '../../models/filterOptions.model';

@Component({
  selector: 'app-offer-results',
  imports: [
    CommonModule,
    FormsModule,
    FontAwesomeModule,
    OfferCardComponent,
    ShareDialogComponent,
  ],
  templateUrl: './offer-results.component.html',
  styleUrl: './offer-results.component.css',
  standalone: true,
})
export class OfferResultsComponent {
  protected faSearch = faSearch;
  protected faShareAlt = faShareAlt;
  protected faClose = faClose;
  protected faRedo = faRedo;
  protected faExclamationTriangle = faExclamationTriangle;

  // Share dialog state
  showShareDialog = signal(false);

  private offers = computed(() => {
    const offersMap = this.state.query()?.offers;
    return offersMap ? Array.from(offersMap.values()) : [];
  });

  filter = signal<FilterOptions>({
    provider: '',
    installation: undefined,
    speedMin: undefined,
    costMax: undefined,
    connectionType: '',
  });

  addressSearch = signal<Address>({
    street: '',
    houseNumber: '',
    city: '',
    zipCode: '',
  });

  error = input<string | null>(null);
  isLoading = input<boolean>(false);
  isShareable = input<boolean>(false);

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
    const currentFilter = this.filter();

    if (currentFilter?.provider && currentFilter.provider.trim() !== '') {
      filters.push({
        type: 'provider',
        label: `Provider: ${currentFilter.provider}`,
        value: currentFilter.provider,
      });
    }

    if (
      currentFilter?.connectionType &&
      currentFilter.connectionType.trim() !== ''
    ) {
      filters.push({
        type: 'connectionType',
        label: `Type: ${currentFilter.connectionType}`,
        value: currentFilter.connectionType,
      });
    }

    if (currentFilter?.speedMin) {
      filters.push({
        type: 'minSpeed',
        label: `Min Speed: ${currentFilter.speedMin} Mbps`,
        value: currentFilter.speedMin,
      });
    }

    if (currentFilter?.costMax) {
      filters.push({
        type: 'maxCost',
        label: `Max Cost: €${currentFilter.costMax}`,
        value: currentFilter.costMax,
      });
    }

    if (currentFilter?.installation) {
      filters.push({
        type: 'installation',
        label: 'Installation Required',
        value: true,
      });
    }

    return filters;
  });

  constructor(
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

  resetComponent() {
    this.filter.set({
      provider: '',
      installation: undefined,
      speedMin: undefined,
      costMax: undefined,
      connectionType: '',
    });

    this.addressSearch.set({
      street: '',
      houseNumber: '',
      city: '',
      zipCode: '',
    });

    this.state.resetState();
  }

  retrySearch() {
    this.router.navigate(['/']);
  }

  clearFilters() {
    this.filter.set({
      provider: '',
      installation: undefined,
      speedMin: undefined,
      costMax: undefined,
      connectionType: '',
    });
  }

  removeFilter(filterType: string) {
    switch (filterType) {
      case 'provider':
        this.filter.set({
          ...this.filter(),
          provider: '',
        });
        break;
      case 'connectionType':
        this.filter.set({
          ...this.filter(),
          connectionType: '',
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

  updateFilter(field: string, value: any) {
    const currentFilter = this.filter();
    this.filter.set({
      ...currentFilter,
      [field]: value,
    });
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

    this.resetComponent();

    // Navigate to search with the new address
    this.router.navigate(['/'], { skipLocationChange: true }).then(() => {
      this.router.navigate(['/offers'], {
        queryParams: {
          street: newAddress.street,
          houseNumber: newAddress.houseNumber,
          city: newAddress.city,
          zipCode: newAddress.zipCode,
        },
      });
    });
  }

  shareQuery() {
    const currentQuery = this.state.query();
    if (!currentQuery) {
      console.error('No query to share');
      return;
    }

    this.showShareDialog.set(true);
  }

  closeShareDialog() {
    this.showShareDialog.set(false);
  }
}
