import 'package:flutter_riverpod/flutter_riverpod.dart';

enum ViewMode { grid, list }

class ViewModeNotifier extends StateNotifier<ViewMode> {
  ViewModeNotifier() : super(ViewMode.grid);

  void setViewMode(ViewMode mode) {
    state = mode;
  }

  void toggle() {
    state = state == ViewMode.grid ? ViewMode.list : ViewMode.grid;
  }
}

final viewModeProvider = StateNotifierProvider<ViewModeNotifier, ViewMode>((ref) {
  return ViewModeNotifier();
});
