import { Search, Plus, MoreHorizontal } from 'lucide-react';

const PRODUCTS = [
  { name: 'Espresso', category: 'Coffee', price: '$4.50', stock: 120, status: 'Active' },
  { name: 'Cappuccino', category: 'Coffee', price: '$5.00', stock: 85, status: 'Active' },
  { name: 'Latte', category: 'Coffee', price: '$5.50', stock: 90, status: 'Active' },
  { name: 'Iced Tea', category: 'Beverages', price: '$3.50', stock: 60, status: 'Active' },
  { name: 'Croissant', category: 'Pastry', price: '$4.00', stock: 30, status: 'Active' },
  { name: 'Muffin', category: 'Pastry', price: '$3.50', stock: 0, status: 'Inactive' },
  { name: 'Mineral Water', category: 'Beverages', price: '$2.00', stock: 200, status: 'Active' },
  { name: 'Sandwich', category: 'Food', price: '$7.50', stock: 15, status: 'Active' },
];

export default function ProductsPage() {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="relative flex-1 max-w-xs">
          <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-neutral-400" />
          <input
            type="text"
            placeholder="Search products..."
            className="w-full rounded-lg border border-neutral-200 bg-white py-2 pl-9 pr-3 text-sm outline-none focus:border-neutral-400"
          />
        </div>
        <button className="flex items-center gap-2 rounded-lg bg-primary px-3 py-2 text-sm text-white hover:bg-primary-hover">
          <Plus size={16} />
          Add Product
        </button>
      </div>

      <div className="rounded-lg border border-neutral-200 bg-white">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-neutral-100 text-left text-xs text-neutral-500">
                <th className="px-4 py-3 font-medium">Product</th>
                <th className="px-4 py-3 font-medium">Category</th>
                <th className="px-4 py-3 font-medium">Price</th>
                <th className="px-4 py-3 font-medium">Stock</th>
                <th className="px-4 py-3 font-medium">Status</th>
                <th className="px-4 py-3" />
              </tr>
            </thead>
            <tbody>
              {PRODUCTS.map(({ name, category, price, stock, status }) => (
                <tr key={name} className="border-b border-neutral-100 last:border-0 hover:bg-neutral-50">
                  <td className="px-4 py-3 font-medium text-neutral-900">{name}</td>
                  <td className="px-4 py-3 text-neutral-500">{category}</td>
                  <td className="px-4 py-3 text-neutral-600">{price}</td>
                  <td className="px-4 py-3">
                    <span className={stock === 0 ? 'text-danger' : 'text-neutral-600'}>{stock}</span>
                  </td>
                  <td className="px-4 py-3">
                    <span className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${
                      status === 'Active' ? 'bg-secondary/10 text-secondary' : 'bg-neutral-100 text-neutral-500'
                    }`}>
                      {status}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <button className="rounded-md p-1 text-neutral-400 hover:bg-neutral-100 hover:text-neutral-600">
                      <MoreHorizontal size={16} />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
