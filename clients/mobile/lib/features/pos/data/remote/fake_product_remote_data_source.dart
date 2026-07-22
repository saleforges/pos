import 'dart:math';
import '../../../../../shared/models/product.dart';
import 'product_remote_data_source.dart';

class FakeProductRemoteDataSource extends ProductRemoteDataSource {
  @override
  Future<List<Product>> getProducts() async {
    await Future.delayed(Duration(milliseconds: 500 + Random().nextInt(500)));
    return _generateProducts();
  }

  @override
  Future<Product> getProduct(String id) async {
    await Future.delayed(const Duration(milliseconds: 300));
    return _generateProducts().firstWhere((p) => p.id == id);
  }

  List<Product> _generateProducts() {
    final now = DateTime.now();
    return [
      // Coffee (8 products)
      Product(id: 'cof-01', name: 'Espresso', sku: 'COF-001', barcode: '8991001001', description: 'Single shot espresso with rich crema', sellingPrice: 25000, costPrice: 15000, stock: 100, category: 'Coffee', unit: 'cup', createdAt: now, updatedAt: now),
      Product(id: 'cof-02', name: 'Cappuccino', sku: 'COF-002', barcode: '8991001002', description: 'Espresso with steamed milk foam', sellingPrice: 35000, costPrice: 20000, stock: 80, category: 'Coffee', unit: 'cup', createdAt: now, updatedAt: now),
      Product(id: 'cof-03', name: 'Latte', sku: 'COF-003', barcode: '8991001003', description: 'Smooth espresso with steamed milk', sellingPrice: 35000, costPrice: 20000, stock: 90, category: 'Coffee', unit: 'cup', createdAt: now, updatedAt: now),
      Product(id: 'cof-04', name: 'Mocha', sku: 'COF-004', barcode: '8991001004', description: 'Espresso with chocolate and steamed milk', sellingPrice: 40000, costPrice: 22000, stock: 60, category: 'Coffee', unit: 'cup', createdAt: now, updatedAt: now),
      Product(id: 'cof-05', name: 'Cold Brew', sku: 'COF-005', barcode: '8991001005', description: 'Slow-steeped cold brew coffee', sellingPrice: 30000, costPrice: 18000, stock: 45, category: 'Coffee', unit: 'cup', createdAt: now, updatedAt: now),
      Product(id: 'cof-06', name: 'Americano', sku: 'COF-006', barcode: '8991001006', description: 'Espresso diluted with hot water', sellingPrice: 25000, costPrice: 14000, stock: 120, category: 'Coffee', unit: 'cup', createdAt: now, updatedAt: now),
      Product(id: 'cof-07', name: 'Flat White', sku: 'COF-007', barcode: '8991001007', description: 'Espresso with microfoam milk', sellingPrice: 38000, costPrice: 21000, stock: 55, category: 'Coffee', unit: 'cup', createdAt: now, updatedAt: now),
      Product(id: 'cof-08', name: 'Affogato', sku: 'COF-008', barcode: '8991001008', description: 'Espresso poured over vanilla ice cream', sellingPrice: 45000, costPrice: 25000, stock: 30, category: 'Coffee', unit: 'cup', createdAt: now, updatedAt: now),

      // Tea (6 products)
      Product(id: 'tea-01', name: 'Green Tea', sku: 'TEA-001', barcode: '8992002001', description: 'Premium Japanese green tea', sellingPrice: 18000, costPrice: 8000, stock: 100, category: 'Tea', unit: 'cup', createdAt: now, updatedAt: now),
      Product(id: 'tea-02', name: 'Earl Grey', sku: 'TEA-002', barcode: '8992002002', description: 'Classic bergamot black tea', sellingPrice: 20000, costPrice: 9000, stock: 85, category: 'Tea', unit: 'cup', createdAt: now, updatedAt: now),
      Product(id: 'tea-03', name: 'Jasmine Tea', sku: 'TEA-003', barcode: '8992002003', description: 'Fragrant jasmine green tea', sellingPrice: 22000, costPrice: 10000, stock: 70, category: 'Tea', unit: 'cup', createdAt: now, updatedAt: now),
      Product(id: 'tea-04', name: 'Matcha Latte', sku: 'TEA-004', barcode: '8992002004', description: 'Creamy matcha with steamed milk', sellingPrice: 38000, costPrice: 20000, stock: 50, category: 'Tea', unit: 'cup', createdAt: now, updatedAt: now),
      Product(id: 'tea-05', name: 'Chai Tea', sku: 'TEA-005', barcode: '8992002005', description: 'Spiced Indian chai with milk', sellingPrice: 25000, costPrice: 12000, stock: 65, category: 'Tea', unit: 'cup', createdAt: now, updatedAt: now),
      Product(id: 'tea-06', name: 'Lemon Tea', sku: 'TEA-006', barcode: '8992002006', description: 'Refreshing lemon-infused black tea', sellingPrice: 15000, costPrice: 7000, stock: 110, category: 'Tea', unit: 'cup', createdAt: now, updatedAt: now),

      // Bakery (6 products)
      Product(id: 'bak-01', name: 'Croissant', sku: 'BAK-001', barcode: '8993003001', description: 'Buttery flaky French croissant', sellingPrice: 15000, costPrice: 7000, stock: 40, category: 'Bakery', unit: 'pcs', createdAt: now, updatedAt: now),
      Product(id: 'bak-02', name: 'Banana Bread', sku: 'BAK-002', barcode: '8993003002', description: 'Moist banana loaf', sellingPrice: 12000, costPrice: 6000, stock: 25, category: 'Bakery', unit: 'slice', createdAt: now, updatedAt: now),
      Product(id: 'bak-03', name: 'Cinnamon Roll', sku: 'BAK-003', barcode: '8993003003', description: 'Soft cinnamon roll with cream cheese glaze', sellingPrice: 20000, costPrice: 10000, stock: 35, category: 'Bakery', unit: 'pcs', createdAt: now, updatedAt: now),
      Product(id: 'bak-04', name: 'Muffin Blueberry', sku: 'BAK-004', barcode: '8993003004', description: 'Fresh blueberry muffin with crumb topping', sellingPrice: 15000, costPrice: 7000, stock: 30, category: 'Bakery', unit: 'pcs', createdAt: now, updatedAt: now),
      Product(id: 'bak-05', name: 'Baguette', sku: 'BAK-005', barcode: '8993003005', description: 'Crusty French baguette', sellingPrice: 18000, costPrice: 8000, stock: 20, category: 'Bakery', unit: 'pcs', createdAt: now, updatedAt: now),
      Product(id: 'bak-06', name: 'Danish Pastry', sku: 'BAK-006', barcode: '8993003006', description: 'Flaky pastry with fruit filling', sellingPrice: 22000, costPrice: 11000, stock: 28, category: 'Bakery', unit: 'pcs', createdAt: now, updatedAt: now),

      // Dessert (6 products)
      Product(id: 'des-01', name: 'Cheesecake', sku: 'DES-001', barcode: '8994004001', description: 'Creamy New York style cheesecake', sellingPrice: 35000, costPrice: 18000, stock: 15, category: 'Dessert', unit: 'slice', createdAt: now, updatedAt: now),
      Product(id: 'des-02', name: 'Chocolate Lava', sku: 'DES-002', barcode: '8994004002', description: 'Warm chocolate cake with molten center', sellingPrice: 30000, costPrice: 15000, stock: 20, category: 'Dessert', unit: 'pcs', createdAt: now, updatedAt: now),
      Product(id: 'des-03', name: 'Tiramisu', sku: 'DES-003', barcode: '8994004003', description: 'Classic Italian coffee-flavored dessert', sellingPrice: 32000, costPrice: 16000, stock: 18, category: 'Dessert', unit: 'slice', createdAt: now, updatedAt: now),
      Product(id: 'des-04', name: 'Panna Cotta', sku: 'DES-004', barcode: '8994004004', description: 'Silky Italian cream dessert with berry sauce', sellingPrice: 28000, costPrice: 14000, stock: 22, category: 'Dessert', unit: 'pcs', createdAt: now, updatedAt: now),
      Product(id: 'des-05', name: 'Ice Cream Sundae', sku: 'DES-005', barcode: '8994004005', description: 'Vanilla ice cream with chocolate syrup and nuts', sellingPrice: 25000, costPrice: 12000, stock: 35, category: 'Dessert', unit: 'pcs', createdAt: now, updatedAt: now),
      Product(id: 'des-06', name: 'Fruit Tart', sku: 'DES-006', barcode: '8994004006', description: 'Crispy tart shell with custard and fresh fruits', sellingPrice: 30000, costPrice: 15000, stock: 12, category: 'Dessert', unit: 'slice', createdAt: now, updatedAt: now),

      // Grocery (6 products)
      Product(id: 'gro-01', name: 'Mineral Water 600ml', sku: 'GRO-001', barcode: '8995005001', description: 'Pure drinking water 600ml bottle', sellingPrice: 5000, costPrice: 2500, stock: 200, category: 'Grocery', unit: 'bottle', createdAt: now, updatedAt: now),
      Product(id: 'gro-02', name: 'Sparkling Water', sku: 'GRO-002', barcode: '8995005002', description: 'Carbonated mineral water 330ml', sellingPrice: 8000, costPrice: 4000, stock: 150, category: 'Grocery', unit: 'can', createdAt: now, updatedAt: now),
      Product(id: 'gro-03', name: 'Orange Juice', sku: 'GRO-003', barcode: '8995005003', description: 'Fresh squeezed orange juice 250ml', sellingPrice: 12000, costPrice: 6000, stock: 80, category: 'Grocery', unit: 'bottle', createdAt: now, updatedAt: now),
      Product(id: 'gro-04', name: 'Mixed Nuts', sku: 'GRO-004', barcode: '8995005004', description: 'Roasted mixed nuts 100g pack', sellingPrice: 15000, costPrice: 9000, stock: 60, category: 'Grocery', unit: 'pack', createdAt: now, updatedAt: now),
      Product(id: 'gro-05', name: 'Protein Bar', sku: 'GRO-005', barcode: '8995005005', description: 'Chocolate protein bar 50g', sellingPrice: 18000, costPrice: 11000, stock: 45, category: 'Grocery', unit: 'pcs', createdAt: now, updatedAt: now),
      Product(id: 'gro-06', name: 'Potato Chips', sku: 'GRO-006', barcode: '8995005006', description: 'Classic salted potato chips 75g', sellingPrice: 10000, costPrice: 5000, stock: 120, category: 'Grocery', unit: 'pack', createdAt: now, updatedAt: now),
    ];
  }
}
