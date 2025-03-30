package gdp

import (
	"log"
	"strings"
)

func (g *GDP) InputValidation() {

	if g.config.Dest != "null" {

		dest, _, found := strings.Cut(g.config.Dest, ":")

		if !found {
			log.Fatalf("InputValidation XTCP Dest must contain ':' chars:%s", g.config.Dest)
		}

		if strings.Count(g.config.Dest, ":") != 2 {
			log.Fatalf("InputValidation XTCP Dest must contain x2 ':' chars:%s", g.config.Dest)
		}

		if _, ok := g.Destinations.Load(dest); !ok {
			log.Fatalf("InputValidation XTCP Dest must start with one of:%s dest:%s :%s", validDestinations(), dest, g.config.Dest)
		}
	}

}
