package domain

var masterControlledWorldSettingKeys = map[string]struct{}{
	"spawnmode":               {},
	"ghostenabled":            {},
	"ghostsanitydrain":        {},
	"portalresurection":       {},
	"resettime":               {},
	"beefaloheat":             {},
	"krampus":                 {},
	"roads":                   {},
	"touchstone":              {},
	"boons":                   {},
	"basicresource_regrowth":  {},
	"extrastartingitems":      {},
	"seasonalstartingitems":   {},
	"spawnprotection":         {},
	"dropeverythingondespawn": {},
	"healthpenalty":           {},
	"lessdamagetaken":         {},
	"temperaturedamage":       {},
	"hunger":                  {},
	"darkness":                {},
	"shadowcreatures":         {},
	"brightmarecreatures":     {},
	"specialevent":            {},
	"crow_carnival":           {},
	"hallowed_nights":         {},
	"winters_feast":           {},
	"year_of_the_gobbler":     {},
	"year_of_the_varg":        {},
	"year_of_the_pig":         {},
	"year_of_the_carrat":      {},
	"year_of_the_beefalo":     {},
	"year_of_the_catcoon":     {},
	"year_of_the_bunnyman":    {},
	"year_of_the_dragonfly":   {},
	"year_of_the_snake":       {},
	"year_of_the_knight":      {},
}

func IsMasterControlledWorldSettingKey(key string) bool {
	_, ok := masterControlledWorldSettingKeys[key]
	return ok
}
