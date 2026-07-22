import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../core/network/api_client.dart';
import '../../../../core/storage/session_local_data_source.dart';
import '../../../../shared/models/user.dart';
import '../../data/auth_repository.dart';
import '../../data/remote/auth_remote_data_source.dart';
import '../../data/local/auth_local_data_source.dart';

final apiClientProvider = Provider<ApiClient>((ref) {
  return ApiClient();
});

final sessionLocalDataSourceProvider = Provider<SessionLocalDataSource>((ref) {
  return SessionLocalDataSource();
});

final authRemoteDataSourceProvider = Provider<AuthRemoteDataSource>((ref) {
  return AuthRemoteDataSource(ref.read(apiClientProvider));
});

final authLocalDataSourceProvider = Provider<AuthLocalDataSource>((ref) {
  return AuthLocalDataSource();
});

final authRepositoryProvider = Provider<AuthRepository>((ref) {
  return AuthRepository(
    remoteDataSource: ref.read(authRemoteDataSourceProvider),
    localDataSource: ref.read(authLocalDataSourceProvider),
    sessionDataSource: ref.read(sessionLocalDataSourceProvider),
  );
});

enum AuthStatus { initial, loading, authenticated, unauthenticated, error }

class AuthState {
  final AuthStatus status;
  final User? user;
  final Branch? selectedBranch;
  final String? error;

  AuthState({
    this.status = AuthStatus.initial,
    this.user,
    this.selectedBranch,
    this.error,
  });

  AuthState copyWith({
    AuthStatus? status,
    User? user,
    Branch? selectedBranch,
    String? error,
  }) {
    return AuthState(
      status: status ?? this.status,
      user: user ?? this.user,
      selectedBranch: selectedBranch ?? this.selectedBranch,
      error: error ?? this.error,
    );
  }
}

class AuthNotifier extends StateNotifier<AuthState> {
  final AuthRepository _repository;

  AuthNotifier(this._repository) : super(AuthState()) {
    _checkAuth();
  }

  Future<void> _checkAuth() async {
    final isLoggedIn = await _repository.isLoggedIn();
    if (isLoggedIn) {
      try {
        final user = await _repository.getCurrentUser();
        final uniqueBranches = _getUniqueBranches(user);
        final savedBranchId = await _repository.getSelectedBranchId();

        Branch? selectedBranch;
        if (savedBranchId != null) {
          selectedBranch = uniqueBranches.where((b) => b.id == savedBranchId).firstOrNull;
        }

        if (selectedBranch == null && uniqueBranches.length == 1) {
          selectedBranch = uniqueBranches.first;
          await _repository.saveSelectedBranchId(selectedBranch.id);
        }

        state = AuthState(
          status: AuthStatus.authenticated,
          user: user,
          selectedBranch: selectedBranch,
        );
      } catch (_) {
        state = AuthState(status: AuthStatus.unauthenticated);
      }
    } else {
      state = AuthState(status: AuthStatus.unauthenticated);
    }
  }

  Future<void> login(String username, String password) async {
    state = state.copyWith(status: AuthStatus.loading, error: null);
    try {
      await _repository.login(username, password);
      final user = await _repository.getCurrentUser();
      final uniqueBranches = _getUniqueBranches(user);

      Branch? selectedBranch;
      if (uniqueBranches.length == 1) {
        selectedBranch = uniqueBranches.first;
        await _repository.saveSelectedBranchId(selectedBranch.id);
      }

      state = AuthState(
        status: AuthStatus.authenticated,
        user: user,
        selectedBranch: selectedBranch,
      );
    } catch (e) {
      state = AuthState(
        status: AuthStatus.error,
        error: e.toString().replaceFirst('Exception: ', ''),
      );
    }
  }

  Future<void> selectBranch(Branch? branch) async {
    if (branch != null) {
      await _repository.saveSelectedBranchId(branch.id);
    } else {
      await _repository.clearSelectedBranchId();
    }
    state = state.copyWith(selectedBranch: branch);
  }

  Future<void> logout() async {
    await _repository.logout();
    state = AuthState(status: AuthStatus.unauthenticated);
  }

  List<Branch> _getUniqueBranches(User user) {
    final seen = <int>{};
    return user.roles.where((r) => seen.add(r.branch.id)).map((r) => r.branch).toList();
  }
}

final authProvider = StateNotifierProvider<AuthNotifier, AuthState>((ref) {
  return AuthNotifier(ref.read(authRepositoryProvider));
});
