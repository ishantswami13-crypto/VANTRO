class SavingPot {
  final String id;
  final String name;
  final int targetCents;
  final int savedCents;

  SavingPot({required this.id, required this.name, required this.targetCents, required this.savedCents});

  factory SavingPot.fromJson(Map<String, dynamic> j) => SavingPot(
        id: j['id']?.toString() ?? '',
        name: j['name']?.toString() ?? '',
        targetCents: (j['target_cents'] ?? 0) as int,
        savedCents: (j['saved_cents'] ?? 0) as int,
      );

  double get progress => targetCents == 0 ? 0 : savedCents / targetCents;
}
