package db

import (
	"net/url"
	"reflect"
	"testing"
)

func TestWhereFromConstraints(t *testing.T) {
	testCases := []struct {
		constraints []SearchConstraint
		want        string
		wantP       int
	}{
		{
			[]SearchConstraint{
				SearchConstraint{"agent", false, OpIn, []string{"jim"}},
			},
			"agent in (?)",
			1,
		},
		{
			// give a dodgy column name
			[]SearchConstraint{
				SearchConstraint{"ag22; drop table users", false, OpIn, []string{"jim"}},
			}, "", 0,
		},
		{
			// give a dodgy op
			[]SearchConstraint{
				SearchConstraint{"agent", false, 22, []string{"jim"}},
			}, "", 0,
		},
		{
			// give a negated op
			[]SearchConstraint{
				SearchConstraint{"agent", true, OpIn, []string{"jim"}},
			}, "agent not in (?)", 1,
		},
	}

	for i, c := range testCases {
		got, gotP, err := WhereFromConstraints(c.constraints)
		if got != c.want {
			t.Errorf("Case %d: \"%s\" != \"%s\"", i, got, c.want)
		}
		if len(gotP) != c.wantP {
			t.Errorf("Case %d: Param len %d != %d", i, len(gotP), c.wantP)
		}
		if err != nil && c.want != "" {
			t.Fatalf("Case %d: Got unexpected error: \"%+v\"", i, err)
		}
	}
}

func TestWhereOpToSQL(t *testing.T) {
	testCases := []struct {
		op     WhereOp
		pCount int
		want   string
	}{
		{OpIn, 1, "in (?)"},
		{OpGtE, 1, ">= ?"},
		{OpLtE, 1, "<= ?"},
		{12, 1, ""},
		{OpGtE, 10, ""},
		{OpLtE, 10, ""},
	}

	for i, c := range testCases {
		got, err := WhereOpToSQL(c.op, c.pCount)
		if got != c.want {
			t.Fatalf("Case %d got %s want %s", i, got, c.want)
		}
		if got == "" && err == nil {
			t.Fatalf("Case %d missing expected error", i)
		}
	}
}

func TestWhereFromQuery(t *testing.T) {
	testCases := []struct {
		q     string
		want  string
		wantP []string
	}{
		{"agent=bob", "agent in (?)", []string{"bob"}},
		{"agent=bob&agent=jane", "agent in (?,?)", []string{"bob", "jane"}},
		{"agent=bob&hour__gte=9", "agent in (?) AND hour >= ?", []string{"bob", "9"}},
		{"agent__in=bob&hour__lte=9", "agent in (?) AND hour <= ?", []string{"bob", "9"}},
		{"agent=bob&!agent=jane", "agent in (?) AND agent not in (?)", []string{"bob", "jane"}},
		{"agent__sub=12", "", nil},
	}

	for i, c := range testCases {
		query, err := url.ParseQuery(c.q)
		if err != nil {
			if c.want == "" {
				continue
			}
			t.Fatal(err)
		}
		sql, params, err := WhereFromQuery(query)
		if err != nil && c.want != "" {
			t.Fatalf("Unexpected error in case %d: %s", i, err.Error())
		}

		if sql != c.want {
			t.Errorf("Case %d: SQL mismatch \"%s\" != \"%s\"", i, sql, c.want)
		}
		if !reflect.DeepEqual(params, c.wantP) {
			t.Errorf("Case %d: Params mismatch \"%v\" != \"%v\"", i, params, c.wantP)
		}
	}
}
