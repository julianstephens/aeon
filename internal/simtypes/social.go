package simtypes

type Sex int

const (
	SexMale Sex = iota
	SexFemale
)

type Occupation int

const (
	OccupationFarmer Occupation = iota
	OccupationHunter
	OccupationCarpenter
	OccupationFisher
	OccupationBlacksmith
	OccupationMiner
	OccupationMason
	OccupationWeaver
	OccupationMerchant
	OccupationHealer
	OccupationTeacher
	OccupationGuard
)

var Occupations = []Occupation{
	OccupationFarmer,
	OccupationHunter,
	OccupationCarpenter,
	OccupationFisher,
	OccupationBlacksmith,
	OccupationMiner,
	OccupationMason,
	OccupationWeaver,
	OccupationMerchant,
	OccupationHealer,
	OccupationTeacher,
	OccupationGuard,
}

func (o Occupation) String() string {
	if o < 0 || int(o) >= len(Occupations) {
		return "unknown"
	}

	labels := [...]string{
		"farmer",
		"hunter",
		"carpenter",
		"fisher",
		"blacksmith",
		"miner",
		"mason",
		"weaver",
		"merchant",
		"healer",
		"teacher",
		"guard",
	}

	return labels[o]
}
