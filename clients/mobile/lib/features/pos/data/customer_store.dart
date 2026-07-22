import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../shared/models/customer.dart';

class CustomerStore {
  List<Customer> _customers = [
    Customer(
      id: 'C001', name: 'Toko ABC', phone: '081234567890',
      customPrices: {'1': 3000, '3': 3500},
    ),
    Customer(
      id: 'C002', name: 'Warung Jaya', phone: '081234567891',
      customPrices: {'2': 2800, '4': 4500},
    ),
    Customer(
      id: 'C003', name: 'Budi (Member)', phone: '081234567892',
      customPrices: {'5': 16000, '6': 20000},
    ),
    Customer(
      id: 'C004', name: 'Sari Snack', phone: '081234567893',
      customPrices: {'1': 3200, '2': 3200, '3': 3800, '4': 4800},
    ),
    Customer(
      id: 'C005', name: 'Rina Catering', phone: '081234567894',
      customPrices: {'7': 13000, '8': 30000},
    ),
    Customer(
      id: 'C006', name: 'Umum', phone: '',
    ),
  ];

  List<Customer> get customers => List.unmodifiable(_customers);

  List<Customer> search(String query) {
    if (query.isEmpty) return [];
    final q = query.toLowerCase();
    return _customers.where((c) => c.name.toLowerCase().contains(q) && c.name != 'Umum').take(5).toList();
  }

  Customer? findByName(String name) {
    return _customers.cast<Customer?>().firstWhere(
      (c) => c!.name.toLowerCase() == name.toLowerCase(),
      orElse: () => null,
    );
  }

  Customer add(String name, String phone) {
    final id = 'C${DateTime.now().millisecondsSinceEpoch}';
    final c = Customer(id: id, name: name, phone: phone);
    _customers.add(c);
    return c;
  }

  void setCustomPrice(String customerId, String productId, int price) {
    final idx = _customers.indexWhere((c) => c.id == customerId);
    if (idx == -1) return;
    final updated = Map<String, int>.from(_customers[idx].customPrices);
    updated[productId] = price;
    _customers[idx] = Customer(
      id: _customers[idx].id,
      name: _customers[idx].name,
      phone: _customers[idx].phone,
      customPrices: updated,
    );
  }
}

final customerStoreProvider = Provider<CustomerStore>((ref) => CustomerStore());
