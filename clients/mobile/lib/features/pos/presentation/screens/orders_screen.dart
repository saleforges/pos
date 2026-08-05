import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../shared/models/order.dart';
import '../../../../core/config/translations.dart';
import '../../data/print_service.dart';
import 'main_shell.dart';

class OrdersScreen extends StatelessWidget {
  const OrdersScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return const OrdersScreenBody();
  }
}

class OrdersScreenBody extends ConsumerStatefulWidget {
  const OrdersScreenBody({super.key});

  @override
  ConsumerState<OrdersScreenBody> createState() => _OrdersScreenBodyState();
}

class _OrdersScreenBodyState extends ConsumerState<OrdersScreenBody> {
  final _searchController = TextEditingController();
  String _query = '';
  OrderStatus? _filter;

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  List<Order> _filterOrders(List<Order> orders) {
    var result = orders;
    if (_filter != null) {
      result = result.where((o) => o.status == _filter).toList();
    }
    if (_query.isNotEmpty) {
      final q = _query.toLowerCase();
      result = result.where((o) =>
        o.id.toLowerCase().contains(q) ||
        o.items.any((i) => i.name.toLowerCase().contains(q))
      ).toList();
    }
    return result;
  }

  @override
  Widget build(BuildContext context) {
    final t = ref.watch(translationsProvider);
    final allOrders = ref.watch(orderStoreProvider).orders;
    final orders = _filterOrders(allOrders);

    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
          child: TextField(
            controller: _searchController,
            onChanged: (v) => setState(() => _query = v),
            decoration: InputDecoration(
              hintText: t.tr('search_orders'),
              hintStyle: TextStyle(color: Colors.grey.shade400),
              prefixIcon: const Icon(Icons.search),
              suffixIcon: _query.isNotEmpty
                  ? IconButton(icon: const Icon(Icons.clear), onPressed: () { _searchController.clear(); setState(() => _query = ''); })
                  : null,
              filled: true,
              fillColor: Colors.white,
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(16),
                borderSide: BorderSide.none,
              ),
            ),
          ),
        ),
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 16),
          child: Row(
            children: [
              _FilterChip(label: t.tr('all_orders'), selected: _filter == null, onTap: () => setState(() => _filter = null)),
              const SizedBox(width: 8),
              _FilterChip(label: t.tr('unpaid'), selected: _filter == OrderStatus.unpaid, onTap: () => setState(() => _filter = OrderStatus.unpaid)),
              const SizedBox(width: 8),
              _FilterChip(label: t.tr('paid'), selected: _filter == OrderStatus.paid, onTap: () => setState(() => _filter = OrderStatus.paid)),
            ],
          ),
        ),
        const SizedBox(height: 8),
        Expanded(
          child: orders.isEmpty
              ? Center(child: Text(t.tr('no_orders'), style: TextStyle(color: Colors.grey.shade400)))
              : ListView.builder(
                  padding: const EdgeInsets.all(16),
                  itemCount: orders.length,
                  itemBuilder: (context, index) {
                    final order = orders[index];
                    return Card(
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                      child: Padding(
                        padding: const EdgeInsets.all(16),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Text(order.id, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
                                Container(
                                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                                  decoration: BoxDecoration(
                                    color: order.status == OrderStatus.unpaid ? Colors.orange.shade50 : Colors.green.shade50,
                                    borderRadius: BorderRadius.circular(8),
                                    border: Border.all(color: order.status == OrderStatus.unpaid ? Colors.orange.shade200 : Colors.green.shade200),
                                  ),
                                  child: Text(
                                    order.status == OrderStatus.unpaid ? t.tr('unpaid').toUpperCase() : t.tr('paid').toUpperCase(),
                                    style: TextStyle(
                                      color: order.status == OrderStatus.unpaid ? Colors.orange.shade700 : Colors.green.shade700,
                                      fontSize: 11, fontWeight: FontWeight.bold,
                                    ),
                                  ),
                                ),
                              ],
                            ),
                            const SizedBox(height: 4),
                            Text(order.paymentMethod, style: TextStyle(color: Colors.grey.shade500, fontSize: 13)),
                            const SizedBox(height: 12),
                            ...order.items.map((item) => Padding(
                              padding: const EdgeInsets.only(bottom: 4),
                              child: Row(
                                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                children: [
                                  Flexible(child: Text('${item.name} x${item.quantity}', style: const TextStyle(fontSize: 14), overflow: TextOverflow.ellipsis)),
                                  Text('Rp ${(item.price * item.quantity).toStringAsFixed(0)}', style: const TextStyle(fontSize: 14)),
                                ],
                              ),
                            )),
                            const Divider(),
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Text(t.tr('total'), style: const TextStyle(fontWeight: FontWeight.bold)),
                                Text('Rp ${order.total.toStringAsFixed(0)}', style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16, color: Color(0xFF6366F1))),
                              ],
                            ),
                            if (order.status == OrderStatus.unpaid) ...[
                              const SizedBox(height: 12),
                              SizedBox(
                                width: double.infinity,
                                child: FilledButton(
                                  onPressed: () {
                                    ref.read(orderStoreProvider).markPaid(order.id);
                                    ScaffoldMessenger.of(context).showSnackBar(
                                      SnackBar(content: Text('${order.id} ${t.tr('paid')}')),
                                    );
                                  },
                                  style: FilledButton.styleFrom(shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12))),
                                  child: Text(t.tr('mark_paid')),
                                ),
                              ),
                            ],
                            const SizedBox(height: 8),
                            SizedBox(
                              width: double.infinity,
                              child: OutlinedButton.icon(
                                onPressed: () => PrintService.printReceipt(context, order),
                                icon: const Icon(Icons.receipt, size: 18),
                                label: Text(t.tr('print_receipt')),
                                style: OutlinedButton.styleFrom(shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12))),
                              ),
                            ),
                          ],
                        ),
                      ),
                    );
                  },
                ),
        ),
      ],
    );
  }
}

class _FilterChip extends StatelessWidget {
  final String label;
  final bool selected;
  final VoidCallback onTap;

  const _FilterChip({required this.label, required this.selected, required this.onTap});

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 6),
        decoration: BoxDecoration(
          color: selected ? const Color(0xFF6366F1) : Colors.grey.shade100,
          borderRadius: BorderRadius.circular(20),
        ),
        child: Text(
          label,
          style: TextStyle(
            color: selected ? Colors.white : Colors.grey.shade700,
            fontWeight: FontWeight.w600,
            fontSize: 13,
          ),
        ),
      ),
    );
  }
}
