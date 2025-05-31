import { Offer } from './offer.model';

export interface SortOption {
  value: string;
  label: string;
  field: 'price' | 'speed';
  direction: 'asc' | 'desc';
}

export const SORT_OPTIONS: SortOption[] = [
  {
    value: 'price-asc',
    label: 'Price: Low to High',
    field: 'price',
    direction: 'asc',
  },
  {
    value: 'price-desc',
    label: 'Price: High to Low',
    field: 'price',
    direction: 'desc',
  },
  {
    value: 'speed-asc',
    label: 'Speed: Low to High',
    field: 'speed',
    direction: 'asc',
  },
  {
    value: 'speed-desc',
    label: 'Speed: High to Low',
    field: 'speed',
    direction: 'desc',
  },
];

export type SortOptionValue =
  | ''
  | 'price-asc'
  | 'price-desc'
  | 'speed-asc'
  | 'speed-desc';

export function sortOffers(offers: Offer[], sortOption: SortOption): Offer[] {
  return offers.sort((a, b) => {
    let comparison = 0;

    if (sortOption.field === 'price') {
      comparison = a.monthlyCostInCent - b.monthlyCostInCent;
    } else if (sortOption.field === 'speed') {
      comparison = a.speed - b.speed;
    }

    return sortOption.direction === 'asc' ? comparison : -comparison;
  });
}
