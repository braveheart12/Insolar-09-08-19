// Code generated by "stringer -type=Type"; DO NOT EDIT.

package payload

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[TypeUnknown-0]
	_ = x[TypeError-100]
	_ = x[TypeID-101]
	_ = x[TypeJet-102]
	_ = x[TypeGetObject-103]
	_ = x[TypeObjIndex-104]
	_ = x[TypeObjState-105]
}

const (
	_Type_name_0 = "TypeUnknown"
	_Type_name_1 = "TypeErrorTypeIDTypeJetTypeGetObjectTypeObjIndexTypeObjState"
)

var (
	_Type_index_1 = [...]uint8{0, 9, 15, 22, 35, 47, 59}
)

func (i Type) String() string {
	switch {
	case i == 0:
		return _Type_name_0
	case 100 <= i && i <= 105:
		i -= 100
		return _Type_name_1[_Type_index_1[i]:_Type_index_1[i+1]]
	default:
		return "Type(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}