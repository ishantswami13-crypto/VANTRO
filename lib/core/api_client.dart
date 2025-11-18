import 'package:dio/dio.dart';
import 'package:supabase_flutter/supabase_flutter.dart';

class ApiClient {
  final Dio dio;
  static const String baseUrl = String.fromEnvironment('API_BASE', defaultValue: 'https://vantro.onrender.com');

  ApiClient() : dio = Dio(BaseOptions(
    baseUrl: baseUrl,
    headers: {'Content-Type': 'application/json'},
    connectTimeout: const Duration(seconds: 8),
    receiveTimeout: const Duration(seconds: 8),
  )) {
    dio.interceptors.add(InterceptorsWrapper(
      onRequest: (options, handler) async {
        final tok = Supabase.instance.client.auth.currentSession?.accessToken;
        if (tok != null) {
          options.headers['Authorization'] = 'Bearer $tok';
        }
        handler.next(options);
      },
    ));
  }
}
