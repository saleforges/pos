import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../auth/presentation/providers/auth_provider.dart';
import '../../data/order_store.dart';
import '../../../../core/config/translations.dart';
import 'pos_screen.dart';
import 'orders_screen.dart';
import '../../data/bluetooth_printer_service.dart';

final orderStoreProvider = Provider<OrderStore>((ref) => OrderStore());

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

  @override
  Widget build(BuildContext context) {
    final t = ref.watch(translationsProvider);
    final titles = [t.tr('catalog'), t.tr('orders'), t.tr('history'), t.tr('customers'), t.tr('settings')];
    final wide = MediaQuery.of(context).size.width >= 720;

    return Scaffold(
      appBar: wide ? null : AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        scrolledUnderElevation: 0.5,
        title: Text(titles[_selectedIndex], style: const TextStyle(fontWeight: FontWeight.w700, color: Color(0xFF1E293B))),
        centerTitle: true,
        actions: [
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
      body: _buildBody(titles[_selectedIndex]),
      bottomNavigationBar: wide ? null : _buildBottomNav(t),
    );
  }

  Widget _buildBody(String title) {
    switch (_selectedIndex) {
      case 0: return const PosScreen();
      case 1: return const OrdersScreen();
      case 2: return _buildPlaceholder(title);
      case 3: return _buildPlaceholder(title);
      case 4: return const _SettingsScreen();
      default: return const PosScreen();
    }
  }

  Widget _buildPlaceholder(String title) {
    final t = ref.read(translationsProvider);
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.construction, size: 64, color: Colors.grey.shade300),
          const SizedBox(height: 16),
          Text(title, style: TextStyle(color: Colors.grey.shade500, fontSize: 18)),
          const SizedBox(height: 8),
          Text(t.tr('coming_soon'), style: TextStyle(color: Colors.grey.shade400, fontSize: 14)),
        ],
      ),
    );
  }

  Widget _buildBottomNav(Translations t) {
    final titles = [t.tr('catalog'), t.tr('orders'), t.tr('history'), t.tr('customers'), t.tr('settings')];
    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        boxShadow: [
          BoxShadow(color: Colors.black.withValues(alpha: 0.06), blurRadius: 12, offset: const Offset(0, -3)),
        ],
      ),
      child: SafeArea(
        top: false,
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: 4),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceAround,
            children: List.generate(_navItems.length, (i) {
              final selected = _selectedIndex == i;
              final item = _navItems[i];
              return GestureDetector(
                onTap: () => setState(() => _selectedIndex = i),
                behavior: HitTestBehavior.opaque,
                child: Container(
                  padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                  decoration: BoxDecoration(
                    color: selected ? const Color(0xFF6366F1).withValues(alpha: 0.1) : Colors.transparent,
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(
                        selected ? item.activeIcon : item.icon,
                        color: selected ? const Color(0xFF6366F1) : Colors.grey.shade400,
                        size: 24,
                      ),
                      const SizedBox(height: 2),
                      Text(
                        titles[i],
                        style: TextStyle(
                          fontSize: 11,
                          fontWeight: selected ? FontWeight.w700 : FontWeight.w500,
                          color: selected ? const Color(0xFF6366F1) : Colors.grey.shade500,
                        ),
                      ),
                    ],
                  ),
                ),
              );
            }),
          ),
        ),
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

    return Padding(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
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
          const Spacer(),
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
    );
  }

  bool _hasMultipleBranches(WidgetRef ref) {
    final user = ref.read(authProvider).user;
    if (user == null) return false;
    final seen = <int>{};
    return user.roles.where((r) => seen.add(r.branch.id)).length > 1;
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


