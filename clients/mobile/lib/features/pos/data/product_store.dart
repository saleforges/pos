import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../shared/models/product.dart';

class ProductStore extends StateNotifier<List<Product>> {
  ProductStore() : super(_defaultProducts);

  static final List<Product> _defaultProducts = [
    Product(
      id: '1',
      name: 'Indomie Goreng',
      sellingPrice: 3500,
      stock: 50,
      category: 'Food',
      image: 'https://upload.wikimedia.org/wikipedia/commons/thumb/9/94/Indomie_Mi_Goreng_Aceh.jpg/330px-Indomie_Mi_Goreng_Aceh.jpg',
    ),
    Product(
      id: '2',
      name: 'Indomie Kuah Soto',
      sellingPrice: 3500,
      stock: 30,
      category: 'Food',
      image: 'https://upload.wikimedia.org/wikipedia/commons/thumb/d/da/Indomie_Soto_Spesial.jpg/330px-Indomie_Soto_Spesial.jpg',
    ),
    Product(
      id: '3',
      name: 'Teh Pucuk 350ml',
      sellingPrice: 4000,
      stock: 24,
      category: 'Beverage',
      image: 'https://upload.wikimedia.org/wikipedia/commons/thumb/7/72/Minuman_Kemasan_Teh_Pucuk_Harum.jpg/330px-Minuman_Kemasan_Teh_Pucuk_Harum.jpg',
    ),
    Product(
      id: '4',
      name: 'Aqua 600ml',
      sellingPrice: 5000,
      stock: 20,
      category: 'Beverage',
      image: 'https://upload.wikimedia.org/wikipedia/commons/thumb/6/6e/Air_Minum_Dalam_Botol_Kemasan.jpg/330px-Air_Minum_Dalam_Botol_Kemasan.jpg',
    ),
    Product(
      id: '5',
      name: 'Rokok Surya 12',
      sellingPrice: 18000,
      stock: 15,
      category: 'Merchandise',
      image: 'https://upload.wikimedia.org/wikipedia/commons/thumb/a/a2/Rokok_Kretek_Kudus.jpg/330px-Rokok_Kretek_Kudus.jpg',
    ),
    Product(
      id: '6',
      name: 'Rokok Djarum Super',
      sellingPrice: 22000,
      stock: 10,
      category: 'Merchandise',
      image: 'https://upload.wikimedia.org/wikipedia/commons/thumb/e/e5/Djarum-blacks-kretek.jpg/250px-Djarum-blacks-kretek.jpg',
    ),
    Product(
      id: '7',
      name: 'Kopi Susu',
      sellingPrice: 18000,
      stock: 40,
      category: 'Beverage',
      image: 'https://upload.wikimedia.org/wikipedia/commons/thumb/8/8e/Kopi_Susu_Sakarek.jpg/330px-Kopi_Susu_Sakarek.jpg',
    ),
    Product(
      id: '8',
      name: 'Es Teh Manis',
      sellingPrice: 8000,
      stock: 35,
      category: 'Beverage',
      image: 'https://upload.wikimedia.org/wikipedia/commons/thumb/6/64/Es_teh_manis.jpg/330px-Es_teh_manis.jpg',
    ),
    Product(
      id: '9',
      name: 'Nasi Goreng',
      sellingPrice: 25000,
      stock: 20,
      category: 'Food',
      image: 'https://upload.wikimedia.org/wikipedia/commons/thumb/3/3e/Nasi_goreng_indonesia.jpg/330px-Nasi_goreng_indonesia.jpg',
    ),
    Product(
      id: '10',
      name: 'Mie Ayam',
      sellingPrice: 22000,
      stock: 18,
      category: 'Food',
      image: 'https://upload.wikimedia.org/wikipedia/commons/thumb/8/82/Mi_ayam_jamur.JPG/330px-Mi_ayam_jamur.JPG',
    ),
  ];

  List<Product> filter({String? category, String? searchQuery}) {
    var list = state;
    if (category != null && category != 'all') {
      list = list.where((p) => p.category.toLowerCase() == category.toLowerCase()).toList();
    }
    if (searchQuery != null && searchQuery.isNotEmpty) {
      final q = searchQuery.toLowerCase();
      list = list.where((p) => p.name.toLowerCase().contains(q)).toList();
    }
    return list;
  }
}

final productStoreProvider = StateNotifierProvider<ProductStore, List<Product>>((ref) {
  return ProductStore();
});
