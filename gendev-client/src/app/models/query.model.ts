import { Address } from './address.model';
import { Offer } from './offer.model';

export interface Query {
  offers: Map<string, Offer>;
  address: Address;
  timestamp: number;
  sessionID: string;
  addressHash: string;
}
