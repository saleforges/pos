import 'package:flutter/material.dart';

class PosScreen extends StatefulWidget {
  const PosScreen({super.key});

  @override
  State<PosScreen> createState() => _PosScreenState();
}

class _PosScreenState extends State<PosScreen> {
  final List<Map<String, dynamic>> _products = [
    {'id': '1', 'name': 'Indomie Goreng', 'price': 3500, 'stock': 50},
    {'id': '2', 'name': 'Indomie Kuah Soto', 'price': 3500, 'stock': 30},
    {'id': '3', 'name': 'Teh Pucuk 350ml', 'price': 4000, 'stock': 24},
    {'id': '4', 'name': 'Aqua 600ml', 'price': 5000, 'stock': 20},
    {'id': '5', 'name': 'Rokok Surya 12', 'price': 18000, 'stock': 15},
    {'id': '6', 'name': 'Rokok Djarum Super', 'price': 22000, 'stock': 10},
  ];

  final List<Map<String, dynamic>> _cart = [];
  String _searchQuery = '';

  List<Map<String, dynamic>> get _filteredProducts {
    if (_searchQuery.isEmpty) return _products;
    return _products.where((p) =>
      p['name'].toLowerCase().contains(_searchQuery.toLowerCase())
    ).toList();
  }

  int get _totalPrice => _cart.fold(0, (sum, item) => sum + (item['price'] as int) * (item['quantity'] as int));

  void _addToCart(Map<String, dynamic> product) {
    setState(() {
      final existingIndex = _cart.indexWhere((item) => item['id'] == product['id']);
      if (existingIndex >= 0) {
        _cart[existingIndex]['quantity']++;
      } else {
        _cart.add({...product, 'quantity': 1});
      }
    });
  }

  void _removeFromCart(int index) {
    setState(() {
      if (_cart[index]['quantity'] > 1) {
        _cart[index]['quantity']--;
      } else {
        _cart.removeAt(index);
      }
    });
  }

  void _clearCart() {
    setState(() {
      _cart.clear();
    });
  }

  void _checkout() {
    if (_cart.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Cart is empty')),
      );
      return;
    }

    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Checkout'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Total: Rp ${_totalPrice.toStringAsFixed(0)}'),
            const SizedBox(height: 8),
            Text('Items: ${_cart.length}'),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () {
              Navigator.pop(context);
              _clearCart();
              ScaffoldMessenger.of(context).showSnackBar(
                const SnackBar(content: Text('Transaction successful!')),
              );
            },
            child: const Text('Pay'),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('POS Cashier'),
        actions: [
          IconButton(
            icon: const Icon(Icons.bluetooth),
            onPressed: () {},
            tooltip: 'Printer',
          ),
          IconButton(
            icon: const Icon(Icons.qr_code_scanner),
            onPressed: () {},
            tooltip: 'Scan Barcode',
          ),
        ],
      ),
      body: Row(
        children: [
          Expanded(
            flex: 3,
            child: Column(
              children: [
                Padding(
                  padding: const EdgeInsets.all(8.0),
                  child: TextField(
                    onChanged: (value) {
                      setState(() {
                        _searchQuery = value;
                      });
                    },
                    decoration: InputDecoration(
                      hintText: 'Search products...',
                      prefixIcon: const Icon(Icons.search),
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(8),
                      ),
                    ),
                  ),
                ),
                Expanded(
                  child: GridView.builder(
                    padding: const EdgeInsets.all(8),
                    gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                      crossAxisCount: 3,
                      childAspectRatio: 1.2,
                      crossAxisSpacing: 8,
                      mainAxisSpacing: 8,
                    ),
                    itemCount: _filteredProducts.length,
                    itemBuilder: (context, index) {
                      final product = _filteredProducts[index];
                      return Card(
                        child: InkWell(
                          onTap: () => _addToCart(product),
                          borderRadius: BorderRadius.circular(12),
                          child: Padding(
                            padding: const EdgeInsets.all(12),
                            child: Column(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                Icon(
                                  Icons.inventory_2_outlined,
                                  size: 40,
                                  color: Theme.of(context).colorScheme.primary,
                                ),
                                const SizedBox(height: 8),
                                Text(
                                  product['name'],
                                  textAlign: TextAlign.center,
                                  style: const TextStyle(fontWeight: FontWeight.bold),
                                  maxLines: 2,
                                  overflow: TextOverflow.ellipsis,
                                ),
                                const SizedBox(height: 4),
                                Text(
                                  'Rp ${product['price'].toStringAsFixed(0)}',
                                  style: TextStyle(
                                    color: Theme.of(context).colorScheme.primary,
                                    fontWeight: FontWeight.bold,
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ),
                      );
                    },
                  ),
                ),
              ],
            ),
          ),
          Expanded(
            flex: 2,
            child: Container(
              decoration: BoxDecoration(
                color: Theme.of(context).colorScheme.surface,
                border: Border(
                  left: BorderSide(
                    color: Theme.of(context).dividerColor,
                  ),
                ),
              ),
              child: Column(
                children: [
                  Container(
                    padding: const EdgeInsets.all(16),
                    color: Theme.of(context).colorScheme.primaryContainer,
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text(
                          'Cart (${_cart.length})',
                          style: TextStyle(
                            fontWeight: FontWeight.bold,
                            color: Theme.of(context).colorScheme.onPrimaryContainer,
                          ),
                        ),
                        if (_cart.isNotEmpty)
                          TextButton.icon(
                            onPressed: _clearCart,
                            icon: const Icon(Icons.delete_outline, size: 18),
                            label: const Text('Clear'),
                            style: TextButton.styleFrom(
                              foregroundColor: Colors.red,
                            ),
                          ),
                      ],
                    ),
                  ),
                  Expanded(
                    child: _cart.isEmpty
                        ? const Center(
                            child: Column(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                Icon(Icons.shopping_cart_outlined, size: 64, color: Colors.grey),
                                SizedBox(height: 16),
                                Text('Cart is empty', style: TextStyle(color: Colors.grey)),
                                Text('Tap product to add', style: TextStyle(color: Colors.grey, fontSize: 12)),
                              ],
                            ),
                          )
                        : ListView.builder(
                            padding: const EdgeInsets.all(8),
                            itemCount: _cart.length,
                            itemBuilder: (context, index) {
                              final item = _cart[index];
                              return Card(
                                child: ListTile(
                                  title: Text(
                                    item['name'],
                                    style: const TextStyle(fontWeight: FontWeight.bold),
                                  ),
                                  subtitle: Text(
                                    'Rp ${item['price'].toStringAsFixed(0)} x ${item['quantity']}',
                                  ),
                                  trailing: Row(
                                    mainAxisSize: MainAxisSize.min,
                                    children: [
                                      IconButton(
                                        icon: const Icon(Icons.remove_circle_outline),
                                        onPressed: () => _removeFromCart(index),
                                        color: Colors.red,
                                      ),
                                      Text(
                                        '${item['quantity']}',
                                        style: const TextStyle(fontWeight: FontWeight.bold),
                                      ),
                                      IconButton(
                                        icon: const Icon(Icons.add_circle_outline),
                                        onPressed: () => _addToCart(item),
                                        color: Colors.green,
                                      ),
                                    ],
                                  ),
                                ),
                              );
                            },
                          ),
                  ),
                  Container(
                    padding: const EdgeInsets.all(16),
                    decoration: BoxDecoration(
                      color: Theme.of(context).colorScheme.surface,
                      border: Border(
                        top: BorderSide(color: Theme.of(context).dividerColor),
                      ),
                    ),
                    child: Column(
                      children: [
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            const Text('Total', style: TextStyle(fontSize: 18)),
                            Text(
                              'Rp ${_totalPrice.toStringAsFixed(0)}',
                              style: TextStyle(
                                fontSize: 24,
                                fontWeight: FontWeight.bold,
                                color: Theme.of(context).colorScheme.primary,
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 16),
                        SizedBox(
                          width: double.infinity,
                          height: 56,
                          child: FilledButton.icon(
                            onPressed: _checkout,
                            icon: const Icon(Icons.payment),
                            label: const Text(
                              'PAY',
                              style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}
