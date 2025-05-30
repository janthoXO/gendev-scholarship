import { Injectable, signal, computed } from '@angular/core';
import { Address } from '../models/address.model';
import { Offer } from '../models/offer.model';
import {
  NdjsonResponse,
  QueryResponse,
  OfferResponse,
} from '../models/response.model';
import { Query } from '../models/query.model';

@Injectable({
  providedIn: 'root',
})
export class State {
  // Signals for state management
  private readonly _query = signal<Query | null>(null);
  private readonly _isLoading = signal<boolean>(false);
  private readonly _error = signal<string | null>(null);
  private readonly _sessionId = signal<string>(this.generateSessionId());

  // Read-only computed signals
  readonly query = this._query.asReadonly();
  readonly isLoading = this._isLoading.asReadonly();
  readonly error = this._error.asReadonly();
  readonly sessionId = this._sessionId.asReadonly();

  // Computed values
  readonly offerCount = computed(() => this._query()?.offers?.size ?? 0);
  readonly hasOffers = computed(() => this.offerCount() > 0);

  // Actions
  setSessionId(sessionId: string) {
    this._sessionId.set(sessionId);
  }

  setQuery(query: Query) {
    this._query.set(query);
  }

  addOffer(offer: Offer) {
    const currentQuery = this._query();
    if (!currentQuery) {
      console.log('No current query to add offer to');
      return;
    }

    const updatedOffers = new Map(currentQuery.offers);
    updatedOffers.set(offer.offerHash, offer);

    this._query.set({
      ...currentQuery,
      offers: updatedOffers,
      timestamp: currentQuery.timestamp,
      sessionID: currentQuery.sessionID,
      addressHash: currentQuery.addressHash,
    });
  }

  setLoading(loading: boolean) {
    this._isLoading.set(loading);
  }

  setError(error: string | null) {
    this._error.set(error);
  }

  // Handle NDJSON response
  handleNdjsonResponse(response: NdjsonResponse) {
    if ('query' in response) {
      // Handle query response
      const queryResponse = response as QueryResponse;
      this.setQuery(queryResponse.query);
      this.setSessionId(queryResponse.query.sessionID);
    } else if ('offer' in response) {
      // Handle individual offer response
      const offerResponse = response as OfferResponse;
      this.addOffer(offerResponse.offer);
    }
  }

  resetState() {
    this._query.set(null);
    this._isLoading.set(false);
    this._error.set(null);
    this._sessionId.set('');
  }

  generateSessionId(): string {
    return (
      Math.random().toString(36).substring(2, 15) +
      Math.random().toString(36).substring(2, 15)
    );
  }
}
