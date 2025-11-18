import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:supabase_flutter/supabase_flutter.dart';
import 'features/auth/login_screen.dart';
import 'features/home/home_screen.dart';
import 'features/expenses/expenses_screen.dart';
import 'features/pots/pots_screen.dart';
import 'features/coach/coach_screen.dart';
import 'features/profile/profile_screen.dart';

GoRouter createRouter() {
  return GoRouter(
    initialLocation: '/',
    redirect: (context, state) {
      final hasSession = Supabase.instance.client.auth.currentSession != null;
      final loggingIn = state.matchedLocation == '/login';
      if (!hasSession && !loggingIn) return '/login';
      if (hasSession && loggingIn) return '/';
      return null;
    },
    routes: [
      GoRoute(path: '/login', pageBuilder: (_, __) => const MaterialPage(child: LoginScreen())),
      GoRoute(
        path: '/',
        pageBuilder: (context, state) => const MaterialPage(child: HomeScreen()),
        routes: [
          GoRoute(path: 'expenses', pageBuilder: (_, __) => const MaterialPage(child: ExpensesScreen())),
          GoRoute(path: 'pots', pageBuilder: (_, __) => const MaterialPage(child: PotsScreen())),
          GoRoute(path: 'coach', pageBuilder: (_, __) => const MaterialPage(child: CoachScreen())),
          GoRoute(path: 'profile', pageBuilder: (_, __) => const MaterialPage(child: ProfileScreen())),
        ],
      ),
    ],
  );
}
