import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../shared/models/order.dart';
import '../../../../shared/models/product.dart';
import '../../../../shared/models/customer.dart';
import '../../../../core/config/translations.dart';
import '../../data/bluetooth_printer_service.dart';
import '../../data/customer_store.dart';
import '../../data/providers.dart';
import '../../data/cart_store.dart';
import '../../data/view_mode_store.dart';
import '../widgets/search_bar.dart';
import '../widgets/category_chips.dart';
import '../widgets/view_mode_switcher.dart';
import '../widgets/product_grid.dart';
import '../widgets/product_list.dart';
import '../widgets/checkout_bar.dart';
import 'main_shell.dart';
import 'payment_complete_screen.dart';

const Color kPrimary = Color(0xFF6366F1);
const Color kBg = Color(0xFFF7F5F8);

const List<Map<String, String>> _categories = [
  {'label': 'all', 'icon': 'grid_view'},
  {'label': 'coffee', 'icon': 'local_cafe'},
  {'label': 'tea', 'icon': 'local_cafe'},
  {'label': 'bakery', 'icon': 'restaurant'},
  {'label': 'dessert', 'icon': 'cake'},
  {'label': 'grocery', 'icon': 'shopping_bag'},
];

String _categoryKeyToDb(String key) {
  switch (key) {
    case 'coffee': return 'Coffee';
    case 'tea': return 'Tea';
    case 'bakery': return 'Bakery';
    case 'dessert': return 'Dessert';
    case 'grocery': return 'Grocery';
    default: return 'All';
  }
}

class PosScreen extends StatelessWidget {
  const PosScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return const Scaffold(
      body: PosScreenBody(),
    );
  }
}

class PosScreenBody extends ConsumerStatefulWidget {
  const PosScreenBody({super.key});

  @override
  ConsumerState<PosScreenBody> createState() => _PosBodyState();
}

class _PosBodyState extends ConsumerState<PosScreenBody> {
  String _searchQuery = '';
  String _selectedCategory = 'all';
  String _customerName = '';
  final TextEditingController _customerController = TextEditingController();

  final List<Map<String, dynamic>> _payments = [];

  static const double _taxRate = 0.08;
  static const int _discount = 5000;

  int get _subtotal => ref.read(cartProvider).totalPrice;
  int get _tax => (_subtotal * _taxRate).round();
  int get _grandTotal => _subtotal + _tax - _discount;

  List<Product> _applyFilters(List<Product> products) {
    var list = products;
    if (_selectedCategory != 'all') {
      final cat = _categoryKeyToDb(_selectedCategory);
      list = list.where((p) => p.category == cat).toList();
    }
    if (_searchQuery.isNotEmpty) {
      final q = _searchQuery.toLowerCase();
      list = list.where((p) => p.name.toLowerCase().contains(q)).toList();
    }
    return list;
  }

  void _addToCart(Product product) {
    ref.read(cartProvider.notifier).addItem(product);
  }

  void _clearCart() {
    ref.read(cartProvider.notifier).clear();
  }

  Order _saveOrder(String method, OrderStatus status) {
    final container = ProviderScope.containerOf(context, listen: false);
    final store = container.read(orderStoreProvider);
    final cart = container.read(cartProvider);
    final order = Order(
      id: store.nextId(),
      items: cart.items.map((item) => OrderItem(
        name: item.product.name,
        price: item.product.price,
        quantity: item.quantity,
      )).toList(),
      total: _grandTotal,
      status: status,
      paymentMethod: method,
      createdAt: DateTime.now(),
      discount: _discount,
      tax: _tax,
    );
    store.add(order);
    _clearCart();
    return order;
  }

  Translations _t() => ProviderScope.containerOf(context, listen: false).read(translationsProvider);

  void _checkout() {
    final t = _t();
    final cart = ref.read(cartProvider);
    if (cart.totalItems == 0) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(t.tr('cart_empty'))),
      );
      return;
    }
    _payments.clear();
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
      builder: (ctx) {
        int paidTotal() => _payments.fold(0, (s, p) => s + (p['amount'] as int));
        int remaining() => _grandTotal - paidTotal();
        bool fullyPaid() => remaining() <= 0;

        return StatefulBuilder(
          builder: (ctx, setSheetState) {
            final tt = ProviderScope.containerOf(ctx).read(translationsProvider);

            void addPayment(String method) {
              final maxAmount = remaining();
              if (maxAmount <= 0) return;
              final ctrl = TextEditingController(text: maxAmount.toString());
              showModalBottomSheet<int>(
                context: ctx,
                isScrollControlled: true,
                shape: const RoundedRectangleBorder(borderRadius: BorderRadius.vertical(top: Radius.circular(24))),
                builder: (ctx2) => StatefulBuilder(
                  builder: (ctx2, setAmtState) => Padding(
                    padding: EdgeInsets.only(left: 20, right: 20, top: 16, bottom: MediaQuery.of(ctx2).viewInsets.bottom + 20),
                    child: Column(
                      mainAxisSize: MainAxisSize.min,
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        Container(width: 40, height: 4, decoration: BoxDecoration(color: Colors.grey.shade300, borderRadius: BorderRadius.circular(2))),
                        const SizedBox(height: 16),
                        Text('$method · Rp ${maxAmount.toStringAsFixed(0)}', style: const TextStyle(fontWeight: FontWeight.w700, fontSize: 18)),
                        const SizedBox(height: 4),
                        Text(tt.tr('remaining'), style: TextStyle(fontSize: 13, color: Colors.grey.shade500)),
                        const SizedBox(height: 16),
                        TextField(
                          controller: ctrl,
                          keyboardType: TextInputType.number,
                          autofocus: true,
                          onChanged: (_) => setAmtState(() {}),
                          decoration: InputDecoration(
                            labelText: '${tt.tr('total')} (max Rp ${maxAmount.toStringAsFixed(0)})',
                            prefixText: 'Rp ',
                            border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
                          ),
                        ),
                        const SizedBox(height: 16),
                        Row(
                          children: [
                            Expanded(
                              child: OutlinedButton(
                                onPressed: () => Navigator.pop(ctx2, 0),
                                style: OutlinedButton.styleFrom(
                                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                                  padding: const EdgeInsets.symmetric(vertical: 14),
                                  side: BorderSide(color: Colors.grey.shade300),
                                ),
                                child: Text(tt.tr('cancel'), style: TextStyle(color: Colors.grey.shade700)),
                              ),
                            ),
                            const SizedBox(width: 12),
                            Expanded(
                              flex: 2,
                              child: FilledButton(
                                onPressed: () {
                                  final raw = int.tryParse(ctrl.text) ?? 0;
                                  Navigator.pop(ctx2, (raw > 0 && raw <= maxAmount) ? raw : 0);
                                },
                                style: FilledButton.styleFrom(backgroundColor: kPrimary, shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)), padding: const EdgeInsets.symmetric(vertical: 14)),
                                child: Text('Bayar Rp ${(int.tryParse(ctrl.text) ?? maxAmount).toStringAsFixed(0)}', style: const TextStyle(fontSize: 16)),
                              ),
                            ),
                          ],
                        ),
                      ],
                    ),
                  ),
                ),
              ).then((amount) {
                if (amount != null && amount > 0) {
                  _payments.add({'method': method, 'amount': amount});
                  setSheetState(() {});
                }
              });
            }

            void selectPaymentMethod(String method) {
              addPayment(method);
            }

            return DraggableScrollableSheet(
            initialChildSize: 0.92,
            minChildSize: 0.5,
            maxChildSize: 0.95,
            expand: false,
            builder: (ctx, scrollCtrl) => Column(
              children: [
                Expanded(
                  child: ListView(
                    controller: scrollCtrl,
                    padding: const EdgeInsets.fromLTRB(20, 16, 20, 0),
                    children: [
                      Center(
                        child: Container(
                          width: 40, height: 4,
                          decoration: BoxDecoration(color: Colors.grey.shade300, borderRadius: BorderRadius.circular(2)),
                        ),
                      ),
                      const SizedBox(height: 16),
                      Row(
                        children: [
                          GestureDetector(
                            onTap: () => Navigator.pop(ctx),
                            child: Container(
                              width: 36, height: 36,
                              decoration: BoxDecoration(color: Colors.grey.shade100, borderRadius: BorderRadius.circular(10)),
                              child: Icon(Icons.close, size: 20, color: Colors.grey.shade600),
                            ),
                          ),
                          const SizedBox(width: 12),
                          Text(tt.tr('select_payment'), style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 20)),
                          const Spacer(),
                          Container(
                            width: 36, height: 36,
                            decoration: BoxDecoration(color: Colors.grey.shade100, borderRadius: BorderRadius.circular(10)),
                            child: Icon(Icons.help_outline, size: 20, color: Colors.grey.shade500),
                          ),
                        ],
                      ),
                      const SizedBox(height: 20),
                      Container(
                        padding: const EdgeInsets.all(20),
                        decoration: BoxDecoration(
                          color: Colors.grey.shade100,
                          borderRadius: BorderRadius.circular(16),
                        ),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(tt.tr('total_due'), style: TextStyle(fontSize: 14, color: Colors.grey.shade600)),
                            const SizedBox(height: 6),
                            Text('Rp ${_grandTotal.toStringAsFixed(0)}', style: const TextStyle(fontSize: 32, fontWeight: FontWeight.w800, color: Color(0xFF1E293B))),
                            if (_payments.isNotEmpty) ...[
                              const SizedBox(height: 14),
                              Divider(color: Colors.grey.shade300, height: 1),
                              const SizedBox(height: 10),
                              ..._payments.map((p) => Padding(
                                padding: const EdgeInsets.only(bottom: 6),
                                child: Row(
                                  children: [
                                    Icon(Icons.check_circle, size: 14, color: Colors.green.shade600),
                                    const SizedBox(width: 6),
                                    Text('${p['method']}', style: TextStyle(fontSize: 13, color: Colors.grey.shade700)),
                                    const Spacer(),
                                    Text('Rp ${(p['amount'] as int).toStringAsFixed(0)}', style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600)),
                                  ],
                                ),
                              )),
                              const SizedBox(height: 4),
                              Divider(color: Colors.grey.shade300, height: 1),
                              const SizedBox(height: 10),
                            ],
                            Row(
                              children: [
                                Expanded(
                                  child: Column(
                                    crossAxisAlignment: CrossAxisAlignment.start,
                                    children: [
                                      Text(fullyPaid() ? tt.tr('paid') : tt.tr('remaining'), style: TextStyle(fontSize: 13, color: fullyPaid() ? Colors.green.shade700 : Colors.grey.shade500)),
                                      const SizedBox(height: 2),
                                      Text(
                                        fullyPaid() ? 'Rp ${_grandTotal.toStringAsFixed(0)}' : 'Rp ${remaining().toStringAsFixed(0)}',
                                        style: TextStyle(fontSize: 16, fontWeight: FontWeight.w700, color: fullyPaid() ? Colors.green.shade700 : Color(0xFFEF4444)),
                                      ),
                                    ],
                                  ),
                                ),
                              ],
                            ),
                          ],
                        ),
                      ),
                      const SizedBox(height: 24),
                      Row(
                        children: [
                          Text(tt.tr('payment_methods'), style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 18)),
                          const Spacer(),
                          Text(tt.tr('tap_to_select'), style: TextStyle(fontSize: 13, color: Colors.grey.shade500)),
                        ],
                      ),
                      const SizedBox(height: 16),
                      GridView.count(
                        crossAxisCount: 2,
                        shrinkWrap: true,
                        physics: const NeverScrollableScrollPhysics(),
                        mainAxisSpacing: 12,
                        crossAxisSpacing: 12,
                        childAspectRatio: 1.5,
                        children: [
                          _PaymentMethodCard(icon: Icons.money, label: tt.tr('cash'), onTap: () => selectPaymentMethod(tt.tr('cash'))),
                          _PaymentMethodCard(icon: Icons.qr_code, label: tt.tr('qris'), onTap: () => selectPaymentMethod(tt.tr('qris'))),
                          _PaymentMethodCard(icon: Icons.account_balance, label: tt.tr('bank_transfer'), onTap: () => selectPaymentMethod(tt.tr('bank_transfer'))),
                          _PaymentMethodCard(icon: Icons.credit_card, label: tt.tr('debit_card'), onTap: () => selectPaymentMethod(tt.tr('debit_card'))),
                          _PaymentMethodCard(icon: Icons.credit_card, label: tt.tr('credit_card'), onTap: () => selectPaymentMethod(tt.tr('credit_card'))),
                          _PaymentMethodCard(icon: Icons.account_balance_wallet, label: tt.tr('e_wallet'), onTap: () => selectPaymentMethod(tt.tr('e_wallet'))),
                        ],
                      ),
                      if (!fullyPaid() && _payments.isNotEmpty) ...[
                        const SizedBox(height: 16),
                        OutlinedButton.icon(
                          onPressed: () => addPayment(tt.tr('add_payment')),
                          icon: const Icon(Icons.add, size: 18),
                          label: Text(tt.tr('add_payment')),
                          style: OutlinedButton.styleFrom(
                            foregroundColor: Colors.grey.shade700,
                            side: BorderSide(color: Colors.grey.shade300),
                            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                            padding: const EdgeInsets.symmetric(vertical: 14),
                          ),
                        ),
                      ],
                      const SizedBox(height: 10),
                      OutlinedButton.icon(
                        onPressed: null,
                        icon: const Icon(Icons.swap_horiz, size: 18),
                        label: Text(tt.tr('split_bill'), style: TextStyle(color: Colors.grey.shade400)),
                        style: OutlinedButton.styleFrom(
                          foregroundColor: Colors.grey.shade400,
                          side: BorderSide(color: Colors.grey.shade200),
                          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                          padding: const EdgeInsets.symmetric(vertical: 14),
                        ),
                      ),
                      const SizedBox(height: 20),
                      Container(
                        padding: const EdgeInsets.all(20),
                        decoration: BoxDecoration(
                          border: Border.all(color: Colors.grey.shade300, width: 1.5),
                          borderRadius: BorderRadius.circular(16),
                        ),
                        child: Column(
                          children: [
                            Icon(Icons.info_outline, color: Colors.grey.shade400, size: 32),
                            const SizedBox(height: 12),
                            Text(
                              tt.tr('payment_info_hint'),
                              textAlign: TextAlign.center,
                              style: TextStyle(color: Colors.grey.shade500, fontSize: 14, height: 1.5),
                            ),
                          ],
                        ),
                      ),
                      const SizedBox(height: 20),
                    ],
                  ),
                ),
                Container(
                  padding: const EdgeInsets.fromLTRB(20, 12, 20, 20),
                  decoration: BoxDecoration(
                    color: Colors.white,
                    boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 8, offset: const Offset(0, -2))],
                  ),
                  child: SafeArea(
                    top: false,
                    child: Row(
                      children: [
                        TextButton(
                          onPressed: () => Navigator.pop(ctx),
                          style: TextButton.styleFrom(
                            side: BorderSide(color: Colors.grey.shade300),
                            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 12),
                          ),
                          child: Text(tt.tr('cancel'), style: TextStyle(color: Colors.grey.shade700)),
                        ),
                        const Spacer(),
                        Flexible(
                          child: FilledButton(
                            onPressed: fullyPaid() ? () {
                              final methods = _payments.map((p) => p['method'] as String).join(', ');
                              Navigator.pop(ctx);
                              final o = _saveOrder(methods, OrderStatus.paid);
                              _navigateToComplete(o);
                            } : null,
                            style: FilledButton.styleFrom(
                              backgroundColor: fullyPaid() ? kPrimary : Colors.grey.shade300,
                              disabledBackgroundColor: Colors.grey.shade300,
                              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 12),
                            ),
                            child: Text(
                              fullyPaid() ? '${tt.tr('complete_payment')} · Rp ${_grandTotal.toStringAsFixed(0)}' : tt.tr('complete_payment'),
                              style: TextStyle(color: fullyPaid() ? Colors.white : Colors.white38), overflow: TextOverflow.ellipsis,
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ],
            ),
          );
        },
      );
    },
    );
  }

  void _navigateToComplete(Order order) {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (_) => PaymentCompleteScreen(
          order: order,
          customerName: _customerName,
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final t = _t();
    final viewMode = ref.watch(viewModeProvider);
    final allProducts = ref.watch(productListProvider);
    final filteredProducts = _applyFilters(allProducts);

    return Column(
      children: [
        PosSearchBar(
          hintText: t.tr('search_hint'),
          onChanged: (v) => setState(() => _searchQuery = v),
        ),
        Row(
          children: [
            Expanded(
              child: CategoryChips(
                categories: _categories,
                selectedCategory: _selectedCategory,
                onCategorySelected: (cat) => setState(() => _selectedCategory = cat),
              ),
            ),
            ViewModeSwitcher(
              currentMode: viewMode,
              onModeChanged: (mode) => ref.read(viewModeProvider.notifier).setViewMode(mode),
            ),
            const SizedBox(width: 16),
          ],
        ),
        const SizedBox(height: 8),
        Expanded(
          child: viewMode == ViewMode.grid
              ? ProductGrid(
                  products: filteredProducts,
                  onProductTap: _addToCart,
                )
              : ProductList(
                  products: filteredProducts,
                  onProductTap: _addToCart,
                ),
        ),
        CheckoutBar(onCheckout: _showCartSheet),
      ],
    );
  }

  void _showCartSheet() {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
      builder: (ctx) {
        final t2 = ProviderScope.containerOf(ctx).read(translationsProvider);
        return StatefulBuilder(
          builder: (ctx, setSheetState) {
            return DraggableScrollableSheet(
              initialChildSize: 0.85,
              minChildSize: 0.4,
              maxChildSize: 0.95,
              expand: false,
              builder: (ctx, scrollCtrl) => Container(
                decoration: const BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
                ),
                child: Column(
                children: [
                  Expanded(
                    child: ListView(
                      controller: scrollCtrl,
                      padding: const EdgeInsets.fromLTRB(20, 16, 20, 0),
                      children: [
                        Center(
                          child: Container(width: 40, height: 4, decoration: BoxDecoration(color: Colors.grey.shade300, borderRadius: BorderRadius.circular(2))),
                        ),
                        const SizedBox(height: 16),
                        Row(
                          children: [
                            Text(t2.tr('current_order'), style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 20)),
                            const Spacer(),
                            Consumer(builder: (ctx, ref, _) {
                              final cart = ref.watch(cartProvider);
                              return cart.totalItems > 0
                                  ? GestureDetector(
                                      onTap: () { _clearCart(); setSheetState(() {}); },
                                      child: Text(t2.tr('clear'), style: TextStyle(color: Colors.red.shade400, fontWeight: FontWeight.w600)),
                                    )
                                  : const SizedBox.shrink();
                            }),
                            const SizedBox(width: 12),
                            GestureDetector(
                              onTap: () => Navigator.pop(ctx),
                              child: Container(
                                width: 32, height: 32,
                                decoration: BoxDecoration(color: Colors.grey.shade100, borderRadius: BorderRadius.circular(8)),
                                child: Icon(Icons.close, size: 18, color: Colors.grey.shade600),
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 12),
                        _buildCustomerField(ctx, t2),
                        const SizedBox(height: 12),
                        Consumer(builder: (ctx, ref, _) {
                          final cart = ref.watch(cartProvider);
                          if (cart.totalItems == 0) {
                            return Padding(
                              padding: const EdgeInsets.symmetric(vertical: 40),
                              child: Center(child: Text(t2.tr('empty_cart'), style: TextStyle(color: Colors.grey.shade400, fontSize: 16))),
                            );
                          }
                          return Column(
                            children: cart.items.map((item) {
                              return Padding(
                                padding: const EdgeInsets.only(bottom: 8),
                                child: Container(
                                  padding: const EdgeInsets.all(12),
                                  decoration: BoxDecoration(
                                    color: Colors.white,
                                    borderRadius: BorderRadius.circular(14),
                                    border: Border.all(color: Colors.grey.shade100),
                                  ),
                                  child: Row(
                                    crossAxisAlignment: CrossAxisAlignment.start,
                                    children: [
                                      Container(
                                        width: 44, height: 50,
                                        decoration: BoxDecoration(
                                          color: kPrimary.withValues(alpha: 0.08),
                                          borderRadius: BorderRadius.circular(10),
                                        ),
                                        child: Center(child: Icon(Icons.inventory_2_outlined, size: 22, color: kPrimary.withValues(alpha: 0.4))),
                                      ),
                                      const SizedBox(width: 10),
                                      Expanded(
                                        child: Column(
                                          crossAxisAlignment: CrossAxisAlignment.start,
                                          children: [
                                            Text(item.product.name, style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 14, color: Color(0xFF1E293B)), maxLines: 1, overflow: TextOverflow.ellipsis),
                                            const SizedBox(height: 2),
                                            GestureDetector(
                                              onTap: () => _editPrice(item.product.id, item.product.price),
                                              child: Row(
                                                mainAxisSize: MainAxisSize.min,
                                                children: [
                                                  Text('Rp ${item.product.price.toStringAsFixed(0)}', style: TextStyle(fontSize: 13, color: kPrimary, fontWeight: FontWeight.w600)),
                                                  const SizedBox(width: 4),
                                                  Icon(Icons.edit_note, size: 13, color: kPrimary),
                                                ],
                                              ),
                                            ),
                                            if (item.notes.isNotEmpty)
                                              Padding(
                                                padding: const EdgeInsets.only(top: 4),
                                                child: Container(
                                                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                                                  decoration: BoxDecoration(color: Colors.amber.shade50, borderRadius: BorderRadius.circular(6)),
                                                  child: Text(item.notes, style: TextStyle(fontSize: 11, color: Colors.amber.shade800)),
                                                ),
                                              ),
                                            Padding(
                                              padding: const EdgeInsets.only(top: 4),
                                              child: GestureDetector(
                                                onTap: () => _showNoteInput(ctx, item.product.id),
                                                child: Row(
                                                  mainAxisSize: MainAxisSize.min,
                                                  children: [
                                                    Icon(Icons.edit_outlined, size: 13, color: Colors.grey.shade400),
                                                    const SizedBox(width: 4),
                                                    Text(t2.tr('add_note'), style: TextStyle(fontSize: 12, color: Colors.grey.shade400)),
                                                  ],
                                                ),
                                              ),
                                            ),
                                          ],
                                        ),
                                      ),
                                      Container(
                                        height: 32,
                                        decoration: BoxDecoration(
                                          color: kPrimary.withValues(alpha: 0.08),
                                          borderRadius: BorderRadius.circular(16),
                                        ),
                                        child: Row(
                                          mainAxisSize: MainAxisSize.min,
                                          children: [
                                            GestureDetector(
                                              onTap: () {
                                                if (item.quantity > 1) {
                                                  ref.read(cartProvider.notifier).updateQuantity(item.product.id, item.quantity - 1);
                                                } else {
                                                  ref.read(cartProvider.notifier).removeItem(item.product.id);
                                                }
                                                setSheetState(() {});
                                              },
                                              child: Container(
                                                width: 32, height: 32,
                                                decoration: BoxDecoration(
                                                  color: kPrimary.withValues(alpha: 0.12),
                                                  borderRadius: const BorderRadius.horizontal(left: Radius.circular(16)),
                                                ),
                                                child: Center(child: Icon(Icons.remove, size: 16, color: kPrimary)),
                                              ),
                                            ),
                                            Padding(
                                              padding: const EdgeInsets.symmetric(horizontal: 10),
                                              child: Text('${item.quantity}', style: const TextStyle(fontWeight: FontWeight.w700, fontSize: 14, color: Color(0xFF1E293B))),
                                            ),
                                            GestureDetector(
                                              onTap: () {
                                                ref.read(cartProvider.notifier).updateQuantity(item.product.id, item.quantity + 1);
                                                setSheetState(() {});
                                              },
                                              child: Container(
                                                width: 32, height: 32,
                                                decoration: BoxDecoration(
                                                  color: kPrimary,
                                                  borderRadius: const BorderRadius.horizontal(right: Radius.circular(16)),
                                                ),
                                                child: const Center(child: Icon(Icons.add, size: 16, color: Colors.white)),
                                              ),
                                            ),
                                          ],
                                        ),
                                      ),
                                      const SizedBox(width: 8),
                                      GestureDetector(
                                        onTap: () {
                                          ref.read(cartProvider.notifier).removeItem(item.product.id);
                                          setSheetState(() {});
                                        },
                                        child: Container(
                                          padding: const EdgeInsets.all(6),
                                          decoration: BoxDecoration(
                                            color: Colors.red.shade50,
                                            borderRadius: BorderRadius.circular(8),
                                          ),
                                          child: Icon(Icons.delete_outline, size: 16, color: Colors.red.shade400),
                                        ),
                                      ),
                                    ],
                                  ),
                                ),
                              );
                            }).toList(),
                          );
                        }),
                      ],
                    ),
                  ),
                  Consumer(builder: (ctx, ref, _) {
                    final cart = ref.watch(cartProvider);
                    if (cart.totalItems == 0) return const SizedBox.shrink();
                    return Container(
                      padding: const EdgeInsets.fromLTRB(20, 12, 20, 16),
                      decoration: BoxDecoration(
                        color: Colors.white,
                        boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 8, offset: const Offset(0, -2))],
                      ),
                      child: SafeArea(
                        top: false,
                        child: Column(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            _summaryRow(t2.tr('subtotal'), _subtotal),
                            const SizedBox(height: 6),
                            _summaryRow('${t2.tr('tax')} (8%)', _tax),
                            const SizedBox(height: 6),
                            _summaryRow(t2.tr('discount'), -_discount),
                            const Divider(height: 16),
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Flexible(child: Text(t2.tr('grand_total'), style: TextStyle(fontSize: 18, fontWeight: FontWeight.w700, color: const Color(0xFF1E293B)), overflow: TextOverflow.ellipsis)),
                                Text('Rp ${_grandTotal.toStringAsFixed(0)}', style: TextStyle(fontSize: 18, fontWeight: FontWeight.w800, color: kPrimary)),
                              ],
                            ),
                            const SizedBox(height: 16),
                            SizedBox(
                              width: double.infinity,
                              height: 56,
                              child: FilledButton.icon(
                                onPressed: () { if (mounted) { setState(() {}); Navigator.pop(ctx); _checkout(); } },
                                icon: const Icon(Icons.arrow_forward),
                                label: Text(t2.tr('proceed_to_payment'), style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
                                style: FilledButton.styleFrom(
                                  backgroundColor: kPrimary,
                                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(28)),
                                ),
                              ),
                            ),
                          ],
                        ),
                      ),
                    );
                  }),
                ],
              ),
            ),
            );
          },
        );
      },
    );
  }

  Widget _buildCustomerField(BuildContext sheetCtx, Translations t) {
    final suggestions = _filteredCustomers(_customerName);
    final exactMatch = suggestions.any((c) => c.name.toLowerCase() == _customerName.toLowerCase());

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        TextField(
          controller: _customerController,
          onChanged: (v) { _customerName = v; },
          decoration: InputDecoration(
            hintText: t.tr('customer_hint'),
            prefixIcon: const Icon(Icons.person_outline, size: 20),
            filled: true,
            fillColor: Colors.grey.shade50,
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: BorderSide.none),
            contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
          ),
        ),
        if (_customerName.isNotEmpty && suggestions.isNotEmpty)
          Container(
            margin: const EdgeInsets.only(top: 4),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: Colors.grey.shade200),
              boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 8, offset: const Offset(0, 2))],
            ),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                ...suggestions.take(5).map((c) {
                  final selected = _customerName == c.name;
                  return GestureDetector(
                    onTap: selected ? null : () => _selectCustomer(c),
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                      decoration: BoxDecoration(
                        color: selected ? kPrimary.withValues(alpha: 0.08) : Colors.transparent,
                        border: Border(bottom: BorderSide(color: Colors.grey.shade100)),
                      ),
                      child: Row(
                        children: [
                          Icon(Icons.store, size: 16, color: kPrimary.withValues(alpha: 0.6)),
                          const SizedBox(width: 8),
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(c.name, style: TextStyle(fontWeight: FontWeight.w600, fontSize: 13, color: selected ? kPrimary : const Color(0xFF1E293B))),
                                if (c.phone.isNotEmpty)
                                  Text(c.phone, style: TextStyle(fontSize: 11, color: Colors.grey.shade500)),
                              ],
                            ),
                          ),
                          if (selected) Icon(Icons.check_circle, size: 16, color: kPrimary),
                        ],
                      ),
                    ),
                  );
                }),
                if (!exactMatch)
                  GestureDetector(
                    onTap: () => _saveNewCustomer(sheetCtx),
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 12),
                      decoration: BoxDecoration(border: Border(top: BorderSide(color: Colors.grey.shade200))),
                      child: Row(
                        children: [
                          Container(
                            width: 28, height: 28,
                            decoration: BoxDecoration(color: kPrimary, borderRadius: BorderRadius.circular(8)),
                            child: const Icon(Icons.person_add, size: 16, color: Colors.white),
                          ),
                          const SizedBox(width: 10),
                          Flexible(child: Text('Simpan "$_customerName" + kontak', style: TextStyle(fontWeight: FontWeight.w600, fontSize: 14, color: kPrimary), overflow: TextOverflow.ellipsis)),
                        ],
                      ),
                    ),
                  ),
              ],
            ),
          ),
      ],
    );
  }

  void _selectCustomer(Customer c) {
    _customerController.text = c.name;
    _customerName = c.name;
    setState(() {});
  }

  List<Customer> _filteredCustomers(String query) {
    final store = ProviderScope.containerOf(context, listen: false).read(customerStoreProvider);
    return store.search(query);
  }

  void _saveNewCustomer(BuildContext sheetCtx) {
    final phoneController = TextEditingController();
    showModalBottomSheet(
      context: sheetCtx,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(borderRadius: BorderRadius.vertical(top: Radius.circular(24))),
      builder: (ctx) => Padding(
        padding: EdgeInsets.only(left: 20, right: 20, top: 16, bottom: MediaQuery.of(ctx).viewInsets.bottom + 20),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Container(width: 40, height: 4, decoration: BoxDecoration(color: Colors.grey.shade300, borderRadius: BorderRadius.circular(2))),
            const SizedBox(height: 16),
            const Text('Simpan Pelanggan', style: TextStyle(fontWeight: FontWeight.w700, fontSize: 18)),
            const SizedBox(height: 16),
            TextField(
              controller: TextEditingController(text: _customerName),
              readOnly: true,
              decoration: InputDecoration(
                labelText: 'Nama',
                border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
              ),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: phoneController,
              keyboardType: TextInputType.phone,
              decoration: InputDecoration(
                labelText: 'No. Telepon',
                border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
              ),
            ),
            const SizedBox(height: 16),
            FilledButton(
              onPressed: () {
                final store = ProviderScope.containerOf(ctx, listen: false).read(customerStoreProvider);
                store.add(_customerName, phoneController.text);
                Navigator.pop(ctx);
              },
              style: FilledButton.styleFrom(backgroundColor: kPrimary, shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12))),
              child: const Text('Simpan'),
            ),
          ],
        ),
      ),
    );
  }

  void _editPrice(String productId, int currentPrice) {
    final controller = TextEditingController(text: '$currentPrice');
    final t = _t();
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(borderRadius: BorderRadius.vertical(top: Radius.circular(24))),
      builder: (ctx) => Padding(
        padding: EdgeInsets.only(left: 20, right: 20, top: 16, bottom: MediaQuery.of(ctx).viewInsets.bottom + 20),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Container(width: 40, height: 4, decoration: BoxDecoration(color: Colors.grey.shade300, borderRadius: BorderRadius.circular(2))),
            const SizedBox(height: 16),
            Text('${t.tr('edit')} Product', style: const TextStyle(fontWeight: FontWeight.w700, fontSize: 18)),
            const SizedBox(height: 16),
            TextField(
              controller: controller,
              keyboardType: TextInputType.number,
              autofocus: true,
              decoration: InputDecoration(
                labelText: 'Harga',
                prefixText: 'Rp ',
                border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
              ),
            ),
            const SizedBox(height: 16),
            FilledButton(
              onPressed: () {
                Navigator.pop(ctx);
              },
              style: FilledButton.styleFrom(backgroundColor: kPrimary, shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)), padding: const EdgeInsets.symmetric(vertical: 16)),
              child: Text(t.tr('save'), style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
            ),
          ],
        ),
      ),
    );
  }

  void _showNoteInput(BuildContext parentCtx, String productId) {
    final controller = TextEditingController();
    final nt = _t();
    showModalBottomSheet(
      context: parentCtx,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(borderRadius: BorderRadius.vertical(top: Radius.circular(24))),
      builder: (ctx) => Padding(
        padding: EdgeInsets.only(left: 20, right: 20, top: 16, bottom: MediaQuery.of(ctx).viewInsets.bottom + 20),
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Container(width: 40, height: 4, decoration: BoxDecoration(color: Colors.grey.shade300, borderRadius: BorderRadius.circular(2))),
              const SizedBox(height: 16),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(nt.tr('item_note'), style: const TextStyle(fontWeight: FontWeight.w700, fontSize: 18)),
                  GestureDetector(
                    onTap: () => Navigator.pop(ctx),
                    child: Container(
                      width: 32, height: 32,
                      decoration: BoxDecoration(color: Colors.grey.shade100, borderRadius: BorderRadius.circular(8)),
                      child: Icon(Icons.close, size: 18, color: Colors.grey.shade600),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 16),
              TextField(
                controller: controller,
                autofocus: true,
                maxLines: 3,
                decoration: InputDecoration(
                  hintText: nt.tr('special_instructions'),
                  border: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: BorderSide.none),
                  filled: true,
                  fillColor: Colors.grey.shade50,
                ),
              ),
              const SizedBox(height: 16),
              FilledButton(
                onPressed: () {
                  ref.read(cartProvider.notifier).updateNotes(productId, controller.text);
                  Navigator.pop(ctx);
                },
                style: FilledButton.styleFrom(
                  backgroundColor: kPrimary,
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
                  padding: const EdgeInsets.symmetric(vertical: 16),
                ),
                child: Text(nt.tr('save'), style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _summaryRow(String label, int amount) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Flexible(child: Text(label, style: TextStyle(fontSize: 14, color: Colors.grey.shade600), overflow: TextOverflow.ellipsis)),
        Text(
          '${amount >= 0 ? 'Rp ' : '-Rp '}${amount.abs().toStringAsFixed(0)}',
          style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: amount < 0 ? Colors.green.shade700 : const Color(0xFF1E293B)),
        ),
      ],
    );
  }
}

void showPrinterDialog(BuildContext context) {
  if (!BluetoothPrinterService.supported) {
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text('Bluetooth printer hanya untuk Android.')),
    );
    return;
  }
  showModalBottomSheet(
    context: context,
    isScrollControlled: true,
    shape: const RoundedRectangleBorder(borderRadius: BorderRadius.vertical(top: Radius.circular(20))),
    builder: (_) => const _PrinterSetupSheet(),
  );
}

class _PrinterSetupSheet extends StatefulWidget {
  const _PrinterSetupSheet();
  @override
  State<_PrinterSetupSheet> createState() => _PrinterSetupSheetState();
}

class _PrinterSetupSheetState extends State<_PrinterSetupSheet> {
  final _bt = BluetoothPrinterService();
  List<BtPrinter> _devices = [];
  StreamSubscription? _sub;
  bool _scanning = false;

  @override
  void initState() {
    super.initState();
    _devices = _bt.devices;
    _sub = _bt.onDevicesChanged.listen((d) { if (mounted) setState(() => _devices = d); });
  }

  @override
  void dispose() {
    _sub?.cancel();
    if (_scanning) _bt.stopScan();
    super.dispose();
  }

  void _toggleScan() {
    if (_scanning) { _bt.stopScan(); setState(() => _scanning = false); }
    else { setState(() => _scanning = true); _bt.startScan(); }
  }

  Future<void> _connect(BtPrinter d) async {
    try {
      await _bt.connect(d);
      if (mounted) { Navigator.pop(context); ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Connected to ${d.name}'))); }
    } catch (e) {
      if (mounted) { ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Failed: $e'))); }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.only(bottom: MediaQuery.of(context).viewInsets.bottom),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 16, 16, 0),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                const Text('Printer', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
                Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    if (_bt.isConnected)
                      TextButton.icon(
                        onPressed: () { _bt.disconnect(); setState(() {}); },
                        icon: const Icon(Icons.link_off, size: 18),
                        label: const Text('Disconnect'),
                        style: TextButton.styleFrom(foregroundColor: Colors.red),
                      ),
                    TextButton(onPressed: _toggleScan, child: Text(_scanning ? 'Stop' : 'Scan')),
                  ],
                ),
              ],
            ),
          ),
          if (_bt.isConnected)
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: Row(
                children: [
                  const Icon(Icons.check_circle, size: 16, color: Colors.green),
                  const SizedBox(width: 6),
                  Text('Connected', style: TextStyle(color: Colors.green.shade700)),
                ],
              ),
            ),
          if (_scanning) const Padding(padding: EdgeInsets.all(16), child: Center(child: CircularProgressIndicator())),
          if (_devices.isEmpty && !_scanning) const Padding(padding: EdgeInsets.all(24), child: Center(child: Text('No devices found'))),
          ConstrainedBox(
            constraints: BoxConstraints(maxHeight: MediaQuery.of(context).size.height * 0.4),
            child: ListView.builder(
              shrinkWrap: true,
              itemCount: _devices.length,
              itemBuilder: (context, i) {
                final d = _devices[i];
                return ListTile(leading: const Icon(Icons.print), title: Text(d.name), subtitle: Text(d.address), onTap: () => _connect(d));
              },
            ),
          ),
          const SizedBox(height: 8),
        ],
      ),
    );
  }
}

class _PaymentMethodCard extends StatelessWidget {
  final IconData icon;
  final String label;
  final VoidCallback onTap;

  const _PaymentMethodCard({required this.icon, required this.label, required this.onTap});

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(16),
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: 12, horizontal: 12),
        decoration: BoxDecoration(
          color: Colors.grey.shade50,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: Colors.grey.shade200),
        ),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Container(
              width: 44, height: 44,
              decoration: BoxDecoration(
                color: kPrimary.withValues(alpha: 0.08),
                borderRadius: BorderRadius.circular(12),
              ),
              child: Icon(icon, color: kPrimary, size: 22),
            ),
            const SizedBox(height: 8),
            Flexible(child: Text(label, style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600), textAlign: TextAlign.center, overflow: TextOverflow.ellipsis)),
          ],
        ),
      ),
    );
  }
}
