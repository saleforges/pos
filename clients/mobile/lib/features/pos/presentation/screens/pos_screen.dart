import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../shared/models/order.dart';
import '../../../../core/config/app_config.dart';
import '../../../auth/presentation/providers/auth_provider.dart';
import '../../data/print_service.dart';
import 'main_shell.dart';

class PosScreen extends StatelessWidget {
  const PosScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return const Scaffold(
      appBar: _PosAppBar(),
      body: PosScreenBody(),
    );
  }
}

class PosScreenBody extends StatefulWidget {
  const PosScreenBody({super.key});

  @override
  State<PosScreenBody> createState() => _PosBodyState();
}

class _PosAppBar extends StatelessWidget implements PreferredSizeWidget {
  const _PosAppBar();

  @override
  Size get preferredSize => const Size.fromHeight(kToolbarHeight);

  @override
  Widget build(BuildContext context) {
    return AppBar(
      title: Text(AppConfig.appName),
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
        Consumer(builder: (context, ref, _) {
          final user = ref.watch(authProvider).user;
          return Padding(
            padding: const EdgeInsets.only(left: 8),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Icon(Icons.person, size: 20, color: Colors.grey.shade600),
                const SizedBox(width: 6),
                Text(
                  user?.username ?? '',
                  style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w500),
                ),
              ],
            ),
          );
        }),
      ],
    );
  }
}

class _PosBodyState extends State<PosScreenBody> {
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

  Order _saveOrder(String method, OrderStatus status) {
    final store = ProviderScope.containerOf(context, listen: false).read(orderStoreProvider);
    final order = Order(
      id: store.nextId(),
      items: _cart.map((item) => OrderItem(
        name: item['name'],
        price: item['price'],
        quantity: item['quantity'],
      )).toList(),
      total: _totalPrice,
      status: status,
      paymentMethod: method,
      createdAt: DateTime.now(),
    );
    store.add(order);
    _clearCart();
    return order;
  }

  void _checkout() {
    if (_cart.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Cart is empty')),
      );
      return;
    }

    showModalBottomSheet(
      context: context,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (context) => SingleChildScrollView(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
            Center(
              child: Container(
                width: 40, height: 4,
                decoration: BoxDecoration(
                  color: Colors.grey.shade300,
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
            ),
            const SizedBox(height: 24),
            Text(
              'Select Payment',
              style: Theme.of(context).textTheme.titleLarge?.copyWith(
                fontWeight: FontWeight.bold,
                color: const Color(0xFF1E293B),
              ),
            ),
            const SizedBox(height: 8),
            Text(
              'Total: Rp ${_totalPrice.toStringAsFixed(0)}',
              style: TextStyle(color: Colors.grey.shade600, fontSize: 16),
            ),
            const SizedBox(height: 24),
            _PaymentOption(
              icon: Icons.money,
              title: 'Cash',
              subtitle: 'Pay with cash',
              onTap: () {
                Navigator.pop(context);
                final order = _saveOrder('Cash', OrderStatus.paid);
                PrintService.printReceipt(context, order);
              },
            ),
            const SizedBox(height: 12),
            _PaymentOption(
              icon: Icons.qr_code,
              title: 'QRIS',
              subtitle: 'Scan QR code',
              onTap: () {
                Navigator.pop(context);
                showDialog(
                  context: context,
                  builder: (context) => AlertDialog(
                    title: const Text('QRIS Payment'),
                    content: Column(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Container(
                          width: 200, height: 200,
                          decoration: BoxDecoration(
                            color: Colors.grey.shade100,
                            borderRadius: BorderRadius.circular(12),
                          ),
                          child: const Center(child: Icon(Icons.qr_code, size: 120)),
                        ),
                        const SizedBox(height: 16),
                        Text('Total: Rp ${_totalPrice.toStringAsFixed(0)}'),
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
                          final order = _saveOrder('QRIS', OrderStatus.paid);
                          PrintService.printReceipt(context, order);
                        },
                        child: const Text('Paid'),
                      ),
                    ],
                  ),
                );
              },
            ),
            const SizedBox(height: 12),
            _PaymentOption(
              icon: Icons.bookmark_border,
              title: 'Pay Later',
              subtitle: 'Customer will pay later',
              onTap: () {
                Navigator.pop(context);
                _saveOrder('Pay Later', OrderStatus.unpaid);
                ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(content: Text('Order saved as unpaid')),
                );
              },
            ),
            const SizedBox(height: 16),
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('Cancel'),
            ),
          ],
          ),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final wide = MediaQuery.of(context).size.width >= 720;

    if (!wide) return _buildMobile();
    return _buildDesktop();
  }

  Widget _buildDesktop() {
    return Row(
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
                    ),
                  ),
                ),
                Expanded(
                  child: LayoutBuilder(
                    builder: (context, constraints) {
                      final crossAxisCount = constraints.maxWidth > 600 ? 4 : constraints.maxWidth > 400 ? 3 : 2;
                      return GridView.builder(
                        padding: const EdgeInsets.all(8),
                        gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
                          crossAxisCount: crossAxisCount,
                          childAspectRatio: 0.85,
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
                            padding: const EdgeInsets.all(8),
                            child: Column(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                Icon(Icons.inventory_2_outlined, size: 24, color: Theme.of(context).colorScheme.primary),
                                const SizedBox(height: 6),
                                Text(
                                  product['name'],
                                  textAlign: TextAlign.center,
                                  style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 12),
                                  maxLines: 2,
                                  overflow: TextOverflow.ellipsis,
                                ),
                                const SizedBox(height: 2),
                                Text(
                                  'Rp ${product['price'].toStringAsFixed(0)}',
                                  style: TextStyle(
                                    color: Theme.of(context).colorScheme.primary,
                                    fontWeight: FontWeight.bold,
                                    fontSize: 12,
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ),
                      );
                    },
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
                color: Colors.white,
                border: Border(
                  left: BorderSide(color: Colors.grey.shade200),
                ),
              ),
              child: Column(
                children: [
                  Container(
                    padding: const EdgeInsets.all(16),
                    decoration: BoxDecoration(
                      color: Colors.white,
                      border: Border(bottom: BorderSide(color: Colors.grey.shade200)),
                    ),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Row(
                          children: [
                            Icon(Icons.shopping_cart_outlined, size: 20, color: Colors.grey.shade600),
                            const SizedBox(width: 8),
                            Text(
                              'Cart',
                              style: TextStyle(
                                fontWeight: FontWeight.w600,
                                fontSize: 16,
                                color: Colors.grey.shade800,
                              ),
                            ),
                            const SizedBox(width: 8),
                            Container(
                              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                              decoration: BoxDecoration(
                                color: const Color(0xFF6366F1),
                                borderRadius: BorderRadius.circular(12),
                              ),
                              child: Text(
                                '${_cart.length}',
                                style: const TextStyle(
                                  color: Colors.white,
                                  fontSize: 12,
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                            ),
                          ],
                        ),
                        if (_cart.isNotEmpty)
                          TextButton(
                            onPressed: _clearCart,
                            child: const Text('Clear', style: TextStyle(color: Colors.red)),
                          ),
                      ],
                    ),
                  ),
                  Expanded(
                    child: _cart.isEmpty
                        ? Center(
                            child: Column(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                Icon(Icons.shopping_cart_outlined, size: 64, color: Colors.grey.shade300),
                                const SizedBox(height: 16),
                                Text('Cart is empty', style: TextStyle(color: Colors.grey.shade400)),
                              ],
                            ),
                          )
                        : ListView.builder(
                            padding: const EdgeInsets.all(12),
                            itemCount: _cart.length,
                            itemBuilder: (context, index) {
                              final item = _cart[index];
                              return Padding(
                                padding: const EdgeInsets.only(bottom: 8),
                                child: Container(
                                  padding: const EdgeInsets.all(12),
                                  decoration: BoxDecoration(
                                    color: Colors.grey.shade50,
                                    borderRadius: BorderRadius.circular(12),
                                  ),
                                  child: Row(
                                    children: [
                                      Expanded(
                                        child: Column(
                                          crossAxisAlignment: CrossAxisAlignment.start,
                                          children: [
                                            Text(
                                              item['name'],
                                              style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 14),
                                            ),
                                            const SizedBox(height: 4),
                                            Text(
                                              'Rp ${((item['price'] as int) * (item['quantity'] as int)).toStringAsFixed(0)}',
                                              style: TextStyle(
                                                color: const Color(0xFF6366F1),
                                                fontWeight: FontWeight.bold,
                                              ),
                                            ),
                                          ],
                                        ),
                                      ),
                                      Row(
                                        mainAxisSize: MainAxisSize.min,
                                        children: [
                                          IconButton(
                                            icon: const Icon(Icons.remove_circle_outline, size: 22),
                                            onPressed: () => _removeFromCart(index),
                                            color: Colors.red.shade400,
                                          ),
                                          Text(
                                            '${item['quantity']}',
                                            style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
                                          ),
                                          IconButton(
                                            icon: const Icon(Icons.add_circle_outline, size: 22),
                                            onPressed: () => _addToCart(item),
                                            color: Colors.green.shade400,
                                          ),
                                        ],
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
                      color: Colors.white,
                      border: Border(top: BorderSide(color: Colors.grey.shade200)),
                    ),
                    child: Column(
                      children: [
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            const Text('Total', style: TextStyle(fontSize: 16, color: Colors.grey)),
                            Text(
                              'Rp ${_totalPrice.toStringAsFixed(0)}',
                              style: const TextStyle(
                                fontSize: 24,
                                fontWeight: FontWeight.bold,
                                color: Color(0xFF1E293B),
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
    );
  }

  Widget _buildMobile() {
    return Column(
      children: [
        Expanded(
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
                  decoration: const InputDecoration(
                    hintText: 'Search products...',
                    prefixIcon: Icon(Icons.search),
                  ),
                ),
              ),
              Expanded(
                child: LayoutBuilder(
                  builder: (context, constraints) {
                    final crossAxisCount = constraints.maxWidth > 400 ? 3 : 2;
                    return GridView.builder(
                      padding: const EdgeInsets.all(8),
                      gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
                        crossAxisCount: crossAxisCount,
                        childAspectRatio: 0.85,
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
                              padding: const EdgeInsets.all(8),
                              child: Column(
                                mainAxisAlignment: MainAxisAlignment.center,
                                children: [
                                  Icon(Icons.inventory_2_outlined, size: 24, color: Theme.of(context).colorScheme.primary),
                                  const SizedBox(height: 6),
                                  Text(
                                    product['name'],
                                    textAlign: TextAlign.center,
                                    style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 12),
                                    maxLines: 2,
                                    overflow: TextOverflow.ellipsis,
                                  ),
                                  const SizedBox(height: 2),
                                  Text(
                                    'Rp ${product['price'].toStringAsFixed(0)}',
                                    style: TextStyle(color: Theme.of(context).colorScheme.primary, fontWeight: FontWeight.bold, fontSize: 12),
                                  ),
                                ],
                              ),
                            ),
                          ),
                        );
                      },
                    );
                  },
                ),
              ),
            ],
          ),
        ),
        if (_cart.isNotEmpty)
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: Colors.white,
              border: Border(top: BorderSide(color: Colors.grey.shade200)),
            ),
            child: SafeArea(
              top: false,
              child: Row(
                children: [
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Text('${_cart.length} items', style: TextStyle(color: Colors.grey.shade600, fontSize: 13)),
                        Text('Rp ${_totalPrice.toStringAsFixed(0)}', style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 18, color: Color(0xFF1E293B))),
                      ],
                    ),
                  ),
                  const SizedBox(width: 12),
                  FilledButton.icon(
                    onPressed: _showCartSheet,
                    icon: const Icon(Icons.shopping_cart_outlined),
                    label: const Text('View Cart'),
                  ),
                ],
              ),
            ),
          ),
      ],
    );
  }

  void _showCartSheet() {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (context) => DraggableScrollableSheet(
        initialChildSize: 0.6,
        minChildSize: 0.3,
        maxChildSize: 0.9,
        expand: false,
        builder: (context, scrollController) => Padding(
          padding: const EdgeInsets.fromLTRB(16, 16, 16, 0),
          child: Column(
            children: [
              Container(
                width: 40, height: 4,
                decoration: BoxDecoration(color: Colors.grey.shade300, borderRadius: BorderRadius.circular(2)),
              ),
              const SizedBox(height: 16),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text('Cart (${_cart.length})', style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 18)),
                  if (_cart.isNotEmpty)
                    TextButton(
                      onPressed: () { setState(() { _cart.clear(); }); Navigator.pop(context); },
                      child: const Text('Clear', style: TextStyle(color: Colors.red)),
                    ),
                ],
              ),
              const SizedBox(height: 8),
              Expanded(
                child: _cart.isEmpty
                    ? const Center(child: Text('Cart is empty'))
                    : ListView.builder(
                        controller: scrollController,
                        padding: const EdgeInsets.only(bottom: 80),
                        itemCount: _cart.length,
                        itemBuilder: (context, index) {
                          final item = _cart[index];
                          return Padding(
                            padding: const EdgeInsets.only(bottom: 8),
                            child: Container(
                              padding: const EdgeInsets.all(12),
                              decoration: BoxDecoration(
                                color: Colors.grey.shade50,
                                borderRadius: BorderRadius.circular(12),
                              ),
                              child: Row(
                                children: [
                                  Expanded(
                                    child: Column(
                                      crossAxisAlignment: CrossAxisAlignment.start,
                                      children: [
                                        Text(item['name'], style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 14)),
                                        const SizedBox(height: 4),
                                        Text('Rp ${(item['price'] as int).toStringAsFixed(0)}', style: const TextStyle(fontSize: 13, color: Colors.grey)),
                                      ],
                                    ),
                                  ),
                                  Row(
                                    mainAxisSize: MainAxisSize.min,
                                    children: [
                                      IconButton(
                                        icon: Icon(Icons.remove_circle_outline, size: 22),
                                        onPressed: () => _removeFromCart(index),
                                        color: Colors.red.shade400,
                                      ),
                                      Text('${item['quantity']}', style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
                                      IconButton(
                                        icon: Icon(Icons.add_circle_outline, size: 22),
                                        onPressed: () => _addToCart(item),
                                        color: Colors.green.shade400,
                                      ),
                                    ],
                                  ),
                                ],
                              ),
                            ),
                          );
                        },
                      ),
              ),
              Container(
                padding: const EdgeInsets.symmetric(vertical: 16),
                child: SafeArea(
                  top: false,
                  child: SizedBox(
                    width: double.infinity,
                    height: 56,
                    child: FilledButton.icon(
                      onPressed: () {
                        Navigator.pop(context);
                        _checkout();
                      },
                      icon: const Icon(Icons.payment),
                      label: const Text('PAY', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
                    ),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _PaymentOption extends StatelessWidget {
  final IconData icon;
  final String title;
  final String subtitle;
  final VoidCallback onTap;

  const _PaymentOption({
    required this.icon,
    required this.title,
    required this.subtitle,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(16),
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          border: Border.all(color: Colors.grey.shade200),
          borderRadius: BorderRadius.circular(16),
        ),
        child: Row(
          children: [
            Container(
              width: 48, height: 48,
              decoration: BoxDecoration(
                color: const Color(0xFF6366F1).withValues(alpha: 0.1),
                borderRadius: BorderRadius.circular(12),
              ),
              child: Icon(icon, color: const Color(0xFF6366F1), size: 24),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(title, style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 16)),
                  const SizedBox(height: 2),
                  Text(subtitle, style: TextStyle(color: Colors.grey.shade500, fontSize: 13)),
                ],
              ),
            ),
            Icon(Icons.chevron_right, color: Colors.grey.shade400),
          ],
        ),
      ),
    );
  }
}
