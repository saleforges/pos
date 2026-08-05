class Customer {
  final String id;
  final String name;
  final String phone;
  final Map<String, int> customPrices;

  const Customer({
    required this.id,
    required this.name,
    this.phone = '',
    this.customPrices = const {},
  });
}
