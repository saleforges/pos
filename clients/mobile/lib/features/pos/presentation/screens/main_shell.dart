import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../auth/presentation/providers/auth_provider.dart';
import '../../../auth/presentation/screens/branch_selection_screen.dart';
import '../../../auth/presentation/screens/login_screen.dart';
import '../../data/order_store.dart';
import 'pos_screen.dart';
import 'orders_screen.dart';

final orderStoreProvider = Provider<OrderStore>((ref) => OrderStore());

class MainShell extends StatefulWidget {
  const MainShell({super.key});

  @override
  State<MainShell> createState() => _MainShellState();
}

class _MainShellState extends State<MainShell> {
  int _selectedIndex = 0;

  static const _titles = ['Cashier', 'Orders'];

  @override
  Widget build(BuildContext context) {
    final wide = MediaQuery.of(context).size.width >= 720;

    return Scaffold(
      appBar: wide ? null : AppBar(
        title: Text(_titles[_selectedIndex]),
        actions: [
          Consumer(builder: (context, ref, _) {
            final user = ref.watch(authProvider).user;
            return Padding(
              padding: const EdgeInsets.only(right: 8),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(Icons.person, size: 18, color: Colors.grey.shade600),
                  const SizedBox(width: 4),
                  Text(user?.username ?? '', style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w500)),
                ],
              ),
            );
          }),
        ],
      ),
      drawer: wide ? null : _buildDrawer(),
      body: wide ? _buildDesktop() : _buildMobilePage(),
      bottomNavigationBar: wide ? null : BottomNavigationBar(
        currentIndex: _selectedIndex,
        onTap: (i) => setState(() => _selectedIndex = i),
        items: const [
          BottomNavigationBarItem(icon: Icon(Icons.shopping_cart_outlined), label: 'Cashier'),
          BottomNavigationBarItem(icon: Icon(Icons.receipt_long_outlined), label: 'Orders'),
        ],
      ),
    );
  }

  Widget _buildMobilePage() {
    if (_selectedIndex == 0) return const PosScreenBody();
    return const OrdersScreenBody();
  }

  Widget _buildDrawer() {
    return Drawer(
      child: SafeArea(
        child: Column(
          children: [
            Container(
              width: double.infinity,
              padding: const EdgeInsets.all(16),
              color: const Color(0xFF6366F1).withValues(alpha: 0.1),
              child: Consumer(builder: (context, ref, _) {
                final user = ref.watch(authProvider).user;
                final branch = ref.watch(authProvider).selectedBranch;
                return Column(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: const Icon(Icons.point_of_sale, color: Color(0xFF6366F1), size: 32),
                    ),
                    if (user != null) ...[
                      const SizedBox(height: 8),
                      Text(user.username, style: const TextStyle(fontWeight: FontWeight.bold, color: Color(0xFF1E293B))),
                      if (branch != null) Text(branch.name, style: TextStyle(color: Colors.grey.shade600, fontSize: 13)),
                    ],
                  ],
                );
              }),
            ),
            ListTile(
              selected: _selectedIndex == 0,
              leading: const Icon(Icons.shopping_cart_outlined),
              title: const Text('Cashier'),
              onTap: () { setState(() => _selectedIndex = 0); Navigator.pop(context); },
            ),
            ListTile(
              selected: _selectedIndex == 1,
              leading: const Icon(Icons.receipt_long_outlined),
              title: const Text('Orders'),
              onTap: () { setState(() => _selectedIndex = 1); Navigator.pop(context); },
            ),
            const Spacer(),
            ListTile(
              leading: const Icon(Icons.store),
              title: const Text('Switch Branch'),
              onTap: () {
                Navigator.pushAndRemoveUntil(
                  context,
                  MaterialPageRoute(builder: (_) => const BranchSelectionScreen()),
                  (route) => false,
                );
              },
            ),
            ListTile(
              leading: const Icon(Icons.logout),
              title: const Text('Logout'),
              onTap: () {
                ProviderScope.containerOf(context, listen: false)
                  .read(authProvider.notifier)
                  .logout();
                Navigator.pushAndRemoveUntil(
                  context,
                  MaterialPageRoute(builder: (_) => const LoginScreen()),
                  (route) => false,
                );
              },
            ),
            const SizedBox(height: 16),
          ],
        ),
      ),
    );
  }

  Widget _buildDesktop() {
    return Row(
      children: [
        NavigationRail(
          selectedIndex: _selectedIndex,
          onDestinationSelected: (i) => setState(() => _selectedIndex = i),
          labelType: NavigationRailLabelType.all,
          backgroundColor: Colors.white,
            leading: Padding(
            padding: const EdgeInsets.symmetric(vertical: 12),
            child: Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: const Color(0xFF6366F1).withValues(alpha: 0.1),
                borderRadius: BorderRadius.circular(12),
              ),
              child: const Icon(Icons.point_of_sale, color: Color(0xFF6366F1), size: 28),
            ),
          ),
          destinations: const [
            NavigationRailDestination(
              icon: Icon(Icons.shopping_cart_outlined),
              selectedIcon: Icon(Icons.shopping_cart),
              label: Text('Cashier'),
            ),
            NavigationRailDestination(
              icon: Icon(Icons.receipt_long_outlined),
              selectedIcon: Icon(Icons.receipt_long),
              label: Text('Orders'),
            ),
          ],
            trailing: Padding(
            padding: const EdgeInsets.only(bottom: 16),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                IconButton(
                  icon: const Icon(Icons.store),
                  onPressed: () {
                    ProviderScope.containerOf(context, listen: false)
                      .read(authProvider.notifier)
                      .selectBranch(null);
                    Navigator.pushReplacement(
                      context,
                      MaterialPageRoute(builder: (_) => const BranchSelectionScreen()),
                    );
                  },
                  tooltip: 'Switch Branch',
                ),
                IconButton(
                  icon: const Icon(Icons.logout),
                  onPressed: () {
                    ProviderScope.containerOf(context, listen: false)
                      .read(authProvider.notifier)
                      .logout();
                    Navigator.pushAndRemoveUntil(
                      context,
                      MaterialPageRoute(builder: (_) => const LoginScreen()),
                      (route) => false,
                    );
                  },
                  tooltip: 'Logout',
                ),
              ],
            ),
          ),
          ),
          const VerticalDivider(width: 1, thickness: 1),
        Expanded(
          child: _selectedIndex == 0
              ? const PosScreen()
              : const OrdersScreen(),
        ),
      ],
    );
  }
}
