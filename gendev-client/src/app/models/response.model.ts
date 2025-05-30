import { Offer } from "./offer.model";
import { Query } from "./query.model";

// NDJSON response types
export interface QueryResponse {
  query: Query
}

export interface OfferResponse {
  offer: Offer;
}

export type NdjsonResponse = QueryResponse | OfferResponse;