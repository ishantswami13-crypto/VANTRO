class CoachPlan {
  final String weekStart;
  final List<String> rules;
  final String dailyNudge;
  final int healthScore;

  CoachPlan({required this.weekStart, required this.rules, required this.dailyNudge, required this.healthScore});

  factory CoachPlan.fromJson(Map<String, dynamic> j) => CoachPlan(
        weekStart: j['week_start']?.toString() ?? '',
        rules: (j['rules'] as List? ?? const []).map((e) => e.toString()).toList(),
        dailyNudge: j['daily_nudge']?.toString() ?? '',
        healthScore: (j['health_score'] ?? 0) as int,
      );
}
