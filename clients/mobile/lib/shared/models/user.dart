class User {
  final String id;
  final String username;
  final String email;
  final List<String> roles;
  final String type;
  final String status;
  final DateTime createdAt;
  final DateTime updatedAt;
  final List<Merchant> merchants;

  User({
    required this.id,
    required this.username,
    required this.email,
    required this.roles,
    required this.type,
    required this.status,
    required this.createdAt,
    required this.updatedAt,
    required this.merchants,
  });

  factory User.fromJson(Map<String, dynamic> json) {
    return User(
      id: json['id'] ?? '',
      username: json['username'] ?? '',
      email: json['email'] ?? '',
      roles: List<String>.from(json['roles'] ?? []),
      type: json['type'] ?? '',
      status: json['status'] ?? '',
      createdAt: DateTime.parse(json['createdAt'] ?? DateTime.now().toIso8601String()),
      updatedAt: DateTime.parse(json['updatedAt'] ?? DateTime.now().toIso8601String()),
      merchants: (json['merchants'] as List<dynamic>?)
              ?.map((m) => Merchant.fromJson(m))
              .toList() ??
          [],
    );
  }
}

class Merchant {
  final String id;
  final String name;
  final String role;

  Merchant({
    required this.id,
    required this.name,
    required this.role,
  });

  factory Merchant.fromJson(Map<String, dynamic> json) {
    return Merchant(
      id: json['id'] ?? '',
      name: json['name'] ?? '',
      role: json['role'] ?? '',
    );
  }
}
