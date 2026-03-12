package domain

type GradeLevel string

const (
	Grade1 GradeLevel = "1ª SÉRIE"
	Grade2 GradeLevel = "2ª SÉRIE"
	Grade3 GradeLevel = "3ª SÉRIE"
	EJA    GradeLevel = "EJA"
)

func (g GradeLevel) IsValid() bool {
	switch g {
	case Grade1, Grade2, Grade3, EJA:
		return true
	}
	return false
}
