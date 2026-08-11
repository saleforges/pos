import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:image_picker/image_picker.dart';
import '../../../../branding/saleforges_loaders.dart';
import '../../../../core/config/translations.dart';
import '../../../../core/network/connectivity_provider.dart';
import '../../../../core/utils/format.dart';
import '../../../auth/presentation/providers/auth_provider.dart';
import '../../data/order_store.dart';
import '../../data/customer_store.dart';
import '../../data/providers.dart';
import '../../data/bluetooth_printer_service.dart';
import '../../data/remote/shift_remote_data_source.dart';
import '../../data/remote/payment_remote_data_source.dart';
import 'pos_screen.dart';
import 'orders_screen.dart';
import 'history_screen.dart';
import 'customers_screen.dart';
import 'manage_products_screen.dart';
import 'staff_screen.dart';
import 'branches_screen.dart';
import 'audit_log_screen.dart';

final orderStoreProvider = Provider<OrderStore>((ref) => OrderStore());

const Color _kPrimary = Color(0xFF6366F1);

class MainShell extends ConsumerStatefulWidget {
  const MainShell({super.key});
  @override
  ConsumerState<MainShell> createState() => _MainShellState();
}

class _MainShellState extends ConsumerState<MainShell> {
  int _selectedIndex = 0;

  static const _navItems = [
    (icon: Icons.store_outlined, activeIcon: Icons.store),
    (icon: Icons.receipt_long_outlined, activeIcon: Icons.receipt_long),
    (icon: Icons.history_outlined, activeIcon: Icons.history),
    (icon: Icons.people_outline, activeIcon: Icons.people),
    (icon: Icons.settings_outlined, activeIcon: Icons.settings),
  ];

  void _openManageProducts() {
    Navigator.push(
      context,
      MaterialPageRoute(builder: (_) => const ManageProductsScreen()),
    );
  }

  @override
  Widget build(BuildContext context) {
    // Settle queued orders/payments as soon as signal returns, instead of
    // waiting for someone to open the Orders tab.
    ref.listen<bool>(isOfflineProvider, (prev, next) {
      if (prev == true && next == false) {
        ref.read(orderRepositoryProvider).flushQueue();
        ref.read(paymentRepositoryProvider).flushQueue();
        ref.read(productRepositoryProvider).flushQueue().then((_) {
          ref.read(productListProvider.notifier).refresh();
        });
      }
    });
    final t = ref.watch(translationsProvider);
    final titles = [t.tr('catalog'), t.tr('orders'), t.tr('summary'), t.tr('customers'), t.tr('settings')];
    final wide = MediaQuery.of(context).size.width >= 720;
    final body = _buildBody(titles[_selectedIndex]);

    return Scaffold(
      appBar: wide ? null : AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        scrolledUnderElevation: 0.5,
        title: Text(titles[_selectedIndex], style: const TextStyle(fontWeight: FontWeight.w700, color: Color(0xFF1E293B))),
        centerTitle: true,
        actions: [
          const _OfflineBadge(),
          Consumer(builder: (ctx, ref, _) {
            final user = ref.watch(authProvider).user;
            return Padding(
              padding: const EdgeInsets.only(right: 8),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(Icons.person, size: 18, color: Colors.grey.shade600),
                  const SizedBox(width: 4),
                  Text(user?.username ?? '', style: TextStyle(fontSize: 13, fontWeight: FontWeight.w500, color: Colors.grey.shade700)),
                ],
              ),
            );
          }),
        ],
      ),
      bottomNavigationBar: wide ? null : _buildBottomNav(t, titles),
      body: wide
          ? Row(
              children: [
                _buildRail(t, titles),
                const VerticalDivider(width: 1, thickness: 1),
                Expanded(child: body),
              ],
            )
          : body,
    );
  }

  Widget _buildBottomNav(Translations t, List<String> titles) {
    final lowStockCount = ref.watch(lowStockProductsProvider).length;
    final canManageCatalog = ref.watch(authProvider).canManageCatalog;
    final tabs = [
      (index: 0, icon: _navItems[0].icon, activeIcon: _navItems[0].activeIcon, label: titles[0]),
      (index: 1, icon: _navItems[1].icon, activeIcon: _navItems[1].activeIcon, label: titles[1]),
      (index: 4, icon: _navItems[4].icon, activeIcon: _navItems[4].activeIcon, label: titles[4]),
    ];
    final moreActive = _selectedIndex == 2 || _selectedIndex == 3;

    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        boxShadow: [
          BoxShadow(color: Colors.black.withValues(alpha: 0.05), blurRadius: 12, offset: const Offset(0, -2)),
        ],
      ),
      child: SafeArea(
        top: false,
        child: SizedBox(
          height: 60,
          child: Row(
            children: [
              for (final tab in tabs)
                Expanded(
                  child: _BottomNavItem(
                    icon: _selectedIndex == tab.index ? tab.activeIcon : tab.icon,
                    label: tab.label,
                    selected: _selectedIndex == tab.index,
                    onTap: () => setState(() => _selectedIndex = tab.index),
                  ),
                ),
              Expanded(
                child: _BottomNavItem(
                  icon: Icons.more_horiz,
                  label: t.tr('more'),
                  selected: moreActive,
                  badgeCount: canManageCatalog ? lowStockCount : 0,
                  onTap: () => _openMoreSheet(t),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  void _openMoreSheet(Translations t) {
    final canManageCatalog = ref.read(authProvider).canManageCatalog;
    final lowStockCount = ref.read(lowStockProductsProvider).length;
    showModalBottomSheet(
      context: context,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
      builder: (ctx) => SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: 12),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Container(
                width: 40,
                height: 4,
                decoration: BoxDecoration(color: Colors.grey.shade300, borderRadius: BorderRadius.circular(2)),
              ),
              const SizedBox(height: 8),
              ListTile(
                leading: Icon(_navItems[2].icon, color: _kPrimary),
                title: Text(t.tr('summary')),
                onTap: () {
                  Navigator.pop(ctx);
                  setState(() => _selectedIndex = 2);
                },
              ),
              ListTile(
                leading: Icon(_navItems[3].icon, color: _kPrimary),
                title: Text(t.tr('customers')),
                onTap: () {
                  Navigator.pop(ctx);
                  setState(() => _selectedIndex = 3);
                },
              ),
              if (canManageCatalog)
                ListTile(
                  leading: Badge(
                    isLabelVisible: lowStockCount > 0,
                    label: Text('$lowStockCount'),
                    child: const Icon(Icons.inventory_2_outlined, color: _kPrimary),
                  ),
                  title: Text(t.tr('product_stock')),
                  subtitle: lowStockCount > 0 ? Text('$lowStockCount ${t.tr('low_stock_products')}') : null,
                  onTap: () {
                    Navigator.pop(ctx);
                    _openManageProducts();
                  },
                ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildRail(Translations t, List<String> titles) {
    final canManageCatalog = ref.watch(authProvider).canManageCatalog;
    final lowStockCount = ref.watch(lowStockProductsProvider).length;
    return NavigationRail(
      leading: const Padding(
        padding: EdgeInsets.only(top: 12, bottom: 4),
        child: _OfflineBadge(compact: true),
      ),
      selectedIndex: _selectedIndex,
      onDestinationSelected: (i) {
        if (i < _navItems.length) {
          setState(() => _selectedIndex = i);
        } else if (canManageCatalog) {
          _openManageProducts();
        }
      },
      labelType: NavigationRailLabelType.all,
      selectedIconTheme: const IconThemeData(color: Color(0xFF6366F1)),
      selectedLabelTextStyle: const TextStyle(fontWeight: FontWeight.w700, color: Color(0xFF6366F1)),
      destinations: [
        for (var i = 0; i < _navItems.length; i++)
          NavigationRailDestination(
            icon: Icon(_navItems[i].icon),
            selectedIcon: Icon(_navItems[i].activeIcon),
            label: Text(titles[i]),
          ),
        if (canManageCatalog)
          NavigationRailDestination(
            icon: Badge(
              isLabelVisible: lowStockCount > 0,
              label: Text('$lowStockCount'),
              child: const Icon(Icons.inventory_2_outlined),
            ),
            selectedIcon: const Icon(Icons.inventory_2),
            label: Text(t.tr('product_stock')),
          ),
      ],
    );
  }

  Widget _buildBody(String title) {
    switch (_selectedIndex) {
      case 0: return const PosScreen();
      case 1: return const OrdersScreen();
      case 2: return const HistoryScreen();
      case 3: return const CustomersScreen();
      case 4: return const _SettingsScreen();
      default: return const PosScreen();
    }
  }
}

class _OfflineBadge extends ConsumerWidget {
  const _OfflineBadge({this.compact = false});

  final bool compact;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final offline = ref.watch(isOfflineProvider);
    if (!offline) return const SizedBox.shrink();

    final t = ref.watch(translationsProvider);
    final badge = Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: Colors.red.shade50,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.red.shade200),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.cloud_off, size: 13, color: Colors.red.shade600),
          if (!compact) ...[
            const SizedBox(width: 4),
            Text(
              t.tr('offline'),
              style: TextStyle(fontSize: 11, fontWeight: FontWeight.w700, color: Colors.red.shade600),
            ),
          ],
        ],
      ),
    );
    return compact
        ? Padding(padding: const EdgeInsets.symmetric(horizontal: 8), child: badge)
        : Padding(padding: const EdgeInsets.only(right: 8), child: badge);
  }
}

class _BottomNavItem extends StatelessWidget {
  const _BottomNavItem({
    required this.icon,
    required this.label,
    required this.selected,
    required this.onTap,
    this.badgeCount = 0,
  });

  final IconData icon;
  final String label;
  final bool selected;
  final VoidCallback onTap;
  final int badgeCount;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          AnimatedContainer(
            duration: const Duration(milliseconds: 200),
            width: 6,
            height: 6,
            margin: const EdgeInsets.only(bottom: 4),
            decoration: BoxDecoration(
              color: selected ? _kPrimary : Colors.transparent,
              shape: BoxShape.circle,
            ),
          ),
          Badge(
            isLabelVisible: badgeCount > 0,
            label: Text('$badgeCount'),
            child: Icon(icon, size: 22, color: selected ? _kPrimary : Colors.grey.shade500),
          ),
          const SizedBox(height: 2),
          Text(
            label,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
            style: TextStyle(
              fontSize: 11,
              fontWeight: selected ? FontWeight.w700 : FontWeight.w500,
              color: selected ? _kPrimary : Colors.grey.shade500,
            ),
          ),
        ],
      ),
    );
  }
}

class _SettingsScreen extends ConsumerWidget {
  const _SettingsScreen();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final locale = ref.watch(localeProvider);
    final t = ref.watch(translationsProvider);

    return SingleChildScrollView(
      child: Padding(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(t.tr('cashier'), style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Color(0xFF1E293B))),
          const SizedBox(height: 12),
          const _ShiftSection(),
          const SizedBox(height: 24),
          if (ref.watch(authProvider).canManageStaff) ...[
            Text(t.tr('staff_management'), style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Color(0xFF1E293B))),
            const SizedBox(height: 12),
            Card(
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
              child: Column(
                children: [
                  ListTile(
                    leading: const Icon(Icons.people_alt_outlined, color: Color(0xFF6366F1)),
                    title: Text(t.tr('staff_list'), style: const TextStyle(fontWeight: FontWeight.w600)),
                    trailing: const Icon(Icons.chevron_right),
                    onTap: () => Navigator.push(
                      context,
                      MaterialPageRoute(builder: (_) => const StaffScreen()),
                    ),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                  ),
                  const Divider(height: 1, indent: 16, endIndent: 16),
                  ListTile(
                    leading: const Icon(Icons.store_outlined, color: Color(0xFF6366F1)),
                    title: Text(t.tr('branches'), style: const TextStyle(fontWeight: FontWeight.w600)),
                    trailing: const Icon(Icons.chevron_right),
                    onTap: () => Navigator.push(
                      context,
                      MaterialPageRoute(builder: (_) => const BranchesScreen()),
                    ),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                  ),
                  const Divider(height: 1, indent: 16, endIndent: 16),
                  ListTile(
                    leading: const Icon(Icons.history, color: Color(0xFF6366F1)),
                    title: Text(t.tr('login_history'), style: const TextStyle(fontWeight: FontWeight.w600)),
                    trailing: const Icon(Icons.chevron_right),
                    onTap: () => Navigator.push(
                      context,
                      MaterialPageRoute(builder: (_) => const AuditLogScreen()),
                    ),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 24),
          ],
          if (ref.watch(authProvider).canManageCatalog) ...[
            Text(t.tr('qris_static'), style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Color(0xFF1E293B))),
            const SizedBox(height: 12),
            const _QrisStaticSection(),
            const SizedBox(height: 24),
          ],
          Text(t.tr('language'), style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Color(0xFF1E293B))),
          const SizedBox(height: 12),
          Card(
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
            child: Column(
              children: [
                _LangOption(
                  title: t.tr('indonesian'),
                  subtitle: 'Bahasa Indonesia',
                  selected: locale == AppLocale.id,
                  onTap: () => ref.read(localeProvider.notifier).setLocale(AppLocale.id),
                ),
                const Divider(height: 1, indent: 16, endIndent: 16),
                _LangOption(
                  title: t.tr('english'),
                  subtitle: 'English',
                  selected: locale == AppLocale.en,
                  onTap: () => ref.read(localeProvider.notifier).setLocale(AppLocale.en),
                ),
              ],
            ),
          ),
          const SizedBox(height: 24),
          Text(t.tr('printer'), style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Color(0xFF1E293B))),
          const SizedBox(height: 12),
          Card(
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
            child: ListTile(
              leading: const Icon(Icons.bluetooth, color: Color(0xFF6366F1)),
              title: Text(t.tr('printer'), style: const TextStyle(fontWeight: FontWeight.w600)),
              subtitle: Consumer(builder: (ctx, ref, _) {
                final connected = BluetoothPrinterService().isConnected;
                return Text(
                  connected ? t.tr('connected') : '',
                  style: TextStyle(color: connected ? Colors.green : Colors.grey.shade500, fontSize: 13),
                );
              }),
              trailing: const Icon(Icons.chevron_right),
              onTap: () => showPrinterDialog(context),
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
            ),
          ),
          const SizedBox(height: 24),
          Text(t.tr('sync'), style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Color(0xFF1E293B))),
          const SizedBox(height: 12),
          Card(
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
            child: Column(
              children: [
                ListTile(
                  leading: const Icon(Icons.sync, color: Color(0xFF6366F1)),
                  title: Text(t.tr('last_sync'), style: const TextStyle(fontWeight: FontWeight.w600)),
                  subtitle: Consumer(builder: (ctx, ref, _) {
                    final lastSync = ref.watch(lastSyncAtProvider);
                    return Text(
                      lastSync == null
                          ? t.tr('never')
                          : '${lastSync.hour.toString().padLeft(2, '0')}:${lastSync.minute.toString().padLeft(2, '0')}',
                      style: TextStyle(color: Colors.grey.shade500, fontSize: 13),
                    );
                  }),
                  trailing: FilledButton.tonal(
                    onPressed: () async {
                      final container = ProviderScope.containerOf(context, listen: false);
                      await container.read(productListProvider.notifier).refresh();
                      await container.read(categoriesProvider.notifier).refresh();
                      await container
                          .read(customerStoreProvider)
                          .load(container.read(customerRepositoryProvider));
                      if (context.mounted) {
                        ScaffoldMessenger.of(context).showSnackBar(
                          SnackBar(content: Text(t.tr('sync_complete'))),
                        );
                      }
                    },
                    child: Text(t.tr('sync_now')),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 24),
          if (_hasMultipleBranches(ref)) ...[
            SizedBox(
              width: double.infinity,
              child: OutlinedButton.icon(
                onPressed: () async {
                  await ref.read(authProvider.notifier).selectBranch(null);
                },
                icon: const Icon(Icons.store, color: Color(0xFF6366F1)),
                label: Text(t.tr('switch_branch'), style: const TextStyle(color: Color(0xFF6366F1))),
                style: OutlinedButton.styleFrom(
                  side: BorderSide(color: const Color(0xFF6366F1).withValues(alpha: 0.3)),
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                  padding: const EdgeInsets.symmetric(vertical: 16),
                ),
              ),
            ),
            const SizedBox(height: 12),
          ],
          SizedBox(
            width: double.infinity,
            child: OutlinedButton.icon(
              onPressed: () async {
                await ref.read(authProvider.notifier).logout();
              },
              icon: const Icon(Icons.logout, color: Colors.red),
              label: Text(t.tr('logout'), style: TextStyle(color: Colors.red.shade600)),
              style: OutlinedButton.styleFrom(
                side: BorderSide(color: Colors.red.shade200),
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                padding: const EdgeInsets.symmetric(vertical: 16),
              ),
            ),
          ),
          const SizedBox(height: 24),
        ],
      ),
      ),
    );
  }

  bool _hasMultipleBranches(WidgetRef ref) {
    final user = ref.read(authProvider).user;
    if (user == null) return false;
    final seen = <int>{};
    return user.roles.where((r) => r.branch != null && seen.add(r.branch!.id)).length > 1;
  }
}

class _ShiftSection extends ConsumerStatefulWidget {
  const _ShiftSection();

  @override
  ConsumerState<_ShiftSection> createState() => _ShiftSectionState();
}

class _ShiftSectionState extends ConsumerState<_ShiftSection> {
  Shift? _shift;
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    final branchId = ref.read(authProvider).selectedBranch?.id ?? 0;
    if (branchId == 0) {
      if (mounted) setState(() => _loading = false);
      return;
    }
    try {
      final shift = await ref.read(shiftRepositoryProvider).getActiveShift(branchId);
      if (mounted) setState(() { _shift = shift; _loading = false; });
    } catch (_) {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _openShift() async {
    final t = ref.read(translationsProvider);
    final controller = TextEditingController(text: '0');
    final amount = await showDialog<int>(
      context: context,
      builder: (dialogCtx) => AlertDialog(
        title: Text(t.tr('open_shift')),
        content: TextField(
          controller: controller,
          keyboardType: TextInputType.number,
          autofocus: true,
          decoration: InputDecoration(labelText: t.tr('starting_cash'), prefixText: 'Rp '),
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(dialogCtx), child: Text(t.tr('cancel'))),
          FilledButton(
            onPressed: () => Navigator.pop(dialogCtx, int.tryParse(controller.text.trim()) ?? 0),
            child: Text(t.tr('save')),
          ),
        ],
      ),
    );
    if (amount == null || !mounted) return;
    final branchId = ref.read(authProvider).selectedBranch?.id ?? 0;
    try {
      final shift = await ref.read(shiftRepositoryProvider).openShift(branchId, amount);
      if (mounted) setState(() => _shift = shift);
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(t.tr('shift_open_failed')), backgroundColor: Colors.red),
        );
      }
    }
  }

  Future<void> _closeShift() async {
    final shift = _shift;
    if (shift == null) return;
    final t = ref.read(translationsProvider);
    final controller = TextEditingController();
    final amount = await showDialog<int>(
      context: context,
      builder: (dialogCtx) => AlertDialog(
        title: Text(t.tr('close_shift')),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('${t.tr('starting_cash')}: Rp ${formatRupiah(shift.startingCash)}',
                style: TextStyle(color: Colors.grey.shade600)),
            const SizedBox(height: 12),
            TextField(
              controller: controller,
              keyboardType: TextInputType.number,
              autofocus: true,
              decoration: InputDecoration(labelText: t.tr('actual_cash_counted'), prefixText: 'Rp '),
            ),
          ],
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(dialogCtx), child: Text(t.tr('cancel'))),
          FilledButton(
            onPressed: () => Navigator.pop(dialogCtx, int.tryParse(controller.text.trim()) ?? 0),
            child: Text(t.tr('save')),
          ),
        ],
      ),
    );
    if (amount == null || !mounted) return;
    try {
      final closed = await ref.read(shiftRepositoryProvider).closeShift(shift.id, amount);
      if (!mounted) return;
      setState(() => _shift = null);
      _showCloseSummary(closed, t);
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(t.tr('shift_close_failed')), backgroundColor: Colors.red),
        );
      }
    }
  }

  void _showCloseSummary(Shift closed, Translations t) {
    final variance = closed.variance ?? 0;
    final varianceColor = variance == 0
        ? Colors.green.shade700
        : (variance < 0 ? Colors.red.shade600 : Colors.orange.shade700);
    showDialog(
      context: context,
      builder: (dialogCtx) => AlertDialog(
        title: Text(t.tr('shift_closed')),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _kv(t.tr('expected_cash'), closed.expectedCash ?? 0),
            const SizedBox(height: 6),
            _kv(t.tr('actual_cash_counted'), closed.actualCash ?? 0),
            const Divider(height: 20),
            _kv(t.tr('variance'), variance, color: varianceColor),
          ],
        ),
        actions: [
          FilledButton(onPressed: () => Navigator.pop(dialogCtx), child: Text(t.tr('got_it'))),
        ],
      ),
    );
  }

  Widget _kv(String label, int amount, {Color? color}) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Text(label),
        Text('Rp ${formatRupiah(amount)}', style: TextStyle(fontWeight: FontWeight.w700, color: color)),
      ],
    );
  }

  String _formatTime(DateTime dt) => '${dt.hour.toString().padLeft(2, '0')}:${dt.minute.toString().padLeft(2, '0')}';

  @override
  Widget build(BuildContext context) {
    final t = ref.watch(translationsProvider);
    final shift = _shift;
    return Card(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: _loading
            ? const Center(child: Padding(padding: EdgeInsets.all(8), child: SaleForgesRingLoader(size: 32)))
            : Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Icon(Icons.point_of_sale, color: shift != null ? Colors.green.shade600 : Colors.grey.shade400),
                      const SizedBox(width: 8),
                      Text(
                        shift != null ? t.tr('shift_open') : t.tr('shift_closed_status'),
                        style: const TextStyle(fontWeight: FontWeight.w700),
                      ),
                    ],
                  ),
                  if (shift != null) ...[
                    const SizedBox(height: 8),
                    Text('${t.tr('starting_cash')}: Rp ${formatRupiah(shift.startingCash)}',
                        style: TextStyle(fontSize: 13, color: Colors.grey.shade600)),
                    Text('${t.tr('opened_at')}: ${_formatTime(shift.openedAt.toLocal())}',
                        style: TextStyle(fontSize: 13, color: Colors.grey.shade600)),
                  ],
                  const SizedBox(height: 12),
                  SizedBox(
                    width: double.infinity,
                    child: FilledButton(
                      onPressed: shift != null ? _closeShift : _openShift,
                      style: FilledButton.styleFrom(
                        backgroundColor: shift != null ? Colors.red.shade500 : _kPrimary,
                      ),
                      child: Text(shift != null ? t.tr('close_shift') : t.tr('open_shift')),
                    ),
                  ),
                ],
              ),
      ),
    );
  }
}

class _QrisStaticSection extends ConsumerStatefulWidget {
  const _QrisStaticSection();

  @override
  ConsumerState<_QrisStaticSection> createState() => _QrisStaticSectionState();
}

class _QrisStaticSectionState extends ConsumerState<_QrisStaticSection> {
  StaticQR? _qr;
  bool _loading = true;
  bool _uploading = false;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    final branchId = ref.read(authProvider).selectedBranch?.id ?? 0;
    final qr = await ref.read(paymentRepositoryProvider).getStaticQR(branchId);
    if (mounted) setState(() { _qr = qr; _loading = false; });
  }

  Future<void> _pickAndSave() async {
    final picked = await ImagePicker().pickImage(source: ImageSource.gallery, imageQuality: 85, maxWidth: 1024);
    if (picked == null || !mounted) return;
    setState(() => _uploading = true);

    final branchId = ref.read(authProvider).selectedBranch?.id ?? 0;
    final t = ref.read(translationsProvider);
    try {
      final url = await ref.read(productRepositoryProvider).uploadImage(picked.path);
      final qr = await ref.read(paymentRepositoryProvider).updateStaticQR(branchId: branchId, qrImage: url);
      if (mounted) setState(() { _qr = qr; _uploading = false; });
    } catch (_) {
      if (mounted) {
        setState(() => _uploading = false);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(t.tr('image_upload_failed')), backgroundColor: Colors.red),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final t = ref.watch(translationsProvider);

    if (_loading) {
      return const Card(child: Padding(padding: EdgeInsets.all(16), child: LinearProgressIndicator()));
    }

    return Card(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Row(
          children: [
            GestureDetector(
              onTap: _uploading ? null : _pickAndSave,
              child: Container(
                width: 72,
                height: 72,
                decoration: BoxDecoration(
                  color: Colors.grey.shade100,
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(color: Colors.grey.shade200),
                ),
                clipBehavior: Clip.antiAlias,
                child: _uploading
                    ? const Center(child: CircularProgressIndicator(strokeWidth: 2))
                    : (_qr != null && _qr!.qrImage.isNotEmpty)
                        ? CachedNetworkImage(imageUrl: _qr!.qrImage, fit: BoxFit.cover)
                        : Icon(Icons.qr_code_2, color: Colors.grey.shade400, size: 32),
              ),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    (_qr != null && _qr!.qrImage.isNotEmpty) ? t.tr('qris_static_set') : t.tr('qris_static_not_set'),
                    style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 14),
                  ),
                  const SizedBox(height: 4),
                  Text(t.tr('qris_static_hint'), style: TextStyle(fontSize: 12, color: Colors.grey.shade500)),
                  const SizedBox(height: 8),
                  OutlinedButton(
                    onPressed: _uploading ? null : _pickAndSave,
                    child: Text((_qr != null && _qr!.qrImage.isNotEmpty) ? t.tr('change_photo') : t.tr('upload_photo')),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _LangOption extends StatelessWidget {
  final String title;
  final String subtitle;
  final bool selected;
  final VoidCallback onTap;

  const _LangOption({required this.title, required this.subtitle, required this.selected, required this.onTap});

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(16),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
        child: Row(
          children: [
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
            Icon(
              selected ? Icons.radio_button_checked : Icons.radio_button_off,
              color: selected ? const Color(0xFF6366F1) : Colors.grey.shade400,
            ),
          ],
        ),
      ),
    );
  }
}
