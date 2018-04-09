package db

import (
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

var colNameRe = regexp.MustCompile("^[a-zA-Z0-9_]+$")

// WhereOp is used to signify a single where clause operation
type WhereOp int

const (
	// OpIn generates SQL col in (vals)
	OpIn WhereOp = iota
	// OpLtE generates SQL col <= val
	OpLtE
	// OpGtE generates SQL col >= val
	OpGtE
)

// A SearchConstraint represents a single filtering clause when fetching data
type SearchConstraint struct {
	Column  string
	Negated bool
	Op      WhereOp
	Values  []string
}

// contraints sorting helper
type byColname []SearchConstraint

func (c byColname) Len() int {
	return len(c)
}
func (c byColname) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
func (c byColname) Less(i, j int) bool {
	return c[i].Column < c[j].Column
}

// WhereFromQuery takes a url-encoded query string, or POST body, and decodes
// it to a parameterised SQL query and an array of paramaters.
// Paramater values are decoded but left as strings
func WhereFromQuery(query url.Values) (string, []string, error) {
	constraints := make([]SearchConstraint, 0, len(query))

	for key, values := range query {
		c, err := queryKeyToConstraint(key)
		if err != nil {
			return "", nil, err
		}
		c.Values = values
		constraints = append(constraints, c)
	}

	// sort the contraints in column name order for consitency in testing
	sort.Sort(byColname(constraints))

	return WhereFromConstraints(constraints)
}

func queryKeyToConstraint(key string) (SearchConstraint, error) {
	negated := false
	// take off leading ! and negate
	if strings.HasPrefix(key, "!") {
		negated = true
		key = strings.TrimPrefix(key, "!")
	}

	// the default/norm
	col := key
	op := OpIn

	// take off the operator if there is one
	if strings.Contains(key, "__") {
		parts := strings.SplitN(key, "__", 2)
		col = parts[0]
		opStr := parts[1]

		switch opStr {
		case "in":
			op = OpIn
		case "lte":
			op = OpLtE
		case "gte":
			op = OpGtE
		default:
			return SearchConstraint{}, fmt.Errorf("column %s uses unknown filter %s", col, opStr)
		}
	}

	return SearchConstraint{col, negated, op, nil}, nil
}

// WhereFromConstraints takes a list of SearchConstraints and build a
// parameterised SQL query and array of SQL parameters.
// Each constraint represented in constraints is logically ANDed together
func WhereFromConstraints(constraints []SearchConstraint) (string, []string, error) {
	// figure out the final arg count so we can do a single allocation
	argCount := 0
	for _, c := range constraints {
		argCount += len(c.Values)
	}

	// make them the right size to fit all our where clause parts
	sqlParts := make([]string, 0, len(constraints))
	params := make([]string, 0, argCount)

	// build up the SQL where parts and the parameter list
	for i, c := range constraints {
		if !colNameRe.MatchString(c.Column) {
			return "", []string{}, fmt.Errorf("The column name %s is invalid in constraint %d", c.Column, i)
		}

		opStr, err := WhereOpToSQL(c.Op, len(c.Values))
		if err != nil {
			return "", []string{}, fmt.Errorf("Contraint %d on columnt %s is invalid: %s", i, c.Column, err.Error())
		}

		var sep = ""
		if c.Negated {
			sep = " not"
		}

		// colname and operation are valid so add the sql and params
		sqlParts = append(sqlParts, fmt.Sprintf("%s%s %s", c.Column, sep, opStr))
		params = append(params, c.Values...)
	}

	return strings.Join(sqlParts, " AND "), params, nil
}

// WhereOpToSQL takes a single Where Op value and returns the SQL for that
// operation. The number of parameters is required to insert the correct number
// of param placeholders
func WhereOpToSQL(op WhereOp, paramCount int) (string, error) {
	if op == OpIn {
		return "in (" + strings.TrimSuffix(strings.Repeat("?,", paramCount), ",") + ")", nil
	}
	var opStr string
	switch op {
	case OpLtE:
		opStr = "<="
	case OpGtE:
		opStr = ">="
	default:
		return "", fmt.Errorf("The operator %d is unknown", op)
	}

	if paramCount > 1 {
		return "", fmt.Errorf("Cannot use %s with more than 1 parameter", opStr)
	}

	return fmt.Sprintf("%s ?", opStr), nil
}
