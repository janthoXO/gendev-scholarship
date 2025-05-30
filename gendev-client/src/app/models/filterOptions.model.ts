import { Offer } from "./offer.model";

export interface FilterOptions {
  provider?: string;
  installation?: boolean;
  speedMin?: number;
  age?: number;
  costMax?: number;
  connectionType?: string;
}

export function isFilterEmpty(filterOptions: FilterOptions): boolean {
  return (
    !filterOptions.provider &&
    !filterOptions.installation &&
    !filterOptions.speedMin &&
    !filterOptions.age &&
    !filterOptions.costMax &&
    !filterOptions.connectionType
  );
}

export function filterOffer(offer: Offer, filterOptions: FilterOptions): boolean {
  if (filterOptions.provider && offer.provider !== filterOptions.provider) {
    return false;
  }
  if (filterOptions.installation !== undefined && offer.installationService !== filterOptions.installation) {
    return false;
  }
  if (filterOptions.speedMin !== undefined && offer.speed < filterOptions.speedMin) {
    return false;
  }
  if (filterOptions.age !== undefined && offer.maxAgePerson && offer.maxAgePerson < filterOptions.age) {
    return false;
  }
  if (filterOptions.costMax !== undefined && offer.monthlyCostInCent > filterOptions.costMax * 100) {
    return false;
  }
  if (filterOptions.connectionType && offer.connectionType !== filterOptions.connectionType) {
    return false;
  }
  
  return true;
}