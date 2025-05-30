import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { Address } from '../models/address.model';
import { NdjsonResponse } from '../models/response.model';
import { FilterOptions } from '../models/filterOptions.model';
import ndjsonStream from 'can-ndjson-stream';
import { environment } from '../../environments/environment';

@Injectable({
  providedIn: 'root',
})
export class SearchService {
  private readonly apiUrl = environment.apiUrl;

  constructor() {}

  searchOffers(
    address: Address,
    sessionId: string
  ): Observable<NdjsonResponse> {
    const params = new URLSearchParams({
      street: address.street,
      houseNumber: address.houseNumber,
      city: address.city,
      plz: address.zipCode,
      sessionId: sessionId,
    });
    const url = `${this.apiUrl}/offers?${params.toString()}`;

    return new Observable<NdjsonResponse>((observer) => {
      fetch(url, {})
        .then((response) => {
          if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
          }
          return ndjsonStream(response.body); //ndjsonStream parses the response.body
        })
        .then((exampleStream) => {
          const reader = exampleStream.getReader();
          let read: any;
          reader.read().then(
            (read = (result: any) => {
              if (result.done) {
                console.log('Stream completed');
                observer.complete();
                return;
              }

              const jsonData = result.value as NdjsonResponse;
              observer.next(jsonData);
              reader.read().then(read);
            })
          );
        })
        .catch((error: any) => {
          console.error('Fetch failed:', error);
          observer.error(error);
        });

      // Cleanup function
      return () => {};
    });
  }

  shareOffers(
    queryHash: string,
    sessionId: string,
    filters?: FilterOptions
  ): Observable<string> {
    const params = new URLSearchParams({
      sessionId: sessionId,
    });
    const url = `${
      this.apiUrl
    }/offers/shared/${queryHash}?${params.toString()}`;

    // Prepare the request body
    const requestBody = JSON.stringify(filters || {});

    return new Observable<string>((observer) => {
      fetch(url, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: requestBody,
      })
        .then((response) => {
          if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
          }

          return response.json();
        })
        .then((data) => {
          if (data && data.shareId) {
            observer.next(data.shareId);
            observer.complete();
          } else {
            throw new Error('Response does not contain a link field');
          }
        })
        .catch((error) => {
          console.error('Share request failed:', error);
          observer.error(error);
        });

      // Cleanup function
      return () => {};
    });
  }

  getSharedOffers(sharedId: string): Observable<NdjsonResponse> {
    const url = `${this.apiUrl}/offers/shared/${sharedId}`;

    return new Observable<NdjsonResponse>((observer) => {
      fetch(url, {})
        .then((response) => {
          if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
          }
          return ndjsonStream(response.body); //ndjsonStream parses the response.body
        })
        .then((exampleStream) => {
          const reader = exampleStream.getReader();
          let read: any;
          reader.read().then(
            (read = (result: any) => {
              if (result.done) {
                console.log('Stream completed');
                observer.complete();
                return;
              }

              const jsonData = result.value as NdjsonResponse;
              observer.next(jsonData);
              reader.read().then(read);
            })
          );
        })
        .catch((error: any) => {
          console.error('Fetch failed:', error);
          observer.error(error);
        });

      // Cleanup function
      return () => {};
    });
  }

  generateSessionId(): string {
    return (
      Math.random().toString(36).substring(2, 15) +
      Math.random().toString(36).substring(2, 15)
    );
  }
}
