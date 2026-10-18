package policy

import _ "embed"

//go:embed model.conf
var ModelConf string

//go:embed policy.csv
var CSV string

//go:embed policy_fixtures.csv
var FixturesCSV string
