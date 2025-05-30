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
  voucherType?: string;
  voucherValue?: number;
  
  isPreliminary: boolean;
  offerHash: string;
}
