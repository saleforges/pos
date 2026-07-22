class Product {
  final String id;
  final String name;
  final int price;
  final int stock;
  final String category;
  final String? image;
  final String? subtitle;
  final bool isFavorite;
  final bool hasVariants;

  const Product({
    required this.id,
    required this.name,
    required this.price,
    this.stock = 0,
    this.category = '',
    this.image,
    this.subtitle,
    this.isFavorite = false,
    this.hasVariants = false,
  });

  factory Product.fromJson(Map<String, dynamic> json) {
    return Product(
      id: json['id']?.toString() ?? '',
      name: json['name'] ?? '',
      price: json['price'] ?? 0,
      stock: json['stock'] ?? 0,
      category: json['category'] ?? '',
      image: json['image'],
      subtitle: json['subtitle'],
      isFavorite: json['isFavorite'] ?? false,
      hasVariants: json['hasVariants'] ?? false,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'name': name,
      'price': price,
      'stock': stock,
      'category': category,
      'image': image,
      'subtitle': subtitle,
      'isFavorite': isFavorite,
      'hasVariants': hasVariants,
    };
  }

  Product copyWith({
    String? id,
    String? name,
    int? price,
    int? stock,
    String? category,
    String? image,
    String? subtitle,
    bool? isFavorite,
    bool? hasVariants,
  }) {
    return Product(
      id: id ?? this.id,
      name: name ?? this.name,
      price: price ?? this.price,
      stock: stock ?? this.stock,
      category: category ?? this.category,
      image: image ?? this.image,
      subtitle: subtitle ?? this.subtitle,
      isFavorite: isFavorite ?? this.isFavorite,
      hasVariants: hasVariants ?? this.hasVariants,
    );
  }
}
