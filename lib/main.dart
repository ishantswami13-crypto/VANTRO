import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:supabase_flutter/supabase_flutter.dart';
import 'theme.dart';
import 'router.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await Supabase.initialize(
    url: const String.fromEnvironment('SUPABASE_URL', defaultValue: 'https://YOUR-REF.supabase.co'),
    anonKey: const String.fromEnvironment('SUPABASE_ANON', defaultValue: 'YOUR_ANON_KEY'),
  );
  final routerConfig = createRouter();
  runApp(ProviderScope(child: VantroApp(router: routerConfig)));
}

class VantroApp extends StatelessWidget {
  final GoRouter router;
  const VantroApp({super.key, required this.router});

  @override
  Widget build(BuildContext context) {
    return MaterialApp.router(
      title: 'VANTRO',
      theme: vantroTheme(),
      routerConfig: router,
      debugShowCheckedModeBanner: false,
    );
  }
}
