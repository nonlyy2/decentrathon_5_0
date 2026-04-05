package models

type DashboardStats struct {
	TotalCandidates        int                        `json:"total_candidates"`
	Analyzed               int                        `json:"analyzed"`
	Pending                int                        `json:"pending"`
	Shortlisted            int                        `json:"shortlisted"`
	Rejected               int                        `json:"rejected"`
	Waitlisted             int                        `json:"waitlisted"`
	AvgScore               float64                    `json:"avg_score"`
	ScoreDistribution      []ScoreBucket              `json:"score_distribution"`
	CategoryCounts         map[string]int             `json:"category_counts"`
	ScoreMean              float64                    `json:"score_mean"`
	ScoreMedian            float64                    `json:"score_median"`
	DimensionMeans         map[string]float64         `json:"dimension_means"`
	DimensionDistributions map[string][]ScoreBucket   `json:"dimension_distributions"`
	// Exam score distributions
	IELTSDistribution  []ScoreBucket `json:"ielts_distribution"`
	TOEFLDistribution  []ScoreBucket `json:"toefl_distribution"`
	UNTDistribution    []ScoreBucket `json:"unt_distribution"`
	NISDistribution    []ScoreBucket `json:"nis_distribution"`
	IELTSCount         int           `json:"ielts_count"`
	TOEFLCount         int           `json:"toefl_count"`
	UNTCount           int           `json:"unt_count"`
	NISCount           int           `json:"nis_count"`
	// Analysis bias detection
	AnalysisCountStats map[int]int   `json:"analysis_count_stats"` // candidate_id -> analysis count
}

type ScoreBucket struct {
	Range string `json:"range"`
	Count int    `json:"count"`
}
