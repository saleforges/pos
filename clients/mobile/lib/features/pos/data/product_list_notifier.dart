import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../shared/models/product.dart';
import 'repositories/product_repository.dart';

class ProductListNotifier extends StateNotifier<List<Product>> {
  final ProductRepository _repository;

  ProductListNotifier(this._repository) : super([]) {
    _init();
  }

  Future<void> _init() async {
    final cached = await _repository.getCachedProducts();
    if (cached.isNotEmpty) {
      state = cached;
    }
    try {
      final fresh = await _repository.refreshProducts();
      if (!_listsEqual(state, fresh)) {
        state = fresh;
      }
    } catch (_) {
      // Keep using cached data when offline
    }
  }

  Future<void> refresh() async {
    try {
      final fresh = await _repository.refreshProducts();
      if (!_listsEqual(state, fresh)) {
        state = fresh;
      }
    } catch (_) {
      // Keep current data when offline
    }
  }

  bool _listsEqual(List<Product> a, List<Product> b) {
    if (a.length != b.length) return false;
    for (var i = 0; i < a.length; i++) {
      if (a[i].id != b[i].id || a[i].updatedAt != b[i].updatedAt) return false;
    }
    return true;
  }
}
