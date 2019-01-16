package meta

import (
	"testing"

	gm "github.com/creativesoftwarefdn/weaviate/graphqlapi/local/getmeta"
)

func Test_QueryBuilder_StringProps(t *testing.T) {
	tests := testCases{
		testCase{
			name: "with only a string, with only topOccurrences.value",
			inputProps: []gm.MetaProperty{
				gm.MetaProperty{
					Name:                "name",
					StatisticalAnalyses: []gm.StatisticalAnalysis{gm.TopOccurrencesValue},
				},
			},
			expectedQuery: `
				.union(
					union(
						groupCount().by("name")
							.order(local).by(values, decr).limit(local, 3)
							.as("topOccurrences").project("topOccurrences").by(select("topOccurrences"))
						)
          .as("name").project("name").by(select("name"))
				)
			`,
		},

		testCase{
			name: "with only a string, with only both topOccurrences.value and .occurs",
			inputProps: []gm.MetaProperty{
				gm.MetaProperty{
					Name:                "name",
					StatisticalAnalyses: []gm.StatisticalAnalysis{gm.TopOccurrencesValue, gm.TopOccurrencesOccurs},
				},
			},
			expectedQuery: `
				.union(
					union(
						groupCount().by("name")
							.order(local).by(values, decr).limit(local, 3)
							.as("topOccurrences").project("topOccurrences").by(select("topOccurrences"))
						)
          .as("name").project("name").by(select("name"))
				)
			`,
		},

		testCase{
			name: "with only a string, with all possible props",
			inputProps: []gm.MetaProperty{
				gm.MetaProperty{
					Name:                "name",
					StatisticalAnalyses: []gm.StatisticalAnalysis{gm.Type, gm.Count, gm.TopOccurrencesValue, gm.TopOccurrencesOccurs},
				},
			},
			expectedQuery: `
				.union(
					union(
						has("name").count()
							.as("count").project("count").by(select("count")),
						groupCount().by("name")
							.order(local).by(values, decr).limit(local, 3)
							.as("topOccurrences").project("topOccurrences").by(select("topOccurrences"))
						)
				.as("name").project("name").by(select("name"))
					)
			`,
		},
	}

	tests.AssertQuery(t, nil)

}