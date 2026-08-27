export interface VariantRow {
  key: string;
  /** Present when the row maps to an existing product item (edit mode). */
  itemId?: number;
  name: string;
  sku: string;
  unitId: string;
  price: string;
  trackInventory: boolean;
}

let tempKey = 0;

export function emptyVariantRow(): VariantRow {
  tempKey += 1;
  return { key: `new-${tempKey}`, name: '', sku: '', unitId: '', price: '', trackInventory: false };
}