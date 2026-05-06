package provider

type FilterOperator int

const (
	FilterOperatorInclude FilterOperator = iota
	FilterOperatorExclude
)

type LabelData struct {
	Label    string
	Operator FilterOperator
	Values   []string
}

type LabelFilter struct {
	filters []LabelData
}

func NewLabelFilter(filters []LabelData) *LabelFilter {
	return &LabelFilter{filters: filters}
}

func (f *LabelFilter) IsSatisfy(labels map[string]string) bool {
	if f == nil || len(f.filters) == 0 {
		return true
	}
	for _, filter := range f.filters {
		val, ok := labels[filter.Label]
		if !ok {
			val = ""
		}
		switch filter.Operator {
		case FilterOperatorInclude:
			if !contains(filter.Values, val) {
				return false
			}
		case FilterOperatorExclude:
			if contains(filter.Values, val) {
				return false
			}
		}
	}
	return true
}

func contains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}
