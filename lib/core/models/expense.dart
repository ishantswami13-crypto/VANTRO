class Expense {
  final String id;
  final int amountCents;
  final String category;
  final String? mood;
  final String? note;
  final DateTime spentAt;

  Expense({
    required this.id,
    required this.amountCents,
    required this.category,
    this.mood,
    this.note,
    required this.spentAt,
  });

  factory Expense.fromJson(Map<String, dynamic> j) => Expense(
        id: j['id']?.toString() ?? '',
        amountCents: (j['amount_cents'] ?? 0) as int,
        category: j['category']?.toString() ?? '',
        mood: j['mood']?.toString(),
        note: j['note']?.toString(),
        spentAt: DateTime.parse(j['spent_at'] ?? DateTime.now().toIso8601String()),
      );
}
