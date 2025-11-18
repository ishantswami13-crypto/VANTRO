import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

class HomeScreen extends StatelessWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text("VANTRO")),
      body: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
          Text("Where Wealth Meets Wisdom.", style: Theme.of(context).textTheme.headlineMedium),
          const SizedBox(height: 18),
          _card(
            child: Row(mainAxisAlignment: MainAxisAlignment.spaceBetween, children: [
              const Text("Expenses"),
              FilledButton(onPressed: () => context.go('/expenses'), child: const Text("Open")),
            ]),
          ),
          const SizedBox(height: 12),
          _card(
            child: Row(mainAxisAlignment: MainAxisAlignment.spaceBetween, children: [
              const Text("Saving Pots"),
              FilledButton(onPressed: () => context.go('/pots'), child: const Text("Open")),
            ]),
          ),
          const SizedBox(height: 12),
          _card(
            child: Row(mainAxisAlignment: MainAxisAlignment.spaceBetween, children: [
              const Text("Coach"),
              FilledButton(onPressed: () => context.go('/coach'), child: const Text("Open")),
            ]),
          ),
        ]),
      ),
      bottomNavigationBar: NavigationBar(
        selectedIndex: 0,
        onDestinationSelected: (i) {
          switch (i) {
            case 0: context.go('/'); break;
            case 1: context.go('/expenses'); break;
            case 2: context.go('/pots'); break;
            case 3: context.go('/coach'); break;
            case 4: context.go('/profile'); break;
          }
        },
        destinations: const [
          NavigationDestination(icon: Icon(Icons.home_outlined), label: 'Home'),
          NavigationDestination(icon: Icon(Icons.receipt_long_outlined), label: 'Expenses'),
          NavigationDestination(icon: Icon(Icons.savings_outlined), label: 'Pots'),
          NavigationDestination(icon: Icon(Icons.auto_awesome_outlined), label: 'Coach'),
          NavigationDestination(icon: Icon(Icons.person_outline), label: 'Profile'),
        ],
      ),
    );
  }

  Widget _card({required Widget child}) => Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: Colors.white, borderRadius: BorderRadius.circular(16),
          boxShadow: const [BoxShadow(color: Color(0x11000000), blurRadius: 16, offset: Offset(0, 8))],
        ),
        child: child,
      );
}
