class User {
  final int id;
  final String username;
  final String email;
  final String type;
  final String status;
  final List<Role> roles;
  final DateTime createdAt;
  final DateTime updatedAt;

  User({
    required this.id,
    required this.username,
    required this.email,
    required this.type,
    required this.status,
    required this.roles,
    required this.createdAt,
    required this.updatedAt,
  });

  factory User.fromJson(Map<String, dynamic> json) {
    return User(
      id: json['id'] ?? 0,
      username: json['username'] ?? '',
      email: json['email'] ?? '',
      type: json['type'] ?? '',
      status: json['status'] ?? '',
      roles: (json['roles'] as List<dynamic>?)
              ?.map((r) => Role.fromJson(r))
              .toList() ??
          [],
      createdAt: DateTime.parse(json['createdAt']),
      updatedAt: DateTime.parse(json['updatedAt']),
    );
  }
}

class Role {
  final int id;
  final String name;
  final Merchant merchant;
  final Branch branch;
  final String branchScope;
  final bool isDefault;

  Role({
    required this.id,
    required this.name,
    required this.merchant,
    required this.branch,
    required this.branchScope,
    required this.isDefault,
  });

  factory Role.fromJson(Map<String, dynamic> json) {
    return Role(
      id: json['id'] ?? 0,
      name: json['name'] ?? '',
      merchant: Merchant.fromJson(json['merchant']),
      branch: Branch.fromJson(json['branch']),
      branchScope: json['branchScope'] ?? '',
      isDefault: json['isDefault'] ?? false,
    );
  }
}

class Merchant {
  final int id;
  final String name;

  Merchant({required this.id, required this.name});

  factory Merchant.fromJson(Map<String, dynamic> json) {
    return Merchant(
      id: json['id'] ?? 0,
      name: json['name'] ?? '',
    );
  }
}

class Branch {
  final int id;
  final String name;

  Branch({required this.id, required this.name});

  factory Branch.fromJson(Map<String, dynamic> json) {
    return Branch(
      id: json['id'] ?? 0,
      name: json['name'] ?? '',
    );
  }
}
