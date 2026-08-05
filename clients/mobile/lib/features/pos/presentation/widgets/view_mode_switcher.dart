import 'package:flutter/material.dart';
import '../../data/view_mode_store.dart';

class ViewModeSwitcher extends StatelessWidget {
  final ViewMode currentMode;
  final ValueChanged<ViewMode> onModeChanged;

  const ViewModeSwitcher({
    super.key,
    required this.currentMode,
    required this.onModeChanged,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: const EdgeInsets.only(right: 16),
      decoration: BoxDecoration(
        color: Colors.grey.shade100,
        borderRadius: BorderRadius.circular(20),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          _buildButton(
            icon: Icons.grid_view,
            mode: ViewMode.grid,
            isSelected: currentMode == ViewMode.grid,
          ),
          _buildButton(
            icon: Icons.view_list,
            mode: ViewMode.list,
            isSelected: currentMode == ViewMode.list,
          ),
        ],
      ),
    );
  }

  Widget _buildButton({
    required IconData icon,
    required ViewMode mode,
    required bool isSelected,
  }) {
    return GestureDetector(
      onTap: () => onModeChanged(mode),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
        decoration: BoxDecoration(
          color: isSelected ? const Color(0xFF6366F1) : Colors.transparent,
          borderRadius: BorderRadius.circular(20),
        ),
        child: Icon(
          icon,
          size: 16,
          color: isSelected ? Colors.white : Colors.grey.shade500,
        ),
      ),
    );
  }
}
