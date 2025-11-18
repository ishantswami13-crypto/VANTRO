import 'package:flutter/material.dart';

ThemeData vantroTheme() {
  final base = ThemeData(
    brightness: Brightness.light,
    scaffoldBackgroundColor: Colors.white,
    colorScheme: ColorScheme.fromSeed(seedColor: const Color(0xFF0A84FF)),
    useMaterial3: true,
  );

  return base.copyWith(
    appBarTheme: const AppBarTheme(
      elevation: 0,
      backgroundColor: Colors.white,
      foregroundColor: Colors.black,
      centerTitle: true,
      titleTextStyle: TextStyle(
        fontSize: 18, fontWeight: FontWeight.w600, color: Colors.black, letterSpacing: -0.2),
    ),
    textTheme: base.textTheme.copyWith(
      headlineMedium: const TextStyle(
        fontWeight: FontWeight.w700, color: Colors.black, letterSpacing: -0.3),
      titleMedium: const TextStyle(fontWeight: FontWeight.w600),
      bodyMedium: const TextStyle(color: Colors.black87, height: 1.4),
      labelLarge: const TextStyle(fontWeight: FontWeight.w600, letterSpacing: 0.1),
    ),
    cardTheme: CardTheme(
      color: Colors.white,
      elevation: 0,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(18)),
      surfaceTintColor: Colors.white,
      shadowColor: const Color(0x14000000),
      margin: EdgeInsets.zero,
    ),
  );
}

/// Apple-ish container card with soft shadow
class MoneyCard extends StatelessWidget {
  final Widget child;
  final EdgeInsets padding;
  final EdgeInsets margin;
  const MoneyCard({
    super.key,
    required this.child,
    this.padding = const EdgeInsets.all(16),
    this.margin = const EdgeInsets.only(bottom: 12),
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: margin,
      padding: padding,
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(18),
        boxShadow: const [BoxShadow(
          color: Color(0x14000000), blurRadius: 18, offset: Offset(0, 10))],
        border: Border.all(color: const Color(0x0F000000)),
      ),
      child: child,
    );
  }
}

/// Rounded pill button (primary = black on white)
class PillButton extends StatelessWidget {
  final String text;
  final VoidCallback? onPressed;
  final bool tonal;
  const PillButton({super.key, required this.text, this.onPressed, this.tonal=false});

  @override
  Widget build(BuildContext context) {
    final style = tonal
        ? OutlinedButton.styleFrom(
            shape: const StadiumBorder(),
            side: const BorderSide(color: Color(0x22000000)),
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12))
        : FilledButton.styleFrom(
            shape: const StadiumBorder(),
            backgroundColor: Colors.black,
            foregroundColor: Colors.white,
            padding: const EdgeInsets.symmetric(horizontal: 18, vertical: 12));
    return tonal
        ? OutlinedButton(onPressed: onPressed, style: style, child: Text(text))
        : FilledButton(onPressed: onPressed, style: style, child: Text(text));
  }
}
