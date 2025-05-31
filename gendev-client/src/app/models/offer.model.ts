export interface VoucherDetails {
  voucherType: 'ABSOLUTE' | 'PERCENTAGE';
  voucherValue: number;
  voucherDescription?: string;
}

export interface Offer {
  provider: string;
  productId?: number;
  productName: string;
  speed: number;
  contractDurationInMonths: number;
  connectionType: string;
  tv?: string;
  limitInGb?: number;
  maxAgePerson?: number;
  monthlyCostInCent: number;
  afterTwoYearsMonthlyCost?: number;
  monthlyCostInCentWithVoucher?: number;
  installationService: boolean;
  voucherDetails?: VoucherDetails;
  
  isPreliminary: boolean;
  offerHash: string;
}
