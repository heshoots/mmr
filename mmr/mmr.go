package mmr

import (
  "math"
)

func expectations(ra float64, rb float64) float64 {
  var expa = (rb - ra)/400
  var ea = 1/(1 + math.Pow(10, expa))
  return ea
}

func NewRating(ra, rb, ga, gb float64) float64 {
    var ea = expectations(ra, rb)
    var scorea = ga/(ga+gb)
    return ra + 32*(scorea - ea)
}
