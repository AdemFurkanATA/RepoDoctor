package main

type TechnicalDebtEstimate struct {
	CircularHours   int `json:"circularHours"`
	LayerHours      int `json:"layerHours"`
	SizeHours       int `json:"sizeHours"`
	GodObjectHours  int `json:"godObjectHours"`
	ComplexityHours int `json:"complexityHours"`
	TotalHours      int `json:"totalHours"`
}

func estimateTechnicalDebt(report *StructuralReport) TechnicalDebtEstimate {
	debt := TechnicalDebtEstimate{
		CircularHours:  len(report.Circular) * 4,
		LayerHours:     len(report.Layer) * 3,
		SizeHours:      len(report.Size) * 1,
		GodObjectHours: len(report.GodObject) * 2,
	}
	debt.ComplexityHours = (report.Complexity.Medium * 1) + (report.Complexity.High * 2)
	debt.TotalHours = debt.CircularHours + debt.LayerHours + debt.SizeHours + debt.GodObjectHours + debt.ComplexityHours
	return debt
}
