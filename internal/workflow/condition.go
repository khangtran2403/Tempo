package workflow

import (
	"fmt"
	"strings"

	"github.com/brokeboycoding/tempo/internal/domain"
)

func EvaluateCondition(condition *domain.Condition, prevResults map[string]interface{}) (bool, error) {
	if condition == nil {
		return true, nil // No condition = always true
	}

	actualValue, err := extractValue(condition.Field, prevResults)
	if err != nil {
		return false, err
	}

	switch condition.Operator {
	case "==":
		return actualValue == condition.Value, nil

	case "!=":
		return actualValue != condition.Value, nil

	case ">":
		return compareGreater(actualValue, condition.Value)

	case "<":
		return compareLess(actualValue, condition.Value)

	case ">=":
		result, err := compareGreater(actualValue, condition.Value)
		if err != nil {
			return false, err
		}
		return result || actualValue == condition.Value, nil

	case "<=":
		result, err := compareLess(actualValue, condition.Value)
		if err != nil {
			return false, err
		}
		return result || actualValue == condition.Value, nil

	case "contains":
		return contains(actualValue, condition.Value)

	default:
		return false, fmt.Errorf("unknown operator: %s", condition.Operator)
	}
}

func extractValue(field string, data map[string]interface{}) (interface{}, error) {
	parts := strings.Split(field, ".")
	current := data

	for i, part := range parts {
		if i == len(parts)-1 {

			if val, ok := current[part]; ok {
				return val, nil
			}
			return nil, fmt.Errorf("field %s not found", field)
		}

		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return nil, fmt.Errorf("field %s is not an object", part)
		}
	}

	return nil, fmt.Errorf("field %s not found", field)
}

func compareGreater(a, b interface{}) (bool, error) {
	aFloat, aOk := toFloat64(a)
	bFloat, bOk := toFloat64(b)

	if !aOk || !bOk {
		return false, fmt.Errorf("cannot compare non-numeric values")
	}

	return aFloat > bFloat, nil
}

func compareLess(a, b interface{}) (bool, error) {
	aFloat, aOk := toFloat64(a)
	bFloat, bOk := toFloat64(b)

	if !aOk || !bOk {
		return false, fmt.Errorf("cannot compare non-numeric values")
	}

	return aFloat < bFloat, nil
}

func contains(a, b interface{}) (bool, error) {
	aStr, aOk := a.(string)
	bStr, bOk := b.(string)

	if !aOk || !bOk {
		return false, fmt.Errorf("contains operator requires string values")
	}

	return strings.Contains(aStr, bStr), nil
}

func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	default:
		return 0, false
	}
}
