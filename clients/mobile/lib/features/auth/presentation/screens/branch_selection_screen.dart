import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/auth_provider.dart';
import '../../../../shared/models/user.dart';
import '../../../../core/config/translations.dart';

class BranchSelectionScreen extends ConsumerWidget {
  const BranchSelectionScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final t = ref.watch(translationsProvider);
    final authState = ref.watch(authProvider);
    final roles = authState.user?.roles ?? [];

    return Scaffold(
      appBar: AppBar(title: Text(t.tr('select_branch'))),
      body: roles.isEmpty
          ? Center(child: Text(t.tr('no_branches')))
          : _BranchList(roles: roles, translations: t),
    );
  }
}

class _BranchList extends StatefulWidget {
  final List<Role> roles;
  final Translations translations;

  const _BranchList({required this.roles, required this.translations});

  @override
  State<_BranchList> createState() => _BranchListState();
}

class _BranchListState extends State<_BranchList> {
  Role? _selected;

  @override
  Widget build(BuildContext context) {
    final t = widget.translations;
    return Column(
      children: [
        const SizedBox(height: 8),
        Expanded(
          child: ListView.builder(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            itemCount: widget.roles.length,
            itemBuilder: (context, index) {
              final role = widget.roles[index];
              final selected = _selected == role;
              return Padding(
                padding: const EdgeInsets.only(bottom: 12),
                child: InkWell(
                  onTap: () => setState(() => _selected = role),
                  borderRadius: BorderRadius.circular(16),
                  child: Container(
                    padding: const EdgeInsets.all(16),
                    decoration: BoxDecoration(
                      color: selected ? const Color(0xFF6366F1).withValues(alpha: 0.06) : Colors.white,
                      borderRadius: BorderRadius.circular(16),
                      border: Border.all(
                        color: selected ? const Color(0xFF6366F1) : Colors.grey.shade200,
                        width: selected ? 2 : 1,
                      ),
                    ),
                    child: Row(
                      children: [
                        Container(
                          width: 48,
                          height: 48,
                          decoration: BoxDecoration(
                            color: selected
                                ? const Color(0xFF6366F1)
                                : Colors.grey.shade100,
                            borderRadius: BorderRadius.circular(12),
                          ),
                          child: Icon(
                            Icons.store,
                            color: selected ? Colors.white : Colors.grey.shade600,
                            size: 24,
                          ),
                        ),
                        const SizedBox(width: 16),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                role.branch.name,
                                style: const TextStyle(
                                  fontWeight: FontWeight.w600,
                                  fontSize: 16,
                                  color: Color(0xFF1E293B),
                                ),
                              ),
                              const SizedBox(height: 4),
                              Text(
                                '${role.merchant.name} \u2014 ${role.name}',
                                style: TextStyle(
                                  color: Colors.grey.shade500,
                                  fontSize: 13,
                                ),
                              ),
                            ],
                          ),
                        ),
                        Radio<Role>(
                          value: role,
                          groupValue: _selected,
                          onChanged: (v) => setState(() => _selected = v),
                          activeColor: const Color(0xFF6366F1),
                        ),
                      ],
                    ),
                  ),
                ),
              );
            },
          ),
        ),
        Padding(
          padding: const EdgeInsets.all(16),
          child: SizedBox(
            width: double.infinity,
            child: FilledButton(
              onPressed: _selected == null ? null : () async {
                await ProviderScope.containerOf(context, listen: false)
                  .read(authProvider.notifier)
                  .selectBranch(_selected!.branch);
              },
              child: Text(t.tr('continue'), style: const TextStyle(fontSize: 16)),
            ),
          ),
        ),
      ],
    );
  }
}
