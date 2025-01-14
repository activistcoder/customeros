package enum

type BilledType string

const (
	BilledTypeNone      BilledType = ""
	BilledTypeMonthly   BilledType = "MONTHLY"
	BilledTypeQuarterly BilledType = "QUARTERLY"
	BilledTypeAnnually  BilledType = "ANNUALLY"
	BilledTypeOnce      BilledType = "ONCE"
	BilledTypeUsage     BilledType = "USAGE"
)

var AllBilledTypes = []BilledType{
	BilledTypeNone,
	BilledTypeMonthly,
	BilledTypeQuarterly,
	BilledTypeAnnually,
	BilledTypeOnce,
	BilledTypeUsage,
}

func DecodeBilledType(s string) BilledType {
	if IsValidBilledType(s) {
		return BilledType(s)
	}
	return BilledTypeNone
}

func IsValidBilledType(s string) bool {
	for _, ms := range AllBilledTypes {
		if ms == BilledType(s) {
			return true
		}
	}
	return false
}

func (bt BilledType) String() string {
	return string(bt)
}

func (bt BilledType) IsRecurrent() bool {
	return bt == BilledTypeMonthly || bt == BilledTypeAnnually || bt == BilledTypeQuarterly
}

func (bt BilledType) InMonths() int64 {
	switch bt {
	case BilledTypeMonthly:
		return 1
	case BilledTypeQuarterly:
		return 3
	case BilledTypeAnnually:
		return 12
	default:
		return 0
	}
}
