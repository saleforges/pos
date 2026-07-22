class Product {
  final String id;
  final String name;
  final String sku;
  final String? barcode;
  final String? description;
  final int sellingPrice;
  final int costPrice;
  final int stock;
  final String category;
  final String? image;
  final String? subtitle;
  final String unit;
  final bool isActive;
  final bool isFavorite;
  final bool hasVariants;
  final DateTime createdAt;
  final DateTime updatedAt;

  Product({
    required this.id,
    required this.name,
    this.sku = '',
    this.barcode,
    this.description,
    this.sellingPrice = 0,
    this.costPrice = 0,
    this.stock = 0,
    this.category = '',
    this.image,
    this.subtitle,
    this.unit = 'pcs',
    this.isActive = true,
    this.isFavorite = false,
    this.hasVariants = false,
    DateTime? createdAt,
    DateTime? updatedAt,
  })  : createdAt = createdAt ?? DateTime.now(),
        updatedAt = updatedAt ?? DateTime.now();

  int get price => sellingPrice;

  factory Product.fromJson(Map<String, dynamic> json) {
    return Product(
      id: json['id']?.toString() ?? '',
      name: json['name'] ?? '',
      sku: json['sku'] ?? '',
      barcode: json['barcode'],
      description: json['description'],
      sellingPrice: json['sellingPrice'] ?? json['price'] ?? 0,
      costPrice: json['costPrice'] ?? 0,
      stock: json['stock'] ?? 0,
      category: json['category'] ?? '',
      image: json['image'],
      subtitle: json['subtitle'],
      unit: json['unit'] ?? 'pcs',
      isActive: json['isActive'] ?? true,
      isFavorite: json['isFavorite'] ?? false,
      hasVariants: json['hasVariants'] ?? false,
      createdAt: json['createdAt'] != null ? DateTime.parse(json['createdAt']) : null,
      updatedAt: json['updatedAt'] != null ? DateTime.parse(json['updatedAt']) : null,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'name': name,
      'sku': sku,
      'barcode': barcode,
      'description': description,
      'sellingPrice': sellingPrice,
      'costPrice': costPrice,
      'stock': stock,
      'category': category,
      'image': image,
      'subtitle': subtitle,
      'unit': unit,
      'isActive': isActive,
      'isFavorite': isFavorite,
      'hasVariants': hasVariants,
      'createdAt': createdAt.toIso8601String(),
      'updatedAt': updatedAt.toIso8601String(),
    };
  }

  Product copyWith({
    String? id,
    String? name,
    String? sku,
    String? barcode,
    String? description,
    int? sellingPrice,
    int? costPrice,
    int? stock,
    String? category,
    String? image,
    String? subtitle,
    String? unit,
    bool? isActive,
    bool? isFavorite,
    bool? hasVariants,
    DateTime? createdAt,
    DateTime? updatedAt,
  }) {
    return Product(
      id: id ?? this.id,
      name: name ?? this.name,
      sku: sku ?? this.sku,
      barcode: barcode ?? this.barcode,
      description: description ?? this.description,
      sellingPrice: sellingPrice ?? this.sellingPrice,
      costPrice: costPrice ?? this.costPrice,
      stock: stock ?? this.stock,
      category: category ?? this.category,
      image: image ?? this.image,
      subtitle: subtitle ?? this.subtitle,
      unit: unit ?? this.unit,
      isActive: isActive ?? this.isActive,
      isFavorite: isFavorite ?? this.isFavorite,
      hasVariants: hasVariants ?? this.hasVariants,
      createdAt: createdAt ?? this.createdAt,
      updatedAt: updatedAt ?? this.updatedAt,
    );
  }
}
