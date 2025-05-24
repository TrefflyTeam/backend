package util

import (
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"strconv"
)

func NumericToFloat64(n pgtype.Numeric) (float64, error) {
	floatValue, err := n.Float64Value()
	if err != nil {
		return 0, fmt.Errorf("numeric conversion error: %w", err)
	}

	if !floatValue.Valid {
		return 0, fmt.Errorf("invalid numeric value")
	}

	return floatValue.Float64, nil
}

func Float64ToNumeric(f float64) pgtype.Numeric {
	str := strconv.FormatFloat(f, 'f', -1, 64)

	var n pgtype.Numeric
	if err := n.Scan(str); err != nil {
		return pgtype.Numeric{Valid: false}
	}

	return n
}