package gdp

import (
	"log"
	"strings"
)

func (g *GDP) InputValidation() {

	if g.Config.Dest != "null" {

		dest, _, found := strings.Cut(g.Config.Dest, ":")

		if !found {
			log.Fatalf("InputValidation XTCP Dest must contain ':' chars:%s", g.Config.Dest)
		}

		if strings.Count(g.Config.Dest, ":") != 2 {
			log.Fatalf("InputValidation XTCP Dest must contain x2 ':' chars:%s", g.Config.Dest)
		}

		if _, ok := g.Destinations.Load(dest); !ok {
			log.Fatalf("InputValidation XTCP Dest must start with one of:%s dest:%s :%s", validDestinations(), dest, g.Config.Dest)
		}
	}

}
