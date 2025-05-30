export interface Address {
  street: string;
  houseNumber: string;
  city: string;
  zipCode: string;
}

export function isAddressEmpty(address: Address): boolean {
  return (
    !address.street &&
    !address.houseNumber &&
    !address.city &&
    !address.zipCode
  );
}
