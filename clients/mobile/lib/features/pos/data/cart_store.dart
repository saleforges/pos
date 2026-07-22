import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../shared/models/product.dart';

class CartItem {
  final Product product;
  final int quantity;
  final String notes;

  const CartItem({
    required this.product,
    this.quantity = 1,
    this.notes = '',
  });

  CartItem copyWith({Product? product, int? quantity, String? notes}) {
    return CartItem(
      product: product ?? this.product,
      quantity: quantity ?? this.quantity,
      notes: notes ?? this.notes,
    );
  }

  int get totalPrice => product.price * quantity;
}

class CartState {
  final List<CartItem> items;

  const CartState({this.items = const []});

  int get totalItems => items.fold(0, (sum, item) => sum + item.quantity);
  int get totalPrice => items.fold(0, (sum, item) => sum + item.totalPrice);

  CartState copyWith({List<CartItem>? items}) {
    return CartState(items: items ?? this.items);
  }
}

class CartNotifier extends StateNotifier<CartState> {
  CartNotifier() : super(const CartState());

  void addItem(Product product) {
    final existingIndex = state.items.indexWhere((item) => item.product.id == product.id);
    if (existingIndex >= 0) {
      final updatedItems = List<CartItem>.from(state.items);
      updatedItems[existingIndex] = updatedItems[existingIndex].copyWith(
        quantity: updatedItems[existingIndex].quantity + 1,
      );
      state = state.copyWith(items: updatedItems);
    } else {
      state = state.copyWith(
        items: [...state.items, CartItem(product: product)],
      );
    }
  }

  void removeItem(String productId) {
    state = state.copyWith(
      items: state.items.where((item) => item.product.id != productId).toList(),
    );
  }

  void updateQuantity(String productId, int quantity) {
    if (quantity <= 0) {
      removeItem(productId);
      return;
    }
    final updatedItems = state.items.map((item) {
      if (item.product.id == productId) {
        return item.copyWith(quantity: quantity);
      }
      return item;
    }).toList();
    state = state.copyWith(items: updatedItems);
  }

  void updateNotes(String productId, String notes) {
    final updatedItems = state.items.map((item) {
      if (item.product.id == productId) {
        return item.copyWith(notes: notes);
      }
      return item;
    }).toList();
    state = state.copyWith(items: updatedItems);
  }

  void clear() {
    state = const CartState();
  }

  int getItemQuantity(String productId) {
    final item = state.items.where((item) => item.product.id == productId);
    return item.isNotEmpty ? item.first.quantity : 0;
  }
}

final cartProvider = StateNotifierProvider<CartNotifier, CartState>((ref) {
  return CartNotifier();
});
