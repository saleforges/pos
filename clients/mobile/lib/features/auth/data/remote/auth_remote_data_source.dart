import 'package:dio/dio.dart';
import '../../../../core/config/api_config.dart';
import '../../../../core/network/api_client.dart';
import '../../../../shared/models/auth_response.dart';
import '../../../../shared/models/user.dart';

class AuthRemoteDataSource {
  final ApiClient _apiClient;

  AuthRemoteDataSource(this._apiClient);

  Future<LoginResponse> login(String username, String password) async {
    try {
      final response = await _apiClient.dio.post(
        '${ApiConfig.authEndpoint}/login',
        data: LoginRequest(username: username, password: password).toJson(),
        options: Options(headers: {'Authorization': null}),
      );

      final apiResponse = ApiResponse.fromJson(response.data, (data) {
        return LoginResponse.fromJson(data as Map<String, dynamic>);
      });

      if (apiResponse.data == null) {
        throw Exception(apiResponse.message);
      }

      return apiResponse.data!;
    } on DioException catch (e) {
      if (e.response?.data != null) {
        throw Exception(e.response!.data['message'] ?? 'Login failed');
      }
      throw Exception('Network error: ${e.message}');
    }
  }

  Future<User> getCurrentUser() async {
    try {
      final response = await _apiClient.dio.get(
        '${ApiConfig.authEndpoint}/me',
      );

      final apiResponse = ApiResponse.fromJson(response.data, (data) {
        return User.fromJson(data as Map<String, dynamic>);
      });

      if (apiResponse.data == null) {
        throw Exception(apiResponse.message);
      }

      return apiResponse.data!;
    } on DioException catch (e) {
      if (e.response?.data != null) {
        throw Exception(e.response!.data['message'] ?? 'Failed to get user');
      }
      throw Exception('Network error: ${e.message}');
    }
  }

  Future<void> logout() async {
    try {
      await _apiClient.dio.post('${ApiConfig.authEndpoint}/logout');
    } catch (e) {
      // Ignore logout errors
    }
  }
}
